package game

import "strings"

// Discovery summary:
// - Built-ins already centralize per-scenario defaults (biome, map size, season sets, location metadata).
// - This is the safest place to attach optional climate profiles without changing scenario loading flow.
// - Climate metadata is currently applied to real-world cold scenarios first (Alaska Paired Exposure).
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

	build := func(id ScenarioID, mode GameMode, name, biome, desc, daunting, motivation string, days int, kit IssuedKit, set SeasonSet) Scenario {
		mapW, mapH := defaultTopologySizeForMode(mode)
		switch mode {
		case ModeNakedAndAfraid:
			mapW, mapH = 100, 100
		case ModeNakedAndAfraidXL:
			mapW, mapH = 125, 125
		case ModeAlone:
			mapW, mapH = 36, 36
		}
		loc := builtInScenarioLocationMeta(id)
		if loc != nil && strings.TrimSpace(loc.Name) == "" {
			loc.Name = name
		}
		return Scenario{
			ID:                 id,
			Name:               name,
			Location:           inferScenarioLocation(name),
			LocationMeta:       loc,
			Climate:            builtInScenarioClimateProfile(id),
			Biome:              biome,
			MapWidthCells:      mapW,
			MapHeightCells:     mapH,
			Description:        desc,
			Daunting:           daunting,
			Motivation:         motivation,
			SupportedModes:     []GameMode{mode},
			DefaultDays:        days,
			IssuedKit:          kit,
			SeasonSets:         []SeasonSet{set},
			DefaultSeasonSetID: set.ID,
		}
	}

	return []Scenario{
		// Isolation Protocol mode scenarios.
		build(ScenarioVancouverIslandID, ModeAlone, "Vancouver Island", "temperate_rainforest",
			"Cold Pacific rain, dense forest, and long isolation windows.", "Persistent rain and limited dry tinder punish poor shelter choices.", "Endure the storm cycles and become the last one standing.", 365, IssuedKit{}, aloneSeasons),
		build(ScenarioArcticID, ModeAlone, "Arctic", "subarctic",
			"Open cold terrain with thin margins for heat and calories.", "Energy drains fast once wind and exposure stack up.", "Master fire discipline and cold-weather routine under pressure.", 365, IssuedKit{}, winterSeasons),
		build("patagonia_argentina", ModeAlone, "Patagonia (Argentina)", "cold_steppe",
			"Dry, windy steppe with abrupt weather shifts and sparse cover.", "Gale-force winds can wreck shelter plans overnight.", "Adapt to volatility and keep momentum despite harsh swings.", 365, IssuedKit{}, aloneSeasons),
		build("mongolia_khentii", ModeAlone, "Khentii Mountains (Mongolia)", "montane_steppe",
			"Highland terrain with cold nights and wide movement distances.", "Fuel and water logistics can dominate each day.", "Win through route planning and consistent resource systems.", 365, IssuedKit{}, drySeasons),
		build("great_slave_lake_100", ModeAlone, "Great Slave Lake (Canada) - 100 Days", "subarctic_lake",
			"Large northern lake environment with prolonged cold exposure.", "Extended duration means every early mistake compounds.", "Play the long game and survive beyond normal limits.", 100, IssuedKit{}, winterSeasons),
		build("great_slave_lake_365", ModeAlone, "Great Slave Lake (Canada) - 365 Days", "subarctic_lake",
			"Large northern lake environment with prolonged cold exposure.", "Extended duration means every early mistake compounds.", "Play the long game and survive beyond normal limits.", 365, IssuedKit{}, winterSeasons),
		build("chilko_lake_bc", ModeAlone, "Chilko Lake (British Columbia)", "mountain_forest",
			"Mountain lake basin with rugged terrain and limited easy calories.", "Terrain and weather can isolate key supply zones.", "Stay methodical and turn terrain into strategic advantage.", 365, IssuedKit{}, aloneSeasons),
		build("labrador_coast", ModeAlone, "Labrador Coast (Canada)", "boreal_coast",
			"Remote boreal coast with exposed shoreline and severe weather.", "Cold moisture and wind chill steadily erode resilience.", "Prove your discipline where comfort never lasts long.", 365, IssuedKit{}, winterSeasons),
		build("reindeer_lake", ModeAlone, "Reindeer Lake (Saskatchewan)", "boreal_lake",
			"Northern boreal lake with mixed forest and seasonal turnover.", "Food acquisition can become highly inconsistent late-game.", "Build stable systems early and outlast the collapse curve.", 365, IssuedKit{}, winterSeasons),
		build("mackenzie_delta", ModeAlone, "Mackenzie River Delta (NWT)", "arctic_delta",
			"Delta channels, cold flats, and broad exposure to elements.", "Waterlogged ground and wind make shelter siting unforgiving.", "Think like an expedition leader and survive the logistics war.", 365, IssuedKit{}, winterSeasons),

		// Paired Exposure mode scenarios.
		build(ScenarioJungleID, ModeNakedAndAfraid, "Jungle", "tropical_jungle",
			"Hot, wet jungle challenge with dense vegetation and high humidity.", "Hydration and foot health can collapse quickly in constant wet heat.", "Stay sharp for 21 days and finish with your partner.", 21, IssuedKit{Firestarter: true}, wetSeasons),
		build("naa_panama", ModeNakedAndAfraid, "Panama Survival (Paired Exposure)", "tropical_jungle",
			"Lowland jungle route with heat, rain, and dense undergrowth.", "High humidity and insects degrade recovery every night.", "Keep your pace and decision quality through relentless discomfort.", 21, IssuedKit{}, wetSeasons),
		build("naa_costa_rica", ModeNakedAndAfraid, "Costa Rica Jungle (Paired Exposure)", "tropical_jungle",
			"Lush tropical terrain with frequent rain and rapid fatigue cycles.", "Wet conditions can stall fire progress for days.", "Push through setbacks and finish a clean extraction run.", 21, IssuedKit{}, wetSeasons),
		build("naa_tanzania", ModeNakedAndAfraid, "Tanzania Savanna (Paired Exposure)", "savanna",
			"Open savanna challenge with sun exposure and scarce cover.", "Water planning errors can become run-ending quickly.", "Execute disciplined water and shade strategy under stress.", 21, IssuedKit{}, drySeasons),
		build("naa_namibia", ModeNakedAndAfraid, "Namib Desert (Paired Exposure)", "desert",
			"Arid desert environment with high daytime heat load.", "Heat and dehydration punish inefficient movement patterns.", "Stay focused, conserve energy, and earn your extraction.", 21, IssuedKit{}, drySeasons),
		build("naa_nicaragua", ModeNakedAndAfraid, "Nicaragua Jungle (Paired Exposure)", "tropical_jungle",
			"Humid jungle basin with aggressive weather and dense canopy.", "Limited visibility and wet nights complicate every task.", "Outlast environmental pressure and stay mission-ready.", 21, IssuedKit{}, wetSeasons),
		build("naa_colombia", ModeNakedAndAfraid, "Colombia Jungle (Paired Exposure)", "tropical_jungle",
			"Steamy jungle conditions with constant moisture load.", "Sleep quality and morale can crater without shelter discipline.", "Keep the team aligned and finish strong at day 21.", 21, IssuedKit{}, wetSeasons),
		build("naa_mexico_yucatan", ModeNakedAndAfraid, "Yucatan (Paired Exposure)", "tropical_dry_forest",
			"Tropical dry forest with heat spikes and scattered resources.", "Water and shade windows can be brutally narrow.", "Demonstrate precision and complete the full challenge.", 21, IssuedKit{}, drySeasons),
		build("naa_florida_everglades", ModeNakedAndAfraid, "Florida Everglades (Paired Exposure)", "wetlands",
			"Flooded wetlands with thick humidity and unstable footing.", "Persistent wet exposure can destroy recovery capacity.", "Maintain composure and solve each day one decision at a time.", 21, IssuedKit{}, wetSeasons),
		build("naa_louisiana_swamp", ModeNakedAndAfraid, "Louisiana Swamp (Paired Exposure)", "swamp",
			"Hot swamp terrain with insects, mud, and difficult routes.", "Progress can stall when routes and shelter both fight you.", "Stay adaptable and extract with confidence.", 21, IssuedKit{}, wetSeasons),
		build("naa_philippines", ModeNakedAndAfraid, "Philippines Island (Paired Exposure)", "tropical_island",
			"Island jungle challenge mixing coast and interior movement.", "Resource pockets are uneven and easy to overcommit.", "Balance risk and consistency to finish the 21-day arc.", 21, IssuedKit{}, wetSeasons),
		build("naa_alaska", ModeNakedAndAfraid, "Alaska Cold Region (Paired Exposure)", "tundra",
			"Short-format cold challenge with severe exposure penalties.", "Cold mismanagement rapidly drains energy reserves.", "Prove you can function under hard-cold constraints.", 21, IssuedKit{}, winterSeasons),

		// Expedition Survival mode scenarios.
		build("naaxl_colombia_40", ModeNakedAndAfraidXL, "Expedition Survival Colombia (40)", "badlands_jungle_edge",
			"Extended-team challenge across mixed jungle and dry ground.", "Long duration amplifies every conflict and resource error.", "Lead with consistency and survive the full XL stretch.", 40, IssuedKit{}, drySeasons),
		build("naaxl_south_africa_40", ModeNakedAndAfraidXL, "Expedition Survival South Africa (40)", "savanna",
			"Group survival in hot savanna with long operational windows.", "Heat load and team fatigue create cascading failures.", "Build resilient systems and carry the team to extraction.", 40, IssuedKit{}, drySeasons),
		build("naaxl_ecuador_40", ModeNakedAndAfraidXL, "Expedition Survival Ecuador (40)", "tropical_jungle",
			"Extended jungle run where wet conditions never let up.", "Sustained humidity can grind morale into the floor.", "Stay mission-first and thrive under prolonged adversity.", 40, IssuedKit{}, wetSeasons),
		build("naaxl_nicaragua_40", ModeNakedAndAfraidXL, "Expedition Survival Nicaragua (40)", "tropical_jungle",
			"Long jungle endurance route with high attrition risk.", "Bad weather windows can undo days of progress.", "Keep rebuilding momentum and outlast the setbacks.", 40, IssuedKit{}, wetSeasons),
		build("naaxl_philippines_40", ModeNakedAndAfraidXL, "Expedition Survival Philippines (40)", "tropical_island",
			"Island XL challenge blending shoreline and inland survival.", "Split terrain increases coordination and logistics burden.", "Execute as a team and finish what others abandon.", 40, IssuedKit{}, wetSeasons),
		build("naaxl_louisiana_60", ModeNakedAndAfraidXL, "Expedition Survival Louisiana (60)", "swamp",
			"Ultra-long swamp campaign with severe environmental drag.", "Sixty days exposes every weakness in routine and planning.", "Outwork the conditions and write a legendary extraction.", 60, IssuedKit{}, wetSeasons),
		build("naaxl_montana_frozen_14", ModeNakedAndAfraidXL, "Expedition Survival Frozen Montana (14)", "cold_mountain",
			"Frozen-team variant emphasizing cold-weather survival execution.", "Cold stress compounds quickly when calories run thin.", "Operate cleanly under pressure and conquer the freeze.", 14, IssuedKit{}, winterSeasons),
	}
}

