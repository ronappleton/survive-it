package game

import "testing"

func TestDeterministicEncounterRollStable(t *testing.T) {
	a := deterministicEncounterRoll(44, 7, 9, 3, TimeBlockDay, "hunt", 2, "species")
	b := deterministicEncounterRoll(44, 7, 9, 3, TimeBlockDay, "hunt", 2, "species")
	if a != b {
		t.Fatalf("expected same deterministic roll, got %f vs %f", a, b)
	}
	if a < 0 || a > 1 {
		t.Fatalf("expected roll in [0,1], got %f", a)
	}
}

func TestFogRevealPermanenceAlone(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 30},
		Seed:        2001,
	})
	if err != nil {
		t.Fatalf("new run: %v", err)
	}
	run.EnsureTopology()
	x, y := run.CurrentMapPosition()
	if !run.IsRevealed(x, y) {
		t.Fatalf("start cell should be revealed")
	}
	targetX, targetY := x+2, y
	run.RevealFog(targetX, targetY, 1)
	if !run.IsRevealed(targetX, targetY) {
		t.Fatalf("target cell expected revealed")
	}
	run.RevealFog(x-2, y, 1)
	if !run.IsRevealed(targetX, targetY) {
		t.Fatalf("revealed cell should remain revealed")
	}
}

func TestGenerateWorldTopologyDeterministicForSameSeed(t *testing.T) {
	a := GenerateWorldTopology(3001, "temperate_rainforest", 36, 36)
	b := GenerateWorldTopology(3001, "temperate_rainforest", 36, 36)
	if a.Width != b.Width || a.Height != b.Height {
		t.Fatalf("expected same dimensions, got %dx%d vs %dx%d", a.Width, a.Height, b.Width, b.Height)
	}
	if len(a.Cells) != len(b.Cells) {
		t.Fatalf("expected same cell count")
	}
	for i := range a.Cells {
		if a.Cells[i] != b.Cells[i] {
			t.Fatalf("topology mismatch at cell %d: %+v vs %+v", i, a.Cells[i], b.Cells[i])
		}
	}
}

func TestHuntPressureReducesPreyEncounterRate(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 30},
		Seed:        4001,
	})
	if err != nil {
		t.Fatalf("new run: %v", err)
	}
	run.EnsureTopology()
	x, y := run.CurrentMapPosition()
	idx, ok := run.topoIndex(x, y)
	if !ok {
		t.Fatalf("expected valid start index")
	}
	if len(run.CellStates) <= idx {
		t.Fatalf("missing cell state for index")
	}

	countPrey := func(huntPressure uint8) int {
		run.CellStates[idx] = CellState{HuntPressure: huntPressure}
		prey := 0
		for i := 0; i < 350; i++ {
			event, ok := run.RollWildlifeEncounter(1, x, y, "hunt", i)
			if ok && event.Prey {
				prey++
			}
		}
		return prey
	}

	low := countPrey(0)
	high := countPrey(220)
	if high >= low {
		t.Fatalf("expected high hunt pressure to reduce prey encounters, low=%d high=%d", low, high)
	}
}
