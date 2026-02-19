package gui

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func ShiftPressedKey(key int32) bool {
	return ShiftKeyPressed(key)
}

func ShiftPressed() bool {
	return shiftDown()
}

func ShiftKeyPressed(key int32) bool {
	if ShiftPressed() && rl.IsKeyPressed(key) {
		return true
	}
	// Accept either key order: Shift then key, or key then Shift.
	if rl.IsKeyDown(key) && (rl.IsKeyPressed(rl.KeyLeftShift) || rl.IsKeyPressed(rl.KeyRightShift)) {
		return true
	}
	return false
}

func HotkeysEnabled(uiState *gameUI) bool {
	if uiState == nil {
		return true
	}
	if uiState.sbuild.Editing || uiState.pcfg.EditingName {
		return false
	}
	// Scenario builder inline-edit mode should only block hotkeys while that
	// screen is active; otherwise a zero-value EditingRow can disable hotkeys.
	if uiState.screen == screenScenarioBuilder && uiState.sb.EditingRow >= 0 {
		return false
	}
	if uiState.screen == screenRun {
		if strings.TrimSpace(uiState.runInput) != "" {
			return false
		}
		if uiState.pendingClarify != nil {
			return false
		}
	}
	return true
}

func ModifiedPressedKey(key int32) bool {
	return (shiftDown() || ctrlDown() || altDown()) && rl.IsKeyPressed(key)
}

func shiftDown() bool {
	return rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
}

func ctrlDown() bool {
	return rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)
}

func altDown() bool {
	return rl.IsKeyDown(rl.KeyLeftAlt) || rl.IsKeyDown(rl.KeyRightAlt)
}
