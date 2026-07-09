// Package system provides functions to collect system load statistics
// (CPU, memory, disk, network) for machine-mode load reporting.
package system

import (
	"runtime"
	"sync"
	"time"
)

// Load contains the system load data sent to the panel.
type Load struct {
	CPU         float64 `json:"cpu"`
	MemTotal    int64   `json:"mem_total"`
	MemUsed     int64   `json:"mem_used"`
	DiskTotal   int64   `json:"disk_total"`
	DiskUsed    int64   `json:"disk_used"`
	NetInSpeed  float64 `json:"net_in_speed"`
	NetOutSpeed float64 `json:"net_out_speed"`
}

// Monitor periodically collects system load statistics.
type Monitor struct {
	mu          sync.Mutex
	prevCPU     cpuTimes
	prevNetIn   uint64
	prevNetOut  uint64
	lastCollect time.Time
	first       bool
}

// NewMonitor creates a new system load monitor.
func NewMonitor() *Monitor {
	return &Monitor{first: true}
}

// Collect gathers the current system load. On first call it returns
// zero values for CPU/network delta metrics (needs a baseline).
func (m *Monitor) Collect() Load {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	load := Load{}

	// Memory
	memTotal, memUsed, memOK := getMemoryInfo()
	if memOK {
		load.MemTotal = memTotal
		load.MemUsed = memUsed
	}

	// Disk (root partition)
	diskTotal, diskUsed, diskOK := getDiskInfo()
	if diskOK {
		load.DiskTotal = diskTotal
		load.DiskUsed = diskUsed
	}

	// CPU (delta-based)
	currentCPU, cpuOK := getCPUTime()
	if cpuOK {
		if m.first {
			m.prevCPU = currentCPU
		} else {
			elapsed := now.Sub(m.lastCollect).Seconds()
			if elapsed > 0 {
				totalDelta := currentCPU.Total() - m.prevCPU.Total()
				idleDelta := currentCPU.Idle + currentCPU.Iowait - (m.prevCPU.Idle + m.prevCPU.Iowait)
				if totalDelta > 0 {
					load.CPU = (1 - idleDelta/totalDelta) * 100
					if load.CPU < 0 {
						load.CPU = 0
					}
					if load.CPU > 100 {
						load.CPU = 100
					}
				}
			}
		}
		m.prevCPU = currentCPU
	}

	// Network (delta-based)
	rxBytes, txBytes, netOK := getNetStats()
	if netOK {
		if m.first {
			m.prevNetIn = rxBytes
			m.prevNetOut = txBytes
		} else {
			elapsed := now.Sub(m.lastCollect).Seconds()
			if elapsed > 0 {
				rxDelta := float64(rxBytes - m.prevNetIn)
				txDelta := float64(txBytes - m.prevNetOut)
				if rxDelta >= 0 {
					load.NetInSpeed = rxDelta / elapsed
				}
				if txDelta >= 0 {
					load.NetOutSpeed = txDelta / elapsed
				}
			}
		}
		m.prevNetIn = rxBytes
		m.prevNetOut = txBytes
	}

	m.lastCollect = now
	m.first = false

	// Fallback for CPU if /proc is not available
	if load.CPU == 0 && !cpuOK {
		numGoroutines := runtime.NumGoroutine()
		numCPU := runtime.NumCPU()
		if numCPU > 0 {
			load.CPU = float64(numGoroutines) / float64(numCPU) * 10
			if load.CPU > 100 {
				load.CPU = 100
			}
		}
	}

	return load
}