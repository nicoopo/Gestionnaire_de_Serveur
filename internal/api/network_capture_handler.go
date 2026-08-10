package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/network"
)

func NetworkInterfacesHandler(w http.ResponseWriter, r *http.Request) {
	options, err := network.ListInterfaceOptions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(options)
}

func CaptureWSHandler(w http.ResponseWriter, r *http.Request) {
	interfaceName := r.URL.Query().Get("iface")
	if interfaceName == "" {
		http.Error(w, "paramètre 'iface' requis", http.StatusBadRequest)
		return
	}

	var protocols []string
	if proto := r.URL.Query().Get("proto"); proto != "" {
		protocols = strings.Split(proto, ",")
	}
	ip := r.URL.Query().Get("ip")

	filter, err := network.BuildBPFFilter(protocols, ip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("erreur d'upgrade websocket (capture):", err)
		return
	}
	defer conn.Close()

	_, cancel := context.WithCancel(r.Context())
	defer cancel()

	stop := make(chan struct{})
	packets := make(chan network.PacketInfo, 200)

	go func() {
		if err := network.CapturePackets(interfaceName, filter, packets, stop); err != nil {
			log.Println("erreur de capture réseau:", err)
		}
	}()

	// Détecte la déconnexion du client
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				close(stop)
				return
			}
		}
	}()

	for {
		select {
		case <-r.Context().Done():
			close(stop)
			return
		case pkt, ok := <-packets:
			if !ok {
				return
			}
			msg, _ := json.Marshal(map[string]interface{}{"type": "packet", "data": pkt})
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				close(stop)
				return
			}
		}
	}
}
