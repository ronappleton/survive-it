package gui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/appengine-ltd/survive-it/internal/game"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type kitPickerState struct {
	Cursor int
}

type scenarioBuilderState struct {
	Cursor       int
	ModeIndex    int
	ListCursor   int
	EditingRow   int
	EditBuffer   string
	Scenario     game.Scenario
	Status       string
	SourceID     game.ScenarioID
	SourceName   string
	SourceCustom bool
}

type scenarioBuilderEntry struct {
	Scenario game.Scenario
	Custom   bool
}

func (ui *gameUI) openKitPicker() {
	ui.kit.Cursor = 0
	ui.screen = screenKitPicker
}

func (ui *gameUI) updateKitPicker() {
	if len(ui.pcfg.Players) == 0 {
		ui.screen = screenPlayerConfig
		return
	}
	if ui.pcfg.PlayerIndex < 0 || ui.pcfg.PlayerIndex >= len(ui.pcfg.Players) {
		ui.pcfg.PlayerIndex = 0
	}
	player := &ui.pcfg.Players[ui.pcfg.PlayerIndex]
	items := game.AllKitItems()
	if len(items) == 0 {
		ui.screen = screenPlayerConfig
		return
	}

	if rl.IsKeyPressed(rl.KeyEscape) {
		ui.screen = screenPlayerConfig
		return
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		ui.kit.Cursor = wrapIndex(ui.kit.Cursor+1, len(items))
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		ui.kit.Cursor = wrapIndex(ui.kit.Cursor-1, len(items))
	}
	if rl.IsKeyPressed(rl.KeyLeft) {
		ui.pcfg.PlayerIndex = wrapIndex(ui.pcfg.PlayerIndex-1, len(ui.pcfg.Players))
	}
	if rl.IsKeyPressed(rl.KeyRight) {
		ui.pcfg.PlayerIndex = wrapIndex(ui.pcfg.PlayerIndex+1, len(ui.pcfg.Players))
	}
	if rl.IsKeyPressed(rl.KeyR) {
		player.Kit = nil
		ui.status = ""
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		selected := items[ui.kit.Cursor]
		if hasKitItem(player.Kit, selected) {
			player.Kit = removeKitItem(player.Kit, selected)
			ui.status = ""
			return
		}
		limit := maxInt(1, player.KitLimit)
		if len(player.Kit) >= limit {
			ui.status = fmt.Sprintf("Kit limit reached (%d)", limit)
			return
		}
		player.Kit = append(player.Kit, selected)
		ui.status = ""
	}
}

func (ui *gameUI) drawKitPicker() {
	left := rl.NewRectangle(20, 20, float32(ui.width)*0.52, float32(ui.height-40))
	right := rl.NewRectangle(left.X+left.Width+16, 20, float32(ui.width)-left.Width-56, float32(ui.height-40))
	drawPanel(left, "Kit Picker")
	drawPanel(right, "Selection")

	if len(ui.pcfg.Players) == 0 {
		drawWrappedText("No players configured.", left, 60, 22, colorWarn)
		return
	}
	if ui.pcfg.PlayerIndex < 0 || ui.pcfg.PlayerIndex >= len(ui.pcfg.Players) {
		ui.pcfg.PlayerIndex = 0
	}
	player := ui.pcfg.Players[ui.pcfg.PlayerIndex]
	items := game.AllKitItems()
	if ui.kit.Cursor < 0 || ui.kit.Cursor >= len(items) {
		ui.kit.Cursor = 0
	}

	rl.DrawText(
		fmt.Sprintf("Player %d/%d: %s  |  Kit %d/%d", ui.pcfg.PlayerIndex+1, len(ui.pcfg.Players), player.Name, len(player.Kit), maxInt(1, player.KitLimit)),
		int32(left.X)+14,
		int32(left.Y)+38,
		20,
		colorAccent,
	)

	y := int32(left.Y) + 70
	for i, item := range items {
		if y > int32(left.Y+left.Height)-42 {
			break
		}
		checked := hasKitItem(player.Kit, item)
		prefix := "[ ] "
		if checked {
			prefix = "[x] "
		}
		if i == ui.kit.Cursor {
			rl.DrawRectangle(int32(left.X)+10, y-4, int32(left.Width)-20, 28, rl.Fade(colorAccent, 0.2))
		}
		clr := colorText
		if checked {
			clr = colorAccent
		}
		rl.DrawText(prefix+string(item), int32(left.X)+18, y, 20, clr)
		y += 30
	}

	selected := make([]string, 0, len(player.Kit))
	for _, item := range player.Kit {
		selected = append(selected, string(item))
	}
	if len(selected) == 0 {
		selected = append(selected, "(none)")
	}
	detail := []string{
		fmt.Sprintf("Player: %s", player.Name),
		fmt.Sprintf("Limit: %d", maxInt(1, player.KitLimit)),
		"",
		"Selected Items:",
		strings.Join(selected, ", "),
		"",
		"Controls:",
		"Up/Down: move",
		"Enter: toggle item",
		"Left/Right: player",
		"R: reset player kit",
		"Esc: back",
	}
	drawLines(right, 46, 20, detail, colorText)
	if strings.TrimSpace(ui.status) != "" {
		rl.DrawText(ui.status, int32(right.X)+14, int32(right.Y+right.Height)-34, 18, colorWarn)
	}
}

