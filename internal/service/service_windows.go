package service

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

type ServiceInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Status      string `json:"status"`
}

func statusToString(s svc.State) string {
	switch s {
	case svc.Stopped:
		return "stopped"
	case svc.StartPending:
		return "start_pending"
	case svc.StopPending:
		return "stop_pending"
	case svc.Running:
		return "running"
	case svc.ContinuePending:
		return "continue_pending"
	case svc.PausePending:
		return "pause_pending"
	case svc.Paused:
		return "paused"
	default:
		return "unknown"
	}
}

func ListServices() ([]ServiceInfo, error) {
	m, err := mgr.Connect()
	if err != nil {
		return nil, fmt.Errorf("connexion au service manager échouée: %w", err)
	}
	defer m.Disconnect()

	names, err := m.ListServices()
	if err != nil {
		return nil, fmt.Errorf("impossible de lister les services: %w", err)
	}

	result := make([]ServiceInfo, 0, len(names))

	for _, name := range names {
		s, err := m.OpenService(name)
		if err != nil {
			// certains services système refusent l'ouverture, on les ignore
			continue
		}

		config, err := s.Config()
		displayName := name
		if err == nil {
			displayName = config.DisplayName
		}

		status, err := s.Query()
		s.Close()
		if err != nil {
			continue
		}

		result = append(result, ServiceInfo{
			Name:        name,
			DisplayName: displayName,
			Status:      statusToString(status.State),
		})
	}

	return result, nil
}

func StartService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connexion au service manager échouée: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("service %s introuvable: %w", name, err)
	}
	defer s.Close()

	if err := s.Start(); err != nil {
		return fmt.Errorf("impossible de démarrer %s: %w", name, err)
	}

	return nil
}

func StopService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connexion au service manager échouée: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("service %s introuvable: %w", name, err)
	}
	defer s.Close()

	status, err := s.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("impossible d'arrêter %s: %w", name, err)
	}

	// Attente courte pour laisser le service passer en "stopped"
	timeout := time.Now().Add(10 * time.Second)
	for status.State != svc.Stopped {
		if time.Now().After(timeout) {
			return fmt.Errorf("timeout en attendant l'arrêt de %s", name)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("erreur en vérifiant l'état de %s: %w", name, err)
		}
	}

	return nil
}
