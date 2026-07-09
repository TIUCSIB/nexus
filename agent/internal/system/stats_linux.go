//go:build linux

package system

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// getMemoryInfo reads /proc/meminfo and returns total, used in bytes.
func getMemoryInfo() (total int64, used int64, ok bool) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, false
	}
	defer f.Close()

	var memTotal, memAvailable, memFree, buffers, cached int64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "MemTotal:"):
			memTotal = parseKB(line)
		case strings.HasPrefix(line, "MemAvailable:"):
			memAvailable = parseKB(line)
		case strings.HasPrefix(line, "MemFree:"):
			memFree = parseKB(line)
		case strings.HasPrefix(line, "Buffers:"):
			buffers = parseKB(line)
		case strings.HasPrefix(line, "Cached:"):
			cached = parseKB(line)
		}
	}

	if memTotal > 0 {
		// Use MemAvailable if available, otherwise calculate
		if memAvailable > 0 {
			used = memTotal - memAvailable
		} else {
			used = memTotal - memFree - buffers - cached
		}
		return memTotal, used, true
	}
	return 0, 0, false
}

func parseKB(line string) int64 {
	fields := strings.Fields(line)
	if len(fields) >= 2 {
		val, err := strconv.ParseInt(fields[1], 10, 64)
		if err == nil {
			return val * 1024 // convert KB to bytes
		}
	}
	return 0
}

// getDiskInfo returns total and used disk space in bytes for the root partition.
func getDiskInfo() (total int64, used int64, ok bool) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return 0, 0, false
	}
	// Available blocks * block size
	total = int64(stat.Blocks) * int64(stat.Bsize)
	free := int64(stat.Bfree) * int64(stat.Bsize)
	used = total - free
	return total, used, true
}

// cpuTimes stores parsed /proc/stat CPU times.
type cpuTimes struct {
	User    float64
	Nice    float64
	System  float64
	Idle    float64
	Iowait  float64
	Irq     float64
	Softirq float64
	Steal   float64
}

func (c cpuTimes) Total() float64 {
	return c.User + c.Nice + c.System + c.Idle + c.Iowait + c.Irq + c.Softirq + c.Steal
}

// getCPUTime reads the aggregate CPU times from /proc/stat.
func getCPUTime() (cpuTimes, bool) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return cpuTimes{}, false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			return cpuTimes{}, false
		}

		parse := func(idx int) float64 {
			if idx >= len(fields) {
				return 0
			}
			v, _ := strconv.ParseFloat(fields[idx], 64)
			return v
		}

		return cpuTimes{
			User:    parse(1),
			Nice:    parse(2),
			System:  parse(3),
			Idle:    parse(4),
			Iowait:  parse(5),
			Irq:     parse(6),
			Softirq: parse(7),
			Steal:   parse(8),
		}, true
	}
	return cpuTimes{}, false
}

// getNetStats reads total received and transmitted bytes from /proc/net/dev.
func getNetStats() (rxBytes uint64, txBytes uint64, ok bool) {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return 0, 0, false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// Skip header lines (first two)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum <= 2 {
			continue
		}
		line := scanner.Text()
		// Skip loopback
		if strings.HasPrefix(line, "  lo:") || strings.HasPrefix(line, "lo:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}
		// fields[0] is interface name (e.g., "eth0:"), fields[1] is rx_bytes
		rx, err1 := strconv.ParseUint(fields[1], 10, 64)
		if err1 != nil {
			continue
		}
		// fields[9] is tx_bytes
		if len(fields) > 9 {
			tx, err2 := strconv.ParseUint(fields[9], 10, 64)
			if err2 == nil {
				rxBytes += rx
				txBytes += tx
			}
		}
	}

	return rxBytes, txBytes, true
}