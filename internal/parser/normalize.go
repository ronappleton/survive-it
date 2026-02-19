package parser

import (
	"regexp"
	"strconv"
	"strings"
)

var multiSpaceRE = regexp.MustCompile(`\s+`)

func normaliseInput(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return ""
	}
	var b strings.Builder
	lastSpace := false
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastSpace = false
			continue
		}
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '-' || r == '_' || r == '/' || r == '\'' {
			if !lastSpace {
				b.WriteByte(' ')
			}
			lastSpace = true
		}
	}
	return strings.TrimSpace(multiSpaceRE.ReplaceAllString(b.String(), " "))
}

func tokenise(normalised string) []string {
	if strings.TrimSpace(normalised) == "" {
		return nil
	}
	return strings.Fields(normalised)
}

func parseQuantityToken(token string) *Quantity {
	token = strings.TrimSpace(strings.ToLower(token))
	if token == "" {
		return nil
	}
	switch token {
	case "all":
		return &Quantity{Raw: token, N: -1, Unit: "all"}
	case "some":
		return &Quantity{Raw: token, N: 0, Unit: "some"}
	}
	if n, err := strconv.Atoi(token); err == nil && n >= 0 {
		return &Quantity{Raw: token, N: n, Unit: "count"}
	}
	if strings.HasSuffix(token, "h") || strings.HasSuffix(token, "hr") || strings.HasSuffix(token, "hours") {
		n := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(token, "hours"), "hr"), "h")
		if v, err := strconv.Atoi(n); err == nil && v >= 0 {
			return &Quantity{Raw: token, N: v, Unit: "hours"}
		}
	}
	if strings.HasSuffix(token, "m") || strings.HasSuffix(token, "min") || strings.HasSuffix(token, "mins") {
		n := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(token, "mins"), "min"), "m")
		if v, err := strconv.Atoi(n); err == nil && v >= 0 {
			return &Quantity{Raw: token, N: v, Unit: "minutes"}
		}
	}
	return nil
}

func isPronoun(token string) bool {
	switch strings.ToLower(strings.TrimSpace(token)) {
	case "it", "that", "them", "this", "those":
		return true
	default:
		return false
	}
}

func mapDirection(token string) string {
	switch strings.ToLower(strings.TrimSpace(token)) {
	case "n", "north":
		return "north"
	case "s", "south":
		return "south"
	case "e", "east":
		return "east"
	case "w", "west":
		return "west"
	default:
		return ""
	}
}
