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
	mux.HandleFunc("GET /api/processes", api.ProcessesHandler)
	mux.HandleFunc("POST /api/processes/{pid}/kill", api.KillProcessHandler)
	mux.HandleFunc("GET /api/services", api.ServicesHandler)
	mux.HandleFunc("POST /api/services/{name}/start", api.StartServiceHandler)
	mux.HandleFunc("POST /api/services/{name}/stop", api.StopServiceHandler)

	fs := http.FileServer(http.Dir("./web"))
	mux.Handle("/", fs)

	log.Println("Serveur démarré sur :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
