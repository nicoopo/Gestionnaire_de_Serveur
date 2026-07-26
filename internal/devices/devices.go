package devices

type DeviceInfo struct {
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer,omitempty"`
	Status       string `json:"status,omitempty"`
	Extra        string `json:"extra,omitempty"`
}

type DevicesSummary struct {
	Keyboards       []DeviceInfo `json:"keyboards"`
	Mice            []DeviceInfo `json:"mice"`
	Cameras         []DeviceInfo `json:"cameras"`
	Monitors        []DeviceInfo `json:"monitors"`
	AudioDevices    []DeviceInfo `json:"audio_devices"`
	USBDevices      []DeviceInfo `json:"usb_devices"`
	StorageDrives   []DeviceInfo `json:"storage_drives"`
	NetworkAdapters []DeviceInfo `json:"network_adapters"`
	Printers        []DeviceInfo `json:"printers"`
	MemoryModules   []DeviceInfo `json:"memory_modules"`
	Motherboard     DeviceInfo   `json:"motherboard"`
	BIOS            DeviceInfo   `json:"bios"`
	Processor       DeviceInfo   `json:"processor"`
	GraphicsCards   []DeviceInfo `json:"graphics_cards"`
	HIDDevices      []DeviceInfo `json:"hid_devices"`
}
