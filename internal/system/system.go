package system

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemInfo struct {
	CPUPercent   float64   `json:"cpu_percent"`
	CPUPerCore   []float64 `json:"cpu_per_core"`
	CPUCoreCount int       `json:"cpu_core_count"`
	RAMPercent   float64   `json:"ram_percent"`
	RAMUsedMB    uint64    `json:"ram_used_mb"`
	RAMTotalMB   uint64    `json:"ram_total_mb"`
	DiskPercent  float64   `json:"disk_percent"`
	DiskUsedGB   uint64    `json:"disk_used_gb"`
	DiskTotalGB  uint64    `json:"disk_total_gb"`
	UptimeSecond uint64    `json:"uptime_seconds"`
	ProcessCount uint64    `json:"process_count"`
}

func GetSystemInfo() (*SystemInfo, error) {
	// Moyenne globale
	cpuPercents, err := cpu.Percent(0, false)
	if err != nil {
		return nil, err
	}

	// Détail par cœur
	cpuPerCore, err := cpu.Percent(0, true)
	if err != nil {
		return nil, err
	}

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	diskStat, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	uptime, err := host.Uptime()
	if err != nil {
		return nil, err
	}

	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}

	var cpuPercent float64
	if len(cpuPercents) > 0 {
		cpuPercent = cpuPercents[0]
	}

	info := &SystemInfo{
		CPUPercent:   cpuPercent,
		CPUPerCore:   cpuPerCore,
		CPUCoreCount: len(cpuPerCore),
		RAMPercent:   vmStat.UsedPercent,
		RAMUsedMB:    vmStat.Used / 1024 / 1024,
		RAMTotalMB:   vmStat.Total / 1024 / 1024,
		DiskPercent:  diskStat.UsedPercent,
		DiskUsedGB:   diskStat.Used / 1024 / 1024 / 1024,
		DiskTotalGB:  diskStat.Total / 1024 / 1024 / 1024,
		UptimeSecond: uptime,
		ProcessCount: hostInfo.Procs,
	}

	return info, nil
}
