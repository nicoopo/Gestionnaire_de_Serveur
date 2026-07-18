package api

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/files"
)

const filesRoot = "./data"                  // dossier racine autorisé, à créer à côté de main.go
const maxReadableFileSize = 2 * 1024 * 1024 // 2 Mo

func ListDirHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	entries, err := files.ListDir(filesRoot, path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func ReadFileHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	content, err := files.ReadTextFile(filesRoot, path, maxReadableFileSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"content": content})
}

func DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	fullPath, err := files.ResolveForDownload(filesRoot, path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(fullPath)+"\"")
	http.ServeFile(w, r, fullPath)
}
