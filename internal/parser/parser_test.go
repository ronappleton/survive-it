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
