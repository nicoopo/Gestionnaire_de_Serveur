package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/api"
	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/auth"
	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/metrics"
	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/system"
	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/ws"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Aucun fichier .env trouvé, utilisation des variables d'environnement système")
	}

	history := metrics.NewHistory(150) // 150 échantillons à 2s = 5 minutes
	alertTracker := metrics.NewAlertTracker()
	mux := http.NewServeMux()

	hub := ws.NewHub()
	go hub.Run()

	go func() {
		// Configuration du logger : écrit à la fois sur la console et dans un fichier
		logFile, err := os.OpenFile("./data/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal("impossible d'ouvrir le fichier de log:", err)
		}
		defer logFile.Close()

		multiWriter := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(multiWriter)
		log.SetFlags(log.Ldate | log.Ltime)

		if err := godotenv.Load(); err != nil {
			log.Println("Aucun fichier .env trouvé, utilisation des variables d'environnement système")
		}
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			info, err := system.GetSystemInfo()
			if err != nil {
				log.Println("erreur récupération system info:", err)
				continue
			}

			var gpuPercent float64
			if gpus, err := system.ListGPUs(); err == nil && len(gpus) > 0 {
				gpuPercent = gpus[0].UsagePercent
			}

			sample := metrics.Sample{
				Timestamp:   time.Now().Unix(),
				CPUPercent:  info.CPUPercent,
				RAMPercent:  info.RAMPercent,
				DiskPercent: info.DiskPercent,
				GPUPercent:  gpuPercent,
			}
			history.Add(sample)

			// Message système (comme avant, mais enveloppé)
			sysMsg, err := json.Marshal(ws.Message{Type: "system", Data: info})
			if err == nil {
				hub.Broadcast(sysMsg)
			}

			// Alertes de seuil
			alerts := alertTracker.Check(sample)
			if len(alerts) > 0 {
				alertMsg, err := json.Marshal(ws.Message{Type: "alerts", Data: alerts})
				if err == nil {
					hub.Broadcast(alertMsg)
				}
			}
		}
	}()

	// Route publique : login
	mux.HandleFunc("POST /api/login", api.LoginHandler)

	// Sous-routeur protégé pour toutes les routes API sensibles
	protectedAPI := http.NewServeMux()
	protectedAPI.HandleFunc("GET /api/system", api.SystemHandler)
	protectedAPI.HandleFunc("GET /api/processes", api.ProcessesHandler)
	protectedAPI.HandleFunc("POST /api/processes/{pid}/kill", api.KillProcessHandler)
	protectedAPI.HandleFunc("GET /api/services", api.ServicesHandler)
	protectedAPI.HandleFunc("POST /api/services/{name}/start", api.StartServiceHandler)
	protectedAPI.HandleFunc("POST /api/services/{name}/stop", api.StopServiceHandler)
	protectedAPI.HandleFunc("GET /api/disks", api.DisksHandler)
	protectedAPI.HandleFunc("GET /api/gpu", api.GPUHandler)
	protectedAPI.HandleFunc("GET /api/history", api.HistoryHandler(history))

	// Fichiers sandboxés (data/)
	protectedAPI.HandleFunc("GET /api/files", api.ListDirHandler)
	protectedAPI.HandleFunc("GET /api/files/read", api.ReadFileHandler)
	protectedAPI.HandleFunc("GET /api/files/download", api.DownloadFileHandler)

	// Explorateur machine complète
	protectedAPI.HandleFunc("GET /api/explorer/roots", api.FileRootsHandler)
	protectedAPI.HandleFunc("GET /api/explorer/list", api.ExplorerListDirHandler)
	protectedAPI.HandleFunc("GET /api/explorer/read", api.ExplorerReadFileHandler)
	protectedAPI.HandleFunc("GET /api/explorer/download", api.ExplorerDownloadFileHandler)
	protectedAPI.HandleFunc("DELETE /api/files/delete", api.DeleteFileHandler)
	protectedAPI.HandleFunc("DELETE /api/explorer/delete", api.ExplorerDeleteFileHandler)
	protectedAPI.HandleFunc("DELETE /api/explorer/delete-dir", api.ExplorerDeleteDirHandler)

	mux.Handle("/ws/logs", auth.Middleware(http.HandlerFunc(api.LogsWSHandler)))

	mux.Handle("/api/", auth.Middleware(protectedAPI))

	// WebSocket protégé aussi
	mux.Handle("/ws", auth.Middleware(api.WSHandler(hub)))

	// Ping reste public (utile pour un check de santé sans authentification)
	mux.HandleFunc("GET /api/ping", api.PingHandler)

	fs := http.FileServer(http.Dir("./web"))
	mux.Handle("/", fs)

	log.Println("Serveur démarré sur :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