var builtInLocationMetaByScenarioID = map[ScenarioID]ScenarioLocation{
	ScenarioVancouverIslandID: {
		Name:      "Vancouver Island",
		BBox:      [4]float64{-128.5, 48.2, -123.0, 50.9},
		ProfileID: "vancouver_island",
	},
	"patagonia_argentina": {
		Name:      "Patagonia (Argentina)",
		BBox:      [4]float64{-72.5, -51.5, -67.0, -46.0},
		ProfileID: "patagonia_argentina",
	},
	"mongolia_khentii": {
		Name:      "Khentii Mountains (Mongolia)",
		BBox:      [4]float64{107.0, 47.0, 111.8, 50.2},
		ProfileID: "khentii_mountains",
	},
	"great_slave_lake_100": {
		Name:      "Great Slave Lake (Canada)",
		BBox:      [4]float64{-117.5, 60.5, -108.0, 63.2},
		ProfileID: "great_slave_lake",
	},
	"great_slave_lake_365": {
		Name:      "Great Slave Lake (Canada)",
		BBox:      [4]float64{-117.5, 60.5, -108.0, 63.2},
		ProfileID: "great_slave_lake",
	},
	"chilko_lake_bc": {
		Name:      "Chilko Lake (British Columbia)",
		BBox:      [4]float64{-125.5, 50.8, -123.4, 52.5},
		ProfileID: "chilko_lake",
	},
	"labrador_coast": {
		Name:      "Labrador Coast (Canada)",
		BBox:      [4]float64{-61.8, 52.5, -55.0, 57.0},
		ProfileID: "labrador_coast",
	},
	"reindeer_lake": {
		Name:      "Reindeer Lake (Saskatchewan)",
		BBox:      [4]float64{-104.7, 55.0, -101.0, 58.8},
		ProfileID: "reindeer_lake",
	},
	"mackenzie_delta": {
		Name:      "Mackenzie River Delta (NWT)",
		BBox:      [4]float64{-137.8, 67.1, -132.0, 70.0},
		ProfileID: "mackenzie_delta",
	},
	"naa_panama": {
		Name:      "Panama Survival (Paired Exposure)",
		BBox:      [4]float64{-79.0, 7.2, -77.2, 9.2},
		ProfileID: "panama_darien",
	},
	"naa_costa_rica": {
		Name:      "Costa Rica Jungle (Paired Exposure)",
		BBox:      [4]float64{-85.8, 9.0, -82.6, 11.2},
		ProfileID: "costa_rica_jungle",
	},
	"naa_tanzania": {
		Name:      "Tanzania Savanna (Paired Exposure)",
		BBox:      [4]float64{33.0, -3.5, 35.6, -1.2},
		ProfileID: "tanzania_savanna",
	},
	"naa_namibia": {
		Name:      "Namib Desert (Paired Exposure)",
		BBox:      [4]float64{14.0, -25.5, 16.8, -22.0},
		ProfileID: "namib_desert",
	},
	"naa_nicaragua": {
		Name:      "Nicaragua Jungle (Paired Exposure)",
		BBox:      [4]float64{-86.8, 10.8, -83.5, 13.2},
		ProfileID: "nicaragua_jungle",
	},
	"naa_colombia": {
		Name:      "Colombia Jungle (Paired Exposure)",
		BBox:      [4]float64{-75.5, 0.5, -72.0, 3.5},
		ProfileID: "colombia_jungle",
	},
	"naa_mexico_yucatan": {
		Name:      "Yucatan (Paired Exposure)",
		BBox:      [4]float64{-90.8, 18.4, -87.2, 21.8},
		ProfileID: "yucatan_dry_forest",
	},
	"naa_florida_everglades": {
		Name:      "Florida Everglades (Paired Exposure)",
		BBox:      [4]float64{-81.7, 24.8, -80.0, 26.5},
		ProfileID: "florida_everglades",
	},
	"naa_louisiana_swamp": {
		Name:      "Louisiana Swamp (Paired Exposure)",
		BBox:      [4]float64{-92.2, 29.0, -89.0, 31.2},
		ProfileID: "louisiana_swamp",
	},
	"naa_philippines": {
		Name:      "Philippines Island (Paired Exposure)",
		BBox:      [4]float64{120.3, 13.0, 123.5, 15.8},
		ProfileID: "philippines_island",
	},
	"naa_alaska": {
		Name:      "Alaska Cold Region (Paired Exposure)",
		BBox:      [4]float64{-153.0, 63.0, -147.0, 67.0},
		ProfileID: "alaska_cold_region",
	},
	"naaxl_colombia_40": {
		Name:      "Expedition Survival Colombia (40)",
		BBox:      [4]float64{-75.8, 1.0, -72.5, 4.0},
		ProfileID: "colombia_jungle",
	},
	"naaxl_south_africa_40": {
		Name:      "Expedition Survival South Africa (40)",
		BBox:      [4]float64{23.0, -26.7, 27.0, -23.0},
		ProfileID: "south_africa_savanna",
	},
	"naaxl_ecuador_40": {
		Name:      "Expedition Survival Ecuador (40)",
		BBox:      [4]float64{-78.9, -2.4, -76.0, 0.7},
		ProfileID: "ecuador_jungle",
	},
	"naaxl_nicaragua_40": {
		Name:      "Expedition Survival Nicaragua (40)",
		BBox:      [4]float64{-86.8, 10.8, -83.5, 13.2},
		ProfileID: "nicaragua_jungle",
	},
	"naaxl_philippines_40": {
		Name:      "Expedition Survival Philippines (40)",
		BBox:      [4]float64{120.3, 13.0, 123.5, 15.8},
		ProfileID: "philippines_island",
	},
	"naaxl_louisiana_60": {
		Name:      "Expedition Survival Louisiana (60)",
		BBox:      [4]float64{-92.2, 29.0, -89.0, 31.2},
		ProfileID: "louisiana_swamp",
	},
	"naaxl_montana_frozen_14": {
		Name:      "Expedition Survival Frozen Montana (14)",
		BBox:      [4]float64{-113.0, 45.6, -109.7, 48.0},
		ProfileID: "montana_frozen",
	},
}

