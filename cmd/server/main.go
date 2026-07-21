package main

import (
	"encoding/json"
	"log"
	"net/http"
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
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		var lastSent, lastRecv uint64
		var lastTime time.Time
		firstTick := true

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

			// Calcul du débit réseau par delta entre deux mesures
			var uploadKBs, downloadKBs float64
			sent, recv, err := system.GetNetworkTotals()
			now := time.Now()
			if err == nil {
				if !firstTick {
					elapsed := now.Sub(lastTime).Seconds()
					if elapsed > 0 {
						uploadKBs = float64(sent-lastSent) / 1024 / elapsed
						downloadKBs = float64(recv-lastRecv) / 1024 / elapsed
					}
				}
				lastSent, lastRecv, lastTime = sent, recv, now
				firstTick = false
			}

			sample := metrics.Sample{
				Timestamp:   now.Unix(),
				CPUPercent:  info.CPUPercent,
				RAMPercent:  info.RAMPercent,
				DiskPercent: info.DiskPercent,
				GPUPercent:  gpuPercent,
				UploadKBs:   uploadKBs,
				DownloadKBs: downloadKBs,
			}
			history.Add(sample)

			// On enrichit le payload broadcast avec le débit réseau
			payload := struct {
				system.SystemInfo
				UploadKBs   float64 `json:"upload_kbs"`
				DownloadKBs float64 `json:"download_kbs"`
			}{
				SystemInfo:  *info,
				UploadKBs:   uploadKBs,
				DownloadKBs: downloadKBs,
			}

			sysMsg, err := json.Marshal(ws.Message{Type: "system", Data: payload})
			if err == nil {
				hub.Broadcast(sysMsg)
			}

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

	protectedAPI.HandleFunc("GET /api/network", api.NetworkHandler)

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
