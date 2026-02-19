package game

import (
	"fmt"
	"math"
	"strings"
)

// Discovery summary:
// - Hunt/fish catch selection was keyed to Scenario.Biome, not the current topo cell biome.
// - Encounter rolls are already cell-based, so catch selection now uses current biome query for coherence.
// - This keeps carcass-first flow intact while aligning intake species with location context.

type HuntResult struct {
	PlayerID      int
	Domain        AnimalDomain
	AnimalID      string
	AnimalName    string
	WeightGrams   int
	CarcassID     string
	CarcassKg     float64
	StoredAt      string
	HoursSpent    float64
	EncounterLogs []string
}

func carcassIDForDomain(domain AnimalDomain) string {
	switch domain {
	case AnimalDomainWater:
		return "fish_carcass"
	case AnimalDomainAir:
		return "bird_carcass"
	default:
		return "small_game_carcass"
	}
}

func (s *RunState) catchWithSkillBonus(playerID int, domain AnimalDomain) (CatchResult, *PlayerState, error) {
	if s == nil {
		return CatchResult{}, nil, fmt.Errorf("run state is nil")
	}
	s.EnsurePlayerRuntimeStats()
	player, ok := s.playerByID(playerID)
	if !ok {
		return CatchResult{}, nil, fmt.Errorf("player %d not found", playerID)
	}
	biome := s.CurrentBiomeQuery()
	if strings.TrimSpace(biome) == "" {
		biome = s.Scenario.Biome
	}
	season, okSeason := s.CurrentSeason()
	if !okSeason {
		season = SeasonAutumn
	}
	pool := AnimalsForBiome(biome, domain)
	pool = filterAnimalsForClimate(pool, domain, s.ActiveClimateProfile(), season, s.Weather.TemperatureC)
	catch, err := randomCatchFromPool(s.Config.Seed, biome, domain, s.Day, playerID, pool)
	if err != nil {
		return CatchResult{}, nil, err
	}

	switch domain {
	case AnimalDomainLand:
		applySkillEffort(&player.Hunting, 18, true)
	case AnimalDomainWater:
		applySkillEffort(&player.Fishing, 18, true)
	default:
		applySkillEffort(&player.Hunting, 12, true)
	}

	bonusPct := 0
	switch domain {
	case AnimalDomainLand:
		bonusPct = player.Hunting/8 + player.Strength + player.Agility + positiveTraitModifier(player.Traits)/2
	case AnimalDomainWater:
		bonusPct = player.Fishing/8 + player.Agility + positiveTraitModifier(player.Traits)/2
	default:
		bonusPct = player.Hunting/10 + player.Agility + positiveTraitModifier(player.Traits)/2
	}
	if bonusPct != 0 {
		adjusted := catch.WeightGrams + (catch.WeightGrams*bonusPct)/100
		catch.WeightGrams = max(80, adjusted)
	}
	catch.EdibleGrams = max(1, int(math.Round(float64(catch.WeightGrams)*catch.Animal.EdibleYieldRatio)))
	return catch, player, nil
}

func filterAnimalsForClimate(pool []AnimalSpec, domain AnimalDomain, climate *ClimateProfile, season SeasonID, tempC int) []AnimalSpec {
	if len(pool) == 0 || climate == nil {
		return pool
	}
	channel := "mammal"
	switch domain {
	case AnimalDomainWater:
		channel = "fish"
	case AnimalDomainAir:
		channel = "bird"
	}
	filtered := make([]AnimalSpec, 0, len(pool))
	for _, animal := range pool {
		tags := encounterSpeciesClimateTags(animal.ID, channel)
		if !climateAllowsFauna(climate, season, tags) {
			continue
		}
		if tempC < climateInsectMinTempC(climate, season) && hasAnyTag(tags, []string{"warm_season"}) {
			continue
		}
		filtered = append(filtered, animal)
	}
	return filtered
}

func randomCatchFromPool(seed int64, biome string, domain AnimalDomain, day, actorID int, pool []AnimalSpec) (CatchResult, error) {
	if len(pool) == 0 {
		return CatchResult{}, fmt.Errorf("no %s animals available in biome %s", domain, biome)
	}
	if day < 1 {
		day = 1
	}
	if actorID < 1 {
		actorID = 1
	}

	rng := seededRNG(seedFromLabel(seed, fmt.Sprintf("catch:%s:%s:%d:%d", normalizeBiome(biome), domain, day, actorID)))
	animal := pool[rng.IntN(len(pool))]

	minKg := animal.WeightMinKg
	maxKg := animal.WeightMaxKg
	if maxKg < minKg {
		maxKg = minKg
	}
	kg := minKg
	if maxKg > minKg {
		kg += rng.Float64() * (maxKg - minKg)
	}
	grams := int(math.Round(kg * 1000))
	if grams < 1 {
		grams = 1
	}
	yield := animal.EdibleYieldRatio
	if yield <= 0 || yield > 0.95 {
		yield = 0.5
	}
	edible := int(math.Round(float64(grams) * yield))
	if edible < 1 {
		edible = 1
	}
	return CatchResult{
		Animal:      animal,
		WeightGrams: grams,
		EdibleGrams: edible,
	}, nil
}

