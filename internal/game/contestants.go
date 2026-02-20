package game

import (
	"math/rand"
	"time"
)

type ContestantStatus string

const (
	ContestantActive             ContestantStatus = "active"
	ContestantTappedOut          ContestantStatus = "tapped_out"
	ContestantMedicallyExtracted ContestantStatus = "medically_extracted"
	ContestantDeceased           ContestantStatus = "deceased"
)

type ContestantState struct {
	ID            int              `json:"id"`
	Name          string           `json:"name"`
	Status        ContestantStatus `json:"status"`
	DaysSurvived  int              `json:"days_survived"`
	TimeSurvived  time.Duration    `json:"time_survived"`
	Energy        int              `json:"energy"`
	Hydration     int              `json:"hydration"`
	Morale        int              `json:"morale"`
	Health        int              `json:"health"`
	RiskTolerance int              `json:"risk_tolerance"` // 1-10, affects how low stats go before tap out
}

func initialContestants(count int, seed int64) []ContestantState {
	if count <= 0 {
		return nil
	}
	r := rand.New(rand.NewSource(seed))
	// Random default names
	firstNames := []string{"John", "Sarah", "Michael", "Emma", "David", "Laura", "Chris", "Anna", "Tom", "Lisa", "Mark", "Jessica", "James", "Emily", "Robert", "Rachel"}
	out := make([]ContestantState, count)
	for i := 0; i < count; i++ {
		name := firstNames[r.Intn(len(firstNames))]
		out[i] = ContestantState{
			ID:            i + 1,
			Name:          name,
			Status:        ContestantActive,
			Energy:        100,
			Hydration:     100,
			Morale:        100,
			Health:        100,
			RiskTolerance: 3 + r.Intn(6), // 3 to 8
		}
	}
	return out
}

func (s *ContestantState) processMacroTick(delta time.Duration, weather WeatherState, r *rand.Rand) (bool, string) {
	if s.Status != ContestantActive {
		return false, ""
	}

	s.TimeSurvived += delta
	s.DaysSurvived = int(s.TimeSurvived.Hours() / 24)

	// Convert delta to hours for decay
	hours := delta.Hours()
	if hours <= 0 {
		return false, ""
	}

	// Base decay per hour
	energyDecay := 1.5 * hours
	hydrationDecay := 2.0 * hours
	moraleDecay := 1.0 * hours

	// Weather impact
	if weather.TemperatureC < 0 {
		energyDecay += 1.0 * hours
		moraleDecay += 0.5 * hours
	} else if weather.TemperatureC > 30 {
		hydrationDecay += 1.5 * hours
		energyDecay += 0.5 * hours
	}

	if weather.Type == WeatherHeavyRain || weather.Type == WeatherBlizzard {
		moraleDecay += 2.0 * hours
	}

	// Apply time-based decay
	s.Energy = max(0, s.Energy-int(energyDecay))
	s.Hydration = max(0, s.Hydration-int(hydrationDecay))
	s.Morale = max(0, s.Morale-int(moraleDecay))

	// Health decay if starving/dehydrated
	if s.Energy == 0 || s.Hydration == 0 {
		s.Health = max(0, s.Health-int(2*hours))
	}

	// Random recovery (simulating foraging/drinking/resting off-screen)
	// They don't always succeed. (Only apply if they haven't tapped out / died yet)
	if r.Float64() > 0.3 { // 70% chance to find water
		s.Hydration = min(100, s.Hydration+int(10*hours))
	}
	if r.Float64() > 0.6 { // 40% chance to find food
		s.Energy = min(100, s.Energy+int(15*hours))
	}
	if r.Float64() > 0.5 { // 50% chance to rest well
		s.Morale = min(100, s.Morale+int(5*hours))
	}

	// Event rolls
	return s.rollForEvents(r)
}

func (s *ContestantState) rollForEvents(r *rand.Rand) (bool, string) {
	// 1. Extreme Medical Emergency (e.g. bear attack, severe fall)
	// Base chance 0.005% per tick, increased if low energy
	riskFactor := 1.0
	if s.Energy < 20 {
		riskFactor = 3.0
	}
	if r.Float64() < 0.00005*riskFactor {
		s.Status = ContestantMedicallyExtracted
		s.Health = 0
		return true, "A contestant has been medically extracted due to a severe injury."
	}

	// 2. Tap Out due to Morale/Condition
	// If stats fall below risk tolerance, chance to tap
	tapThreshold := s.RiskTolerance * 5 // e.g. 15 to 40
	if s.Morale < tapThreshold || s.Energy < tapThreshold || s.Hydration < tapThreshold {
		if r.Float64() < 0.1 { // 10% chance to actually tap when below threshold
			s.Status = ContestantTappedOut
			return true, "A contestant has tapped out."
		}
	}

	// 3. Medical Extraction due to Health
	if s.Health < 20 {
		if r.Float64() < 0.2 {
			s.Status = ContestantMedicallyExtracted
			return true, "A contestant has been medically extracted."
		}
	}

	// 4. Death (very rare, usually extracted first, but possible if health is 0)
	if s.Health == 0 {
		if r.Float64() < 0.05 {
			s.Status = ContestantDeceased
			return true, "A contestant has perished."
		}
	}

	return false, ""
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
