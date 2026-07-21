package network

import (
	"net"
	"strings"
	"sync"
)

type dnsCache struct {
	mu    sync.RWMutex
	cache map[string]string
}

var dnsResolver = &dnsCache{cache: make(map[string]string)}

// Lookup renvoie immédiatement le nom en cache s'il existe, sinon lance une
// résolution en arrière-plan et renvoie une chaîne vide pour ce paquet-ci
// (le nom apparaîtra sur les paquets suivants une fois résolu).
func (d *dnsCache) Lookup(ip string) string {
	d.mu.RLock()
	name, ok := d.cache[ip]
	d.mu.RUnlock()
	if ok {
		return name
	}

	d.mu.Lock()
	d.cache[ip] = "" // marque "en cours" pour éviter de relancer 50 résolutions pour le même IP
	d.mu.Unlock()

	go func() {
		names, err := net.LookupAddr(ip)
		resolved := ""
		if err == nil && len(names) > 0 {
			resolved = strings.TrimSuffix(names[0], ".")
		}
		d.mu.Lock()
		d.cache[ip] = resolved
		d.mu.Unlock()
	}()

	return ""
}
