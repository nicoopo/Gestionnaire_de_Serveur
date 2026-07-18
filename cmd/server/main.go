package main

import (
	"log"
	"net/http"

	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/api"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/ping", api.PingHandler)
	mux.HandleFunc("GET /api/system", api.SystemHandler)

	log.Println("Serveur démarré sur :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
