package system

import (
	"github.com/shirou/gopsutil/v3/disk"
)

type DiskInfo struct {
	Device     string  `json:"device"`
	Mountpoint string  `json:"mountpoint"`
	Fstype     string  `json:"fstype"`
	UsedGB     uint64  `json:"used_gb"`
	TotalGB    uint64  `json:"total_gb"`
	Percent    float64 `json:"percent"`
}

func ListDisks() ([]DiskInfo, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	var result []DiskInfo
	seen := make(map[string]bool)

	for _, p := range partitions {
		if seen[p.Mountpoint] {
			continue
		}

		usage, err := disk.Usage(p.Mountpoint)
		if err != nil || usage.Total == 0 {
			continue
		}
		seen[p.Mountpoint] = true

		result = append(result, DiskInfo{
			Device:     p.Device,
			Mountpoint: p.Mountpoint,
			Fstype:     p.Fstype,
			UsedGB:     usage.Used / 1024 / 1024 / 1024,
			TotalGB:    usage.Total / 1024 / 1024 / 1024,
			Percent:    usage.UsedPercent,
		})
	}

	return result, nil
}
