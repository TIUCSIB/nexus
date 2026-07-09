//go:build !linux

package system

// getMemoryInfo is not supported on this platform.
func getMemoryInfo() (total int64, used int64, ok bool) {
	return 0, 0, false
}

// getDiskInfo is not supported on this platform.
func getDiskInfo() (total int64, used int64, ok bool) {
	return 0, 0, false
}

// cpuTimes is a placeholder for non-Linux platforms.
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
	return 0
}

// getCPUTime is not supported on this platform.
func getCPUTime() (cpuTimes, bool) {
	return cpuTimes{}, false
}

// getNetStats is not supported on this platform.
func getNetStats() (rxBytes uint64, txBytes uint64, ok bool) {
	return 0, 0, false
}