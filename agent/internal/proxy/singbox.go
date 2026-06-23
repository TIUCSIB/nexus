package proxy

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
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

// Start writes the sing-box configuration to disk and starts the process.
func (s *SingboxManager) Start(configJSON string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("sing-box is already running")
	}

	// Write config file
	configPath := s.cfg.ConfigPath
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(s.cfg.WorkingDir, configPath)
	}

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		return fmt.Errorf("write sing-box config: %w", err)
	}
	log.Printf("[singbox] config written to %s", configPath)

	// Build command
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, s.cfg.BinaryPath, "run", "-c", configPath)
	cmd.Dir = s.cfg.WorkingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

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
func (s *SingboxManager) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running || s.cmd == nil || s.cmd.Process == nil {
		s.running = false
		return nil
	}

	log.Printf("[singbox] stopping process PID %d...", s.cmd.Process.Pid)

	// Cancel the context to send termination signal
	if s.cancel != nil {
		s.cancel()
	}

	// Wait up to 5 seconds for graceful shutdown
	done := make(chan struct{})
	go func() {
		// On Windows, cmd.Process.Kill() is the main way to stop.
		// On Unix, Cancel() sends SIGKILL by default with exec.CommandContext.
		// We rely on the context cancellation mechanism.
		for s.running {
			time.Sleep(100 * time.Millisecond)
		}
		close(done)
	}()

	select {
	case <-done:
		log.Printf("[singbox] process stopped gracefully")
	case <-time.After(5 * time.Second):
		log.Printf("[singbox] graceful stop timed out, force killing")
		if err := s.cmd.Process.Kill(); err != nil {
			log.Printf("[singbox] force kill failed: %v", err)
		}
		// On Windows, also try taskkill for clean tree kill
		if runtime.GOOS == "windows" {
			killCmd := exec.Command("taskkill", "/F", "/T", "/PID",
				fmt.Sprintf("%d", s.cmd.Process.Pid))
			killCmd.Run()
		}
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