func builtInScenarioLocationMeta(id ScenarioID) *ScenarioLocation {
	loc, ok := builtInLocationMetaByScenarioID[id]
	if !ok {
		return nil
	}
	copyLoc := loc
	return &copyLoc
}

var builtInClimateByScenarioID = map[ScenarioID]ClimateProfile{
	"naa_alaska": {
		Name:               "Alaska Subarctic Cold",
		AllowedBiomes:      []uint8{TopoBiomeTundra, TopoBiomeBoreal, TopoBiomeMountain, TopoBiomeWetland, TopoBiomeSwamp, TopoBiomeGrassland},
		BaseTempC:          -16,
		TempVarianceC:      9,
		FrozenWaterBelowC:  0,
		DisallowTags:       []string{"desert", "tropical"},
		DefaultInsectMinC:  6,
		DefaultInsectQuiet: "It's quiet. No insect activity in this cold.",
		SeasonRules: map[SeasonID]SeasonRule{
			SeasonWinter: {
				TempBiasC:           -6,
				AllowedFloraTags:    []string{"arctic", "subarctic", "tundra", "boreal", "winter_ok"},
				AllowedFaunaTags:    []string{"arctic", "subarctic", "tundra", "boreal", "cold", "winter_ok", "lake", "river", "coast", "mountain", "forest"},
				DisallowedFloraTags: []string{"desert", "tropical", "savanna", "warm_season"},
				DisallowedFaunaTags: []string{"desert", "tropical", "savanna", "temperate_warm", "warm_season"},
				InsectActivity:      false,
				InsectMinTempC:      6,
				InsectQuietMessage:  "It's quiet. No insect activity in this cold.",
				DisallowWeatherTypes: []WeatherType{
					WeatherHeatwave,
					WeatherRain,
					WeatherHeavyRain,
				},
			},
		},
	},
}

