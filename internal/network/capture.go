package network

import (
	"fmt"
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

// CapturePackets ouvre une interface et pousse chaque paquet capturé dans le channel,
// jusqu'à ce que le channel stop soit fermé.
func CapturePackets(interfaceName string, packets chan<- PacketInfo, stop <-chan struct{}) error {
	handle, err := pcap.OpenLive(interfaceName, 1600, true, pcap.BlockForever)
	if err != nil {
		return fmt.Errorf("impossible d'ouvrir l'interface %s: %w", interfaceName, err)
	}
	defer handle.Close()

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

	case packet.Layer(layers.LayerTypeICMPv4) != nil:
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
