package metrics

import "sync"

type Sample struct {
	Timestamp   int64   `json:"timestamp"`
	CPUPercent  float64 `json:"cpu_percent"`
	RAMPercent  float64 `json:"ram_percent"`
	DiskPercent float64 `json:"disk_percent"`
	GPUPercent  float64 `json:"gpu_percent"`
}

type History struct {
	mu      sync.RWMutex
	samples []Sample
	maxSize int
}

func NewHistory(maxSize int) *History {
	return &History{
		samples: make([]Sample, 0, maxSize),
		maxSize: maxSize,
	}
}

func (h *History) Add(s Sample) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.samples = append(h.samples, s)
	if len(h.samples) > h.maxSize {
		h.samples = h.samples[1:] // retire le plus ancien
	}
}

func (h *History) All() []Sample {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]Sample, len(h.samples))
	copy(result, h.samples)
	return result
}