func builtInScenarioClimateProfile(id ScenarioID) *ClimateProfile {
	profile, ok := builtInClimateByScenarioID[id]
	if !ok {
		return nil
	}
	copyProfile := profile
	if len(profile.AllowedBiomes) > 0 {
		copyProfile.AllowedBiomes = append([]uint8(nil), profile.AllowedBiomes...)
	}
	if len(profile.DisallowTags) > 0 {
		copyProfile.DisallowTags = append([]string(nil), profile.DisallowTags...)
	}
	if profile.SeasonRules != nil {
		copyProfile.SeasonRules = make(map[SeasonID]SeasonRule, len(profile.SeasonRules))
		for season, rule := range profile.SeasonRules {
			ruleCopy := rule
			ruleCopy.AllowedFloraTags = append([]string(nil), rule.AllowedFloraTags...)
			ruleCopy.AllowedFaunaTags = append([]string(nil), rule.AllowedFaunaTags...)
			ruleCopy.DisallowedFloraTags = append([]string(nil), rule.DisallowedFloraTags...)
			ruleCopy.DisallowedFaunaTags = append([]string(nil), rule.DisallowedFaunaTags...)
			ruleCopy.DisallowWeatherTypes = append([]WeatherType(nil), rule.DisallowWeatherTypes...)
			copyProfile.SeasonRules[season] = ruleCopy
		}
	}
	return &copyProfile
}

func inferScenarioLocation(name string) string {
	n := strings.ToLower(name)
	switch {
	case strings.Contains(n, "canada"), strings.Contains(n, "vancouver"), strings.Contains(n, "labrador"), strings.Contains(n, "saskatchewan"), strings.Contains(n, "louisiana"), strings.Contains(n, "alaska"):
		return "North America"
	case strings.Contains(n, "patagonia"), strings.Contains(n, "colombia"), strings.Contains(n, "ecuador"), strings.Contains(n, "panama"), strings.Contains(n, "nicaragua"), strings.Contains(n, "costa rica"), strings.Contains(n, "yucatan"):
		return "South America"
	case strings.Contains(n, "africa"), strings.Contains(n, "tanzania"), strings.Contains(n, "namib"):
		return "Africa"
	case strings.Contains(n, "mongolia"), strings.Contains(n, "philippines"):
		return "Asia-Pacific"
	default:
		return "Wilderness"
	}
}
