package game

import (
	"testing"
)

func TestCalculateMovementRisk(t *testing.T) {
	tests := []struct {
		name         string
		fatigue      int
		energy       int
		hydration    int
		tempC        int
		clockHours   float64
		ailmentCount int
		wantMinScore int
		wantMaxScore int
		wantTier     RiskTier
	}{
		{
			name:         "Optimal Conditions",
			fatigue:      0,
			energy:       100,
			hydration:    100,
			tempC:        20,
			clockHours:   12.0,
			ailmentCount: 0,
			wantMinScore: 0,
			wantMaxScore: 0,
			wantTier:     RiskMinimal,
		},
		{
			name:         "High Fatigue, Low Energy",
			fatigue:      90, // score + 20
			energy:       10, // score + 15
			hydration:    100,
			tempC:        20,
			clockHours:   12.0,
			ailmentCount: 0,
			wantMinScore: 35,
			wantMaxScore: 35,
			wantTier:     RiskModerate,
		},
		{
			name:         "Freezing Temperature in Darkness",
			fatigue:      0,
			energy:       100,
			hydration:    100,
			tempC:        -5,  // score + 10
			clockHours:   2.0, // score + 30
			ailmentCount: 0,
			wantMinScore: 40,
			wantMaxScore: 40,
			wantTier:     RiskHigh,
		},
		{
			name:         "Multiple Ailments and Exhausted",
			fatigue:      100, // score + 25
			energy:       0,   // score + 25
			hydration:    0,   // score + 25
			tempC:        20,
			clockHours:   12.0,
			ailmentCount: 2,   // score + 30
			wantMinScore: 100, // Capped at 100
			wantMaxScore: 100,
			wantTier:     RiskCritical,
		},
		{
			name:         "Dusk and Chilly",
			fatigue:      0,
			energy:       100,
			hydration:    100,
			tempC:        -1,   // score + 2
			clockHours:   18.0, // score + 15
			ailmentCount: 0,
			wantMinScore: 17,
			wantMaxScore: 17,
			wantTier:     RiskMinimal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := &PlayerState{
				Fatigue:   tt.fatigue,
				Energy:    tt.energy,
				Hydration: tt.hydration,
				Ailments:  make([]Ailment, tt.ailmentCount),
			}
			weather := WeatherState{
				TemperatureC: tt.tempC,
			}

			score, tier := CalculateMovementRisk(player, weather, tt.clockHours)

			if score < tt.wantMinScore || score > tt.wantMaxScore {
				t.Errorf("CalculateMovementRisk() score = %d, want between %d and %d", score, tt.wantMinScore, tt.wantMaxScore)
			}
			if tier != tt.wantTier {
				t.Errorf("CalculateMovementRisk() tier = %v, want %v", tier, tt.wantTier)
			}
		})
	}
}
