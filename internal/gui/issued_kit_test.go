package gui

import (
	"testing"

	"github.com/appengine-ltd/survive-it/internal/game"
)

func TestRecommendedIssuedKitForScenarioOnlyHarshBiomes(t *testing.T) {
	normal := game.Scenario{Biome: "temperate_rainforest"}
	if kit := recommendedIssuedKitForScenario(game.ModeAlone, normal); len(kit) != 0 {
		t.Fatalf("expected no issued kit for non-harsh biome, got %v", kit)
	}

	harsh := game.Scenario{Biome: "arctic_tundra"}
	kit := recommendedIssuedKitForScenario(game.ModeAlone, harsh)
	if len(kit) == 0 {
		t.Fatalf("expected issued kit for harsh biome")
	}
	if len(kit) > 2 {
		t.Fatalf("expected alone issued kit to stay compact, got %d", len(kit))
	}
}

func TestRuntimeIssuedKitExcludesPersonalItems(t *testing.T) {
	scenario := game.Scenario{Biome: "arctic_tundra"}
	players := []game.PlayerConfig{
		{Kit: []game.KitItem{game.KitFerroRod, game.KitThermalLayer}},
	}
	issued := runtimeIssuedKit(game.ModeAlone, scenario, players)
	for _, item := range issued {
		if item == game.KitFerroRod || item == game.KitThermalLayer {
			t.Fatalf("issued kit should not include personal item %q", item)
		}
	}
}

func TestRuntimeIssuedKitReturnsNoneWhenAllRecommendedAlreadyCarried(t *testing.T) {
	scenario := game.Scenario{Biome: "desert"}
	players := []game.PlayerConfig{
		{Kit: []game.KitItem{game.KitCanteen}},
		{Kit: []game.KitItem{game.KitWaterFilter}},
	}
	issued := runtimeIssuedKit(game.ModeNakedAndAfraid, scenario, players)
	if len(issued) != 0 {
		t.Fatalf("expected no issued kit when recommended items already carried, got %v", issued)
	}
}
