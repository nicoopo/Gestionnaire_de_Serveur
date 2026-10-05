package api

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/ws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // dev only, à restreindre si tu déploies un jour
	},
}

func WSHandler(hub *ws.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("erreur d'upgrade websocket:", err)
			return
		}

		client := ws.NewClient(hub, conn)
		hub.Register(client)

		go client.WritePump()
		go client.ReadPump()
	}
}
