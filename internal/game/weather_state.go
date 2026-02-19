package game

// Discovery summary:
// - Daily weather is resolved here and cached into RunState.Weather.
// - This is the narrowest integration point to keep climate/season/temp coherent without loop refactors.
// - Constrained weather and temperature now flow through the same deterministic day resolver.

func (s *RunState) EnsureWeather() {
	if s == nil {
		return
	}
	if s.Day < 1 {
		s.Day = 1
	}
	if s.Weather.Day == s.Day && s.Weather.Type != "" {
		return
	}
	s.Weather = s.weatherStateForDay(s.Day)
}

func (s *RunState) weatherStateForDay(day int) WeatherState {
	if day < 1 {
		day = 1
	}

	season, ok := s.SeasonForDay(day)
	if !ok {
		season = SeasonAutumn
	}

	weather := s.weatherTypeForDay(day, season)

	return WeatherState{
		Day:          day,
		Type:         weather,
		TemperatureC: s.temperatureForDay(day, season, weather),
		StreakDays:   s.weatherStreakForDay(day, weather),
	}
}

func (s *RunState) weatherStreakForDay(day int, weather WeatherType) int {
	streak := 1
	for check := day - 1; check >= 1; check-- {
		season, ok := s.SeasonForDay(check)
		if !ok {
			season = SeasonAutumn
		}
		if s.weatherTypeForDay(check, season) != weather {
			break
		}
		streak++
	}
	return streak
}

func (s *RunState) weatherTypeForDay(day int, season SeasonID) WeatherType {
	weather := WeatherForDay(s.Config.Seed, s.Scenario.Biome, season, day)
	return constrainWeatherForClimate(s.Config.Seed, day, season, weather, s.ActiveClimateProfile())
}

func (s *RunState) temperatureForDay(day int, season SeasonID, weather WeatherType) int {
	climate := s.ActiveClimateProfile()
	if climate == nil {
		return TemperatureForDayWithWeatherCelsius(s.Config.Seed, s.Scenario.Biome, season, day, weather)
	}
	return TemperatureForDayWithClimate(s.Config.Seed, day, season, weather, climate)
}
