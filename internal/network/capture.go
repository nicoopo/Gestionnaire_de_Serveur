package network

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type PacketInfo struct {
	Timestamp   string `json:"timestamp"`
	SrcIP       string `json:"src_ip"`
	DstIP       string `json:"dst_ip"`
	SrcHost     string `json:"src_host"`
	DstHost     string `json:"dst_host"`
	SrcPort     string `json:"src_port"`
	DstPort     string `json:"dst_port"`
	Protocol    string `json:"protocol"`
	Length      int    `json:"length"`
	TCPFlags    string `json:"tcp_flags"`
	ProcessName string `json:"process_name"`
}

type InterfaceOption struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func ListInterfaceOptions() ([]InterfaceOption, error) {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return nil, fmt.Errorf("impossible de lister les interfaces (Npcap est-il installé ?): %w", err)
	}

	options := make([]InterfaceOption, 0, len(devices))
	for _, d := range devices {
		desc := d.Description
		if desc == "" {
			desc = d.Name
		}
		options = append(options, InterfaceOption{Name: d.Name, Description: desc})
	}
	return options, nil
}

// BuildBPFFilter construit une expression de filtre BPF (appliquée par Npcap au niveau capture,
// avant que les paquets n'atteignent l'application) à partir d'une liste de protocoles et,
// optionnellement, d'une adresse IP ou d'une plage CIDR.
func BuildBPFFilter(protocols []string, ip string) (string, error) {
	var protoExpr string
	if len(protocols) > 0 {
		var clauses []string
		seen := make(map[string]bool)
		for _, p := range protocols {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}

			var clause string
			switch strings.ToUpper(p) {
			case "TCP":
				clause = "tcp"
			case "UDP":
				clause = "udp"
			case "ICMP":
				clause = "(icmp or icmp6)"
			case "DNS":
				clause = "udp port 53"
			case "AUTRE", "OTHER":
				clause = "(not tcp and not udp and not icmp and not icmp6)"
			default:
				return "", fmt.Errorf("protocole de filtre inconnu: %s", p)
			}

			if !seen[clause] {
				seen[clause] = true
				clauses = append(clauses, clause)
			}
		}
		if len(clauses) > 0 {
			protoExpr = strings.Join(clauses, " or ")
			if len(clauses) > 1 {
				protoExpr = "(" + protoExpr + ")"
			}
		}
	}

	var ipExpr string
	if ip = strings.TrimSpace(ip); ip != "" {
		if strings.Contains(ip, "/") {
			if _, _, err := net.ParseCIDR(ip); err != nil {
				return "", fmt.Errorf("plage IP invalide: %s", ip)
			}
			ipExpr = "net " + ip
		} else {
			if net.ParseIP(ip) == nil {
				return "", fmt.Errorf("adresse IP invalide: %s", ip)
			}
			ipExpr = "host " + ip
		}
	}

	switch {
	case protoExpr != "" && ipExpr != "":
		return protoExpr + " and " + ipExpr, nil
	case protoExpr != "":
		return protoExpr, nil
	case ipExpr != "":
		return ipExpr, nil
	default:
		return "", nil
	}
}

// CapturePackets ouvre une interface et pousse chaque paquet capturé dans le channel,
// jusqu'à ce que le channel stop soit fermé. Si filter n'est pas vide, il est appliqué
// comme filtre BPF au niveau de la capture (voir BuildBPFFilter).
func CapturePackets(interfaceName, filter string, packets chan<- PacketInfo, stop <-chan struct{}) error {
	handle, err := pcap.OpenLive(interfaceName, 1600, true, pcap.BlockForever)
	if err != nil {
		return fmt.Errorf("impossible d'ouvrir l'interface %s: %w", interfaceName, err)
	}
	defer handle.Close()

	if filter != "" {
		if err := handle.SetBPFFilter(filter); err != nil {
			return fmt.Errorf("filtre invalide %q: %w", filter, err)
		}
	}

	source := gopacket.NewPacketSource(handle, handle.LinkType())
	packetChan := source.Packets()

	for {
		select {
		case <-stop:
			return nil
		case packet, ok := <-packetChan:
			if !ok {
				return nil
			}
			info := parsePacket(packet)
			select {
			case packets <- info:
			case <-stop:
				return nil
			}
		}
	}
}

func parsePacket(packet gopacket.Packet) PacketInfo {
	info := PacketInfo{
		Timestamp: time.Now().Format("15:04:05.000"),
		Length:    packet.Metadata().Length,
		Protocol:  "autre",
	}

	if netLayer := packet.NetworkLayer(); netLayer != nil {
		src, dst := netLayer.NetworkFlow().Endpoints()
		info.SrcIP = src.String()
		info.DstIP = dst.String()
		info.SrcHost = dnsResolver.Lookup(info.SrcIP)
		info.DstHost = dnsResolver.Lookup(info.DstIP)
	}

	switch {
	case packet.Layer(layers.LayerTypeDNS) != nil:
		info.Protocol = "DNS"
		if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
			udp, _ := udpLayer.(*layers.UDP)
			info.SrcPort = udp.SrcPort.String()
			info.DstPort = udp.DstPort.String()
		}

	case packet.Layer(layers.LayerTypeTCP) != nil:
		tcp, _ := packet.Layer(layers.LayerTypeTCP).(*layers.TCP)
		info.SrcPort = tcp.SrcPort.String()
		info.DstPort = tcp.DstPort.String()
		info.Protocol = "TCP"
		info.TCPFlags = tcpFlagsString(tcp)
		info.ProcessName = firstNonEmpty(
			LookupProcess("tcp", info.SrcPort),
			LookupProcess("tcp", info.DstPort),
		)

	case packet.Layer(layers.LayerTypeUDP) != nil:
		udp, _ := packet.Layer(layers.LayerTypeUDP).(*layers.UDP)
		info.SrcPort = udp.SrcPort.String()
		info.DstPort = udp.DstPort.String()
		info.Protocol = "UDP"
		info.ProcessName = firstNonEmpty(
			LookupProcess("udp", info.SrcPort),
			LookupProcess("udp", info.DstPort),
		)

	case packet.Layer(layers.LayerTypeICMPv4) != nil, packet.Layer(layers.LayerTypeICMPv6) != nil:
		info.Protocol = "ICMP"
	}

	return info
}

func tcpFlagsString(tcp *layers.TCP) string {
	var flags []string
	if tcp.SYN {
		flags = append(flags, "SYN")
	}
	if tcp.ACK {
		flags = append(flags, "ACK")
	}
	if tcp.FIN {
		flags = append(flags, "FIN")
	}
	if tcp.RST {
		flags = append(flags, "RST")
	}
	if tcp.PSH {
		flags = append(flags, "PSH")
	}
	if tcp.URG {
		flags = append(flags, "URG")
	}
	return strings.Join(flags, ",")
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
