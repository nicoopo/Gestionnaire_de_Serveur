package files

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Entry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"is_dir"`
	SizeKB  int64  `json:"size_kb"`
	ModTime string `json:"mod_time"`
}

// resolveSafePath combine la racine autorisée et le chemin demandé,
// puis vérifie que le résultat reste bien à l'intérieur de la racine.
func resolveSafePath(root, requestedPath string) (string, error) {
	cleaned := filepath.Clean("/" + requestedPath) // force un chemin absolu depuis la racine virtuelle
	fullPath := filepath.Join(root, cleaned)

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	absFull, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(absFull, absRoot) {
		return "", errors.New("chemin en dehors du répertoire autorisé")
	}

	return absFull, nil
}

func ListDir(root, requestedPath string) ([]Entry, error) {
	fullPath, err := resolveSafePath(root, requestedPath)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("impossible de lire le dossier: %w", err)
	}

	// Chemin relatif "propre" recalculé depuis la racine réelle, pas depuis l'input brut
	cleanedRequestPath := strings.TrimPrefix(filepath.Clean("/"+requestedPath), "/")

	result := make([]Entry, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}

		relPath := strings.TrimPrefix(filepath.ToSlash(filepath.Join(cleanedRequestPath, e.Name())), "/")

		result = append(result, Entry{
			Name:    e.Name(),
			Path:    relPath,
			IsDir:   e.IsDir(),
			SizeKB:  info.Size() / 1024,
			ModTime: info.ModTime().Format("2006-01-02 15:04"),
		})
	}

	return result, nil
}

func ReadTextFile(root, requestedPath string, maxBytes int64) (string, error) {
	fullPath, err := resolveSafePath(root, requestedPath)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return "", fmt.Errorf("fichier introuvable: %w", err)
	}
	if info.IsDir() {
		return "", errors.New("le chemin pointe vers un dossier, pas un fichier")
	}
	if info.Size() > maxBytes {
		return "", fmt.Errorf("fichier trop volumineux (%d octets, max %d)", info.Size(), maxBytes)
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("impossible de lire le fichier: %w", err)
	}

	return string(content), nil
}

func ResolveForDownload(root, requestedPath string) (string, error) {
	fullPath, err := resolveSafePath(root, requestedPath)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return "", fmt.Errorf("fichier introuvable: %w", err)
	}
	if info.IsDir() {
		return "", errors.New("impossible de télécharger un dossier")
	}

	return fullPath, nil
}

var _ fs.FileInfo // évite un import inutilisé si tu retires ReadTextFile plus tard
