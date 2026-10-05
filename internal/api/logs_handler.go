package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/logs"
)

const logFilePath = "./data/app.log" // fichier de log à surveiller, à adapter

func LogsWSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("erreur d'upgrade websocket (logs):", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	lines := make(chan string, 100)

	go func() {
		if err := logs.TailFile(ctx, logFilePath, lines); err != nil {
			log.Println("erreur tail de log:", err)
		}
	}()

	// Détecte la déconnexion du client (lecture bloquante sur la connexion)
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-lines:
			if !ok {
				return
			}
			msg, _ := json.Marshal(map[string]string{"type": "log_line", "line": line})
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}
}
