package process

import (
	"sort"

	gopsutilprocess "github.com/shirou/gopsutil/v3/process"
)

type ProcessInfo struct {
	PID        int32   `json:"pid"`
	Name       string  `json:"name"`
	CPUPercent float64 `json:"cpu_percent"`
	RAMPercent float32 `json:"ram_percent"`
	Status     string  `json:"status"`
}

func ListProcesses() ([]ProcessInfo, error) {
	procs, err := gopsutilprocess.Processes()
	if err != nil {
		return nil, err
	}

	result := make([]ProcessInfo, 0, len(procs))

	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			// Certains process système refusent l'accès, on les ignore
			continue
		}

		cpuPercent, err := p.CPUPercent()
		if err != nil {
			cpuPercent = 0
		}

		ramPercent, err := p.MemoryPercent()
		if err != nil {
			ramPercent = 0
		}

		statusSlice, err := p.Status()
		status := "unknown"
		if err == nil && len(statusSlice) > 0 {
			status = statusSlice[0]
		}

		result = append(result, ProcessInfo{
			PID:        p.Pid,
			Name:       name,
			CPUPercent: cpuPercent,
			RAMPercent: ramPercent,
			Status:     status,
		})
	}

	// Tri par utilisation CPU décroissante
	sort.Slice(result, func(i, j int) bool {
		return result[i].CPUPercent > result[j].CPUPercent
	})

	return result, nil
}
