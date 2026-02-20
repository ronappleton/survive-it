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

type MovementCondition string

const (
	ConditionNone  MovementCondition = ""
	ConditionDark  MovementCondition = "dark"
	ConditionTired MovementCondition = "tired"
)

type MovementScale struct {
	DistanceMeters  float64
	DurationMinutes float64
	Tiles           int
	Condition       MovementCondition
}

type Intent struct {
	Raw           string
	Normalised    string
	Kind          IntentKind
	Verb          string
	Args          []string
	Quantity      *Quantity
	Movement      *MovementScale
	Confidence    float64
	Clarify       *ClarifyQuestion
	ConfirmedRisk bool
}

type ClarifyQuestion struct {
	Prompt  string
	Options []Intent
}

type PendingIntent struct {
	OriginalKind   IntentKind
	OriginalVerb   string
	OriginalIntent *Intent
	FilledArgs     []string
	MissingFields  []string
	Prompt         string
	Options        []Intent
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
