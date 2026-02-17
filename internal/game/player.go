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
	ID        int
	Name      string
	Sex       Sex
	BodyType  BodyType
	Energy    int
	Hydration int
	Morale    int
}

type PlayerConfig struct {
	Name     string
	Sex      Sex
	BodyType BodyType
}

func CreatePlayers(cfg RunConfig) []PlayerState {
	players := make([]PlayerState, cfg.PlayerCount)

	seed1 := uint64(cfg.Seed)
	seed2 := seed1 ^ uint64(0x9e3779b97f4a7c15)
	rng := rand.New(rand.NewPCG(seed1, seed2))

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
			Energy:    100,
			Hydration: 100,
			Morale:    100,
		}
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
