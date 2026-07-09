package cert

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/caddyserver/certmagic"

	"nexus-agent/internal/cert/dnsproviders"
)

// Config describes per-node TLS certificate automation settings.
type Config struct {
	CertMode    string            `json:"cert_mode"`
	Domain      string            `json:"domain"`
	Email       string            `json:"email"`
	DNSProvider string            `json:"dns_provider"`
	DNSEnv      map[string]string `json:"dns_env"`
	HTTPPort    int               `json:"http_port"`
	CertFile    string            `json:"cert_file"`
	KeyFile     string            `json:"key_file"`
	CertContent string            `json:"cert_content"`
	KeyContent  string            `json:"key_content"`
	CertDir     string            `json:"cert_dir"`
}

// Material is the resolved certificate and private key PEM.
type Material struct {
	CertPEM []byte
	KeyPEM  []byte
}

type material struct {
	cert []byte
	key  []byte
}

// Manager resolves certificate material from file/content/self/http/dns modes.
type Manager struct {
	cfg Config
	mat atomic.Pointer[material]

	magic       *certmagic.Config
	acmeStarted atomic.Bool
	acmeCancel  context.CancelFunc
	renewed     atomic.Bool
}

func NewManager(cfg Config) *Manager {
	return &Manager{cfg: cfg}
}

func (m *Manager) Start(ctx context.Context) error {
	switch m.mode() {
	case "", "none":
		return nil
	case "file":
		return m.startFile()
	case "content":
		return m.startContent()
	case "self":
		return m.startSelfSigned()
	case "http":
		return m.startACME(ctx, nil)
	case "dns":
		solver, err := m.buildDNSSolver()
		if err != nil {
			return err
		}
		return m.startACME(ctx, solver)
	default:
		return fmt.Errorf("unknown cert_mode %q", m.cfg.CertMode)
	}
}

func (m *Manager) Stop() {
	if m.acmeCancel != nil {
		m.acmeCancel()
		m.acmeCancel = nil
	}
}

func (m *Manager) HasCert() bool {
	mat := m.mat.Load()
	return mat != nil && len(mat.cert) > 0 && len(mat.key) > 0
}

func (m *Manager) Material() Material {
	mat := m.mat.Load()
	if mat == nil {
		return Material{}
	}
	return Material{CertPEM: mat.cert, KeyPEM: mat.key}
}

func (m *Manager) CertRenewed() bool {
	return m.renewed.Swap(false)
}

func (m *Manager) mode() string {
	mode := strings.ToLower(strings.TrimSpace(m.cfg.CertMode))
	if mode != "" {
		return mode
	}
	if m.cfg.CertContent != "" && m.cfg.KeyContent != "" {
		return "content"
	}
	if m.cfg.CertFile != "" && m.cfg.KeyFile != "" {
		return "file"
	}
	return "none"
}

func (m *Manager) startFile() error {
	certPEM, err := os.ReadFile(m.cfg.CertFile)
	if err != nil {
		return fmt.Errorf("read cert_file: %w", err)
	}
	keyPEM, err := os.ReadFile(m.cfg.KeyFile)
	if err != nil {
		return fmt.Errorf("read key_file: %w", err)
	}
	if err := validatePair(certPEM, keyPEM); err != nil {
		return fmt.Errorf("invalid certificate pair: %w", err)
	}
	m.store(certPEM, keyPEM)
	return nil
}

func (m *Manager) startContent() error {
	if m.cfg.CertContent == "" || m.cfg.KeyContent == "" {
		if m.loadPersisted() {
			return nil
		}
		return fmt.Errorf("cert_mode content requires cert_content and key_content")
	}
	certPEM := []byte(m.cfg.CertContent)
	keyPEM := []byte(m.cfg.KeyContent)
	if err := validatePair(certPEM, keyPEM); err != nil {
		return fmt.Errorf("invalid certificate content: %w", err)
	}
	m.store(certPEM, keyPEM)
	return nil
}

func (m *Manager) startSelfSigned() error {
	if m.loadPersisted() {
		return nil
	}
	domain := strings.TrimSpace(m.cfg.Domain)
	if domain == "" {
		domain = "localhost"
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}
	tpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: domain},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	if ip := net.ParseIP(domain); ip != nil {
		tpl.IPAddresses = []net.IP{ip}
	} else {
		tpl.DNSNames = []string{domain}
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &key.PublicKey, key)
	if err != nil {
		return err
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}
	m.store(
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}),
		pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}),
	)
	return nil
}

