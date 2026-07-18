package auth

import (
	"net/http"
	"strings"
)

// Middleware protège un handler en exigeant un token JWT valide
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)

		if token == "" {
			http.Error(w, "authentification requise", http.StatusUnauthorized)
			return
		}

		if err := ValidateToken(token); err != nil {
			http.Error(w, "token invalide ou expiré", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func extractToken(r *http.Request) string {
	// Header Authorization: Bearer <token> (utilisé par les requêtes REST classiques)
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Query param ?token=... (nécessaire pour le WebSocket, qui ne peut pas envoyer de header custom depuis le navigateur)
	return r.URL.Query().Get("token")
}
