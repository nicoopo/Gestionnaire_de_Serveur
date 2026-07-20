package metrics

import "testing"

func TestAlertTracker_Check(t *testing.T) {
	tracker := NewAlertTracker()

	tests := []struct {
		name       string
		sample     Sample
		wantAlerts int // nombre d'alertes attendues à cette étape
	}{
		{
			name:       "premier échantillon sous les seuils, pas d'alerte",
			sample:     Sample{CPUPercent: 50, RAMPercent: 40, DiskPercent: 30},
			wantAlerts: 0,
		},
		{
			name:       "CPU passe en warning, une alerte",
			sample:     Sample{CPUPercent: 85, RAMPercent: 40, DiskPercent: 30},
			wantAlerts: 1,
		},
		{
			name:       "CPU reste en warning, pas de nouvelle alerte (pas de re-spam)",
			sample:     Sample{CPUPercent: 82, RAMPercent: 40, DiskPercent: 30},
			wantAlerts: 0,
		},
		{
			name:       "CPU passe en critical, une nouvelle alerte",
			sample:     Sample{CPUPercent: 95, RAMPercent: 40, DiskPercent: 30},
			wantAlerts: 1,
		},
		{
			name:       "CPU revient à la normale, une alerte de retour",
			sample:     Sample{CPUPercent: 50, RAMPercent: 40, DiskPercent: 30},
			wantAlerts: 1,
		},
		{
			name:       "CPU et RAM dépassent en même temps, deux alertes",
			sample:     Sample{CPUPercent: 90, RAMPercent: 91, DiskPercent: 30},
			wantAlerts: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alerts := tracker.Check(tt.sample)
			if len(alerts) != tt.wantAlerts {
				t.Errorf("Check() a renvoyé %d alerte(s), attendu %d (alertes: %+v)", len(alerts), tt.wantAlerts, alerts)
			}
		})
	}
}

func TestLevelFor(t *testing.T) {
	tests := []struct {
		value float64
		want  string
	}{
		{value: 0, want: "ok"},
		{value: 79.9, want: "ok"},
		{value: 80, want: "warning"},
		{value: 89.9, want: "warning"},
		{value: 90, want: "critical"},
		{value: 100, want: "critical"},
	}

	for _, tt := range tests {
		got := levelFor(tt.value)
		if got != tt.want {
			t.Errorf("levelFor(%v) = %q, attendu %q", tt.value, got, tt.want)
		}
	}
}
