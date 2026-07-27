package devices

import (
	"encoding/json"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/nicoopo/Gestionnaire_de_Serveur/internal/system"
	"github.com/yusufpapurcu/wmi"
)

// Structs miroir des classes WMI qu'on interroge — les noms de champs
// doivent correspondre exactement aux propriétés WMI (casse comprise).

type win32PnPKeyboard struct {
	Name     string
	DeviceID string
	Status   string
}

type win32PnPMouse struct {
	Name     string
	DeviceID string
	Status   string
}

// wmiMonitorIDJSON reçoit le résultat de la requête WmiMonitorID passée par PowerShell
// (voir winMonitorsViaEDID) : les propriétés de type tableau (ManufacturerName,
// UserFriendlyName) sont aplaties en chaîne côté PowerShell avant le JSON, car
// yusufpapurcu/wmi ne supporte pas les types tableau et panique dessus.
type wmiMonitorIDJSON struct {
	InstanceName     string `json:"InstanceName"`
	ManufacturerName string `json:"ManufacturerName"`
	UserFriendlyName string `json:"UserFriendlyName"`
	Active           bool   `json:"Active"`
}

type win32PnPCamera struct {
	Name         string
	Manufacturer string
	Status       string
}

type win32PnPHID struct {
	Name     string
	DeviceID string
	Status   string
}

type win32SoundDevice struct {
	Name         string
	Manufacturer string
	Status       string
}

type win32USBDevice struct {
	Name         string
	Manufacturer string
	Status       string
}

type win32DiskDrive struct {
	Model  string
	Size   uint64
	Status string
}

type win32NetworkAdapter struct {
	Name                string
	Manufacturer        string
	NetConnectionStatus uint16
}

type win32Printer struct {
	Name   string
	Status string
}

type win32PhysicalMemory struct {
	Manufacturer string
	Capacity     uint64
	Speed        uint32
	PartNumber   string
}
type win32VideoController struct {
	Name          string
	AdapterRAM    uint32
	DriverVersion string
	Status        string
}

type win32BaseBoard struct {
	Manufacturer string
	Product      string
}

type win32BIOS struct {
	Manufacturer      string
	SMBIOSBIOSVersion string
}

type win32Processor struct {
	Name                      string
	Manufacturer              string
	NumberOfCores             uint32
	NumberOfLogicalProcessors uint32
}

var genericHIDNames = map[string]bool{
	"Périphérique d'entrée USB":                           true,
	"Périphérique d’entrée USB":                           true,
	"Périphérique fournisseur HID":                        true,
	"Contrôleur système HID":                              true,
	"Contrôleur système HID à plusieurs axes":             true,
	"Contrôleur de jeu HID":                               true,
	"Périphérique HID conforme Bluetooth Low Energy GATT": true,
}

