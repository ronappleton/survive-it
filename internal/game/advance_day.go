package game

func (s *RunState) AdvanceDay() {
	s.EnsurePlayerRuntimeStats()
	s.consumePendingDayMetabolism()
	s.Day++
	s.EnsureWeather()
	season, ok := s.CurrentSeason()
	if !ok {
		season = SeasonAutumn
	}
	weatherImpact := weatherImpactForDay(s.Scenario.Biome, season, s.Weather.Type, s.Weather.StreakDays, s.Weather.TemperatureC)
	campImpact := s.campImpactForDay()

	for i := range s.Players {
		p := &s.Players[i]

		playerWeatherImpact := adjustWeatherImpactForPlayer(weatherImpact, *p, s.Weather.Type)
		playerWeatherImpact = s.applyCraftedWeatherModifiersForPlayer(playerWeatherImpact, *p, s.Weather.Type, s.Weather.TemperatureC)
		p.Energy += playerWeatherImpact.Energy
		p.Hydration += playerWeatherImpact.Hydration
		p.Morale += playerWeatherImpact.Morale
		p.Energy += campImpact.Energy
		p.Hydration += campImpact.Hydration
		p.Morale += campImpact.Morale
		applyDailyAilmentPenalties(p)
		applyDailyDeficiencyEffects(p)

		clampPlayer(p)
		refreshEffectBars(p)
	}
	s.progressCampState()
	s.advanceFoodDegradation()
	s.decayCellStates()
}

func applyDailyAilmentPenalties(playerState *PlayerState) {
	if playerState == nil || len(playerState.Ailments) == 0 {
		return
	}

	active := make([]Ailment, 0, len(playerState.Ailments))
	for _, ailment := range playerState.Ailments {
		if ailment.DaysRemaining <= 0 {
			continue
		}
		playerState.Energy -= ailment.EnergyPenalty
		playerState.Hydration -= ailment.HydrationPenalty
		playerState.Morale -= ailment.MoralePenalty
		ailment.DaysRemaining--
		if ailment.DaysRemaining > 0 {
			active = append(active, ailment)
		}
	}

	playerState.Ailments = active
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
	playerState.Hunger = clamp(playerState.Hunger, 0, 100)
	playerState.Thirst = clamp(playerState.Thirst, 0, 100)
	playerState.Fatigue = clamp(playerState.Fatigue, 0, 100)
	playerState.CaloriesReserveKcal = clamp(playerState.CaloriesReserveKcal, -5000, 12000)
	playerState.ProteinReserveG = clamp(playerState.ProteinReserveG, -250, 800)
	playerState.FatReserveG = clamp(playerState.FatReserveG, -250, 600)
	playerState.SugarReserveG = clamp(playerState.SugarReserveG, -300, 800)
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
		if p.Energy == 0 || p.Hydration == 0 || p.Hunger >= 100 || p.Thirst >= 100 || p.Fatigue >= 100 {
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

	// 3) ModeAlone Victory Condition
	if s.Config.Mode == ModeAlone && len(s.Contestants) > 0 {
		allOut := true
		for _, c := range s.Contestants {
			if c.Status == ContestantActive {
				allOut = false
				break
			}
		}
		if allOut {
			return RunOutcome{
				Status:  RunOutcomeCompleted,
				Message: "All other contestants have tapped out or perished. You have outlasted them all.",
			}
		}
	}

	// 4) Ongoing
	return RunOutcome{Status: RunOutcomeOngoing}
}
