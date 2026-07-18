package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/process"
	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/service"
	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/system"
)

type PingResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type KillResponse struct {
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
	nameFilter := r.URL.Query().Get("name")

	procs, err := process.ListProcesses(nameFilter)
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

func KillProcessHandler(w http.ResponseWriter, r *http.Request) {
	pidStr := r.PathValue("pid")

	pid, err := strconv.ParseInt(pidStr, 10, 32)
	if err != nil {
		http.Error(w, "pid invalide: "+pidStr, http.StatusBadRequest)
		return
	}

	if err := process.KillProcess(int32(pid)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := KillResponse{
		Status:  "ok",
		Message: "process arrêté",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func ServicesHandler(w http.ResponseWriter, r *http.Request) {
	services, err := service.ListServices()
	if err != nil {
		http.Error(w, "failed to list services: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func StartServiceHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	if err := service.StartService(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := KillResponse{Status: "ok", Message: "service démarré"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func StopServiceHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	if err := service.StopService(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := KillResponse{Status: "ok", Message: "service arrêté"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
