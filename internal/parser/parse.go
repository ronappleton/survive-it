package parser

import (
	"fmt"
	"sort"
	"strings"

	"github.com/agnivade/levenshtein"
)

type Parser struct {
	registry *Registry
}

func New() *Parser {
	return &Parser{registry: DefaultRegistry()}
}

func (p *Parser) RegisterCommand(c CommandDef) {
	p.registry.RegisterCommand(c)
}

func (p *Parser) Parse(ctx ParseContext, raw string) Intent {
	intent := Intent{
		Raw:        raw,
		Normalised: normaliseInput(raw),
		Kind:       Unknown,
		Confidence: 0,
	}
	if intent.Normalised == "" {
		intent.Clarify = &ClarifyQuestion{Prompt: "Enter a command or intent.", Options: nil}
		return intent
	}

	tokens := tokenise(intent.Normalised)
	cmdMatch, alternates := p.registry.matchCommand(tokens)
	if cmdMatch.Canonical == "" || cmdMatch.Score < 0.5 {
		inferred := inferFreeTextIntent(ctx, intent.Raw, intent.Normalised)
		if inferred != nil {
			return *inferred
		}
		intent.Clarify = &ClarifyQuestion{
			Prompt: "I couldn't map that to a command. Try help, inventory, look, take, drop, use, craft, eat, drink, sleep, go, inspect.",
		}
		return intent
	}

	if len(alternates) > 0 && (cmdMatch.Score-alternates[0].Score) < 0.05 && alternates[0].Score > 0.65 {
		options := []Intent{
			{
				Raw:        raw,
				Normalised: cmdMatch.Canonical,
				Kind:       commandKind(cmdMatch.Canonical),
				Verb:       cmdMatch.Canonical,
				Confidence: cmdMatch.Score,
			},
			{
				Raw:        raw,
				Normalised: alternates[0].Canonical,
				Kind:       commandKind(alternates[0].Canonical),
				Verb:       alternates[0].Canonical,
				Confidence: alternates[0].Score,
			},
		}
		intent.Clarify = &ClarifyQuestion{
			Prompt:  "Did you mean:",
			Options: options,
		}
		return intent
	}

	intent.Verb = cmdMatch.Canonical
	intent.Kind = commandKind(intent.Verb)
	intent.Confidence = clampScore(cmdMatch.Score)

	argsTokens := tokens
	if cmdMatch.Consumed > 0 && len(tokens) >= cmdMatch.Consumed {
		argsTokens = tokens[cmdMatch.Consumed:]
	}
	argsTokens, q := splitQuantity(argsTokens)
	intent.Quantity = q

	def, _ := p.registry.command(intent.Verb)
	resolvedArgs, clarify, argScore := p.resolveArgs(ctx, def, argsTokens)
	if clarify != nil {
		intent.Clarify = clarify
		intent.Confidence = 0.45
		return intent
	}
	intent.Args = resolvedArgs
	intent.Confidence = clampScore((intent.Confidence * 0.75) + (argScore * 0.25))

	if intent.Kind == Command && len(intent.Args) < def.MinArgs {
		if def.MinArgs > 0 && (def.Canonical == "take" || def.Canonical == "drop" || def.Canonical == "use" || def.Canonical == "inspect") {
			options := buildEntityOptions(ctx, def.Canonical, 5)
			if len(options) > 0 {
				intent.Clarify = &ClarifyQuestion{
					Prompt:  fmt.Sprintf("What should I %s?", def.Canonical),
					Options: options,
				}
				intent.Confidence = 0.46
				return intent
			}
		}
		intent.Clarify = &ClarifyQuestion{Prompt: fmt.Sprintf("%s needs at least %d argument(s).", def.Canonical, def.MinArgs)}
		intent.Confidence = 0.42
		return intent
	}

	if def.MaxArgs > 0 && len(intent.Args) > def.MaxArgs {
		intent.Args = append([]string(nil), intent.Args[:def.MaxArgs]...)
		intent.Confidence = clampScore(intent.Confidence - 0.05)
	}

	if intent.Confidence < 0.52 && intent.Clarify == nil {
		intent.Clarify = &ClarifyQuestion{Prompt: "I have low confidence in that parse. Please rephrase or pick a clearer command."}
	}
	return intent
}

