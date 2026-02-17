package game

func BuiltInScenarios() []Scenario {
	aloneSeasons := SeasonSet{
		ID: SeasonSetAloneDefaultID,
		Phases: []SeasonPhase{
			{Season: SeasonAutumn, Days: 14},
			{Season: SeasonWinter, Days: 0},
		},
	}
	wetSeasons := SeasonSet{
		ID: SeasonSetWetDefaultID,
		Phases: []SeasonPhase{
			{Season: SeasonWet, Days: 0},
		},
	}
	winterSeasons := SeasonSet{
		ID: SeasonSetWinterDefaultID,
		Phases: []SeasonPhase{
			{Season: SeasonWinter, Days: 0},
		},
	}
	drySeasons := SeasonSet{
		ID: SeasonSetDryDefaultID,
		Phases: []SeasonPhase{
			{Season: SeasonDry, Days: 0},
		},
	}

	build := func(id ScenarioID, name, biome string, days int, kit IssuedKit, set SeasonSet) Scenario {
		return Scenario{
			ID:                 id,
			Name:               name,
			Biome:              biome,
			DefaultDays:        days,
			IssuedKit:          kit,
			SeasonSets:         []SeasonSet{set},
			DefaultSeasonSetID: set.ID,
		}
	}

	return []Scenario{
		// Core starter scenarios.
		build(ScenarioVancouverIslandID, "Vancouver Island", "temperate_rainforest", 60, IssuedKit{}, aloneSeasons),
		build(ScenarioJungleID, "Jungle", "tropical_jungle", 60, IssuedKit{Firestarter: true}, wetSeasons),
		build(ScenarioArcticID, "Arctic", "subarctic", 60, IssuedKit{}, winterSeasons),

		// Alone-inspired locations.
		build("patagonia_argentina", "Patagonia (Argentina)", "cold_steppe", 60, IssuedKit{}, aloneSeasons),
		build("mongolia_khentii", "Khentii Mountains (Mongolia)", "montane_steppe", 60, IssuedKit{}, drySeasons),
		build("great_slave_lake", "Great Slave Lake (Canada)", "subarctic_lake", 100, IssuedKit{}, winterSeasons),
		build("chilko_lake_bc", "Chilko Lake (British Columbia)", "mountain_forest", 60, IssuedKit{}, aloneSeasons),
		build("labrador_coast", "Labrador Coast (Canada)", "boreal_coast", 60, IssuedKit{}, winterSeasons),
		build("reindeer_lake", "Reindeer Lake (Saskatchewan)", "boreal_lake", 60, IssuedKit{}, winterSeasons),
		build("mackenzie_delta", "Mackenzie River Delta (NWT)", "arctic_delta", 60, IssuedKit{}, winterSeasons),

		// Naked and Afraid (21-day style) inspired locations.
		build("naa_panama", "Panama Survival (NAA)", "tropical_jungle", 21, IssuedKit{}, wetSeasons),
		build("naa_costa_rica", "Costa Rica Jungle (NAA)", "tropical_jungle", 21, IssuedKit{}, wetSeasons),
		build("naa_tanzania", "Tanzania Savanna (NAA)", "savanna", 21, IssuedKit{}, drySeasons),
		build("naa_namibia", "Namib Desert (NAA)", "desert", 21, IssuedKit{}, drySeasons),
		build("naa_nicaragua", "Nicaragua Jungle (NAA)", "tropical_jungle", 21, IssuedKit{}, wetSeasons),
		build("naa_colombia", "Colombia Jungle (NAA)", "tropical_jungle", 21, IssuedKit{}, wetSeasons),
		build("naa_mexico_yucatan", "Yucatan (NAA)", "tropical_dry_forest", 21, IssuedKit{}, drySeasons),
		build("naa_florida_everglades", "Florida Everglades (NAA)", "wetlands", 21, IssuedKit{}, wetSeasons),
		build("naa_louisiana_swamp", "Louisiana Swamp (NAA)", "swamp", 21, IssuedKit{}, wetSeasons),
		build("naa_philippines", "Philippines Island (NAA)", "tropical_island", 21, IssuedKit{}, wetSeasons),
		build("naa_alaska", "Alaska Cold Region (NAA)", "tundra", 21, IssuedKit{}, winterSeasons),

		// Naked and Afraid XL inspired locations.
		build("naaxl_colombia_40", "NAA XL Colombia (40)", "badlands_jungle_edge", 40, IssuedKit{}, drySeasons),
		build("naaxl_south_africa_40", "NAA XL South Africa (40)", "savanna", 40, IssuedKit{}, drySeasons),
		build("naaxl_ecuador_40", "NAA XL Ecuador (40)", "tropical_jungle", 40, IssuedKit{}, wetSeasons),
		build("naaxl_nicaragua_40", "NAA XL Nicaragua (40)", "tropical_jungle", 40, IssuedKit{}, wetSeasons),
		build("naaxl_philippines_40", "NAA XL Philippines (40)", "tropical_island", 40, IssuedKit{}, wetSeasons),
		build("naaxl_louisiana_60", "NAA XL Louisiana (60)", "swamp", 60, IssuedKit{}, wetSeasons),
		build("naaxl_montana_frozen_14", "NAA XL Frozen Montana (14)", "cold_mountain", 14, IssuedKit{}, winterSeasons),
	}
}