func GetDevicesSummary() (*DevicesSummary, error) {
	summary := &DevicesSummary{}

	queryTo(&summary.Keyboards, "SELECT Name, DeviceID, Status FROM Win32_PnPEntity WHERE PNPClass='Keyboard'", winPnPKeyboards)
	queryTo(&summary.Mice, "SELECT Name, DeviceID, Status FROM Win32_PnPEntity WHERE PNPClass='Mouse'", winPnPMice)
	queryTo(&summary.Cameras, "SELECT Name, Manufacturer, Status FROM Win32_PnPEntity WHERE PNPClass='Camera'", winCameras)
	summary.Monitors = winMonitorsViaEDID()
	queryTo(&summary.AudioDevices, "SELECT Name, Manufacturer, Status FROM Win32_SoundDevice", winAudio)
	queryTo(&summary.USBDevices, "SELECT Name, Manufacturer, Status FROM Win32_PnPEntity WHERE PNPClass='USB'", winUSB)
	queryTo(&summary.StorageDrives, "SELECT Model, Size, Status FROM Win32_DiskDrive", winStorage)
	queryTo(&summary.NetworkAdapters, "SELECT Name, Manufacturer, NetConnectionStatus FROM Win32_NetworkAdapter WHERE PhysicalAdapter=true", winNetAdapters)
	queryTo(&summary.Printers, "SELECT Name, Status FROM Win32_Printer", winPrinters)
	queryTo(&summary.MemoryModules, "SELECT Manufacturer, Capacity, Speed, PartNumber FROM Win32_PhysicalMemory", winMemory)
	queryTo(&summary.GraphicsCards, "SELECT Name, AdapterRAM, DriverVersion, Status FROM Win32_VideoController", winGraphicsCards)
	queryTo(&summary.HIDDevices, "SELECT Name, DeviceID, Status FROM Win32_PnPEntity WHERE PNPClass='HIDClass'", winPnPHID)

	// Enrichit les noms génériques clavier/souris avec les vrais noms trouvés côté HIDClass,
	// en faisant correspondre par marque (VID) puisque les deux classes ne partagent pas
	// le même DeviceID pour un même périphérique physique.
	usedNames := make(map[string]bool)
	enrichWithRealNames(summary.Keyboards, summary.HIDDevices, usedNames)
	enrichWithRealNames(summary.Mice, summary.HIDDevices, usedNames)

	var boards []win32BaseBoard
	if err := wmi.Query("SELECT Manufacturer, Product FROM Win32_BaseBoard", &boards); err == nil && len(boards) > 0 {
		summary.Motherboard = DeviceInfo{Name: boards[0].Product, Manufacturer: boards[0].Manufacturer}
	}

	var biosList []win32BIOS
	if err := wmi.Query("SELECT Manufacturer, SMBIOSBIOSVersion FROM Win32_BIOS", &biosList); err == nil && len(biosList) > 0 {
		summary.BIOS = DeviceInfo{Name: biosList[0].SMBIOSBIOSVersion, Manufacturer: biosList[0].Manufacturer}
	}

	var procs []win32Processor
	if err := wmi.Query("SELECT Name, Manufacturer, NumberOfCores, NumberOfLogicalProcessors FROM Win32_Processor", &procs); err == nil && len(procs) > 0 {
		p := procs[0]
		summary.Processor = DeviceInfo{
			Name:         p.Name,
			Manufacturer: p.Manufacturer,
			Extra:        formatCores(p.NumberOfCores, p.NumberOfLogicalProcessors),
		}
	}

	return summary, nil
}

// enrichWithRealNames remplace le nom générique d'un item (ex: "Souris HID") par le
// vrai nom commercial trouvé dans hidDevices, si un item de même marque existe côté HID
// et que ce nom n'a pas déjà été utilisé (pour éviter d'assigner le même nom deux fois).

func enrichWithRealNames(items []DeviceInfo, hidDevices []DeviceInfo, used map[string]bool) {
	for i := range items {
		if items[i].Manufacturer == "" {
			continue
		}
		for _, hid := range hidDevices {
			if hid.Manufacturer == items[i].Manufacturer && !used[hid.Name] {
				items[i].Name = hid.Name
				used[hid.Name] = true
				break
			}
		}
	}
}

// queryTo exécute une requête WMI et transforme le résultat en []DeviceInfo via le mapper fourni,
// sans faire échouer tout le reste si une catégorie précise échoue (matériel absent, permissions, etc.).
func queryTo[T any](dest *[]DeviceInfo, query string, mapper func([]T) []DeviceInfo) {
	var raw []T
	if err := wmi.Query(query, &raw); err != nil {
		log.Printf("requête WMI échouée (%s): %v", query, err)
		return
	}
	*dest = mapper(raw)
}

func winPnPKeyboards(raw []win32PnPKeyboard) []DeviceInfo {
	out := make([]DeviceInfo, 0, len(raw))
	for _, r := range raw {
		manufacturer := extractVendorFromDeviceID(r.DeviceID)
		out = append(out, DeviceInfo{Name: r.Name, Manufacturer: manufacturer, Status: r.Status})
	}
	return out
}