func commandKind(verb string) IntentKind {
	switch verb {
	case "help":
		return Help
	case "look", "inspect", "inventory":
		return Query
	default:
		return Command
	}
}

func splitQuantity(tokens []string) ([]string, *Quantity) {
	if len(tokens) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(tokens))
	var q *Quantity
	for _, token := range tokens {
		if q == nil {
			if candidate := parseQuantityToken(token); candidate != nil {
				q = candidate
				continue
			}
		}
		out = append(out, token)
	}
	return out, q
}

func (p *Parser) resolveArgs(ctx ParseContext, def CommandDef, args []string) ([]string, *ClarifyQuestion, float64) {
	if len(args) == 0 {
		return nil, nil, 0.9
	}

	resolved := make([]string, 0, len(args))
	score := 0.9
	for i := 0; i < len(args); i++ {
		token := args[i]
		if isPronoun(token) {
			if strings.TrimSpace(ctx.LastEntity) == "" {
				return nil, &ClarifyQuestion{Prompt: "What does that pronoun refer to?"}, 0.4
			}
			resolved = append(resolved, normaliseInput(ctx.LastEntity))
			score -= 0.08
			continue
		}

		if def.Canonical == "go" && i == 0 {
			mapped := mapDirection(token)
			if mapped == "" {
				entity, confidence, tie := resolveDirection(token, ctx.KnownDirections)
				if tie {
					options := []Intent{
						{Kind: Command, Verb: "go", Args: []string{entity[0]}, Confidence: confidence},
						{Kind: Command, Verb: "go", Args: []string{entity[1]}, Confidence: confidence - 0.01},
					}
					return nil, &ClarifyQuestion{Prompt: "Which direction?", Options: options}, 0.5
				}
				if len(entity) > 0 {
					mapped = entity[0]
					score = minScore(score, confidence)
				}
			}
			if mapped != "" {
				resolved = append(resolved, mapped)
				continue
			}
		}

		if expectsEntity(def.Canonical, i) {
			joined := token
			// For multi-token entities, greedily join 2-3 words.
			if i+1 < len(args) {
				try := token + " " + args[i+1]
				if _, s, _ := resolveEntity(try, ctx, def.Canonical); s > 0.9 {
					joined = try
					i++
				}
			}
			entity, confidence, tie := resolveEntity(joined, ctx, def.Canonical)
			if tie && len(entity) >= 2 {
				options := make([]Intent, 0, 2)
				for idx := 0; idx < 2; idx++ {
					options = append(options, Intent{
						Kind:       commandKind(def.Canonical),
						Verb:       def.Canonical,
						Args:       []string{entity[idx]},
						Confidence: confidence - float64(idx)*0.01,
					})
				}
				return nil, &ClarifyQuestion{
					Prompt:  fmt.Sprintf("Did you mean %s?", def.Canonical),
					Options: options,
				}, 0.52
			}
			if len(entity) == 1 {
				resolved = append(resolved, entity[0])
				score = minScore(score, confidence)
				continue
			}
		}

		resolved = append(resolved, token)
		score -= 0.02
	}
	return resolved, nil, clampScore(score)
}

func expectsEntity(verb string, argPos int) bool {
	if argPos > 0 && verb != "use" {
		return false
	}
	switch verb {
	case "take", "drop", "use", "inspect", "craft", "eat", "drink":
		return true
	default:
		return false
	}
}

func resolveDirection(token string, known []string) ([]string, float64, bool) {
	n := normaliseInput(token)
	if d := mapDirection(n); d != "" {
		return []string{d}, 0.98, false
	}
	if len(known) == 0 {
		known = []string{"north", "south", "east", "west"}
	}
	return bestMatches(n, known, nil, nil)
}

