package utils

import (
	"sort"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/shubomifashakin/sysmon/internal/models"
)

func GetMemStats() (models.MemStats,error) {
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return models.MemStats{},err
	}

	usedRamMb := memStat.Used >> 20
	totalRamMb := memStat.Total >> 20
	availableMb := memStat.Available >> 20
	percentageUsed := memStat.UsedPercent

	return models.MemStats{
		UseRamMb:       usedRamMb,
		TotalRamMb:     totalRamMb,
		AvailableRamMb: availableMb,
		PercentUsed:    percentageUsed,
	},nil
}

func GetCpuStats() (models.CpuStats,error){
	// this would take two samples, one at the beginning and one after 500 milliseconds
	// after which it would then calculate the percentage of increase
	percentUsed, err := cpu.Percent(0, true)

	if err != nil {
		return models.CpuStats{},err
	}

	return models.CpuStats{
		PercentUsed: percentUsed,
	},nil
}

func GetUptimeStats() (models.UptimeStats,error){
	sysUptime, err := host.Uptime()
	if err != nil {
		return  models.UptimeStats{},err
	}

	days:=sysUptime/(3600*24)
	hours := sysUptime / 3600
	minutes := (sysUptime % 3600) / 60


	return models.UptimeStats{
		Days: days,
		Hours:   hours,
		Minutes: minutes,
	},nil
}

func GetNetworkStats() (models.NetworkStats,error) {
	totalTcpConnections, err := net.Connections("tcp")
	if err != nil {
		return models.NetworkStats{},err
	}

	stats, err := net.IOCounters(false)
	if err != nil {
		return models.NetworkStats{},err
	}

	return models.NetworkStats{
		TotalTcpConnections: totalTcpConnections,
		Stats:stats,
	},nil
}

func GetTop20ProcessesByCpuAndMem() (models.ProcessStats,error) {
	processes, err := process.Processes()
	if err != nil {
		return models.ProcessStats{},err
	}

	sort.Slice(processes, func(i, j int) bool {
		a, _ := processes[i].CPUPercent()
		b, _ := processes[j].CPUPercent()
		return a > b
	})

	top5ProcessesByCpu := make([]*process.Process, 20)
	copy(top5ProcessesByCpu, processes[:20])

	sort.Slice(processes, func(i, j int) bool {
		a, _ := processes[i].MemoryPercent()
		b, _ := processes[j].MemoryPercent()
		return a > b
	})

	top5ProcessesByMemory := processes[:20]

	return models.ProcessStats{
		Top5ProcessesByMem: top5ProcessesByMemory,
		Top5ProcessesByCpu: top5ProcessesByCpu,
	},nil
}