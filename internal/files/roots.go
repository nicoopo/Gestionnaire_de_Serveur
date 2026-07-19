package files

import (
	"errors"

	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/system"
)

// AllowedRoots renvoie la liste des points de montage/disques valides comme racines de navigation
func AllowedRoots() ([]string, error) {
	disks, err := system.ListDisks()
	if err != nil {
		return nil, err
	}

	roots := make([]string, 0, len(disks))
	for _, d := range disks {
		roots = append(roots, d.Mountpoint)
	}
	return roots, nil
}

// ValidateRoot vérifie que le root demandé fait bien partie des disques réels de la machine
func ValidateRoot(requestedRoot string) (string, error) {
	roots, err := AllowedRoots()
	if err != nil {
		return "", err
	}

	for _, r := range roots {
		if r == requestedRoot {
			return r, nil
		}
	}

	return "", errors.New("racine non autorisée")
}
