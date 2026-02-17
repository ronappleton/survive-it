package game

func BuiltInScenarios() []Scenario {
	return []Scenario{
		{
			ID:          ScenarioVancouverIslandID,
			Name:        "Vancouver Island",
			Biome:       "",
			DefaultDays: 60,
			IssuedKit:   IssuedKit{},
			SeasonSets: []SeasonSet{
				{
					ID: SeasonSetAloneDefaultID,
					Phases: []SeasonPhase{
						{
							Season: SeasonAutumn, Days: 14,
						},
						{
							Season: SeasonWinter, Days: 0,
						},
					},
				},
			},
			DefaultSeasonSetID: SeasonSetAloneDefaultID,
		},
		{
			ID:          ScenarioJungleID,
			Name:        "Jungle",
			Biome:       "",
			DefaultDays: 60,
			IssuedKit: IssuedKit{
				Firestarter: true,
				Pot:         false,
			},
			SeasonSets: []SeasonSet{
				{
					ID: SeasonSetWetDefaultID,
					Phases: []SeasonPhase{
						{
							Season: SeasonWet, Days: 0,
						},
					},
				},
			},
			DefaultSeasonSetID: SeasonSetWetDefaultID,
		},
		{
			ID:          ScenarioArcticID,
			Name:        "Arctic",
			Biome:       "",
			DefaultDays: 60,
			IssuedKit:   IssuedKit{},
			SeasonSets: []SeasonSet{
				{
					ID: SeasonSetWinterDefaultID,
					Phases: []SeasonPhase{
						{
							Season: SeasonWinter, Days: 0,
						},
					},
				},
			},
			DefaultSeasonSetID: SeasonSetWinterDefaultID,
		},
	}
}