func winPnPMice(raw []win32PnPMouse) []DeviceInfo {
	out := make([]DeviceInfo, 0, len(raw))
	for _, r := range raw {
		manufacturer := extractVendorFromDeviceID(r.DeviceID)
		out = append(out, DeviceInfo{Name: r.Name, Manufacturer: manufacturer, Status: r.Status})
	}
	return out
}

func winCameras(raw []win32PnPCamera) []DeviceInfo {
	out := make([]DeviceInfo, 0, len(raw))
	for _, r := range raw {
		out = append(out, DeviceInfo{Name: r.Name, Manufacturer: r.Manufacturer, Status: r.Status})
	}
	return out
}

// winMonitorsViaEDID passe par PowerShell/CIM plutôt que par wmi.QueryNamespace :
// la lib yusufpapurcu/wmi ne supporte pas les propriétés de type tableau
// (ManufacturerName/UserFriendlyName sont des []uint16 côté WMI), ce qui provoquait
// un panic reflect côté Go ("call of reflect.Value.Uint on int32 Value"). PowerShell
// aplatit les tableaux en chaîne avant le JSON, donc plus de mapping struct côté wmi.
func winMonitorsViaEDID() []DeviceInfo {
	psCmd := `Get-CimInstance -Namespace root\wmi -ClassName WmiMonitorID | ForEach-Object {
		[PSCustomObject]@{
			InstanceName     = $_.InstanceName
			ManufacturerName = ($_.ManufacturerName -join ',')
			UserFriendlyName = ($_.UserFriendlyName -join ',')
			Active           = $_.Active
		}
	} | ConvertTo-Json -Compress`

	output, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd).Output()
	if err != nil {
		log.Printf("requête PowerShell échouée (WmiMonitorID): %v", err)
		return nil
	}

	var raw []wmiMonitorIDJSON
	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		return nil
	}
	if strings.HasPrefix(trimmed, "{") {
		// Get-CimInstance renvoie un objet seul (pas un tableau JSON) s'il n'y a qu'un écran
		var single wmiMonitorIDJSON
		if err := json.Unmarshal(output, &single); err != nil {
			log.Printf("parsing JSON échoué (WmiMonitorID): %v", err)
			return nil
		}
		raw = []wmiMonitorIDJSON{single}
	} else if err := json.Unmarshal(output, &raw); err != nil {
		log.Printf("parsing JSON échoué (WmiMonitorID): %v", err)
		return nil
	}

	out := make([]DeviceInfo, 0, len(raw))
	for _, r := range raw {
		name := decodeCommaSeparatedCodes(r.UserFriendlyName)
		manufacturer := decodeCommaSeparatedCodes(r.ManufacturerName)

		status := "inactif"
		if r.Active {
			status = "OK"
		}

		if name == "" {
			name = "Écran (nom indisponible)"
		}

		out = append(out, DeviceInfo{Name: name, Manufacturer: manufacturer, Status: status})
	}
	return out
}

// decodeCommaSeparatedCodes convertit une liste "72,68,77,73,..." (codes ASCII EDID
// séparés par des virgules, aplatis côté PowerShell) en la chaîne lisible correspondante,
// en s'arrêtant au premier zéro de padding.
func decodeCommaSeparatedCodes(s string) string {
	if s == "" {
		return ""
	}
	var sb strings.Builder
	for _, part := range strings.Split(s, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || n == 0 {
			break
		}
		sb.WriteByte(byte(n))
	}
	return strings.TrimSpace(sb.String())
}

func isGenericHIDName(name string) bool {
	if genericHIDNames[name] {
		return true
	}
	return strings.Contains(name, "conforme aux Périphériques d'interface utilisateur")
}

func winPnPHID(raw []win32PnPHID) []DeviceInfo {
	out := make([]DeviceInfo, 0, len(raw))
	seen := make(map[string]bool)

	for _, r := range raw {
		if isGenericHIDName(r.Name) {
			continue // on filtre le bruit générique
		}
		if seen[r.Name] {
			continue // évite les doublons (même périphérique vu plusieurs fois)
		}
		seen[r.Name] = true

		manufacturer := extractVendorFromDeviceID(r.DeviceID)
		out = append(out, DeviceInfo{Name: r.Name, Manufacturer: manufacturer, Status: r.Status})
	}
	return out
}

