package stats

import (
	"runtime"
)

type MemoryStats runtime.MemStats
type RuntimeStatsInfo struct {
	NumCPU       int64
	NumGoroutine int64
	Version      string
	MemoryStats  *MemoryStats
}
type RuntimeStats struct {
	info *RuntimeStatsInfo
}

func NewRuntimeStats() *RuntimeStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return &RuntimeStats{
		info: &RuntimeStatsInfo{
			NumCPU:       int64(runtime.NumCPU()),
			NumGoroutine: int64(runtime.NumGoroutine()),
			Version:      runtime.Version(),
			MemoryStats:  (*MemoryStats)(&memStats),
		},
	}
}

func (stats *RuntimeStats) Info() *RuntimeStatsInfo {
	return stats.info
}
