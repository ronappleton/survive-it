package game

import "testing"

func TestWeatherForDayRespectsBiomeSeasonProfiles(t *testing.T) {
	seed := int64(12345)

	for day := 1; day <= 180; day++ {
		arcticWinter := WeatherForDay(seed, "subarctic", SeasonWinter, day)
		if arcticWinter == WeatherHeatwave || arcticWinter == WeatherRain || arcticWinter == WeatherHeavyRain {
			t.Fatalf("unexpected arctic winter weather on day %d: %s", day, arcticWinter)
		}

		desertDry := WeatherForDay(seed, "desert", SeasonDry, day)
		if desertDry == WeatherSnow || desertDry == WeatherBlizzard {
			t.Fatalf("unexpected desert dry weather on day %d: %s", day, desertDry)
		}

		jungleWet := WeatherForDay(seed, "tropical_jungle", SeasonWet, day)
		if jungleWet == WeatherSnow || jungleWet == WeatherBlizzard || jungleWet == WeatherHeatwave {
			t.Fatalf("unexpected tropical wet weather on day %d: %s", day, jungleWet)
		}
	}
}

func TestTemperatureForDayWithWeatherAdjustsByType(t *testing.T) {
	seed := int64(77)
	day := 8

	sunnyArctic := TemperatureForDayWithWeatherCelsius(seed, "subarctic", SeasonWinter, day, WeatherSunny)
	blizzardArctic := TemperatureForDayWithWeatherCelsius(seed, "subarctic", SeasonWinter, day, WeatherBlizzard)
	if blizzardArctic >= sunnyArctic {
		t.Fatalf("expected blizzard to be colder than sunny in arctic winter, got %d >= %d", blizzardArctic, sunnyArctic)
	}

	rainDesert := TemperatureForDayWithWeatherCelsius(seed, "desert", SeasonDry, day, WeatherRain)
	heatwaveDesert := TemperatureForDayWithWeatherCelsius(seed, "desert", SeasonDry, day, WeatherHeatwave)
	if rainDesert >= heatwaveDesert {
		t.Fatalf("expected rain to be cooler than heatwave in desert dry season, got %d >= %d", rainDesert, heatwaveDesert)
	}
}

func TestWeatherImpactCompoundsForJungleRain(t *testing.T) {
	dayTwo := weatherImpactForDay("tropical_jungle", SeasonWet, WeatherRain, 2, 28)
	dayFour := weatherImpactForDay("tropical_jungle", SeasonWet, WeatherRain, 4, 28)

	if dayFour.Morale >= dayTwo.Morale {
		t.Fatalf("expected 4-day jungle rain streak morale impact to be worse, got day4=%d day2=%d", dayFour.Morale, dayTwo.Morale)
	}
	if dayFour.Energy >= dayTwo.Energy {
		t.Fatalf("expected 4-day jungle rain streak energy impact to be worse, got day4=%d day2=%d", dayFour.Energy, dayTwo.Energy)
	}
}

func TestSpecialBiomeWeatherBonuses(t *testing.T) {
	arcticSunny := weatherImpactForDay("subarctic", SeasonWinter, WeatherSunny, 1, -7)
	if arcticSunny.Morale <= 0 {
		t.Fatalf("expected arctic winter sunny weather to provide morale boost, got %d", arcticSunny.Morale)
	}

	desertRain := weatherImpactForDay("desert", SeasonDry, WeatherRain, 1, 33)
	desertHeat := weatherImpactForDay("desert", SeasonDry, WeatherSunny, 1, 33)
	if desertRain.Hydration <= desertHeat.Hydration {
		t.Fatalf("expected desert dry rain to improve hydration impact vs dry heat, rain=%d sunny=%d", desertRain.Hydration, desertHeat.Hydration)
	}
}

func TestPlayerStatsModifyWeatherImpact(t *testing.T) {
	base := weatherImpactForDay("subarctic", SeasonWinter, WeatherBlizzard, 2, -18)

	strong := adjustWeatherImpactForPlayer(base, PlayerState{
		Endurance: 3,
		Bushcraft: 3,
		Mental:    3,
	}, WeatherBlizzard)
	weak := adjustWeatherImpactForPlayer(base, PlayerState{
		Endurance: -2,
		Bushcraft: -2,
		Mental:    -2,
	}, WeatherBlizzard)

	if strong.Energy <= weak.Energy {
		t.Fatalf("expected strong endurance to mitigate energy penalty, got strong=%d weak=%d", strong.Energy, weak.Energy)
	}
	if strong.Hydration <= weak.Hydration {
		t.Fatalf("expected strong bushcraft to mitigate hydration penalty, got strong=%d weak=%d", strong.Hydration, weak.Hydration)
	}
	if strong.Morale <= weak.Morale {
		t.Fatalf("expected strong mental to mitigate morale penalty, got strong=%d weak=%d", strong.Morale, weak.Morale)
	}
}

func TestRunStateInitializesAndAdvancesWeather(t *testing.T) {
	cfg := RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 7},
		Seed:        42,
	}

	run, err := NewRunState(cfg)
	if err != nil {
		t.Fatalf("new run state: %v", err)
	}

	if run.Weather.Day != 1 || run.Weather.Type == "" {
		t.Fatalf("expected initialized day-1 weather, got %+v", run.Weather)
	}

	run.AdvanceDay()
	if run.Day != 2 {
		t.Fatalf("expected day 2 after advance, got %d", run.Day)
	}
	if run.Weather.Day != 2 || run.Weather.Type == "" {
		t.Fatalf("expected initialized day-2 weather, got %+v", run.Weather)
	}
}