func resolveEntity(token string, ctx ParseContext, verb string) ([]string, float64, bool) {
	n := normaliseInput(token)
	if n == "" {
		return nil, 0, false
	}
	near := make([]string, 0, len(ctx.Nearby))
	for _, item := range ctx.Nearby {
		v := normaliseInput(item)
		if v != "" {
			near = append(near, v)
		}
	}
	inv := make([]string, 0, len(ctx.Inventory))
	for _, item := range ctx.Inventory {
		v := normaliseInput(item)
		if v != "" {
			inv = append(inv, v)
		}
	}
	return bestMatches(n, mergeUnique(near, inv), near, invForVerb(verb, inv))
}

func invForVerb(verb string, inv []string) []string {
	switch verb {
	case "drop", "use", "eat", "drink":
		return inv
	default:
		return nil
	}
}

func bestMatches(token string, all []string, nearbyBoost []string, inventoryBoost []string) ([]string, float64, bool) {
	if len(all) == 0 {
		return nil, 0, false
	}
	type scored struct {
		val   string
		score float64
	}
	nearSet := make(map[string]bool, len(nearbyBoost))
	for _, n := range nearbyBoost {
		nearSet[n] = true
	}
	invSet := make(map[string]bool, len(inventoryBoost))
	for _, n := range inventoryBoost {
		invSet[n] = true
	}

	results := make([]scored, 0, len(all))
	for _, cand := range all {
		score := 0.0
		switch {
		case token == cand:
			score = 1.0
		case strings.HasPrefix(cand, token) && len(token) >= 2:
			score = 0.9
		default:
			dist := levenshtein.ComputeDistance(token, cand)
			if dist > levenshteinLimit(len(cand)) {
				continue
			}
			score = 0.72 - (0.08 * float64(dist))
		}
		if nearSet[cand] {
			score += 0.08
		}
		if invSet[cand] {
			score += 0.08
		}
		results = append(results, scored{val: cand, score: clampScore(score)})
	}
	if len(results) == 0 {
		return nil, 0, false
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].score == results[j].score {
			return results[i].val < results[j].val
		}
		return results[i].score > results[j].score
	})

	best := results[0]
	tie := len(results) > 1 && (best.score-results[1].score) < 0.05 && results[1].score > 0.6
	if tie {
		return []string{best.val, results[1].val}, best.score, true
	}
	return []string{best.val}, best.score, false
}

func buildEntityOptions(ctx ParseContext, verb string, maxOptions int) []Intent {
	pool := make([]string, 0)
	if verb == "take" {
		pool = append(pool, ctx.Nearby...)
	} else {
		pool = append(pool, ctx.Inventory...)
	}
	seen := map[string]bool{}
	options := make([]Intent, 0, maxOptions)
	for _, entity := range pool {
		n := normaliseInput(entity)
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		options = append(options, Intent{
			Kind:       commandKind(verb),
			Verb:       verb,
			Args:       []string{n},
			Confidence: 0.88,
		})
		if len(options) >= maxOptions {
			break
		}
	}
	return options
}

