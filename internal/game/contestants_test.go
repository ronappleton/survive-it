package game

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestContestantSimulationSetup(t *testing.T) {
	run, err := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 365},
		Seed:        1234,
	})
	if err != nil {
		t.Fatalf("failed to create run state: %v", err)
	}

	if len(run.Contestants) != 9 {
		t.Fatalf("expected 9 AI contestants, got %d", len(run.Contestants))
	}
	for _, c := range run.Contestants {
		if c.Status != ContestantActive {
			t.Fatalf("expected all contestants to start active, got %s", c.Status)
		}
		if c.Energy != 100 || c.Hydration != 100 || c.Health != 100 {
			t.Fatalf("expected all stats to start at 100")
		}
	}
}

func TestContestantSimulationDecay(t *testing.T) {
	c := ContestantState{
		ID:            1,
		Status:        ContestantActive,
		Energy:        80, // Start below 100 to allow observing both recovery and decay
		Hydration:     80,
		Morale:        80,
		Health:        100,
		RiskTolerance: 5,
	}

	// Create a predictable RNG that always returns 0 for Float64
	// This ensures r.Float64() > threshold always fails for recovery
	r := rand.New(&mockSource{val: 0})
	weather := WeatherState{TemperatureC: 15, Type: WeatherClear}

	// 10 hours should drop stats
	c.processMacroTick(10*time.Hour, weather, r)

	if c.Energy >= 80 || c.Hydration >= 80 || c.Morale >= 80 {
		t.Fatalf("expected stats to have a net decay over time, got energy:%d hydr:%d morale:%d", c.Energy, c.Hydration, c.Morale)
	}
}

type mockSource struct {
	val int64
}

func (m *mockSource) Int63() int64    { return m.val }
func (m *mockSource) Seed(seed int64) {}

func TestContestantsTapOutOnZeroStats(t *testing.T) {
	c := ContestantState{
		ID:            2,
		Status:        ContestantActive,
		Energy:        0,
		Hydration:     0,
		Morale:        0,
		Health:        100,
		RiskTolerance: 5,
	}

	r := rand.New(rand.NewSource(99))
	weather := WeatherState{TemperatureC: 15, Type: WeatherClear}

	// Process ticks until tap out occurs (since it's a 10% chance per tick when below thresholds)
	tapped := false
	for i := 0; i < 50; i++ {
		if hasEvent, _ := c.processMacroTick(1*time.Hour, weather, r); hasEvent {
			tapped = true
			break
		}
	}

	if !tapped {
		t.Fatalf("expected contestant to tap out due to 0 stats after 50 hourly rolls")
	}
	if c.Status != ContestantTappedOut && c.Status != ContestantMedicallyExtracted {
		t.Fatalf("expected status change, got %s", c.Status)
	}
}

func TestContestantHealthDropsIfStarving(t *testing.T) {
	c := ContestantState{
		Status:    ContestantActive,
		Energy:    0,
		Hydration: 0,
		Health:    100,
	}

	r := rand.New(rand.NewSource(1))
	weather := WeatherState{TemperatureC: 15, Type: WeatherClear}

	c.processMacroTick(10*time.Hour, weather, r)
	if c.Health == 100 {
		t.Fatalf("expected health to drop due to starvation/dehydration")
	}
}

func TestProcessContestantSimulationReturnsMessages(t *testing.T) {
	run, _ := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 365},
		Seed:        123,
	})

	// Force everyone to 0 stats so someone taps out quickly
	for i := range run.Contestants {
		run.Contestants[i].Energy = 0
		run.Contestants[i].Hydration = 0
		run.Contestants[i].Morale = 0
	}

	messages := run.ProcessContestantSimulation(100 * time.Hour)
	if len(messages) == 0 {
		t.Fatalf("expected at least one event message from mass starvation simulation")
	}

	foundTapOut := false
	for _, msg := range messages {
		if strings.Contains(msg, "tapped out") || strings.Contains(msg, "medically extracted") {
			foundTapOut = true
		}
	}
	if !foundTapOut {
		t.Fatalf("expected tap-out or extraction message, got: %v", messages)
	}
}

func TestModeAloneWinCondition(t *testing.T) {
	run, _ := NewRunState(RunConfig{
		Mode:        ModeAlone,
		ScenarioID:  ScenarioVancouverIslandID,
		PlayerCount: 1,
		RunLength:   RunLength{Days: 365},
		Seed:        123,
	})

	// Initially ongoing
	outcome := run.EvaluateRun()
	if outcome.Status != RunOutcomeOngoing {
		t.Fatalf("expected ongoing run, got %s", outcome.Status)
	}

	// Force all contestants out
	for i := range run.Contestants {
		run.Contestants[i].Status = ContestantTappedOut
	}

	outcome = run.EvaluateRun()
	if outcome.Status != RunOutcomeCompleted {
		t.Fatalf("expected victory condition, got %s", outcome.Status)
	}
	if !strings.Contains(outcome.Message, "All other contestants have tapped out") {
		t.Fatalf("expected victory message, got: %s", outcome.Message)
	}
}
