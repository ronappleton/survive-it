package parser

type IntentKind int

const (
	Command IntentKind = iota
	Query
	Help
	Unknown
)

type Quantity struct {
	Raw  string
	N    int
	Unit string
}

type Intent struct {
	Raw        string
	Normalised string
	Kind       IntentKind
	Verb       string
	Args       []string
	Quantity   *Quantity
	Confidence float64
	Clarify    *ClarifyQuestion
}

type ClarifyQuestion struct {
	Prompt  string
	Options []Intent
}

type ParseContext struct {
	Inventory       []string
	Nearby          []string
	KnownDirections []string
	LastEntity      string
}

type CommandDef struct {
	Canonical  string
	Aliases    []string
	MinArgs    int
	MaxArgs    int
	HandlerKey string
}
