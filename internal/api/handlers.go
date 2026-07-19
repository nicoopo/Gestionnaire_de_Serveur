package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/metrics"
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

	// Récupère le nom avant de tuer le process, pour un log plus lisible
	processName := process.GetProcessName(int32(pid))

	if err := process.KillProcess(int32(pid)); err != nil {
		log.Printf("échec arrêt de %s (PID %d): %v", processName, pid, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("processus arrêté : %s (PID %d)", processName, pid)

	resp := KillResponse{Status: "ok", Message: "process arrêté"}
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
		log.Printf("échec démarrage du service %s: %v", name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("service démarré : %s", name)

	resp := KillResponse{Status: "ok", Message: "service démarré"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func StopServiceHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	if err := service.StopService(name); err != nil {
		log.Printf("échec arrêt du service %s: %v", name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("service arrêté : %s", name)

	resp := KillResponse{Status: "ok", Message: "service arrêté"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func DisksHandler(w http.ResponseWriter, r *http.Request) {
	disks, err := system.ListDisks()
	if err != nil {
		http.Error(w, "failed to list disks: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(disks)
}

func GPUHandler(w http.ResponseWriter, r *http.Request) {
	gpus, err := system.ListGPUs()
	if err != nil {
		// pas une erreur bloquante : certaines machines n'ont pas de GPU NVIDIA
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]system.GPUInfo{})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gpus)
}

func HistoryHandler(h *metrics.History) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(h.All())
	}
}
