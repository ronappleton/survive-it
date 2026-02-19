package parser

import (
	"sort"
	"strings"

	"github.com/agnivade/levenshtein"
)

type commandPhrase struct {
	canonical string
	alias     string
	tokens    []string
}

type Registry struct {
	commands map[string]CommandDef
	phrases  []commandPhrase
}

func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]CommandDef),
	}
}

func (r *Registry) RegisterCommand(c CommandDef) {
	c.Canonical = normaliseInput(c.Canonical)
	if c.Canonical == "" {
		return
	}
	if c.HandlerKey == "" {
		c.HandlerKey = c.Canonical
	}
	r.commands[c.Canonical] = c

	r.phrases = append(r.phrases, commandPhrase{
		canonical: c.Canonical,
		alias:     c.Canonical,
		tokens:    tokenise(c.Canonical),
	})
	for _, a := range c.Aliases {
		n := normaliseInput(a)
		if n == "" {
			continue
		}
		r.phrases = append(r.phrases, commandPhrase{
			canonical: c.Canonical,
			alias:     n,
			tokens:    tokenise(n),
		})
	}
}

func (r *Registry) command(canonical string) (CommandDef, bool) {
	canonical = normaliseInput(canonical)
	cmd, ok := r.commands[canonical]
	return cmd, ok
}

type commandCandidate struct {
	Canonical string
	Alias     string
	Consumed  int
	Score     float64
	Source    string
}

func (r *Registry) matchCommand(tokens []string) (commandCandidate, []commandCandidate) {
	if len(tokens) == 0 {
		return commandCandidate{}, nil
	}
	in := strings.Join(tokens, " ")
	cands := make([]commandCandidate, 0, len(r.phrases))
	for _, phrase := range r.phrases {
		if len(phrase.tokens) == 0 {
			continue
		}
		consumed := min(len(tokens), len(phrase.tokens))
		prefix := strings.Join(tokens[:consumed], " ")

		if consumed == len(phrase.tokens) && prefix == phrase.alias {
			score := 1.0
			source := "exact"
			if phrase.alias != phrase.canonical {
				score = 0.97
				source = "alias"
			}
			cands = append(cands, commandCandidate{
				Canonical: phrase.canonical,
				Alias:     phrase.alias,
				Consumed:  consumed,
				Score:     score,
				Source:    source,
			})
			continue
		}

		if len(phrase.tokens) == 1 && strings.HasPrefix(phrase.alias, tokens[0]) && len(tokens[0]) >= 2 {
			cands = append(cands, commandCandidate{
				Canonical: phrase.canonical,
				Alias:     phrase.alias,
				Consumed:  1,
				Score:     0.9,
				Source:    "prefix",
			})
			continue
		}

		// Fuzzy: only when there was no exact/prefix hit for this phrase.
		cut := consumed
		compare := prefix
		if len(phrase.tokens) > 1 && len(tokens) >= len(phrase.tokens) {
			cut = len(phrase.tokens)
			compare = strings.Join(tokens[:cut], " ")
		}
		if cut == 0 || compare == "" {
			continue
		}
		if len(compare) < 3 {
			continue
		}
		dist := levenshtein.ComputeDistance(compare, phrase.alias)
		limit := levenshteinLimit(len(phrase.alias))
		if dist > limit {
			continue
		}
		score := 0.72 - (0.08 * float64(dist))
		if strings.Contains(in, phrase.alias) {
			score += 0.04
		}
		if phrase.alias != phrase.canonical {
			score += 0.03
		}
		cands = append(cands, commandCandidate{
			Canonical: phrase.canonical,
			Alias:     phrase.alias,
			Consumed:  cut,
			Score:     score,
			Source:    "lev",
		})
	}

	sort.SliceStable(cands, func(i, j int) bool {
		if cands[i].Score == cands[j].Score {
			if cands[i].Consumed == cands[j].Consumed {
				return cands[i].Canonical < cands[j].Canonical
			}
			return cands[i].Consumed > cands[j].Consumed
		}
		return cands[i].Score > cands[j].Score
	})

	if len(cands) == 0 {
		return commandCandidate{}, nil
	}
	best := cands[0]
	alts := make([]commandCandidate, 0, 4)
	seen := map[string]bool{best.Canonical: true}
	for _, c := range cands[1:] {
		if seen[c.Canonical] {
			continue
		}
		seen[c.Canonical] = true
		alts = append(alts, c)
		if len(alts) >= 4 {
			break
		}
	}
	return best, alts
}

