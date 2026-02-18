package game

type statDelta struct {
	Energy    int
	Hydration int
	Morale    int
}

func (d statDelta) add(other statDelta) statDelta {
	return statDelta{
		Energy:    d.Energy + other.Energy,
		Hydration: d.Hydration + other.Hydration,
		Morale:    d.Morale + other.Morale,
	}
}

func weatherImpactForDay(biome string, season SeasonID, weather WeatherType, streakDays int, tempC int) statDelta {
	impact := baseWeatherImpact(biome, season, weather)
	impact = impact.add(temperatureStressImpact(tempC))
	impact = impact.add(streakWeatherImpact(biome, weather, streakDays))
	impact = impact.add(specialBiomeWeatherImpact(biome, season, weather))
	return impact
}

func baseWeatherImpact(biome string, season SeasonID, weather WeatherType) statDelta {
	switch weather {
	case WeatherSunny:
		impact := statDelta{Morale: 1}
		if biomeIsHot(biome) || season == SeasonDry {
			impact.Hydration = -1
		}
		return impact
	case WeatherClear:
		return statDelta{}
	case WeatherCloudy:
		return statDelta{Morale: -1}
	case WeatherRain:
		return statDelta{Energy: -1, Hydration: 1, Morale: -1}
	case WeatherHeavyRain:
		return statDelta{Energy: -2, Hydration: 1, Morale: -2}
	case WeatherStorm:
		return statDelta{Energy: -3, Hydration: -1, Morale: -3}
	case WeatherSnow:
		return statDelta{Energy: -2, Hydration: -1, Morale: -2}
	case WeatherBlizzard:
		return statDelta{Energy: -4, Hydration: -2, Morale: -4}
	case WeatherWindy:
		return statDelta{Energy: -1, Morale: -1}
	case WeatherHeatwave:
		return statDelta{Energy: -3, Hydration: -4, Morale: -2}
	default:
		return statDelta{}
	}
}

func temperatureStressImpact(tempC int) statDelta {
	switch {
	case tempC <= -20:
		return statDelta{Energy: -3, Hydration: -1, Morale: -3}
	case tempC <= -5:
		return statDelta{Energy: -2, Morale: -2}
	case tempC <= 2:
		return statDelta{Energy: -1, Morale: -1}
	case tempC >= 40:
		return statDelta{Energy: -3, Hydration: -4, Morale: -2}
	case tempC >= 32:
		return statDelta{Energy: -2, Hydration: -3, Morale: -1}
	case tempC >= 26:
		return statDelta{Energy: -1, Hydration: -2}
	default:
		return statDelta{}
	}
}

func streakWeatherImpact(biome string, weather WeatherType, streakDays int) statDelta {
	if streakDays < 2 {
		return statDelta{}
	}

	extra := clamp(streakDays-1, 1, 4)
	if biomeIsTropicalWet(biome) && isRainyWeather(weather) {
		return statDelta{
			Energy:    -extra,
			Hydration: -((extra + 1) / 2),
			Morale:    -(extra * 2),
		}
	}

	if isAdverseWeather(weather) {
		return statDelta{
			Energy:    -extra,
			Hydration: -(extra / 2),
			Morale:    -extra,
		}
	}

	if weather == WeatherSunny || weather == WeatherClear {
		return statDelta{Morale: clamp(extra, 0, 2)}
	}

	return statDelta{}
}

func specialBiomeWeatherImpact(biome string, season SeasonID, weather WeatherType) statDelta {
	if biomeIsArctic(biome) && season == SeasonWinter && weather == WeatherSunny {
		return statDelta{Energy: 1, Morale: 2}
	}

	if biomeIsDesertOrDry(biome) && isRainyWeather(weather) {
		impact := statDelta{Hydration: 1, Morale: 1}
		if season == SeasonDry {
			impact.Energy++
			impact.Hydration++
		}
		return impact
	}

	return statDelta{}
}

func adjustWeatherImpactForPlayer(impact statDelta, player PlayerState, weather WeatherType) statDelta {
	out := impact

	if out.Energy < 0 {
		out.Energy += clamp(player.Endurance, 0, 3)
		if player.Endurance < 0 {
			out.Energy += player.Endurance
		}
	}
	if out.Hydration < 0 {
		out.Hydration += clamp(player.Bushcraft, 0, 3)
		if player.Bushcraft < 0 {
			out.Hydration += player.Bushcraft
		}
	}
	if out.Morale < 0 {
		out.Morale += clamp(player.Mental, 0, 3)
		if player.Mental < 0 {
			out.Morale += player.Mental
		}
	}

	if out.Morale > 0 && player.Mental < 0 {
		out.Morale += player.Mental / 2
	}

	if isRainyWeather(weather) {
		if player.Bushcraft >= 2 {
			out.Morale++
		}
		if player.Bushcraft <= -2 {
			out.Morale--
		}
	}

	if isSevereWeather(weather) {
		if player.Mental >= 2 {
			out.Morale++
		}
		if player.Endurance <= -2 {
			out.Energy--
		}
	}

	return out
}

func isRainyWeather(weather WeatherType) bool {
	return weather == WeatherRain || weather == WeatherHeavyRain || weather == WeatherStorm
}

func isSevereWeather(weather WeatherType) bool {
	return weather == WeatherStorm || weather == WeatherBlizzard || weather == WeatherHeatwave
}

func isAdverseWeather(weather WeatherType) bool {
	switch weather {
	case WeatherRain, WeatherHeavyRain, WeatherStorm, WeatherSnow, WeatherBlizzard, WeatherHeatwave, WeatherWindy:
		return true
	default:
		return false
	}
}
