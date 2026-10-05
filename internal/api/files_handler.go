package api

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"

	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/files"
)

const filesRoot = "./data"
const maxReadableFileSize = 2 * 1024 * 1024

// --- Explorateur sandboxé sur ./data (existant, inchangé) ---

func ListDirHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	sortBy := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	entries, err := files.ListDir(filesRoot, path, sortBy, order)
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

// --- Explorateur machine complète (nouveau, avec sélection de racine) ---

func FileRootsHandler(w http.ResponseWriter, r *http.Request) {
	roots, err := files.AllowedRoots()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roots)
}

func ExplorerListDirHandler(w http.ResponseWriter, r *http.Request) {
	rootParam := r.URL.Query().Get("root")
	path := r.URL.Query().Get("path")
	sortBy := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	root, err := files.ValidateRoot(rootParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	entries, err := files.ListDir(root, path, sortBy, order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func ExplorerReadFileHandler(w http.ResponseWriter, r *http.Request) {
	rootParam := r.URL.Query().Get("root")
	path := r.URL.Query().Get("path")

	root, err := files.ValidateRoot(rootParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	content, err := files.ReadTextFile(root, path, maxReadableFileSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"content": content})
}

func ExplorerDownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	rootParam := r.URL.Query().Get("root")
	path := r.URL.Query().Get("path")

	root, err := files.ValidateRoot(rootParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fullPath, err := files.ResolveForDownload(root, path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(fullPath)+"\"")
	http.ServeFile(w, r, fullPath)
}

func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	if err := files.DeleteFile(filesRoot, path); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func ExplorerDeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	rootParam := r.URL.Query().Get("root")
	path := r.URL.Query().Get("path")

	root, err := files.ValidateRoot(rootParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := files.DeleteFile(root, path); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func ExplorerDeleteDirHandler(w http.ResponseWriter, r *http.Request) {
	rootParam := r.URL.Query().Get("root")
	path := r.URL.Query().Get("path")

	root, err := files.ValidateRoot(rootParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := files.DeleteDirRecursive(root, path); err != nil {
		log.Printf("échec suppression dossier %s%s: %v", root, path, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("dossier supprimé : %s%s", root, path)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