func levenshteinLimit(length int) int {
	switch {
	case length <= 4:
		return 1
	case length <= 8:
		return 2
	default:
		return 3
	}
}

func DefaultRegistry() *Registry {
	r := NewRegistry()
	commands := []CommandDef{
		{Canonical: "help", Aliases: []string{"h", "commands", "?"}, MinArgs: 0, MaxArgs: 0, HandlerKey: "help"},
		{Canonical: "inventory", Aliases: []string{"inv", "bag", "my bag", "check bag", "check my bag"}, MinArgs: 0, MaxArgs: 2, HandlerKey: "inventory"},
		{Canonical: "look", Aliases: []string{"l", "look around", "where am i"}, MinArgs: 0, MaxArgs: 3, HandlerKey: "look"},
		{Canonical: "take", Aliases: []string{"get", "pickup", "pick up", "grab"}, MinArgs: 1, MaxArgs: 6, HandlerKey: "take"},
		{Canonical: "drop", Aliases: []string{"discard", "leave"}, MinArgs: 1, MaxArgs: 6, HandlerKey: "drop"},
		{Canonical: "use", Aliases: []string{"apply"}, MinArgs: 1, MaxArgs: 8, HandlerKey: "use"},
		{Canonical: "craft", Aliases: []string{"make", "build"}, MinArgs: 1, MaxArgs: 8, HandlerKey: "craft"},
		{Canonical: "eat", Aliases: []string{"consume"}, MinArgs: 0, MaxArgs: 6, HandlerKey: "eat"},
		{Canonical: "drink", Aliases: []string{"sip"}, MinArgs: 0, MaxArgs: 6, HandlerKey: "drink"},
		{Canonical: "sleep", Aliases: []string{"rest", "nap"}, MinArgs: 0, MaxArgs: 3, HandlerKey: "sleep"},
		{Canonical: "go", Aliases: []string{"walk", "move", "head", "travel"}, MinArgs: 1, MaxArgs: 3, HandlerKey: "go"},
		{Canonical: "inspect", Aliases: []string{"examine", "check", "chk"}, MinArgs: 1, MaxArgs: 5, HandlerKey: "inspect"},

		// Existing game/run commands to preserve strict-command behavior.
		{Canonical: "next", MinArgs: 0, MaxArgs: 0, HandlerKey: "next"},
		{Canonical: "save", MinArgs: 0, MaxArgs: 0, HandlerKey: "save"},
		{Canonical: "load", MinArgs: 0, MaxArgs: 0, HandlerKey: "load"},
		{Canonical: "menu", Aliases: []string{"back"}, MinArgs: 0, MaxArgs: 0, HandlerKey: "menu"},
		{Canonical: "hunt", Aliases: []string{"catch"}, MinArgs: 1, MaxArgs: 6, HandlerKey: "hunt"},
		{Canonical: "forage", MinArgs: 0, MaxArgs: 4, HandlerKey: "forage"},
		{Canonical: "wood", MinArgs: 1, MaxArgs: 5, HandlerKey: "wood"},
		{Canonical: "resources", MinArgs: 0, MaxArgs: 0, HandlerKey: "resources"},
		{Canonical: "collect", MinArgs: 1, MaxArgs: 4, HandlerKey: "collect"},
		{Canonical: "fire", MinArgs: 1, MaxArgs: 6, HandlerKey: "fire"},
		{Canonical: "shelter", MinArgs: 1, MaxArgs: 4, HandlerKey: "shelter"},
		{Canonical: "trap", MinArgs: 1, MaxArgs: 4, HandlerKey: "trap"},
		{Canonical: "gut", Aliases: []string{"dress", "clean"}, MinArgs: 1, MaxArgs: 4, HandlerKey: "gut"},
		{Canonical: "cook", Aliases: []string{"roast", "boil"}, MinArgs: 1, MaxArgs: 4, HandlerKey: "cook"},
		{Canonical: "preserve", Aliases: []string{"smoke", "dry", "salt", "cure", "smoke meat", "dry meat", "salt meat"}, MinArgs: 2, MaxArgs: 5, HandlerKey: "preserve"},
		{Canonical: "bark", MinArgs: 1, MaxArgs: 5, HandlerKey: "bark"},
		{Canonical: "plants", MinArgs: 0, MaxArgs: 0, HandlerKey: "plants"},
		{Canonical: "actions", MinArgs: 0, MaxArgs: 2, HandlerKey: "actions"},
	}
	for _, cmd := range commands {
		r.RegisterCommand(cmd)
	}
	return r
}
