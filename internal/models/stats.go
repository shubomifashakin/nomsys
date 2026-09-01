package models

import (
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

type MemStats struct {
	UseRamMb       uint64
	TotalRamMb     uint64
	AvailableRamMb uint64
	PercentUsed    float64
}

type CpuStats struct {
	PercentUsed []float64
}

type UptimeStats struct {
	Days uint64
	Hours   uint64
	Minutes uint64
}

type NetworkStats struct {
	TotalTcpConnections []net.ConnectionStat
	Stats []net.IOCountersStat
}

type ProcessStats struct {
	Top5ProcessesByMem []*process.Process
	Top5ProcessesByCpu []*process.Process
}