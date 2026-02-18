package game

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

	weather := WeatherForDay(s.Config.Seed, s.Scenario.Biome, season, day)

	return WeatherState{
		Day:          day,
		Type:         weather,
		TemperatureC: TemperatureForDayWithWeatherCelsius(s.Config.Seed, s.Scenario.Biome, season, day, weather),
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
		if WeatherForDay(s.Config.Seed, s.Scenario.Biome, season, check) != weather {
			break
		}
		streak++
	}
	return streak
}
