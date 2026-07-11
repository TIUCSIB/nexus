package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"nexus-agent/internal/config"
)

// SingboxManager manages the lifecycle of a sing-box process.
type SingboxManager struct {
	cfg    config.SingboxConfig
	cmd    *exec.Cmd
	cancel context.CancelFunc

	mu      sync.Mutex
	running bool
}

// New creates a new SingboxManager with the given configuration.
func New(cfg config.SingboxConfig) *SingboxManager {
	return &SingboxManager{
		cfg: cfg,
	}
}

// atomicWriteFile writes data to a temp file then renames to target path,
// ensuring the write is atomic and won't produce a partially-written file.
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	// Always resolve to absolute path so rename works regardless of process CWD.
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve config path: %w", err)
	}
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config dir %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, "singbox-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, perm); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("chmod temp file: %w", err)
	}

	if err := os.Rename(tmpPath, absPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename temp file %s -> %s: %w", tmpPath, absPath, err)
	}
	return nil
}

// writeConfig writes the sing-box configuration to disk atomically and returns the resolved path.
func (s *SingboxManager) writeConfig(configJSON string) (string, error) {
	configPath := s.cfg.ConfigPath
	if configPath == "" {
		configPath = "singbox.json"
	}
	if !filepath.IsAbs(configPath) {
		wd := s.cfg.WorkingDir
		if wd == "" {
			wd = "."
		}
		absWd, err := filepath.Abs(wd)
		if err != nil {
			return "", fmt.Errorf("resolve working_dir: %w", err)
		}
		if err := os.MkdirAll(absWd, 0755); err != nil {
			return "", fmt.Errorf("create working_dir %s: %w", absWd, err)
		}
		configPath = filepath.Join(absWd, configPath)
	} else {
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return "", fmt.Errorf("create config dir: %w", err)
		}
	}
	return configPath, atomicWriteFile(configPath, []byte(configJSON), 0644)
}

// Start writes the sing-box configuration to disk and starts the process.
func (s *SingboxManager) Start(configJSON string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("sing-box is already running")
	}

	// Write config file atomically
	configPath, err := s.writeConfig(configJSON)
	if err != nil {
		return fmt.Errorf("write sing-box config: %w", err)
	}
	log.Printf("[singbox] config written to %s", configPath)

// Build command
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, s.cfg.BinaryPath, "run", "-c", configPath)
	if s.cfg.WorkingDir != "" {
		if absWd, err := filepath.Abs(s.cfg.WorkingDir); err == nil {
			cmd.Dir = absWd
		} else {
			cmd.Dir = s.cfg.WorkingDir
		}
	} else {
		cmd.Dir = filepath.Dir(configPath)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"ENABLE_DEPRECATED_LEGACY_DNS_FAKEIP_OPTIONS=true",
		"ENABLE_DEPRECATED_LEGACY_DNS_SERVERS=true",
		"ENABLE_DEPRECATED_OUTBOUND_DNS_RULE_ITEM=true",
		"ENABLE_DEPRECATED_MISSING_DOMAIN_RESOLVER=true",
	)

	if err := cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("start sing-box: %w", err)
	}

	s.cmd = cmd
	s.cancel = cancel
	s.running = true

	log.Printf("[singbox] started with PID %d", cmd.Process.Pid)

	// Monitor the process in background
	go func() {
		err := cmd.Wait()
		s.mu.Lock()
		s.running = false
		s.cancel = nil
		s.cmd = nil
		s.mu.Unlock()

		if err != nil {
			if ctx.Err() != nil {
				// Process was killed intentionally
				log.Printf("[singbox] process stopped (killed)")
			} else {
				log.Printf("[singbox] process exited with error: %v", err)
			}
		} else {
			log.Printf("[singbox] process exited normally")
		}
	}()

	return nil
}

// Stop gracefully terminates the sing-box process.
// It sends SIGTERM first, waits up to 5 seconds, then force-kills.
func (s *SingboxManager) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running || s.cmd == nil || s.cmd.Process == nil {
		s.running = false
		return nil
	}

	log.Printf("[singbox] stopping process PID %d...", s.cmd.Process.Pid)

	// Step 1: Send SIGTERM on Unix for graceful shutdown
	if runtime.GOOS != "windows" {
		log.Printf("[singbox] sending SIGTERM to PID %d", s.cmd.Process.Pid)
		s.cmd.Process.Signal(syscall.SIGTERM)
	} else {
		// Cancel context (sends kill on Windows)
		if s.cancel != nil {
			s.cancel()
		}
	}

	// Step 2: Wait up to 5 seconds for graceful shutdown
	done := make(chan struct{})
	go func() {
		s.cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("[singbox] process stopped gracefully")
		s.running = false
		return nil
	case <-time.After(5 * time.Second):
		log.Printf("[singbox] graceful stop timed out, force killing")
	}

	// Step 3: Force kill with SIGKILL
	if runtime.GOOS != "windows" {
		s.cmd.Process.Kill()
	} else {
		// On Windows, use taskkill for clean tree kill
		killCmd := exec.Command("taskkill", "/F", "/T", "/PID",
			fmt.Sprintf("%d", s.cmd.Process.Pid))
		killCmd.Run()
	}

	// Wait another 2 seconds for the process to die after force kill
	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}

	s.running = false
	return nil
}

// Restart stops the current process and starts a new one with the provided config.
func (s *SingboxManager) Restart(configJSON string) error {
	if err := s.Stop(); err != nil {
		log.Printf("[singbox] stop error during restart: %v", err)
	}
	// Brief pause to let the OS release resources
	time.Sleep(500 * time.Millisecond)
	return s.Start(configJSON)
}

// ReloadConfig sends the new config to sing-box via PUT /configs API.
// This replaces the running configuration without restarting the process,
// avoiding connection drops for user-list-only changes.
// Returns error if sing-box rejects the config or is not running.
func (s *SingboxManager) ReloadConfig(configJSON string) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return fmt.Errorf("sing-box is not running")
	}
	statsURL := s.cfg.StatsURL
	s.mu.Unlock()

	// PUT /configs to sing-box Clash API
	url := statsURL + "/configs"
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBufferString(configJSON))
	if err != nil {
		return fmt.Errorf("build reload request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("PUT /configs failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sing-box rejected config: HTTP %d", resp.StatusCode)
	}

// Update config file on disk atomically for consistency
		if _, err := s.writeConfig(configJSON); err != nil {
			log.Printf("[singbox] warning: failed to update config file on disk: %v", err)
		}

	log.Printf("[singbox] config hot-reloaded via PUT /configs")
	return nil
}

// IsRunning returns true if the sing-box process is currently active.
func (s *SingboxManager) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// GetStatsURL returns the sing-box statistics API URL.
func (s *SingboxManager) GetStatsURL() string {
	return s.cfg.StatsURL
}
