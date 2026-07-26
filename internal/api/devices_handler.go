package api

import (
	"encoding/json"
	"net/http"

	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/devices"
)

func DevicesHandler(w http.ResponseWriter, r *http.Request) {
	summary, err := devices.GetDevicesSummary()
	if err != nil {
		http.Error(w, "failed to get devices: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
