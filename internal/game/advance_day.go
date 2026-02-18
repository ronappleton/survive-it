package game

func (s *RunState) AdvanceDay() {
	s.Day++
	s.EnsureWeather()
	season, ok := s.CurrentSeason()
	if !ok {
		season = SeasonAutumn
	}
	weatherImpact := weatherImpactForDay(s.Scenario.Biome, season, s.Weather.Type, s.Weather.StreakDays, s.Weather.TemperatureC)

	for i := range s.Players {
		p := &s.Players[i]
		physiologyProfile := PhysiologyFor(p.BodyType)

		p.Energy -= physiologyProfile.EnergyDrainPerDay
		p.Hydration -= physiologyProfile.HydrationDrainPerDay
		p.Morale -= physiologyProfile.MoraleDrainPerDay

		playerWeatherImpact := adjustWeatherImpactForPlayer(weatherImpact, *p, s.Weather.Type)
		p.Energy += playerWeatherImpact.Energy
		p.Hydration += playerWeatherImpact.Hydration
		p.Morale += playerWeatherImpact.Morale

		clampPlayer(p)
	}
}

func clamp(number, min, max int) int {
	if number < min {
		return min
	}

	if number > max {
		return max
	}

	return number
}

func clampPlayer(playerState *PlayerState) {
	playerState.Energy = clamp(playerState.Energy, 0, 100)
	playerState.Hydration = clamp(playerState.Hydration, 0, 100)
	playerState.Morale = clamp(playerState.Morale, 0, 100)
}

type RunOutcomeStatus string

const (
	RunOutcomeOngoing   RunOutcomeStatus = "ongoing"
	RunOutcomeCompleted RunOutcomeStatus = "completed"
	RunOutcomeCritical  RunOutcomeStatus = "critical"
)

type RunOutcome struct {
	Status            RunOutcomeStatus
	Message           string
	CriticalPlayerIDs []int
}

func (s *RunState) EvaluateRun() RunOutcome {
	// 1) Completion by day limit (fixed-length runs)
	if !s.Config.RunLength.OpenEnded && s.Config.RunLength.Days > 0 {
		if s.Day > s.Config.RunLength.Days {
			return RunOutcome{
				Status:  RunOutcomeCompleted,
				Message: "Run completed.",
			}
		}
	}

	// 2) Critical condition (v1: no death yet)
	// Collect all critical players
	criticalIDs := make([]int, 0)
	for _, p := range s.Players {
		if p.Energy == 0 || p.Hydration == 0 {
			criticalIDs = append(criticalIDs, p.ID)
		}
	}

	if len(criticalIDs) > 0 {
		return RunOutcome{
			Status:            RunOutcomeCritical,
			Message:           "One or more players are in critical condition.",
			CriticalPlayerIDs: criticalIDs,
		}
	}

	// 3) Ongoing
	return RunOutcome{Status: RunOutcomeOngoing}
}
