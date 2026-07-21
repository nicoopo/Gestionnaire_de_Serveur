package network

import (
	"strconv"
	"strings"
	"sync"
	"time"

	gopsutilnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

type processMap struct {
	mu     sync.RWMutex
	byPort map[string]string // ex: "tcp:51820" -> "firefox.exe"
}

var procMap = &processMap{byPort: make(map[string]string)}

// StartProcessMapRefresher lance une goroutine qui reconstruit la table port->processus
// toutes les 3 secondes (les connexions changent en permanence).
func StartProcessMapRefresher() {
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		refreshProcessMap()
		for range ticker.C {
			refreshProcessMap()
		}
	}()
}

func refreshProcessMap() {
	conns, err := gopsutilnet.Connections("inet")
	if err != nil {
		return
	}

	newMap := make(map[string]string)
	for _, c := range conns {
		if c.Pid == 0 || c.Laddr.Port == 0 {
			continue
		}
		proto := "tcp"
		if c.Type == 2 { // SOCK_DGRAM
			proto = "udp"
		}
		key := proto + ":" + strconv.Itoa(int(c.Laddr.Port))
		newMap[key] = processNameCached(c.Pid)
	}

	procMap.mu.Lock()
	procMap.byPort = newMap
	procMap.mu.Unlock()
}

var nameCacheMu sync.Mutex
var nameCache = make(map[int32]string)

func processNameCached(pid int32) string {
	nameCacheMu.Lock()
	defer nameCacheMu.Unlock()
	if n, ok := nameCache[pid]; ok {
		return n
	}
	p, err := process.NewProcess(pid)
	name := "inconnu"
	if err == nil {
		if n, err := p.Name(); err == nil {
			name = n
		}
	}
	nameCache[pid] = name
	return name
}

func LookupProcess(protocol, port string) string {
	key := strings.ToLower(protocol) + ":" + port
	procMap.mu.RLock()
	defer procMap.mu.RUnlock()
	return procMap.byPort[key]
}