func (s *RunState) HuntAndCollectCarcass(playerID int, domain AnimalDomain, action string) (HuntResult, error) {
	catch, player, err := s.catchWithSkillBonus(playerID, domain)
	if err != nil {
		return HuntResult{}, err
	}
	if strings.TrimSpace(action) == "" {
		action = "hunt"
	}
	x, y := s.CurrentMapPosition()
	s.applyCellStateAction(x, y, action)

	carcassID := carcassIDForDomain(domain)
	carcass, ok := carcassCatalog[carcassID]
	if !ok {
		return HuntResult{}, fmt.Errorf("no carcass profile for domain %s", domain)
	}
	kg := math.Round(max(0.1, float64(catch.WeightGrams)/1000.0)*10) / 10
	item := InventoryItem{
		ID:       carcassID,
		Name:     carcass.Name,
		Unit:     "kg",
		Qty:      kg,
		WeightKg: 1.2,
		Category: "carcass",
		AgeDays:  0,
	}

	storedAt := ""
	if err := s.AddPersonalInventoryItem(playerID, item); err == nil {
		storedAt = "personal"
	} else if err := s.addCampInventoryItem(item); err == nil {
		storedAt = "camp"
	} else {
		// If full-carcass storage fails, keep a partial field-dressed yield that fits available space.
		tryStorePartial := func() bool {
			maxCampQty := math.Floor((s.campFreeKg()/item.WeightKg)*10) / 10
			if maxCampQty >= 0.1 {
				partial := item
				partial.Qty = maxCampQty
				if err := s.addCampInventoryItem(partial); err == nil {
					item = partial
					storedAt = "camp (partial)"
					return true
				}
			}
			personalFreeKg := max(0.0, s.playerCarryLimitKg(player)-inventoryWeightKg(player.PersonalItems))
			maxPersonalQty := math.Floor((personalFreeKg/item.WeightKg)*10) / 10
			if maxPersonalQty >= 0.1 {
				partial := item
				partial.Qty = maxPersonalQty
				if err := s.AddPersonalInventoryItem(playerID, partial); err == nil {
					item = partial
					storedAt = "personal (partial)"
					return true
				}
			}
			return false
		}
		if !tryStorePartial() {
			return HuntResult{}, fmt.Errorf("caught %s (%.1fkg), but no storage space", catch.Animal.Name, kg)
		}
	}
	kg = item.Qty

	baseHours := 1.8
	switch domain {
	case AnimalDomainWater:
		baseHours = 1.6
	case AnimalDomainAir:
		baseHours = 1.4
	}
	skillFactor := float64(player.Hunting) / 40.0
	if domain == AnimalDomainWater {
		skillFactor = float64(player.Fishing) / 40.0
	}
	hours := clampFloat(baseHours+(kg*0.05)-(skillFactor*0.25), 0.5, 10)
	_ = s.AdvanceActionClock(hours)

	player.Energy = clamp(player.Energy-int(math.Ceil(hours*2.0)), 0, 100)
	player.Hydration = clamp(player.Hydration-int(math.Ceil(hours*1.4)), 0, 100)
	player.Morale = clamp(player.Morale+1, 0, 100)

	encounterLogs := make([]string, 0, 2)
	event, ok := s.RollWildlifeEncounter(playerID, x, y, action, 0)
	if ok {
		encounterLogs = append(encounterLogs, event.Message)
		player.Energy = clamp(player.Energy+event.EnergyDelta, 0, 100)
		player.Hydration = clamp(player.Hydration+event.HydrationDelta, 0, 100)
		player.Morale = clamp(player.Morale+event.MoraleDelta, 0, 100)
	}
	refreshEffectBars(player)

	return HuntResult{
		PlayerID:      playerID,
		Domain:        domain,
		AnimalID:      catch.Animal.ID,
		AnimalName:    catch.Animal.Name,
		WeightGrams:   catch.WeightGrams,
		CarcassID:     carcassID,
		CarcassKg:     kg,
		StoredAt:      storedAt,
		HoursSpent:    hours,
		EncounterLogs: encounterLogs,
	}, nil
}

func (s *RunState) CatchAndConsume(playerID int, domain AnimalDomain, choice MealChoice) (CatchResult, MealOutcome, error) {
	catch, player, err := s.catchWithSkillBonus(playerID, domain)
	if err != nil {
		return CatchResult{}, MealOutcome{}, err
	}
	outcome := ConsumeCatch(s.Config.Seed, s.Day, player, catch, choice)
	return catch, outcome, nil
}

func (s *RunState) ActiveAilmentNames(playerID int) []string {
	for i := range s.Players {
		if s.Players[i].ID != playerID {
			continue
		}
		out := make([]string, 0, len(s.Players[i].Ailments))
		for _, ailment := range s.Players[i].Ailments {
			name := ailment.Name
			if name == "" {
				name = string(ailment.Type)
			}
			out = append(out, name)
		}
		return out
	}
	return nil
}
