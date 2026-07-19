package files

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type Entry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"is_dir"`
	SizeKB  int64  `json:"size_kb"`
	ModTime string `json:"mod_time"`
	ModUnix int64  `json:"-"` // utilisé pour le tri, pas exposé en JSON
}

func resolveSafePath(root, requestedPath string) (string, error) {
	normalizedRoot := root
	// Sur Windows, "C:" doit être normalisé en "C:\" pour pointer vers la vraie racine du disque
	if len(normalizedRoot) == 2 && normalizedRoot[1] == ':' {
		normalizedRoot += string(filepath.Separator)
	}

	absRoot, err := filepath.Abs(normalizedRoot)
	if err != nil {
		return "", err
	}

	cleaned := filepath.Clean(string(filepath.Separator) + requestedPath)
	fullPath := filepath.Join(absRoot, cleaned)

	absFull, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(absFull, absRoot) {
		return "", errors.New("chemin en dehors du répertoire autorisé")
	}

	return absFull, nil
}

const maxConcurrentStats = 16

// ListDir lit un dossier et récupère les métadonnées de chaque entrée en parallèle
// (worker pool borné), ce qui accélère nettement les dossiers avec beaucoup de fichiers.
func ListDir(root, requestedPath, sortBy, order string) ([]Entry, error) {
	fullPath, err := resolveSafePath(root, requestedPath)
	if err != nil {
		return nil, err
	}

	dirEntries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("impossible de lire le dossier: %w", err)
	}

	cleanedRequestPath := strings.TrimPrefix(filepath.Clean("/"+requestedPath), "/")

	results := make([]Entry, len(dirEntries))
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrentStats) // limite le nombre de goroutines simultanées

	for i, e := range dirEntries {
		wg.Add(1)
		sem <- struct{}{} // prend un "jeton" du pool, bloque si le pool est plein

		go func(i int, e os.DirEntry) {
			defer wg.Done()
			defer func() { <-sem }() // relâche le jeton en fin de traitement

			info, err := e.Info()
			if err != nil {
				return // entrée laissée à zéro-valeur, filtrée ensuite
			}

			relPath := filepath.ToSlash(strings.TrimPrefix(filepath.Join(cleanedRequestPath, e.Name()), "/"))

			results[i] = Entry{
				Name:    e.Name(),
				Path:    relPath,
				IsDir:   e.IsDir(),
				SizeKB:  info.Size() / 1024,
				ModTime: info.ModTime().Format("2006-01-02 15:04"),
				ModUnix: info.ModTime().Unix(),
			}
		}(i, e)
	}

	wg.Wait()

	// Filtre les entrées qui auraient échoué (Name vide = jamais rempli)
	valid := results[:0]
	for _, r := range results {
		if r.Name != "" {
			valid = append(valid, r)
		}
	}

	sortEntries(valid, sortBy, order)

	return valid, nil
}

func sortEntries(entries []Entry, sortBy, order string) {
	desc := order == "desc"

	sort.Slice(entries, func(i, j int) bool {
		// Les dossiers restent toujours groupés avant les fichiers, quel que soit le tri
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}

		var less bool
		switch sortBy {
		case "size":
			less = entries[i].SizeKB < entries[j].SizeKB
		case "date":
			less = entries[i].ModUnix < entries[j].ModUnix
		default: // "name"
			less = strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
		}

		if desc {
			return !less
		}
		return less
	})
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

func DeleteFile(root, requestedPath string) error {
	fullPath, err := resolveSafePath(root, requestedPath)
	if err != nil {
		return err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("fichier introuvable: %w", err)
	}
	if info.IsDir() {
		return errors.New("suppression de dossier non autorisée (fichiers uniquement)")
	}

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("impossible de supprimer le fichier: %w", err)
	}

	return nil
}

func DeleteDirRecursive(root, requestedPath string) error {
	if requestedPath == "" || requestedPath == "/" {
		return errors.New("impossible de supprimer la racine du disque")
	}

	fullPath, err := resolveSafePath(root, requestedPath)
	if err != nil {
		return err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("dossier introuvable: %w", err)
	}
	if !info.IsDir() {
		return errors.New("le chemin ne pointe pas vers un dossier")
	}

	if err := os.RemoveAll(fullPath); err != nil {
		return fmt.Errorf("impossible de supprimer le dossier: %w", err)
	}

	return nil
}
