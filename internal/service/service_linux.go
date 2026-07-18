package service

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

func ListServices() ([]ServiceInfo, error) {
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--all", "--no-pager", "--no-legend", "--plain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("impossible de lister les services systemd: %w", err)
	}

	var result []ServiceInfo
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Format typique d'une ligne :
		// nom.service    loaded    active    running    Description du service
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		name := fields[0]
		activeState := fields[2] // active / inactive / failed
		subState := fields[3]    // running / dead / exited...

		status := activeState
		if activeState == "active" {
			status = subState // ex: "running"
		}

		// Description = tout ce qui reste après les 4 premiers champs
		displayName := name
		if len(fields) > 4 {
			displayName = strings.Join(fields[4:], " ")
		}

		result = append(result, ServiceInfo{
			Name:        name,
			DisplayName: displayName,
			Status:      status,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("erreur de lecture de la sortie systemctl: %w", err)
	}

	return result, nil
}

func StartService(name string) error {
	cmd := exec.Command("systemctl", "start", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("impossible de démarrer %s: %w (%s)", name, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func StopService(name string) error {
	cmd := exec.Command("systemctl", "stop", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("impossible d'arrêter %s: %w (%s)", name, err, strings.TrimSpace(string(output)))
	}
	return nil
}
