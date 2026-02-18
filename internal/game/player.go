package game

import (
	"fmt"
	"math/rand/v2"
)

type Sex string

const (
	SexMale      = "male"
	SexFemale    = "female"
	SexNonBinary = "non-binary"
	SexOther     = "other"
)

type BodyType string

const (
	BodyTypeNeutral BodyType = "neutral"
	BodyTypeMale    BodyType = "male"
	BodyTypeFemale  BodyType = "female"
)

var maleNames = []string{
	"Jack", "Tom", "Ben", "Luke", "Sam",
	"James", "Owen", "Ethan", "Noah", "Leo",
	"Harry", "Daniel", "Liam", "Alex", "Ryan",
	"Jacob", "Nathan", "Adam", "Callum", "Mason",
	"Henry", "Finn", "Oscar", "Isaac", "Aaron",
	"Connor", "Zach", "Evan", "Matthew", "Reece",
}
var femaleNames = []string{
	"Emma", "Olivia", "Isla", "Ava", "Sophia",
	"Mia", "Amelia", "Grace", "Freya", "Lily",
	"Chloe", "Ella", "Ruby", "Harper", "Evie",
	"Scarlett", "Aria", "Zoe", "Hannah", "Layla",
	"Willow", "Maya", "Sienna", "Ivy", "Luna",
	"Erin", "Brooke", "Jasmine", "Phoebe", "Nina",
}
var neutralNames = []string{
	"Rowan", "Avery", "Kai", "Riley", "Quinn",
	"Jordan", "Morgan", "Taylor", "Reese", "Casey",
	"Blake", "Jamie", "Cameron", "Dakota", "Skyler",
	"Phoenix", "Sage", "River", "Emery", "Finley",
	"Hayden", "Charlie", "Alexis", "Micah", "Indigo",
	"Robin", "Shay", "Jules", "Marley", "Kendall",
}
var anyNames = append(append([]string{}, maleNames...), append(femaleNames, neutralNames...)...)

var romanNumerals = []string{
	"", "II", "III", "IV", "V", "VI", "VII", "VIII",
}

type PlayerState struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Sex       Sex       `json:"sex"`
	BodyType  BodyType  `json:"body_type"`
	WeightKg  int       `json:"weight_kg"`
	HeightFt  int       `json:"height_ft"`
	HeightIn  int       `json:"height_in"`
	Endurance int       `json:"endurance"`
	Bushcraft int       `json:"bushcraft"`
	Mental    int       `json:"mental"`
	KitLimit  int       `json:"kit_limit"`
	Kit       []KitItem `json:"kit"`
	Energy    int       `json:"energy"`
	Hydration int       `json:"hydration"`
	Morale    int       `json:"morale"`

	// Runtime-only survival reserves and bars. These are not editable in setup.
	CaloriesReserveKcal int `json:"calories_reserve_kcal"`
	ProteinReserveG     int `json:"protein_reserve_g"`
	FatReserveG         int `json:"fat_reserve_g"`
	Hunger              int `json:"hunger"`  // 0 = satiated, 100 = starving
	Thirst              int `json:"thirst"`  // 0 = hydrated, 100 = severely thirsty
	Fatigue             int `json:"fatigue"` // 0 = rested, 100 = exhausted

	Nutrition NutritionTotals `json:"nutrition"`
	Ailments  []Ailment       `json:"ailments"`
}

type PlayerConfig struct {
	Name      string
	Sex       Sex
	BodyType  BodyType
	WeightKg  int
	HeightFt  int
	HeightIn  int
	Endurance int
	Bushcraft int
	Mental    int
	KitLimit  int
	Kit       []KitItem
}

func CreatePlayers(cfg RunConfig) []PlayerState {
	players := make([]PlayerState, cfg.PlayerCount)

	rng := seededRNG(cfg.Seed)

	used := make(map[string]int)

	for i := range players {
		var pc PlayerConfig

		if i < len(cfg.Players) {
			pc = cfg.Players[i]
		}

		if pc.Sex == "" {
			pc.Sex = SexOther
		}

		if pc.BodyType == "" {
			switch pc.Sex {
			case SexMale:
				pc.BodyType = BodyTypeMale
			case SexFemale:
				pc.BodyType = BodyTypeFemale
			default:
				pc.BodyType = BodyTypeNeutral
			}
		}

		if pc.WeightKg <= 0 {
			pc.WeightKg = 75
		}
		if pc.HeightFt <= 0 {
			pc.HeightFt = 5
		}
		if pc.HeightIn < 0 || pc.HeightIn > 11 {
			pc.HeightIn = 10
		}
		if pc.KitLimit <= 0 {
			pc.KitLimit = 1
		}

		name := pc.Name
		if name == "" {
			name = generateName(rng, pc.Sex, used)
		} else {
			count := used[name]
			used[name]++
			if count > 0 {
				name = fmt.Sprintf("%s %s", name, romanSuffix(count))
			}
		}

		players[i] = PlayerState{
			ID:        i + 1,
			Name:      name,
			Sex:       pc.Sex,
			BodyType:  pc.BodyType,
			WeightKg:  pc.WeightKg,
			HeightFt:  pc.HeightFt,
			HeightIn:  pc.HeightIn,
			Endurance: clamp(pc.Endurance, -3, 3),
			Bushcraft: clamp(pc.Bushcraft, -3, 3),
			Mental:    clamp(pc.Mental, -3, 3),
			KitLimit:  pc.KitLimit,
			Kit:       append([]KitItem(nil), pc.Kit...),
			Energy:    100,
			Hydration: 100,
			Morale:    100,
		}
		initializeRuntimeBars(&players[i])
	}

	return players
}

func generateName(rng *rand.Rand, sex Sex, used map[string]int) string {
	var pool []string

	switch sex {
	case SexMale:
		pool = maleNames
	case SexFemale:
		pool = femaleNames
	case SexNonBinary:
		pool = neutralNames
	default:
		pool = anyNames
	}

	base := pool[rng.IntN(len(pool))]
	count := used[base]
	used[base]++

	if count > 0 {
		return fmt.Sprintf("%s %s", base, romanSuffix(count))
	}

	return base
}

func romanSuffix(n int) string {
	if n > 0 && n < len(romanNumerals) {
		return romanNumerals[n]
	}
	return fmt.Sprintf("%d", n+1)
}
