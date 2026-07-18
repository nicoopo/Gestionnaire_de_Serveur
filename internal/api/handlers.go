package api

import (
	"encoding/json"
	"net/http"

	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/process"
	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/system"
)

type PingResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	resp := PingResponse{
		Status:  "ok",
		Message: "server-manager is alive",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func SystemHandler(w http.ResponseWriter, r *http.Request) {
	info, err := system.GetSystemInfo()
	if err != nil {
		http.Error(w, "failed to get system info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func ProcessesHandler(w http.ResponseWriter, r *http.Request) {
	procs, err := process.ListProcesses()
	if err != nil {
		http.Error(w, "failed to list processes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(procs); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
