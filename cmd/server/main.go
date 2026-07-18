package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/api"
	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/system"
	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/ws"
)

func main() {
	mux := http.NewServeMux()

	// Hub WebSocket
	hub := ws.NewHub()
	go hub.Run()

	// Goroutine qui pousse les métriques système toutes les 2 secondes
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			info, err := system.GetSystemInfo()
			if err != nil {
				log.Println("erreur récupération system info:", err)
				continue
			}
			data, err := json.Marshal(info)
			if err != nil {
				continue
			}
			hub.Broadcast(data)
		}
	}()

	// API REST
	mux.HandleFunc("GET /api/ping", api.PingHandler)
	mux.HandleFunc("GET /api/system", api.SystemHandler)
	mux.HandleFunc("GET /api/processes", api.ProcessesHandler)
	mux.HandleFunc("POST /api/processes/{pid}/kill", api.KillProcessHandler)
	mux.HandleFunc("GET /api/services", api.ServicesHandler)
	mux.HandleFunc("POST /api/services/{name}/start", api.StartServiceHandler)
	mux.HandleFunc("POST /api/services/{name}/stop", api.StopServiceHandler)

	// WebSocket
	mux.HandleFunc("GET /ws", api.WSHandler(hub))

	// Fichiers statiques
	fs := http.FileServer(http.Dir("./web"))
	mux.Handle("/", fs)

	log.Println("Serveur démarré sur :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