func hasKitItem(items []game.KitItem, item game.KitItem) bool {
	for _, x := range items {
		if x == item {
			return true
		}
	}
	return false
}

func removeKitItem(items []game.KitItem, item game.KitItem) []game.KitItem {
	out := make([]game.KitItem, 0, len(items))
	removed := false
	for _, x := range items {
		if !removed && x == item {
			removed = true
			continue
		}
		out = append(out, x)
	}
	return out
}

func (ui *gameUI) openScenarioBuilder() {
	ui.sb = scenarioBuilderState{
		Cursor:       0,
		ModeIndex:    ui.setup.ModeIndex,
		ListCursor:   0,
		EditingRow:   -1,
		Scenario:     newScenarioTemplate(modeOptions()[clampInt(ui.setup.ModeIndex, 0, len(modeOptions())-1)]),
		SourceID:     "",
		SourceName:   "",
		SourceCustom: false,
	}
	ui.screen = screenScenarioBuilder
}

func (ui *gameUI) updateScenarioBuilder() {
	if ui.sb.EditingRow >= 0 {
		captureTextInput(&ui.sb.EditBuffer, 220)
		if rl.IsKeyPressed(rl.KeyEnter) {
			ui.commitScenarioEdit()
		}
		if rl.IsKeyPressed(rl.KeyEscape) {
			ui.sb.EditingRow = -1
		}
		return
	}

	if rl.IsKeyPressed(rl.KeyEscape) {
		ui.enterMenu()
		return
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		ui.sb.Cursor = wrapIndex(ui.sb.Cursor+1, 12)
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		ui.sb.Cursor = wrapIndex(ui.sb.Cursor-1, 12)
	}
	if rl.IsKeyPressed(rl.KeyLeft) {
		ui.adjustScenarioBuilder(-1)
	}
	if rl.IsKeyPressed(rl.KeyRight) {
		ui.adjustScenarioBuilder(1)
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		switch ui.sb.Cursor {
		case 2, 3, 4, 5, 6:
			ui.startScenarioEdit(ui.sb.Cursor)
		case 8:
			ui.loadSelectedScenario()
		case 9:
			ui.saveScenarioFromBuilder()
		case 10:
			ui.deleteSelectedCustomScenario()
		case 11:
			ui.enterMenu()
		}
	}
}

