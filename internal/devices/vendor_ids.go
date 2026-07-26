package devices

import "strings"

// Table des VID USB les plus courants pour claviers/souris.
// Liste non exhaustive (il en existe des milliers), mais couvre les fabricants grand public.
var usbVendorIDs = map[string]string{
	"046D": "Logitech",
	"045E": "Microsoft",
	"1532": "Razer",
	"1038": "SteelSeries",
	"0951": "Kingston (HyperX)",
	"04D9": "Holtek (générique)",
	"258A": "SINOWEALTH (générique gaming)",
	"05AC": "Apple",
	"0B05": "ASUS",
	"1E7D": "ROCCAT",
	"3554": "Trust",
	"2717": "Xiaomi",
	"04B4": "Cypress (générique)",
	"093A": "Pixart (générique)",
}

// extractVendorFromDeviceID cherche un VID (Vendor ID) dans un DeviceID Windows
// du type "HID\VID_046D&PID_C52B\..." et renvoie le nom du fabricant si connu.
func extractVendorFromDeviceID(deviceID string) string {
	upper := strings.ToUpper(deviceID)
	idx := strings.Index(upper, "VID_")
	if idx == -1 || idx+8 > len(upper) {
		return ""
	}
	vid := upper[idx+4 : idx+8]
	if name, ok := usbVendorIDs[vid]; ok {
		return name
	}
	return ""
}