func (m *Manager) startACME(ctx context.Context, dnsSolver *certmagic.DNS01Solver) error {
	if m.acmeStarted.Load() {
		return nil
	}
	domain := strings.TrimSpace(m.cfg.Domain)
	if domain == "" {
		return fmt.Errorf("cert_config.domain is required for ACME")
	}
	certDir := m.certDir()
	if err := os.MkdirAll(certDir, 0o755); err != nil {
		return err
	}

	storage := &certmagic.FileStorage{Path: certDir}
	var magic *certmagic.Config
	cache := certmagic.NewCache(certmagic.CacheOptions{
		GetConfigForCert: func(_ certmagic.Certificate) (*certmagic.Config, error) {
			return magic, nil
		},
	})
	magic = certmagic.New(cache, certmagic.Config{Storage: storage})
	issuer := certmagic.ACMEIssuer{
		CA:    certmagic.LetsEncryptProductionCA,
		Email: m.cfg.Email,
	}
	if dnsSolver != nil {
		issuer.DNS01Solver = dnsSolver
		issuer.DisableHTTPChallenge = true
		issuer.DisableTLSALPNChallenge = true
	} else {
		httpPort := m.cfg.HTTPPort
		if httpPort == 0 {
			httpPort = 80
		}
		issuer.AltHTTPPort = httpPort
		issuer.DisableTLSALPNChallenge = true
	}
	magic.Issuers = []certmagic.Issuer{certmagic.NewACMEIssuer(magic, issuer)}

	acmeCtx, cancel := context.WithCancel(ctx)
	if err := magic.ObtainCertSync(acmeCtx, domain); err != nil {
		cancel()
		return err
	}
	issuerKey := magic.Issuers[0].IssuerKey()
	if err := m.loadPEMFromStorage(acmeCtx, storage, issuerKey, domain); err != nil {
		cancel()
		return err
	}
	if err := magic.ManageAsync(acmeCtx, []string{domain}); err != nil {
		cancel()
		return err
	}
	m.magic = magic
	m.acmeCancel = cancel
	m.acmeStarted.Store(true)
	return nil
}

func (m *Manager) buildDNSSolver() (*certmagic.DNS01Solver, error) {
	provider, err := m.newDNSProvider()
	if err != nil {
		return nil, err
	}
	return &certmagic.DNS01Solver{DNSManager: certmagic.DNSManager{DNSProvider: provider}}, nil
}

func (m *Manager) newDNSProvider() (certmagic.DNSProvider, error) {
	name := strings.TrimSpace(m.cfg.DNSProvider)
	if name == "" {
		return nil, fmt.Errorf("dns_provider is required for cert_mode=dns")
	}
	p, ok := dnsproviders.Get(name)
	if !ok {
		return nil, fmt.Errorf("unsupported dns_provider %q (supported: %s)", name, strings.Join(dnsproviders.CanonicalNames(), ", "))
	}
	env := m.cfg.DNSEnv
	if env == nil {
		env = map[string]string{}
	}
	return p.Build(env)
}

func (m *Manager) loadPEMFromStorage(ctx context.Context, storage certmagic.Storage, issuerKey, domain string) error {
	certPEM, err := storage.Load(ctx, certmagic.StorageKeys.SiteCert(issuerKey, domain))
	if err != nil {
		return err
	}
	keyPEM, err := storage.Load(ctx, certmagic.StorageKeys.SitePrivateKey(issuerKey, domain))
	if err != nil {
		return err
	}
	if err := validatePair(certPEM, keyPEM); err != nil {
		return err
	}
	m.store(certPEM, keyPEM)
	return nil
}

func (m *Manager) store(certPEM, keyPEM []byte) {
	m.mat.Store(&material{cert: certPEM, key: keyPEM})
	m.persist(certPEM, keyPEM)
}

func (m *Manager) certDir() string {
	if m.cfg.CertDir != "" {
		return m.cfg.CertDir
	}
	name := strings.TrimSpace(m.cfg.Domain)
	if name == "" {
		name = "default"
	}
	return filepath.Join("certs", name)
}

func (m *Manager) persist(certPEM, keyPEM []byte) {
	dir := m.certDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(dir, "cert.pem"), certPEM, 0o644)
	_ = os.WriteFile(filepath.Join(dir, "key.pem"), keyPEM, 0o600)
}

func (m *Manager) loadPersisted() bool {
	certPEM, err := os.ReadFile(filepath.Join(m.certDir(), "cert.pem"))
	if err != nil {
		return false
	}
	keyPEM, err := os.ReadFile(filepath.Join(m.certDir(), "key.pem"))
	if err != nil {
		return false
	}
	if validatePair(certPEM, keyPEM) != nil {
		return false
	}
	m.mat.Store(&material{cert: certPEM, key: keyPEM})
	return true
}

func validatePair(certPEM, keyPEM []byte) error {
	_, err := tls.X509KeyPair(certPEM, keyPEM)
	return err
}