func (ui *gameUI) adjustScenarioBuilder(delta int) {
	modes := modeOptions()
	switch ui.sb.Cursor {
	case 0:
		ui.sb.ModeIndex = wrapIndex(ui.sb.ModeIndex+delta, len(modes))
		ui.sb.Scenario = newScenarioTemplate(modes[ui.sb.ModeIndex])
		ui.sb.SourceID = ""
		ui.sb.SourceName = ""
		ui.sb.SourceCustom = false
		ui.sb.ListCursor = 0
	case 7:
		ui.sb.Scenario.DefaultDays = clampInt(ui.sb.Scenario.DefaultDays+delta*3, 1, 365)
	case 8, 10:
		list := ui.scenarioBuilderEntriesForMode(modes[ui.sb.ModeIndex])
		if len(list) > 0 {
			ui.sb.ListCursor = wrapIndex(ui.sb.ListCursor+delta, len(list))
		}
	}
}

func (ui *gameUI) drawScenarioBuilder() {
	left := rl.NewRectangle(20, 20, float32(ui.width)*0.46, float32(ui.height-40))
	right := rl.NewRectangle(left.X+left.Width+16, 20, float32(ui.width)-left.Width-56, float32(ui.height-40))
	drawPanel(left, "Scenario Builder")
	drawPanel(right, "Scenarios")

	mode := modeOptions()[clampInt(ui.sb.ModeIndex, 0, len(modeOptions())-1)]
	loadedLabel := "New Scenario"
	if strings.TrimSpace(ui.sb.SourceName) != "" {
		loadedLabel = ui.sb.SourceName
		if ui.sb.SourceCustom {
			loadedLabel += " (Custom)"
		} else {
			loadedLabel += " (Built-in)"
		}
	}
	rows := []struct {
		label string
		value string
	}{
		{"Mode", modeLabel(mode)},
		{"Loaded", loadedLabel},
		{"Name", ui.sb.Scenario.Name},
		{"Biome", ui.sb.Scenario.Biome},
		{"Description", ui.sb.Scenario.Description},
		{"Daunting", ui.sb.Scenario.Daunting},
		{"Motivation", ui.sb.Scenario.Motivation},
		{"Default Days", fmt.Sprintf("%d", ui.sb.Scenario.DefaultDays)},
		{"Load Selected", "Enter"},
		{"Save Scenario", "Enter"},
		{"Delete Selected", "Enter"},
		{"Back", "Enter"},
	}

	y := int32(left.Y) + 54
	for i, row := range rows {
		if i == ui.sb.Cursor {
			rl.DrawRectangle(int32(left.X)+10, y-5, int32(left.Width)-20, 30, rl.Fade(colorAccent, 0.2))
		}
		value := row.value
		if len(value) > 34 {
			value = value[:34] + "..."
		}
		rl.DrawText(row.label, int32(left.X)+16, y, 19, colorText)
		rl.DrawText(value, int32(left.X)+188, y, 19, colorAccent)
		y += 34
	}
	rl.DrawText("Left/Right change mode/days/list or selected row", int32(left.X)+14, int32(left.Y+left.Height)-64, 17, colorDim)
	rl.DrawText("Built-in edits must be saved with a new name", int32(left.X)+14, int32(left.Y+left.Height)-40, 17, colorDim)

	list := ui.scenarioBuilderEntriesForMode(mode)
	ly := int32(right.Y) + 52
	if len(list) == 0 {
		rl.DrawText("No scenarios for this mode yet.", int32(right.X)+14, ly, 20, colorWarn)
	} else {
		for i, entry := range list {
			if ly > int32(right.Y+right.Height)-48 {
				break
			}
			if i == ui.sb.ListCursor {
				rl.DrawRectangle(int32(right.X)+10, ly-5, int32(right.Width)-20, 30, rl.Fade(colorAccent, 0.2))
			}
			source := "Built-in"
			if entry.Custom {
				source = "Custom"
			}
			s := entry.Scenario
			line := fmt.Sprintf("%s [%s]  (%s, %dd)", s.Name, source, s.Biome, s.DefaultDays)
			rl.DrawText(line, int32(right.X)+16, ly, 19, colorText)
			ly += 34
		}
		sel := list[clampInt(ui.sb.ListCursor, 0, len(list)-1)].Scenario
		drawWrappedText("Selected Description: "+safeText(sel.Description), right, int32(right.Height)-220, 19, colorDim)
		drawWrappedText("Daunting: "+safeText(sel.Daunting), right, int32(right.Height)-150, 19, colorWarn)
	}

	if ui.sb.EditingRow >= 0 {
		r := rl.NewRectangle(left.X+18, left.Y+left.Height-138, left.Width-36, 110)
		rl.DrawRectangleRounded(r, 0.16, 8, rl.Fade(colorPanel, 0.96))
		rl.DrawRectangleRoundedLinesEx(r, 0.16, 8, 2, colorAccent)
		rl.DrawText("Editing (Enter apply, Esc cancel)", int32(r.X)+12, int32(r.Y)+10, 18, colorAccent)
		drawWrappedText(ui.sb.EditBuffer+"_", r, 40, 21, colorText)
	}
	if strings.TrimSpace(ui.sb.Status) != "" {
		rl.DrawText(ui.sb.Status, int32(left.X)+14, int32(left.Y+left.Height)-20, 17, colorWarn)
	}
}