func winAudio(raw []win32SoundDevice) []DeviceInfo {
	out := make([]DeviceInfo, 0, len(raw))
	for _, r := range raw {
		out = append(out, DeviceInfo{Name: r.Name, Manufacturer: r.Manufacturer, Status: r.Status})
	}
	return out
}

func winUSB(raw []win32USBDevice) []DeviceInfo {
	out := make([]DeviceInfo, 0, len(raw))
	for _, r := range raw {
		out = append(out, DeviceInfo{Name: r.Name, Manufacturer: r.Manufacturer, Status: r.Status})
	}
	return out
}

func winStorage(raw []win32DiskDrive) []DeviceInfo {
	out := make([]DeviceInfo, 0, len(raw))
	for _, r := range raw {
		gb := r.Size / 1024 / 1024 / 1024
		out = append(out, DeviceInfo{Name: r.Model, Status: r.Status, Extra: formatGB(gb)})
	}
	return out
}

func winNetAdapters(raw []win32NetworkAdapter) []DeviceInfo {
	out := make([]DeviceInfo, 0, len(raw))
	for _, r := range raw {
		status := "inconnu"
		switch r.NetConnectionStatus {
		case 2:
			status = "connecté"
		case 7:
			status = "déconnecté (media)"
		case 0:
			status = "déconnecté"
		}
		out = append(out, DeviceInfo{Name: r.Name, Manufacturer: r.Manufacturer, Status: status})
	}
	return out
}

func winPrinters(raw []win32Printer) []DeviceInfo {
	out := make([]DeviceInfo, 0, len(raw))
	for _, r := range raw {
		out = append(out, DeviceInfo{Name: r.Name, Status: r.Status})
	}
	return out
}

func winMemory(raw []win32PhysicalMemory) []DeviceInfo {
	out := make([]DeviceInfo, 0, len(raw))
	for _, r := range raw {
		gb := r.Capacity / 1024 / 1024 / 1024
		out = append(out, DeviceInfo{
			Name:         r.PartNumber,
			Manufacturer: r.Manufacturer,
			Extra:        formatMemory(gb, r.Speed),
		})
	}
	return out
}

func winGraphicsCards(raw []win32VideoController) []DeviceInfo {
	// nvidia-smi donne la vraie VRAM (WMI plafonne à ~4 Go sur les GPU récents, bug connu)
	nvidiaGPUs, _ := system.ListGPUs()

	out := make([]DeviceInfo, 0, len(raw))
	for _, r := range raw {
		extra := "Pilote " + r.DriverVersion

		if realVRAM := findRealVRAM(r.Name, nvidiaGPUs); realVRAM > 0 {
			extra = formatGB(realVRAM) + " VRAM · " + extra
		} else if vramGB := uint64(r.AdapterRAM) / 1024 / 1024 / 1024; vramGB > 0 {
			extra = formatGB(vramGB) + " VRAM (approx.) · " + extra
		}

		out = append(out, DeviceInfo{Name: r.Name, Status: r.Status, Extra: extra})
	}
	return out
}

func findRealVRAM(name string, gpus []system.GPUInfo) uint64 {
	for _, g := range gpus {
		if strings.Contains(name, g.Name) || strings.Contains(g.Name, name) {
			return g.MemTotalMB / 1024
		}
	}
	return 0
}

func formatResolution(w, h uint32) string {
	return itoa(int(w)) + "x" + itoa(int(h))
}

func formatGB(gb uint64) string {
	return itoa(int(gb)) + " GB"
}

func formatMemory(gb uint64, speed uint32) string {
	return itoa(int(gb)) + " GB @ " + itoa(int(speed)) + " MHz"
}

func formatCores(physical, logical uint32) string {
	return itoa(int(physical)) + " cœurs / " + itoa(int(logical)) + " threads"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}
