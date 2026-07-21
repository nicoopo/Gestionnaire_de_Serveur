package system

import (
	"github.com/shirou/gopsutil/v3/net"
)

type InterfaceInfo struct {
	Name        string `json:"name"`
	BytesSentMB uint64 `json:"bytes_sent_mb"`
	BytesRecvMB uint64 `json:"bytes_recv_mb"`
}

func ListInterfaces() ([]InterfaceInfo, error) {
	counters, err := net.IOCounters(true) // true = détail par interface
	if err != nil {
		return nil, err
	}

	result := make([]InterfaceInfo, 0, len(counters))
	for _, c := range counters {
		if c.BytesSent == 0 && c.BytesRecv == 0 {
			continue // interface inactive, on l'ignore pour ne pas polluer l'affichage
		}
		result = append(result, InterfaceInfo{
			Name:        c.Name,
			BytesSentMB: c.BytesSent / 1024 / 1024,
			BytesRecvMB: c.BytesRecv / 1024 / 1024,
		})
	}

	return result, nil
}

// GetNetworkTotals renvoie les compteurs cumulés (octets depuis le démarrage machine),
// tous adaptateurs confondus. À utiliser en calculant un delta entre deux appels pour obtenir un débit.
func GetNetworkTotals() (bytesSent, bytesRecv uint64, err error) {
	counters, err := net.IOCounters(false) // false = total agrégé, une seule ligne
	if err != nil {
		return 0, 0, err
	}
	if len(counters) == 0 {
		return 0, 0, nil
	}
	return counters[0].BytesSent, counters[0].BytesRecv, nil
}
