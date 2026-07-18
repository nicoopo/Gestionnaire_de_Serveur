package process

import (
	"fmt"
	"sort"
	"strings"

	gopsutilprocess "github.com/shirou/gopsutil/v3/process"
)

type ProcessInfo struct {
	PID        int32   `json:"pid"`
	Name       string  `json:"name"`
	CPUPercent float64 `json:"cpu_percent"`
	RAMPercent float32 `json:"ram_percent"`
	Status     string  `json:"status"`
}

func ListProcesses(nameFilter string) ([]ProcessInfo, error) {
	procs, err := gopsutilprocess.Processes()
	if err != nil {
		return nil, err
	}

	result := make([]ProcessInfo, 0, len(procs))
	nameFilter = strings.ToLower(strings.TrimSpace(nameFilter))

	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}

		// Si un filtre est fourni, on ignore les process qui ne matchent pas
		if nameFilter != "" && !strings.Contains(strings.ToLower(name), nameFilter) {
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

	sort.Slice(result, func(i, j int) bool {
		return result[i].CPUPercent > result[j].CPUPercent
	})

	return result, nil
}

func KillProcess(pid int32) error {
	p, err := gopsutilprocess.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("process %d introuvable: %w", pid, err)
	}

	if err := p.Kill(); err != nil {
		return fmt.Errorf("impossible de tuer le process %d: %w", pid, err)
	}

	return nil
}
