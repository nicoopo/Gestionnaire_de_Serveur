package api

import (
	"encoding/json"
	"net/http"
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
