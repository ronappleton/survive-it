package theme

import rl "github.com/gen2brain/raylib-go/raylib"

// Brand palette for the wilderness field-log UI.
var (
	BG            = rl.NewColor(0x14, 0x1A, 0x1F, 255) // #141A1F
	Panel         = rl.NewColor(0x1C, 0x23, 0x29, 255) // #1C2329
	PanelRaised   = rl.NewColor(0x21, 0x2A, 0x31, 255) // #212A31
	Border        = rl.NewColor(0x2E, 0x3A, 0x40, 255) // #2E3A40
	Divider       = rl.NewColor(0x26, 0x30, 0x38, 255) // #263038
	TextPrimary   = rl.NewColor(0xE8, 0xE2, 0xD8, 255) // #E8E2D8
	TextSecondary = rl.NewColor(0xA6, 0xAD, 0xB1, 255) // #A6ADB1
	TextMuted     = rl.NewColor(0x7D, 0x85, 0x8A, 255) // #7D858A
	AccentEmber   = rl.NewColor(0xD4, 0x6A, 0x1E, 255) // #D46A1E
	AccentForest  = rl.NewColor(0x2F, 0x5D, 0x42, 255) // #2F5D42
	WarningAmber  = rl.NewColor(0xC1, 0x8B, 0x2F, 255) // #C18B2F
	Danger        = rl.NewColor(0xB8, 0x4A, 0x3A, 255) // #B84A3A
	DisabledPanel = rl.NewColor(0x16, 0x1C, 0x21, 255)
	DisabledText  = TextMuted
)
