package game

type RiskTier int

const (
	RiskMinimal RiskTier = iota
	RiskModerate
	RiskHigh
	RiskSevere
	RiskCritical
)

func (t RiskTier) String() string {
	switch t {
	case RiskMinimal:
		return "Minimal"
	case RiskModerate:
		return "Moderate"
	case RiskHigh:
		return "High"
	case RiskSevere:
		return "Severe"
	case RiskCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// CalculateMovementRisk determines movement danger based on player state and environment.
func CalculateMovementRisk(player *PlayerState, weather WeatherState, clockHours float64) (int, RiskTier) {
	if player == nil {
		return 0, RiskMinimal
	}

	score := 0

	// Fatigue / Energy
	if player.Fatigue > 50 {
		score += (player.Fatigue - 50) / 2
	}
	if player.Energy < 25 {
		score += (25 - player.Energy)
	}
	if player.Hydration < 25 {
		score += (25 - player.Hydration)
	}

	// Temperature (Cold Exposure)
	if weather.TemperatureC <= 0 {
		score += (0 - weather.TemperatureC) * 2 // Colder is worse
	} else if weather.TemperatureC >= 35 {
		score += (weather.TemperatureC - 35) // Heat also risky
	}

	// TODO: Wetness - Not directly tracked on player yet, hook in when available.

	// Daylight
	// Assuming night is roughly 19:00 to 05:00.
	if clockHours < 5 || clockHours >= 19 {
		score += 30
	} else if clockHours >= 17 && clockHours < 19 {
		// Dusk
		score += 15
	}

	// Ailments (Injury)
	if len(player.Ailments) > 0 {
		score += len(player.Ailments) * 15
	}

	// TODO: Terrain Difficulty / Weather severity hook (e.g., storms).

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	var tier RiskTier
	switch {
	case score < 20:
		tier = RiskMinimal
	case score < 40:
		tier = RiskModerate
	case score < 60:
		tier = RiskHigh
	case score < 80:
		tier = RiskSevere
	default:
		tier = RiskCritical
	}

	return score, tier
}
