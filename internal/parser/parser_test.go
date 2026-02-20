package parser

import "testing"

func TestNormalisationTable(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "  INVENTRY  ", want: "inventry"},
		{in: "pick-up   STIC!!", want: "pick up stic"},
		{in: "go   N", want: "go n"},
	}
	for _, tc := range tests {
		got := normaliseInput(tc.in)
		if got != tc.want {
			t.Fatalf("normaliseInput(%q)=%q want=%q", tc.in, got, tc.want)
		}
	}
}

func TestAliasInvMapsToInventory(t *testing.T) {
	p := New()
	intent := p.Parse(ParseContext{}, "inv")
	if intent.Verb != "inventory" {
		t.Fatalf("expected inventory verb, got %q", intent.Verb)
	}
	if intent.Clarify != nil {
		t.Fatalf("did not expect clarify: %+v", intent.Clarify)
	}
}

func TestTypoInventryMapsToInventory(t *testing.T) {
	p := New()
	intent := p.Parse(ParseContext{}, "inventry")
	if intent.Verb != "inventory" {
		t.Fatalf("expected inventory verb, got %q", intent.Verb)
	}
	if intent.Confidence < 0.6 {
		t.Fatalf("expected decent confidence for typo correction, got %.2f", intent.Confidence)
	}
}

func TestInScopeBoostResolvesStick(t *testing.T) {
	p := New()
	ctx := ParseContext{
		Nearby: []string{"stick", "stone"},
	}
	intent := p.Parse(ctx, "pick up stic")
	if intent.Verb != "take" {
		t.Fatalf("expected take verb, got %q", intent.Verb)
	}
	if len(intent.Args) == 0 || intent.Args[0] != "stick" {
		t.Fatalf("expected first arg stick, got %+v", intent.Args)
	}
}

func TestAmbiguityReturnsClarify(t *testing.T) {
	p := New()
	ctx := ParseContext{
		Nearby: []string{"stick", "stone"},
	}
	intent := p.Parse(ctx, "take")
	if intent.Clarify == nil {
		t.Fatalf("expected clarify for ambiguous target-less take")
	}
	if len(intent.Clarify.Options) < 2 {
		t.Fatalf("expected at least 2 clarify options, got %d", len(intent.Clarify.Options))
	}
}

func TestFreeTextBagInference(t *testing.T) {
	p := New()
	intent := p.Parse(ParseContext{}, "i need to check my bag")
	if intent.Verb != "inventory" {
		t.Fatalf("expected inventory inference, got %q", intent.Verb)
	}
}

func TestPreserveSmokeAlias(t *testing.T) {
	p := New()
	intent := p.Parse(ParseContext{}, "smoke raw fish meat")
	if intent.Verb != "preserve" {
		t.Fatalf("expected preserve command, got %q", intent.Verb)
	}
}

func TestFishCommandParses(t *testing.T) {
	p := New()
	intent := p.Parse(ParseContext{}, "fish")
	if intent.Verb != "fish" {
		t.Fatalf("expected fish verb, got %q", intent.Verb)
	}
}

func TestPronounResolutionUseIt(t *testing.T) {
	p := New()
	ctx := ParseContext{
		Inventory:  []string{"ferro rod"},
		LastEntity: "ferro rod",
	}
	intent := p.Parse(ctx, "use it")
	if intent.Clarify != nil {
		t.Fatalf("unexpected clarify: %+v", intent.Clarify)
	}
	if intent.Verb != "use" {
		t.Fatalf("expected use verb, got %q", intent.Verb)
	}
	if len(intent.Args) == 0 || intent.Args[0] != "ferro rod" {
		t.Fatalf("expected pronoun to resolve to ferro rod, got %+v", intent.Args)
	}
}

func TestMovementScaleParsing(t *testing.T) {
	p := New()
	ctx := ParseContext{}

	tests := []struct {
		in       string
		wantDir  string
		wantM    float64
		wantMin  float64
		wantCond MovementCondition
	}{
		{"go north 500m", "north", 500, 0, ConditionNone},
		{"walk east for 2 hours", "east", 0, 120, ConditionNone},
		{"walk south 1.5km", "south", 1500, 0, ConditionNone},
		{"head west until dark", "west", 0, 0, ConditionDark},
		{"go north until exhausted", "north", 0, 0, ConditionTired},
	}

	for _, tc := range tests {
		intent := p.Parse(ctx, tc.in)
		if intent.Clarify != nil {
			t.Fatalf("unexpected clarify for %q: %s", tc.in, intent.Clarify.Prompt)
		}
		if intent.Verb != "go" {
			t.Fatalf("expected go verb for %q, got %q", tc.in, intent.Verb)
		}
		if len(intent.Args) == 0 || intent.Args[0] != tc.wantDir {
			t.Fatalf("expected direction %q for %q, got args %v", tc.wantDir, tc.in, intent.Args)
		}
		if intent.Movement == nil {
			t.Fatalf("expected movement scale for %q, got nil", tc.in)
		}
		if intent.Movement.DistanceMeters != tc.wantM {
			t.Errorf("expected %f meters for %q, got %f", tc.wantM, tc.in, intent.Movement.DistanceMeters)
		}
		if intent.Movement.DurationMinutes != tc.wantMin {
			t.Errorf("expected %f minutes for %q, got %f", tc.wantMin, tc.in, intent.Movement.DurationMinutes)
		}
		if intent.Movement.Condition != tc.wantCond {
			t.Errorf("expected condition %q for %q, got %q", tc.wantCond, tc.in, intent.Movement.Condition)
		}
	}
}

func TestMovementAmbiguity(t *testing.T) {
	p := New()

	tests := []struct {
		in     string
		prompt string
	}{
		{"go north", "How far or how long? (e.g. 500m, 1km, 10min, until dark)"},
		{"go 5", "5 what: meters, km, tiles, or minutes?"},
		{"walk 10 minutes", "Which direction?"},
	}

	for _, tc := range tests {
		intent := p.Parse(ParseContext{}, tc.in)
		if intent.Clarify == nil {
			t.Fatalf("expected clarify prompt for %q, got none", tc.in)
		}
		if intent.Clarify.Prompt != tc.prompt {
			t.Errorf("expected prompt %q for %q, got %q", tc.prompt, tc.in, intent.Clarify.Prompt)
		}
	}
}