func (ui *gameUI) startScenarioEdit(row int) {
	ui.sb.EditingRow = row
	switch row {
	case 2:
		ui.sb.EditBuffer = ui.sb.Scenario.Name
	case 3:
		ui.sb.EditBuffer = ui.sb.Scenario.Biome
	case 4:
		ui.sb.EditBuffer = ui.sb.Scenario.Description
	case 5:
		ui.sb.EditBuffer = ui.sb.Scenario.Daunting
	case 6:
		ui.sb.EditBuffer = ui.sb.Scenario.Motivation
	}
}

func (ui *gameUI) commitScenarioEdit() {
	value := strings.TrimSpace(ui.sb.EditBuffer)
	switch ui.sb.EditingRow {
	case 2:
		if value != "" {
			ui.sb.Scenario.Name = value
		}
	case 3:
		if value != "" {
			ui.sb.Scenario.Biome = value
			ui.sb.Scenario.Wildlife = game.WildlifeForBiome(value)
		}
	case 4:
		ui.sb.Scenario.Description = value
	case 5:
		ui.sb.Scenario.Daunting = value
	case 6:
		ui.sb.Scenario.Motivation = value
	}
	ui.sb.EditingRow = -1
}

func (ui *gameUI) customScenariosForMode(mode game.GameMode) []game.Scenario {
	out := make([]game.Scenario, 0, len(ui.customScenarios))
	for _, scenario := range ui.customScenarios {
		for _, supported := range scenario.SupportedModes {
			if supported == mode {
				out = append(out, scenario)
				break
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func supportsMode(s game.Scenario, mode game.GameMode) bool {
	for _, supported := range s.SupportedModes {
		if supported == mode {
			return true
		}
	}
	return false
}

func (ui *gameUI) scenarioBuilderEntriesForMode(mode game.GameMode) []scenarioBuilderEntry {
	entries := make([]scenarioBuilderEntry, 0, len(game.BuiltInScenarios())+len(ui.customScenarios))
	for _, scenario := range game.BuiltInScenarios() {
		if supportsMode(scenario, mode) {
			entries = append(entries, scenarioBuilderEntry{
				Scenario: scenario,
				Custom:   false,
			})
		}
	}
	for _, scenario := range ui.customScenarios {
		if supportsMode(scenario, mode) {
			entries = append(entries, scenarioBuilderEntry{
				Scenario: scenario,
				Custom:   true,
			})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if strings.EqualFold(entries[i].Scenario.Name, entries[j].Scenario.Name) {
			if entries[i].Custom == entries[j].Custom {
				return entries[i].Scenario.ID < entries[j].Scenario.ID
			}
			// Built-in entries first when names collide.
			return !entries[i].Custom
		}
		return entries[i].Scenario.Name < entries[j].Scenario.Name
	})
	return entries
}

func (ui *gameUI) loadSelectedScenario() {
	mode := modeOptions()[clampInt(ui.sb.ModeIndex, 0, len(modeOptions())-1)]
	list := ui.scenarioBuilderEntriesForMode(mode)
	if len(list) == 0 {
		ui.sb.Status = "No scenarios to load"
		return
	}
	sel := list[clampInt(ui.sb.ListCursor, 0, len(list)-1)]
	ui.sb.Scenario = sel.Scenario
	ui.sb.SourceID = sel.Scenario.ID
	ui.sb.SourceName = sel.Scenario.Name
	ui.sb.SourceCustom = sel.Custom
	if sel.Custom {
		ui.sb.Status = "Loaded custom scenario into editor"
	} else {
		ui.sb.Status = "Loaded built-in scenario template. Save as a new name."
	}
}

func (ui *gameUI) saveScenarioFromBuilder() {
	mode := modeOptions()[clampInt(ui.sb.ModeIndex, 0, len(modeOptions())-1)]
	scenario := ui.sb.Scenario
	if strings.TrimSpace(scenario.Name) == "" {
		ui.sb.Status = "Name is required"
		return
	}
	if strings.TrimSpace(scenario.Biome) == "" {
		scenario.Biome = "temperate_forest"
	}
	if scenario.DefaultDays <= 0 {
		scenario.DefaultDays = 30
	}
	normalizeScenarioForMode(&scenario, mode)

	editingBuiltIn := ui.sb.SourceID != "" && !ui.sb.SourceCustom
	if editingBuiltIn && strings.EqualFold(strings.TrimSpace(scenario.Name), strings.TrimSpace(ui.sb.SourceName)) {
		ui.sb.Status = "Built-in edits must be saved with a new scenario name"
		return
	}

	editingCustom := ui.sb.SourceCustom && ui.sb.SourceID != ""
	if editingCustom {
		scenario.ID = ui.sb.SourceID
	} else {
		// New scenarios and built-in edits always create a custom copy.
		existing := append([]game.Scenario{}, game.BuiltInScenarios()...)
		existing = append(existing, ui.customScenarios...)
		scenario.ID = game.ScenarioID(generateScenarioID(scenario.Name, mode, existing))
	}

	replaced := false
	if editingCustom {
		for i := range ui.customScenarios {
			if ui.customScenarios[i].ID == scenario.ID {
				ui.customScenarios[i] = scenario
				replaced = true
				break
			}
		}
	}
	if !replaced {
		ui.customScenarios = append(ui.customScenarios, scenario)
	}

	if err := saveCustomScenarios(defaultCustomScenariosFile, ui.customScenarios); err != nil {
		ui.sb.Status = "Save failed: " + err.Error()
		return
	}
	game.SetExternalScenarios(ui.customScenarios)
	ui.syncScenarioSelection()
	ui.sb.Status = "Scenario saved"
	ui.sb.Scenario = scenario
	ui.sb.SourceID = scenario.ID
	ui.sb.SourceName = scenario.Name
	ui.sb.SourceCustom = true
}

func (ui *gameUI) deleteSelectedCustomScenario() {
	mode := modeOptions()[clampInt(ui.sb.ModeIndex, 0, len(modeOptions())-1)]
	list := ui.scenarioBuilderEntriesForMode(mode)
	if len(list) == 0 {
		ui.sb.Status = "No scenarios to delete"
		return
	}
	sel := list[clampInt(ui.sb.ListCursor, 0, len(list)-1)]
	if !sel.Custom {
		ui.sb.Status = "Only custom scenarios can be deleted"
		return
	}

	next := make([]game.Scenario, 0, len(ui.customScenarios)-1)
	for _, scenario := range ui.customScenarios {
		if scenario.ID == sel.Scenario.ID {
			continue
		}
		next = append(next, scenario)
	}
	ui.customScenarios = next
	if err := saveCustomScenarios(defaultCustomScenariosFile, ui.customScenarios); err != nil {
		ui.sb.Status = "Delete save failed: " + err.Error()
		return
	}
	game.SetExternalScenarios(ui.customScenarios)
	ui.sb.ListCursor = 0
	ui.syncScenarioSelection()
	ui.sb.Status = "Scenario deleted"
}
