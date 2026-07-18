package metrics

import "strconv"

type Alert struct {
	Level   string  `json:"level"` // "ok", "warning" ou "critical"
	Metric  string  `json:"metric"`
	Value   float64 `json:"value"`
	Message string  `json:"message"`
}

const (
	warningThreshold  = 80.0
	criticalThreshold = 90.0
)

type AlertTracker struct {
	states map[string]string // dernier niveau connu par métrique
}

func NewAlertTracker() *AlertTracker {
	return &AlertTracker{states: make(map[string]string)}
}

// Check ne renvoie une alerte que si le niveau a changé depuis le dernier appel
func (t *AlertTracker) Check(s Sample) []Alert {
	var alerts []Alert

	alerts = append(alerts, t.checkMetric("CPU", s.CPUPercent)...)
	alerts = append(alerts, t.checkMetric("RAM", s.RAMPercent)...)
	alerts = append(alerts, t.checkMetric("Disque", s.DiskPercent)...)
	if s.GPUPercent > 0 {
		alerts = append(alerts, t.checkMetric("GPU", s.GPUPercent)...)
	}

	return alerts
}

func (t *AlertTracker) checkMetric(name string, value float64) []Alert {
	level := levelFor(value)
	previous := t.states[name]

	if level == previous {
		return nil // pas de changement, on ne renvoie rien
	}
	t.states[name] = level

	if level == "ok" {
		return []Alert{{
			Level:   "ok",
			Metric:  name,
			Value:   value,
			Message: name + " revenu à la normale (" + strconv.Itoa(int(value)) + "%)",
		}}
	}

	return []Alert{{
		Level:   level,
		Metric:  name,
		Value:   value,
		Message: name + " " + levelLabel(level) + " : " + strconv.Itoa(int(value)) + "%",
	}}
}

func levelFor(value float64) string {
	if value >= criticalThreshold {
		return "critical"
	}
	if value >= warningThreshold {
		return "warning"
	}
	return "ok"
}

func levelLabel(level string) string {
	if level == "critical" {
		return "critique"
	}
	return "élevé"
}
