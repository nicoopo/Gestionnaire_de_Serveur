package system

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type GPUInfo struct {
	Name         string  `json:"name"`
	UsagePercent float64 `json:"usage_percent"`
	MemUsedMB    uint64  `json:"mem_used_mb"`
	MemTotalMB   uint64  `json:"mem_total_mb"`
	TempC        float64 `json:"temp_c"`
}

func ListGPUs() ([]GPUInfo, error) {
	cmd := exec.Command("nvidia-smi",
		"--query-gpu=name,utilization.gpu,memory.used,memory.total,temperature.gpu",
		"--format=csv,noheader,nounits")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("nvidia-smi indisponible: %w", err)
	}

	var result []GPUInfo
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) < 5 {
			continue
		}

		name := strings.TrimSpace(fields[0])
		usage, _ := strconv.ParseFloat(strings.TrimSpace(fields[1]), 64)
		memUsed, _ := strconv.ParseUint(strings.TrimSpace(fields[2]), 10, 64)
		memTotal, _ := strconv.ParseUint(strings.TrimSpace(fields[3]), 10, 64)
		temp, _ := strconv.ParseFloat(strings.TrimSpace(fields[4]), 64)

		result = append(result, GPUInfo{
			Name:         name,
			UsagePercent: usage,
			MemUsedMB:    memUsed,
			MemTotalMB:   memTotal,
			TempC:        temp,
		})
	}

	return result, nil
}
