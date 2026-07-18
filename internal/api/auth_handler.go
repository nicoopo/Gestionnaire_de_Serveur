package api

import (
	"encoding/json"
	"net/http"

	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/auth"
)

type LoginRequest struct {
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "requête invalide", http.StatusBadRequest)
		return
	}

	if !auth.CheckPassword(req.Password) {
		http.Error(w, "mot de passe incorrect", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken()
	if err != nil {
		http.Error(w, "erreur lors de la génération du token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{Token: token})
}
