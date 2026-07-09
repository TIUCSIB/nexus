package model

import "encoding/json"

// CertConfig holds per-node TLS certificate automation settings.
//
// Modes (CertMode):
//   - "none"    — No TLS. Node handles plain connections.
//   - "file"    — Use user-provided cert/key file paths (CertFile/KeyFile).
//   - "content" — Use PEM content pushed from panel (CertContent/KeyContent).
//   - "self"    — Self-signed certificate generated in memory.
//   - "http"    — ACME HTTP-01 challenge (needs port 80 reachable).
//   - "dns"     — ACME DNS-01 challenge via DNSProvider + DNSEnv.
//
// When CertMode is empty, resolution order is:
//  1. "http"   (if AutoTLS is true)
//  2. "content" (if both CertContent and KeyContent are provided)
//  3. "file"    (if both CertFile and KeyFile are provided)
//  4. "none"
type CertConfig struct {
	CertMode    string            `json:"cert_mode"`    // none/file/content/self/http/dns
	Domain      string            `json:"domain"`       // Certificate domain (ACME/self)
	Email       string            `json:"email"`        // ACME email
	DNSProvider string            `json:"dns_provider"` // dns mode: cloudflare, alidns, ...
	DNSEnv      map[string]string `json:"dns_env"`      // Provider-specific API keys/tokens
	HTTPPort    int               `json:"http_port"`    // ACME HTTP-01 local port (default 80)
	CertFile    string            `json:"cert_file"`    // file mode: path to cert
	KeyFile     string            `json:"key_file"`     // file mode: path to key
	CertContent string            `json:"cert_content"` // content mode: cert PEM text
	KeyContent  string            `json:"key_content"`  // content mode: key PEM text
	CertDir     string            `json:"cert_dir"`     // storage dir for ACME/self persistence
}

// ParseCertConfig decodes a JSON string into a CertConfig.
// Returns an empty CertConfig if the input is empty or invalid.
func ParseCertConfig(s string) CertConfig {
	var c CertConfig
	if s == "" {
		return c
	}
	_ = json.Unmarshal([]byte(s), &c)
	return c
}

// String serialises the CertConfig back to JSON.
func (c CertConfig) String() string {
	b, err := json.Marshal(c)
	if err != nil {
		return "{}"
	}
	return string(b)
}