func inferFreeTextIntent(ctx ParseContext, raw string, normalised string) *Intent {
	n := normalised
	makeIntent := func(kind IntentKind, verb string, args []string, confidence float64) *Intent {
		return &Intent{
			Raw:        raw,
			Normalised: normalised,
			Kind:       kind,
			Verb:       verb,
			Args:       args,
			Confidence: clampScore(confidence),
		}
	}

	if containsAnyPhrase(n,
		"check my bag", "check bag", "what do i have", "what i have", "what have i got", "my inventory", "open bag",
	) {
		return makeIntent(Query, "inventory", nil, 0.92)
	}
	if n == "inventory" || n == "inv" {
		return makeIntent(Query, "inventory", nil, 0.98)
	}

	if containsAnyPhrase(n, "i need a fire", "i need fire", "make a fire", "build fire", "starting fire", "start fire", "im freezing", "i m freezing") {
		return makeIntent(Command, "fire", []string{"build"}, 0.84)
	}
	if containsAnyPhrase(n, "smoke meat", "preserve meat", "cure meat", "dry meat", "keep meat from spoiling") {
		return makeIntent(Command, "preserve", nil, 0.8)
	}

	if containsAnyPhrase(n, "where am i", "look around", "look about", "where i am") {
		return makeIntent(Query, "look", nil, 0.88)
	}

	if dir := inferDirectionFromText(n); dir != "" {
		return makeIntent(Command, "go", []string{dir}, 0.86)
	}

	if containsWord(n, "eat") {
		return makeIntent(Command, "eat", nil, 0.78)
	}
	if containsWord(n, "drink") {
		return makeIntent(Command, "drink", nil, 0.78)
	}
	if containsWord(n, "sleep") || containsWord(n, "rest") {
		return makeIntent(Command, "sleep", nil, 0.8)
	}

	// Free text "pick up stic" pattern fallback.
	if containsAnyPhrase(n, "pick up", "pickup") || strings.HasPrefix(n, "grab ") {
		tokens := tokenise(n)
		if len(tokens) > 1 {
			entity := strings.Join(tokens[1:], " ")
			if strings.HasPrefix(entity, "up ") {
				entity = strings.TrimPrefix(entity, "up ")
			}
			if entity != "" {
				m, confidence, tie := resolveEntity(entity, ctx, "take")
				if tie && len(m) >= 2 {
					return &Intent{
						Raw:        raw,
						Normalised: normalised,
						Kind:       Command,
						Verb:       "take",
						Confidence: 0.52,
						Clarify: &ClarifyQuestion{
							Prompt: "Did you mean:",
							Options: []Intent{
								{Kind: Command, Verb: "take", Args: []string{m[0]}, Confidence: confidence},
								{Kind: Command, Verb: "take", Args: []string{m[1]}, Confidence: confidence - 0.01},
							},
						},
					}
				}
				if len(m) == 1 {
					return makeIntent(Command, "take", []string{m[0]}, confidence)
				}
				return makeIntent(Command, "take", []string{entity}, 0.62)
			}
		}
	}

	return nil
}

func inferDirectionFromText(normalised string) string {
	tokens := tokenise(normalised)
	if len(tokens) == 0 {
		return ""
	}
	for i, token := range tokens {
		mapped := mapDirection(token)
		if mapped == "" {
			continue
		}
		// "go n", "walk north", "head east", etc.
		if i > 0 {
			prev := tokens[i-1]
			if prev == "go" || prev == "walk" || prev == "head" || prev == "travel" || prev == "move" {
				return mapped
			}
		}
		if i == 0 && len(tokens) == 1 {
			return mapped
		}
	}
	if strings.Contains(normalised, "go ") {
		parts := strings.Split(normalised, "go ")
		if len(parts) > 1 {
			next := strings.Fields(parts[1])
			if len(next) > 0 {
				if mapped := mapDirection(next[0]); mapped != "" {
					return mapped
				}
			}
		}
	}
	return ""
}

func containsAnyPhrase(value string, phrases ...string) bool {
	for _, phrase := range phrases {
		if containsPhrase(value, phrase) {
			return true
		}
	}
	return false
}

func containsPhrase(value, phrase string) bool {
	p := normaliseInput(phrase)
	if p == "" {
		return false
	}
	return strings.Contains(" "+value+" ", " "+p+" ")
}

func containsWord(value, word string) bool {
	w := normaliseInput(word)
	if w == "" {
		return false
	}
	return strings.Contains(" "+value+" ", " "+w+" ")
}

func mergeUnique(a, b []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(a)+len(b))
	add := func(list []string) {
		for _, v := range list {
			n := normaliseInput(v)
			if n == "" || seen[n] {
				continue
			}
			seen[n] = true
			out = append(out, n)
		}
	}
	add(a)
	add(b)
	return out
}

func minScore(a, b float64) float64 {
	if b < a {
		return b
	}
	return a
}

func clampScore(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func IntentToCommandString(intent Intent) string {
	verb := normaliseInput(intent.Verb)
	if verb == "" {
		return ""
	}
	args := make([]string, 0, len(intent.Args)+1)
	for _, arg := range intent.Args {
		n := normaliseInput(arg)
		if n != "" {
			args = append(args, n)
		}
	}
	if intent.Quantity != nil && intent.Quantity.Raw != "" {
		args = append(args, normaliseInput(intent.Quantity.Raw))
	}
	if len(args) == 0 {
		return verb
	}
	return verb + " " + strings.Join(args, " ")
}
