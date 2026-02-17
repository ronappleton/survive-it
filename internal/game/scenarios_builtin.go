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

	build := func(id ScenarioID, mode GameMode, name, biome, desc, daunting, motivation string, days int, kit IssuedKit, set SeasonSet) Scenario {
		return Scenario{
			ID:                 id,
			Name:               name,
			Biome:              biome,
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
		// Alone mode scenarios.
		build(ScenarioVancouverIslandID, ModeAlone, "Vancouver Island", "temperate_rainforest",
			"Cold Pacific rain, dense forest, and long isolation windows.", "Persistent rain and limited dry tinder punish poor shelter choices.", "Endure the storm cycles and become the last one standing.", 60, IssuedKit{}, aloneSeasons),
		build(ScenarioArcticID, ModeAlone, "Arctic", "subarctic",
			"Open cold terrain with thin margins for heat and calories.", "Energy drains fast once wind and exposure stack up.", "Master fire discipline and cold-weather routine under pressure.", 60, IssuedKit{}, winterSeasons),
		build("patagonia_argentina", ModeAlone, "Patagonia (Argentina)", "cold_steppe",
			"Dry, windy steppe with abrupt weather shifts and sparse cover.", "Gale-force winds can wreck shelter plans overnight.", "Adapt to volatility and keep momentum despite harsh swings.", 60, IssuedKit{}, aloneSeasons),
		build("mongolia_khentii", ModeAlone, "Khentii Mountains (Mongolia)", "montane_steppe",
			"Highland terrain with cold nights and wide movement distances.", "Fuel and water logistics can dominate each day.", "Win through route planning and consistent resource systems.", 60, IssuedKit{}, drySeasons),
		build("great_slave_lake", ModeAlone, "Great Slave Lake (Canada)", "subarctic_lake",
			"Large northern lake environment with prolonged cold exposure.", "Extended duration means every early mistake compounds.", "Play the long game and survive beyond normal limits.", 100, IssuedKit{}, winterSeasons),
		build("chilko_lake_bc", ModeAlone, "Chilko Lake (British Columbia)", "mountain_forest",
			"Mountain lake basin with rugged terrain and limited easy calories.", "Terrain and weather can isolate key supply zones.", "Stay methodical and turn terrain into strategic advantage.", 60, IssuedKit{}, aloneSeasons),
		build("labrador_coast", ModeAlone, "Labrador Coast (Canada)", "boreal_coast",
			"Remote boreal coast with exposed shoreline and severe weather.", "Cold moisture and wind chill steadily erode resilience.", "Prove your discipline where comfort never lasts long.", 60, IssuedKit{}, winterSeasons),
		build("reindeer_lake", ModeAlone, "Reindeer Lake (Saskatchewan)", "boreal_lake",
			"Northern boreal lake with mixed forest and seasonal turnover.", "Food acquisition can become highly inconsistent late-game.", "Build stable systems early and outlast the collapse curve.", 60, IssuedKit{}, winterSeasons),
		build("mackenzie_delta", ModeAlone, "Mackenzie River Delta (NWT)", "arctic_delta",
			"Delta channels, cold flats, and broad exposure to elements.", "Waterlogged ground and wind make shelter siting unforgiving.", "Think like an expedition leader and survive the logistics war.", 60, IssuedKit{}, winterSeasons),

		// Naked and Afraid mode scenarios.
		build(ScenarioJungleID, ModeNakedAndAfraid, "Jungle", "tropical_jungle",
			"Hot, wet jungle challenge with dense vegetation and high humidity.", "Hydration and foot health can collapse quickly in constant wet heat.", "Stay sharp for 21 days and finish with your partner.", 21, IssuedKit{Firestarter: true}, wetSeasons),
		build("naa_panama", ModeNakedAndAfraid, "Panama Survival (NAA)", "tropical_jungle",
			"Lowland jungle route with heat, rain, and dense undergrowth.", "High humidity and insects degrade recovery every night.", "Keep your pace and decision quality through relentless discomfort.", 21, IssuedKit{}, wetSeasons),
		build("naa_costa_rica", ModeNakedAndAfraid, "Costa Rica Jungle (NAA)", "tropical_jungle",
			"Lush tropical terrain with frequent rain and rapid fatigue cycles.", "Wet conditions can stall fire progress for days.", "Push through setbacks and finish a clean extraction run.", 21, IssuedKit{}, wetSeasons),
		build("naa_tanzania", ModeNakedAndAfraid, "Tanzania Savanna (NAA)", "savanna",
			"Open savanna challenge with sun exposure and scarce cover.", "Water planning errors can become run-ending quickly.", "Execute disciplined water and shade strategy under stress.", 21, IssuedKit{}, drySeasons),
		build("naa_namibia", ModeNakedAndAfraid, "Namib Desert (NAA)", "desert",
			"Arid desert environment with high daytime heat load.", "Heat and dehydration punish inefficient movement patterns.", "Stay focused, conserve energy, and earn your extraction.", 21, IssuedKit{}, drySeasons),
		build("naa_nicaragua", ModeNakedAndAfraid, "Nicaragua Jungle (NAA)", "tropical_jungle",
			"Humid jungle basin with aggressive weather and dense canopy.", "Limited visibility and wet nights complicate every task.", "Outlast environmental pressure and stay mission-ready.", 21, IssuedKit{}, wetSeasons),
		build("naa_colombia", ModeNakedAndAfraid, "Colombia Jungle (NAA)", "tropical_jungle",
			"Steamy jungle conditions with constant moisture load.", "Sleep quality and morale can crater without shelter discipline.", "Keep the team aligned and finish strong at day 21.", 21, IssuedKit{}, wetSeasons),
		build("naa_mexico_yucatan", ModeNakedAndAfraid, "Yucatan (NAA)", "tropical_dry_forest",
			"Tropical dry forest with heat spikes and scattered resources.", "Water and shade windows can be brutally narrow.", "Demonstrate precision and complete the full challenge.", 21, IssuedKit{}, drySeasons),
		build("naa_florida_everglades", ModeNakedAndAfraid, "Florida Everglades (NAA)", "wetlands",
			"Flooded wetlands with thick humidity and unstable footing.", "Persistent wet exposure can destroy recovery capacity.", "Maintain composure and solve each day one decision at a time.", 21, IssuedKit{}, wetSeasons),
		build("naa_louisiana_swamp", ModeNakedAndAfraid, "Louisiana Swamp (NAA)", "swamp",
			"Hot swamp terrain with insects, mud, and difficult routes.", "Progress can stall when routes and shelter both fight you.", "Stay adaptable and extract with confidence.", 21, IssuedKit{}, wetSeasons),
		build("naa_philippines", ModeNakedAndAfraid, "Philippines Island (NAA)", "tropical_island",
			"Island jungle challenge mixing coast and interior movement.", "Resource pockets are uneven and easy to overcommit.", "Balance risk and consistency to finish the 21-day arc.", 21, IssuedKit{}, wetSeasons),
		build("naa_alaska", ModeNakedAndAfraid, "Alaska Cold Region (NAA)", "tundra",
			"Short-format cold challenge with severe exposure penalties.", "Cold mismanagement rapidly drains energy reserves.", "Prove you can function under hard-cold constraints.", 21, IssuedKit{}, winterSeasons),

		// Naked and Afraid XL mode scenarios.
		build("naaxl_colombia_40", ModeNakedAndAfraidXL, "NAA XL Colombia (40)", "badlands_jungle_edge",
			"Extended-team challenge across mixed jungle and dry ground.", "Long duration amplifies every conflict and resource error.", "Lead with consistency and survive the full XL stretch.", 40, IssuedKit{}, drySeasons),
		build("naaxl_south_africa_40", ModeNakedAndAfraidXL, "NAA XL South Africa (40)", "savanna",
			"Group survival in hot savanna with long operational windows.", "Heat load and team fatigue create cascading failures.", "Build resilient systems and carry the team to extraction.", 40, IssuedKit{}, drySeasons),
		build("naaxl_ecuador_40", ModeNakedAndAfraidXL, "NAA XL Ecuador (40)", "tropical_jungle",
			"Extended jungle run where wet conditions never let up.", "Sustained humidity can grind morale into the floor.", "Stay mission-first and thrive under prolonged adversity.", 40, IssuedKit{}, wetSeasons),
		build("naaxl_nicaragua_40", ModeNakedAndAfraidXL, "NAA XL Nicaragua (40)", "tropical_jungle",
			"Long jungle endurance route with high attrition risk.", "Bad weather windows can undo days of progress.", "Keep rebuilding momentum and outlast the setbacks.", 40, IssuedKit{}, wetSeasons),
		build("naaxl_philippines_40", ModeNakedAndAfraidXL, "NAA XL Philippines (40)", "tropical_island",
			"Island XL challenge blending shoreline and inland survival.", "Split terrain increases coordination and logistics burden.", "Execute as a team and finish what others abandon.", 40, IssuedKit{}, wetSeasons),
		build("naaxl_louisiana_60", ModeNakedAndAfraidXL, "NAA XL Louisiana (60)", "swamp",
			"Ultra-long swamp campaign with severe environmental drag.", "Sixty days exposes every weakness in routine and planning.", "Outwork the conditions and write a legendary extraction.", 60, IssuedKit{}, wetSeasons),
		build("naaxl_montana_frozen_14", ModeNakedAndAfraidXL, "NAA XL Frozen Montana (14)", "cold_mountain",
			"Frozen-team variant emphasizing cold-weather survival execution.", "Cold stress compounds quickly when calories run thin.", "Operate cleanly under pressure and conquer the freeze.", 14, IssuedKit{}, winterSeasons),
	}
}
