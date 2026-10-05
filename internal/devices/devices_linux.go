package devices

func GetDevicesSummary() (*DevicesSummary, error) {
	// WMI est spécifique à Windows ; sur Linux on renvoie une structure vide pour l'instant.
	// Une future version pourrait parser lspci/lsusb/lshw pour un équivalent partiel.
	return &DevicesSummary{}, nil
}
