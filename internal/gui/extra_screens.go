package gui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/appengine-ltd/survive-it/internal/game"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type kitPickerTarget int

const (
	kitTargetPersonal kitPickerTarget = iota
	kitTargetIssued
)

type kitPickerFocus int

const (
	kitFocusCategories kitPickerFocus = iota
	kitFocusItems
)

type kitCategory struct {
	Label string
	Items []game.KitItem
}

type kitPickerState struct {
	Target      kitPickerTarget
	Focus       kitPickerFocus
	ReturnTo    screen
	CategoryIdx int
	ItemIdx     int
}

type scenarioBuilderState struct {
	Cursor          int
	ModeIndex       int
	ListCursor      int
	PickingScenario bool
	EditingRow      int
	EditBuffer      string
	Scenario        game.Scenario
	Status          string
	SourceID        game.ScenarioID
	SourceName      string
	SourceCustom    bool
}

type scenarioBuilderEntry struct {
	Scenario game.Scenario
	Custom   bool
}

func (ui *gameUI) openPersonalKitPicker(returnTo screen) {
	ui.kit = kitPickerState{
		Target:      kitTargetPersonal,
		Focus:       kitFocusCategories,
		ReturnTo:    returnTo,
		CategoryIdx: 0,
		ItemIdx:     0,
	}
	ui.ensureSetupPlayers()
	ui.screen = screenKitPicker
}

func (ui *gameUI) openIssuedKitPicker(returnTo screen) {
	ui.kit = kitPickerState{
		Target:      kitTargetIssued,
		Focus:       kitFocusCategories,
		ReturnTo:    returnTo,
		CategoryIdx: 0,
		ItemIdx:     0,
	}
	ui.ensureSetupPlayers()
	ui.screen = screenKitPicker
}

func (ui *gameUI) kitPickerItems() []game.KitItem {
	if ui.kit.Target == kitTargetIssued {
		items := append([]game.KitItem(nil), issuedKitOptionsForMode(ui.selectedMode())...)
		for _, selected := range ui.setup.IssuedKit {
			if hasKitItem(items, selected) {
				continue
			}
			items = append(items, selected)
		}
		return items
	}
	return game.AllKitItems()
}

func categorizeKitItems(items []game.KitItem) []kitCategory {
	order := []string{
		"Cutting / Tools",
		"Fire",
		"Water",
		"Food / Hunting",
		"Shelter / Warmth",
		"Navigation / Signal",
		"Protection / Medical",
		"Utility",
	}
	buckets := make(map[string][]game.KitItem, len(order))
	for _, item := range items {
		label := kitCategoryLabel(item)
		buckets[label] = append(buckets[label], item)
	}
	out := make([]kitCategory, 0, len(order))
	for _, label := range order {
		if len(buckets[label]) == 0 {
			continue
		}
		out = append(out, kitCategory{Label: label, Items: buckets[label]})
	}
	return out
}

func kitCategoryLabel(item game.KitItem) string {
	name := strings.ToLower(string(item))
	switch {
	case strings.Contains(name, "knife"),
		strings.Contains(name, "hatchet"),
		strings.Contains(name, "saw"),
		strings.Contains(name, "machete"),
		strings.Contains(name, "shovel"),
		strings.Contains(name, "multi-tool"):
		return "Cutting / Tools"
	case strings.Contains(name, "ferro"),
		strings.Contains(name, "fire"),
		strings.Contains(name, "magnifying"):
		return "Fire"
	case strings.Contains(name, "water"),
		strings.Contains(name, "canteen"),
		strings.Contains(name, "cup"),
		strings.Contains(name, "purification"):
		return "Water"
	case strings.Contains(name, "fishing"),
		strings.Contains(name, "gill"),
		strings.Contains(name, "snare"),
		strings.Contains(name, "bow"),
		strings.Contains(name, "spear"),
		strings.Contains(name, "ration"):
		return "Food / Hunting"
	case strings.Contains(name, "tarp"),
		strings.Contains(name, "sleep"),
		strings.Contains(name, "blanket"),
		strings.Contains(name, "thermal"),
		strings.Contains(name, "rain"),
		strings.Contains(name, "paracord"),
		strings.Contains(name, "rope"),
		strings.Contains(name, "dry bag"):
		return "Shelter / Warmth"
	case strings.Contains(name, "compass"),
		strings.Contains(name, "map"),
		strings.Contains(name, "headlamp"),
		strings.Contains(name, "signal"),
		strings.Contains(name, "whistle"):
		return "Navigation / Signal"
	case strings.Contains(name, "first aid"),
		strings.Contains(name, "insect"),
		strings.Contains(name, "net"),
		strings.Contains(name, "salt"):
		return "Protection / Medical"
	default:
		return "Utility"
	}
}

func (ui *gameUI) normalizeKitPicker(categories []kitCategory) {
	if len(categories) == 0 {
		ui.kit.CategoryIdx = 0
		ui.kit.ItemIdx = 0
		ui.kit.Focus = kitFocusCategories
		return
	}
	ui.kit.CategoryIdx = clampInt(ui.kit.CategoryIdx, 0, len(categories)-1)
	items := categories[ui.kit.CategoryIdx].Items
	if len(items) == 0 {
		ui.kit.ItemIdx = 0
		ui.kit.Focus = kitFocusCategories
		return
	}
	ui.kit.ItemIdx = clampInt(ui.kit.ItemIdx, 0, len(items)-1)
	if ui.kit.Focus != kitFocusCategories && ui.kit.Focus != kitFocusItems {
		ui.kit.Focus = kitFocusCategories
	}
}

func (ui *gameUI) updateKitPicker() {
	ui.ensureSetupPlayers()
	categories := categorizeKitItems(ui.kitPickerItems())
	if len(categories) == 0 {
		ui.screen = ui.kit.ReturnTo
		ui.status = "No kit items available."
		return
	}
	ui.normalizeKitPicker(categories)
	currentItems := categories[ui.kit.CategoryIdx].Items

	if rl.IsKeyPressed(rl.KeyEscape) {
		if ui.kit.ReturnTo == 0 {
			ui.kit.ReturnTo = screenPlayerConfig
		}
		ui.screen = ui.kit.ReturnTo
		return
	}
	if rl.IsKeyPressed(rl.KeyLeft) {
		ui.kit.Focus = kitFocusCategories
	}
	if rl.IsKeyPressed(rl.KeyRight) && len(currentItems) > 0 {
		ui.kit.Focus = kitFocusItems
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		if ui.kit.Focus == kitFocusCategories {
			ui.kit.CategoryIdx = wrapIndex(ui.kit.CategoryIdx-1, len(categories))
			ui.normalizeKitPicker(categories)
		} else if len(currentItems) > 0 {
			ui.kit.ItemIdx = wrapIndex(ui.kit.ItemIdx-1, len(currentItems))
		}
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		if ui.kit.Focus == kitFocusCategories {
			ui.kit.CategoryIdx = wrapIndex(ui.kit.CategoryIdx+1, len(categories))
			ui.normalizeKitPicker(categories)
		} else if len(currentItems) > 0 {
			ui.kit.ItemIdx = wrapIndex(ui.kit.ItemIdx+1, len(currentItems))
		}
	}
	if ShiftPressedKey(rl.KeyR) {
		switch ui.kit.Target {
		case kitTargetPersonal:
			if len(ui.pcfg.Players) > 0 {
				ui.pcfg.Players[ui.pcfg.PlayerIndex].Kit = nil
			}
		case kitTargetIssued:
			ui.resetIssuedKitRecommendations()
		}
		ui.status = ""
	}
	if rl.IsKeyPressed(rl.KeySpace) {
		if ui.kit.Focus == kitFocusItems && len(currentItems) > 0 {
			ui.toggleKitPickerSelection(currentItems[ui.kit.ItemIdx])
		}
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		if ui.kit.Focus == kitFocusCategories {
			if len(currentItems) == 0 {
				ui.status = "No items in this category."
				return
			}
			ui.kit.Focus = kitFocusItems
			return
		}
		if len(currentItems) > 0 {
			ui.toggleKitPickerSelection(currentItems[ui.kit.ItemIdx])
		}
	}
}

func (ui *gameUI) toggleKitPickerSelection(item game.KitItem) {
	switch ui.kit.Target {
	case kitTargetPersonal:
		if len(ui.pcfg.Players) == 0 {
			return
		}
		player := &ui.pcfg.Players[ui.pcfg.PlayerIndex]
		if hasKitItem(player.Kit, item) {
			player.Kit = removeKitItem(player.Kit, item)
			ui.status = ""
			return
		}
		limit := maxInt(1, player.KitLimit)
		if len(player.Kit) >= limit {
			ui.status = fmt.Sprintf("Kit limit reached (%d)", limit)
			return
		}
		player.Kit = append(player.Kit, item)
		ui.status = ""
	case kitTargetIssued:
		if hasKitItem(ui.setup.IssuedKit, item) {
			ui.setup.IssuedKit = removeKitItem(ui.setup.IssuedKit, item)
			ui.setup.IssuedCustom = true
			ui.status = ""
			return
		}
		if !hasKitItem(issuedKitOptionsForMode(ui.selectedMode()), item) {
			ui.status = "Item not allowed for selected mode."
			return
		}
		ui.setup.IssuedKit = append(ui.setup.IssuedKit, item)
		ui.setup.IssuedCustom = true
		ui.status = ""
	}
}

func (ui *gameUI) drawKitPicker() {
	ui.ensureSetupPlayers()
	categories := categorizeKitItems(ui.kitPickerItems())
	if len(categories) == 0 {
		panel := rl.NewRectangle(20, 20, float32(ui.width-40), float32(ui.height-40))
		drawPanel(panel, "Kit Picker")
		drawWrappedText("No kit items available.", panel, 60, 22, colorWarn)
		return
	}
	ui.normalizeKitPicker(categories)

	outer := rl.NewRectangle(20, 20, float32(ui.width-40), float32(ui.height-40))
	left := rl.NewRectangle(outer.X, outer.Y, outer.Width*0.25, outer.Height)
	mid := rl.NewRectangle(left.X+left.Width+10, outer.Y, outer.Width*0.34, outer.Height)
	right := rl.NewRectangle(mid.X+mid.Width+10, outer.Y, outer.X+outer.Width-(mid.X+mid.Width+10), outer.Height)
	drawPanel(left, "Categories")
	drawPanel(mid, "Items")
	drawPanel(right, "Selection")

	selected := map[game.KitItem]bool{}
	selectedCount := 0
	limit := 0
	subtitle := "Issued kit selection"
	if ui.kit.Target == kitTargetPersonal && len(ui.pcfg.Players) > 0 {
		player := ui.pcfg.Players[ui.pcfg.PlayerIndex]
		for _, item := range player.Kit {
			selected[item] = true
		}
		selectedCount = len(player.Kit)
		limit = maxInt(1, player.KitLimit)
		subtitle = fmt.Sprintf("Player %d/%d %s", ui.pcfg.PlayerIndex+1, len(ui.pcfg.Players), player.Name)
	} else {
		for _, item := range ui.setup.IssuedKit {
			selected[item] = true
		}
		selectedCount = len(ui.setup.IssuedKit)
		subtitle = fmt.Sprintf("Issued Kit (%s)", modeLabel(ui.selectedMode()))
	}
	drawText(subtitle, int32(outer.X)+14, int32(outer.Y)-2, 18, colorDim)

	cy := int32(left.Y) + 48
	for i, category := range categories {
		if cy > int32(left.Y+left.Height)-34 {
			break
		}
		if i == ui.kit.CategoryIdx {
			rl.DrawRectangle(int32(left.X)+10, cy-4, int32(left.Width)-20, 28, rl.Fade(colorAccent, 0.2))
		}
		label := fmt.Sprintf("%s (%d)", category.Label, len(category.Items))
		clr := colorText
		if i == ui.kit.CategoryIdx && ui.kit.Focus == kitFocusCategories {
			clr = colorAccent
		}
		drawText(label, int32(left.X)+16, cy, 18, clr)
		cy += 30
	}

	current := categories[ui.kit.CategoryIdx]
	iy := int32(mid.Y) + 48
	for i, item := range current.Items {
		if iy > int32(mid.Y+mid.Height)-34 {
			break
		}
		if i == ui.kit.ItemIdx {
			rl.DrawRectangle(int32(mid.X)+10, iy-4, int32(mid.Width)-20, 28, rl.Fade(colorAccent, 0.2))
		}
		prefix := "[ ] "
		if selected[item] {
			prefix = "[*] "
		}
		clr := colorText
		if i == ui.kit.ItemIdx && ui.kit.Focus == kitFocusItems {
			clr = colorAccent
		}
		drawText(prefix+string(item), int32(mid.X)+16, iy, 18, clr)
		iy += 30
	}

	selectedLines := make([]string, 0, selectedCount+8)
	selectedLines = append(selectedLines, fmt.Sprintf("Selected: %d", selectedCount))
	if limit > 0 {
		selectedLines[0] = fmt.Sprintf("Selected: %d/%d", selectedCount, limit)
	}
	selectedLines = append(selectedLines, "")
	selectedLines = append(selectedLines, "Selected Items:")
	if selectedCount == 0 {
		selectedLines = append(selectedLines, "- none")
	} else {
		for _, category := range categories {
			for _, item := range category.Items {
				if selected[item] {
					selectedLines = append(selectedLines, "- "+string(item))
				}
			}
		}
	}
	selectedLines = append(selectedLines, "", "Focused Category:")
	selectedLines = append(selectedLines, current.Label)
	selectedLines = append(selectedLines, "", "Focused Item:")
	if len(current.Items) > 0 {
		activeItem := current.Items[ui.kit.ItemIdx]
		selectedLines = append(selectedLines, string(activeItem))
		selectedLines = append(selectedLines, kitItemFlavorText(activeItem))
	}
	if ui.kit.Target == kitTargetIssued {
		selectedLines = append(selectedLines, "", "Shift+R resets to recommendation")
	} else {
		selectedLines = append(selectedLines, "", "Shift+R resets personal kit")
	}
	drawLines(right, 44, 18, selectedLines, colorText)

	help := "Up/Down move  Left/Right pane  Enter select/toggle  Space toggle  Shift+R reset  Esc back"
	drawText(help, int32(outer.X)+14, int32(outer.Y+outer.Height)-30, 17, colorDim)
	if strings.TrimSpace(ui.status) != "" {
		drawText(ui.status, int32(outer.X)+14, int32(outer.Y+outer.Height)-52, 17, colorWarn)
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

func kitItemFlavorText(item game.KitItem) string {
	name := strings.ToLower(string(item))
	switch {
	case strings.Contains(name, "knife"), strings.Contains(name, "hatchet"), strings.Contains(name, "saw"), strings.Contains(name, "machete"):
		return "Tool item for shelter build and wood prep."
	case strings.Contains(name, "water"), strings.Contains(name, "canteen"), strings.Contains(name, "purification"):
		return "Water item for hydration and treatment."
	case strings.Contains(name, "ferro"), strings.Contains(name, "fire"), strings.Contains(name, "magnifying"):
		return "Fire item for ignition reliability."
	case strings.Contains(name, "tarp"), strings.Contains(name, "blanket"), strings.Contains(name, "sleep"), strings.Contains(name, "thermal"):
		return "Shelter item for exposure management."
	case strings.Contains(name, "fishing"), strings.Contains(name, "snare"), strings.Contains(name, "bow"):
		return "Food item supporting catch strategy."
	case strings.Contains(name, "first aid"), strings.Contains(name, "insect"), strings.Contains(name, "net"):
		return "Protection item reducing risk load."
	default:
		return "General survival item."
	}
}

type locationBiomeOption struct {
	Location string
	Biomes   []string
}

func scenarioLocationBiomeOptions() []locationBiomeOption {
	return []locationBiomeOption{
		{Location: "North America", Biomes: []string{"temperate_rainforest", "boreal_coast", "subarctic_lake", "mountain_forest", "arctic_delta"}},
		{Location: "South America", Biomes: []string{"tropical_jungle", "wetlands", "river_delta_wetlands", "montane_steppe", "coast"}},
		{Location: "Africa", Biomes: []string{"savanna", "desert", "swamp", "badlands_jungle_edge", "coast"}},
		{Location: "Asia-Pacific", Biomes: []string{"tropical_island", "tropical_jungle", "subarctic", "cold_mountain", "coast"}},
		{Location: "Wilderness", Biomes: []string{"temperate_forest", "mountain", "coast", "jungle", "desert"}},
	}
}

func biomeOptionsForLocation(location string) []string {
	options := scenarioLocationBiomeOptions()
	for _, option := range options {
		if strings.EqualFold(strings.TrimSpace(option.Location), strings.TrimSpace(location)) {
			return append([]string(nil), option.Biomes...)
		}
	}
	return append([]string(nil), options[len(options)-1].Biomes...)
}

func currentLocationBiomeIndex(location string) int {
	options := scenarioLocationBiomeOptions()
	for i, option := range options {
		if strings.EqualFold(strings.TrimSpace(option.Location), strings.TrimSpace(location)) {
			return i
		}
	}
	return len(options) - 1
}

func (ui *gameUI) openScenarioBuilder() {
	ui.sb = scenarioBuilderState{
		Cursor:          0,
		ModeIndex:       ui.setup.ModeIndex,
		ListCursor:      0,
		PickingScenario: false,
		EditingRow:      -1,
		Scenario:        newScenarioTemplate(modeOptions()[clampInt(ui.setup.ModeIndex, 0, len(modeOptions())-1)]),
		SourceID:        "",
		SourceName:      "",
		SourceCustom:    false,
	}
	if strings.TrimSpace(ui.sb.Scenario.Location) == "" {
		ui.sb.Scenario.Location = "Wilderness"
	}
	biomes := biomeOptionsForLocation(ui.sb.Scenario.Location)
	if len(biomes) > 0 {
		ui.sb.Scenario.Biome = biomes[0]
	}
	ui.ensureScenarioSeasonSet()
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
	if ui.sb.PickingScenario {
		mode := modeOptions()[clampInt(ui.sb.ModeIndex, 0, len(modeOptions())-1)]
		list := ui.scenarioBuilderEntriesForMode(mode)
		if len(list) == 0 {
			ui.sb.PickingScenario = false
			ui.sb.Status = "No scenarios available for this mode."
			return
		}
		if rl.IsKeyPressed(rl.KeyEscape) || rl.IsKeyPressed(rl.KeyEnter) {
			ui.sb.PickingScenario = false
			return
		}
		if rl.IsKeyPressed(rl.KeyDown) {
			ui.sb.ListCursor = wrapIndex(ui.sb.ListCursor+1, len(list))
			ui.loadSelectedScenario()
		}
		if rl.IsKeyPressed(rl.KeyUp) {
			ui.sb.ListCursor = wrapIndex(ui.sb.ListCursor-1, len(list))
			ui.loadSelectedScenario()
		}
		return
	}

	if rl.IsKeyPressed(rl.KeyEscape) {
		ui.enterMenu()
		return
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		ui.sb.Cursor = wrapIndex(ui.sb.Cursor+1, 19)
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		ui.sb.Cursor = wrapIndex(ui.sb.Cursor-1, 19)
	}
	if rl.IsKeyPressed(rl.KeyLeft) {
		ui.adjustScenarioBuilder(-1)
	}
	if rl.IsKeyPressed(rl.KeyRight) {
		ui.adjustScenarioBuilder(1)
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		switch ui.sb.Cursor {
		case 3, 6, 7, 8:
			ui.startScenarioEdit(ui.sb.Cursor)
		case 12:
			ui.openPhaseEditor()
		case 13:
			ui.setup.ModeIndex = ui.sb.ModeIndex
			ui.ensureSetupPlayers()
			ui.openStatsBuilder(screenScenarioBuilder)
		case 14:
			ui.setup.ModeIndex = ui.sb.ModeIndex
			ui.ensureSetupPlayers()
			ui.preparePlayerConfig()
			ui.pcfg.ReturnTo = screenScenarioBuilder
			ui.screen = screenPlayerConfig
		case 15:
			ui.sb.PickingScenario = true
			ui.loadSelectedScenario()
		case 16:
			ui.saveScenarioFromBuilder()
		case 17:
			ui.deleteSelectedCustomScenario()
		case 18:
			ui.enterMenu()
		}
	}
}

func (ui *gameUI) adjustScenarioBuilder(delta int) {
	modes := modeOptions()
	switch ui.sb.Cursor {
	case 0:
		ui.sb.ModeIndex = wrapIndex(ui.sb.ModeIndex+delta, len(modes))
		ui.setup.ModeIndex = ui.sb.ModeIndex
		ui.sb.Scenario = newScenarioTemplate(modes[ui.sb.ModeIndex])
		if strings.TrimSpace(ui.sb.Scenario.Location) == "" {
			ui.sb.Scenario.Location = "Wilderness"
		}
		ui.sb.SourceID = ""
		ui.sb.SourceName = ""
		ui.sb.SourceCustom = false
		ui.setup.PlayerCount = defaultPlayerCountForMode(modes[ui.sb.ModeIndex])
		ui.setup.IssuedCustom = false
		ui.sb.ListCursor = 0
	case 1:
		ui.setup.PlayerCount = clampInt(ui.setup.PlayerCount+delta, 1, 8)
	case 4:
		locations := scenarioLocationBiomeOptions()
		locIdx := currentLocationBiomeIndex(ui.sb.Scenario.Location)
		locIdx = wrapIndex(locIdx+delta, len(locations))
		ui.sb.Scenario.Location = locations[locIdx].Location
		biomes := biomeOptionsForLocation(ui.sb.Scenario.Location)
		if len(biomes) > 0 {
			ui.sb.Scenario.Biome = biomes[0]
			ui.sb.Scenario.Wildlife = game.WildlifeForBiome(ui.sb.Scenario.Biome)
		}
	case 5:
		biomes := biomeOptionsForLocation(ui.sb.Scenario.Location)
		if len(biomes) > 0 {
			current := 0
			for i, biome := range biomes {
				if strings.EqualFold(biome, ui.sb.Scenario.Biome) {
					current = i
					break
				}
			}
			current = wrapIndex(current+delta, len(biomes))
			ui.sb.Scenario.Biome = biomes[current]
			ui.sb.Scenario.Wildlife = game.WildlifeForBiome(ui.sb.Scenario.Biome)
		}
	case 9:
		ui.sb.Scenario.DefaultDays = clampInt(ui.sb.Scenario.DefaultDays+delta*3, 1, 365)
	case 10:
		ui.sb.Scenario.MapWidthCells += delta * 2
		ui.sb.Scenario.MapWidthCells, ui.sb.Scenario.MapHeightCells = clampScenarioMapSize(modeOptions()[clampInt(ui.sb.ModeIndex, 0, len(modeOptions())-1)], ui.sb.Scenario.MapWidthCells, ui.sb.Scenario.MapHeightCells)
	case 11:
		ui.sb.Scenario.MapHeightCells += delta * 2
		ui.sb.Scenario.MapWidthCells, ui.sb.Scenario.MapHeightCells = clampScenarioMapSize(modeOptions()[clampInt(ui.sb.ModeIndex, 0, len(modeOptions())-1)], ui.sb.Scenario.MapWidthCells, ui.sb.Scenario.MapHeightCells)
	}
	ui.setup.ModeIndex = ui.sb.ModeIndex
	ui.ensureScenarioSeasonSet()
	ui.ensureSetupPlayers()
}

func (ui *gameUI) drawScenarioBuilder() {
	left := rl.NewRectangle(20, 20, float32(ui.width)*0.56, float32(ui.height-40))
	right := rl.NewRectangle(left.X+left.Width+16, 20, float32(ui.width)-left.Width-56, float32(ui.height-40))
	drawPanel(left, "Scenario Builder")
	if ui.sb.PickingScenario {
		drawPanel(right, "Scenario Browser (ACTIVE)")
	} else {
		drawPanel(right, "Scenarios")
	}

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
		{"Player Count", fmt.Sprintf("%d", ui.setup.PlayerCount)},
		{"Loaded", loadedLabel},
		{"Name", ui.sb.Scenario.Name},
		{"Location", ui.sb.Scenario.Location},
		{"Biome", ui.sb.Scenario.Biome},
		{"Description", ui.sb.Scenario.Description},
		{"Daunting", ui.sb.Scenario.Daunting},
		{"Motivation", ui.sb.Scenario.Motivation},
		{"Default Days", fmt.Sprintf("%d", ui.sb.Scenario.DefaultDays)},
		{"Map Width (cells)", fmt.Sprintf("%d", ui.sb.Scenario.MapWidthCells)},
		{"Map Height (cells)", fmt.Sprintf("%d", ui.sb.Scenario.MapHeightCells)},
		{"Phase Builder", fmt.Sprintf("%d phase(s)", ui.scenarioPhaseCount())},
		{"Player Stats Builder", "Enter"},
		{"Player Editor", "Enter"},
		{"Load Scenario", "Enter"},
		{"Save Scenario", "Enter"},
		{"Delete Selected", "Enter"},
		{"Back", "Enter"},
	}

	labelX := int32(left.X) + 16
	valueX := int32(left.X) + int32(left.Width*0.58)
	maxValueChars := maxInt(6, int((left.Width*0.40)/8))
	y := int32(left.Y) + 54
	for i, row := range rows {
		if i == ui.sb.Cursor {
			rl.DrawRectangle(int32(left.X)+10, y-5, int32(left.Width)-20, 30, rl.Fade(colorAccent, 0.2))
		}
		value := row.value
		value = truncateForUI(value, maxValueChars)
		drawText(row.label, labelX, y, 19, colorText)
		drawText(value, valueX, y, 19, colorAccent)
		y += 34
	}
	drawText("Left/Right change mode/player count/location/biome/days/map size", int32(left.X)+14, int32(left.Y+left.Height)-64, 17, colorDim)
	drawText("Load Scenario opens right pane browse mode (auto-load while scrolling).", int32(left.X)+14, int32(left.Y+left.Height)-40, 17, colorDim)

	list := ui.scenarioBuilderEntriesForMode(mode)
	ly := int32(right.Y) + 52
	if len(list) == 0 {
		drawText("No scenarios for this mode yet.", int32(right.X)+14, ly, 20, colorWarn)
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
			line := fmt.Sprintf("%s [%s]  (%s, %s, %dd, %dx%d)", s.Name, source, s.Location, s.Biome, s.DefaultDays, s.MapWidthCells, s.MapHeightCells)
			clr := colorText
			if ui.sb.PickingScenario && i == ui.sb.ListCursor {
				clr = colorAccent
			}
			drawText(truncateForUI(line, int((right.Width-36)/8)), int32(right.X)+16, ly, 19, clr)
			ly += 34
		}
		sel := list[clampInt(ui.sb.ListCursor, 0, len(list)-1)].Scenario
		land := animalsPreviewForBiome(sel.Biome, game.AnimalDomainLand)
		fish := animalsPreviewForBiome(sel.Biome, game.AnimalDomainWater)
		birds := animalsPreviewForBiome(sel.Biome, game.AnimalDomainAir)
		drawWrappedText("Selected Name: "+sel.Name, right, int32(right.Height)-258, 19, colorAccent)
		drawWrappedText("Location: "+safeText(sel.Location), right, int32(right.Height)-230, 19, colorText)
		drawWrappedText("Biome: "+sel.Biome, right, int32(right.Height)-202, 19, colorText)
		drawWrappedText("Animals: "+land, right, int32(right.Height)-174, 19, colorDim)
		drawWrappedText("Fish: "+fish, right, int32(right.Height)-146, 19, colorDim)
		drawWrappedText("Birds: "+birds, right, int32(right.Height)-118, 19, colorDim)
		drawWrappedText("Description: "+safeText(sel.Description), right, int32(right.Height)-90, 18, colorDim)
	}
	if ui.sb.PickingScenario {
		drawText("Right pane browse active: Up/Down scroll, Enter confirm, Esc cancel", int32(right.X)+14, int32(right.Y+right.Height)-20, 17, colorAccent)
	}

	if ui.sb.EditingRow >= 0 {
		r := rl.NewRectangle(left.X+18, left.Y+left.Height-138, left.Width-36, 110)
		rl.DrawRectangleRounded(r, 0.16, 8, rl.Fade(colorPanel, 0.96))
		rl.DrawRectangleRoundedLinesEx(r, 0.16, 8, 2, colorAccent)
		drawText("Editing (Enter apply, Esc cancel)", int32(r.X)+12, int32(r.Y)+10, 18, colorAccent)
		drawWrappedText(ui.sb.EditBuffer+"_", r, 40, 21, colorText)
	}
	if strings.TrimSpace(ui.sb.Status) != "" {
		drawText(ui.sb.Status, int32(left.X)+14, int32(left.Y+left.Height)-20, 17, colorWarn)
	}
}

func (ui *gameUI) startScenarioEdit(row int) {
	ui.sb.EditingRow = row
	switch row {
	case 3:
		ui.sb.EditBuffer = ui.sb.Scenario.Name
	case 6:
		ui.sb.EditBuffer = ui.sb.Scenario.Description
	case 7:
		ui.sb.EditBuffer = ui.sb.Scenario.Daunting
	case 8:
		ui.sb.EditBuffer = ui.sb.Scenario.Motivation
	}
}

func (ui *gameUI) commitScenarioEdit() {
	value := strings.TrimSpace(ui.sb.EditBuffer)
	switch ui.sb.EditingRow {
	case 3:
		if value != "" {
			ui.sb.Scenario.Name = value
		}
	case 6:
		ui.sb.Scenario.Description = value
	case 7:
		ui.sb.Scenario.Daunting = value
	case 8:
		ui.sb.Scenario.Motivation = value
	}
	ui.sb.EditingRow = -1
}

func animalsPreviewForBiome(biome string, domain game.AnimalDomain) string {
	specs := game.AnimalsForBiome(biome, domain)
	if len(specs) == 0 {
		return "none"
	}
	maxItems := 5
	parts := make([]string, 0, min(len(specs), maxItems))
	for i := 0; i < len(specs) && i < maxItems; i++ {
		parts = append(parts, specs[i].Name)
	}
	if len(specs) > maxItems {
		parts = append(parts, fmt.Sprintf("+%d more", len(specs)-maxItems))
	}
	return strings.Join(parts, ", ")
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
	if strings.TrimSpace(ui.sb.Scenario.Location) == "" {
		ui.sb.Scenario.Location = "Wilderness"
	}
	biomes := biomeOptionsForLocation(ui.sb.Scenario.Location)
	if len(biomes) > 0 {
		found := false
		for _, biome := range biomes {
			if strings.EqualFold(strings.TrimSpace(ui.sb.Scenario.Biome), strings.TrimSpace(biome)) {
				found = true
				break
			}
		}
		if !found {
			ui.sb.Scenario.Biome = biomes[0]
		}
	}
	ui.sb.SourceID = sel.Scenario.ID
	ui.sb.SourceName = sel.Scenario.Name
	ui.sb.SourceCustom = sel.Custom
	ui.sb.Scenario.MapWidthCells, ui.sb.Scenario.MapHeightCells = clampScenarioMapSize(mode, ui.sb.Scenario.MapWidthCells, ui.sb.Scenario.MapHeightCells)
	ui.ensureScenarioSeasonSet()
	if sel.Custom {
		ui.sb.Status = "Loaded custom scenario into editor"
	} else {
		ui.sb.Status = "Loaded built-in scenario template. Save as a new name."
	}
}

func (ui *gameUI) saveScenarioFromBuilder() {
	mode := modeOptions()[clampInt(ui.sb.ModeIndex, 0, len(modeOptions())-1)]
	scenario := ui.sb.Scenario
	ui.ensureScenarioSeasonSet()
	scenario = ui.sb.Scenario
	if strings.TrimSpace(scenario.Name) == "" {
		ui.sb.Status = "Name is required"
		return
	}
	if strings.TrimSpace(scenario.Location) == "" {
		scenario.Location = "Wilderness"
	}
	if strings.TrimSpace(scenario.Biome) == "" {
		biomes := biomeOptionsForLocation(scenario.Location)
		if len(biomes) > 0 {
			scenario.Biome = biomes[0]
		} else {
			scenario.Biome = "temperate_forest"
		}
	}
	if scenario.DefaultDays <= 0 {
		scenario.DefaultDays = 30
	}
	scenario.MapWidthCells, scenario.MapHeightCells = clampScenarioMapSize(mode, scenario.MapWidthCells, scenario.MapHeightCells)
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

func clampScenarioMapSize(mode game.GameMode, width, height int) (int, int) {
	defaultW, defaultH := 72, 72
	switch mode {
	case game.ModeAlone:
		defaultW, defaultH = 36, 36
	case game.ModeNakedAndAfraid:
		defaultW, defaultH = 100, 100
	case game.ModeNakedAndAfraidXL:
		defaultW, defaultH = 125, 125
	}
	if width <= 0 {
		width = defaultW
	}
	if height <= 0 {
		height = defaultH
	}
	switch mode {
	case game.ModeAlone:
		width = clampInt(width, 28, 46)
		height = clampInt(height, 28, 46)
	case game.ModeNakedAndAfraid:
		width = clampInt(width, 88, 125)
		height = clampInt(height, 88, 125)
	case game.ModeNakedAndAfraidXL:
		width = clampInt(width, 100, 150)
		height = clampInt(height, 100, 150)
	default:
		width = clampInt(width, 50, 140)
		height = clampInt(height, 50, 140)
	}
	return width, height
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

const maxScenarioPhases = 24

type phaseRowKind int

const (
	phaseRowNewSeason phaseRowKind = iota
	phaseRowNewDays
	phaseRowAddPhase
	phaseRowRemoveLast
	phaseRowBack
)

type phaseEditorRow struct {
	Label  string
	Value  string
	Kind   phaseRowKind
	Active bool
}

func builderSeasonOptions() []game.SeasonID {
	return []game.SeasonID{
		game.SeasonAutumn,
		game.SeasonWinter,
		game.SeasonWet,
		game.SeasonDry,
	}
}

func builderSeasonLabel(id game.SeasonID) string {
	switch id {
	case game.SeasonAutumn:
		return "Autumn"
	case game.SeasonWinter:
		return "Winter"
	case game.SeasonWet:
		return "Wet"
	case game.SeasonDry:
		return "Dry"
	default:
		return string(id)
	}
}

func (ui *gameUI) scenarioDefaultSeasonSetIndex() int {
	if len(ui.sb.Scenario.SeasonSets) == 0 {
		return -1
	}
	target := ui.sb.Scenario.DefaultSeasonSetID
	if target == "" {
		ui.sb.Scenario.DefaultSeasonSetID = ui.sb.Scenario.SeasonSets[0].ID
		return 0
	}
	for i := range ui.sb.Scenario.SeasonSets {
		if ui.sb.Scenario.SeasonSets[i].ID == target {
			return i
		}
	}
	ui.sb.Scenario.DefaultSeasonSetID = ui.sb.Scenario.SeasonSets[0].ID
	return 0
}

func (ui *gameUI) ensureScenarioSeasonSet() {
	mode := modeOptions()[clampInt(ui.sb.ModeIndex, 0, len(modeOptions())-1)]
	fallback := defaultSeasonSetForMode(mode)

	if len(ui.sb.Scenario.SeasonSets) == 0 {
		ui.sb.Scenario.SeasonSets = []game.SeasonSet{fallback}
		ui.sb.Scenario.DefaultSeasonSetID = fallback.ID
	}

	for i := range ui.sb.Scenario.SeasonSets {
		set := &ui.sb.Scenario.SeasonSets[i]
		if strings.TrimSpace(string(set.ID)) == "" {
			set.ID = game.SeasonSetID(fmt.Sprintf("custom_profile_%d", i+1))
		}
		if len(set.Phases) == 0 {
			set.Phases = append([]game.SeasonPhase(nil), fallback.Phases...)
		}
		for j := range set.Phases {
			if strings.TrimSpace(string(set.Phases[j].Season)) == "" {
				set.Phases[j].Season = game.SeasonAutumn
			}
			if set.Phases[j].Days < 0 {
				set.Phases[j].Days = 0
			}
		}
	}

	if ui.sb.Scenario.DefaultSeasonSetID == "" {
		ui.sb.Scenario.DefaultSeasonSetID = ui.sb.Scenario.SeasonSets[0].ID
		return
	}
	if ui.scenarioDefaultSeasonSetIndex() < 0 {
		ui.sb.Scenario.DefaultSeasonSetID = ui.sb.Scenario.SeasonSets[0].ID
	}
}

func (ui *gameUI) scenarioPhaseCount() int {
	ui.ensureScenarioSeasonSet()
	idx := ui.scenarioDefaultSeasonSetIndex()
	if idx < 0 {
		return 0
	}
	return len(ui.sb.Scenario.SeasonSets[idx].Phases)
}

func (ui *gameUI) openPhaseEditor() {
	ui.ensureScenarioSeasonSet()
	ui.phase = phaseEditorState{
		Cursor:       0,
		Adding:       false,
		NewSeasonIdx: 0,
		NewDays:      "7",
	}
	ui.screen = screenPhaseEditor
}

func (ui *gameUI) phaseEditorRows() []phaseEditorRow {
	seasons := builderSeasonOptions()
	if len(seasons) == 0 {
		return nil
	}
	ui.phase.NewSeasonIdx = clampInt(ui.phase.NewSeasonIdx, 0, len(seasons)-1)
	rows := make([]phaseEditorRow, 0, 5)
	if ui.phase.Adding {
		rows = append(rows,
			phaseEditorRow{
				Label:  "New Phase Season",
				Value:  builderSeasonLabel(seasons[ui.phase.NewSeasonIdx]),
				Kind:   phaseRowNewSeason,
				Active: true,
			},
			phaseEditorRow{
				Label:  "New Phase Days",
				Value:  ui.phase.NewDays,
				Kind:   phaseRowNewDays,
				Active: true,
			},
		)
	}

	addLabel := "Add Phase"
	if ui.phase.Adding {
		addLabel = "Confirm Add Phase"
	}
	rows = append(rows,
		phaseEditorRow{Label: addLabel, Kind: phaseRowAddPhase, Active: true},
		phaseEditorRow{Label: "Remove Last Phase", Kind: phaseRowRemoveLast, Active: ui.scenarioPhaseCount() > 1},
		phaseEditorRow{Label: "Back", Kind: phaseRowBack, Active: true},
	)
	return rows
}

func (ui *gameUI) updatePhaseEditor() {
	ui.ensureScenarioSeasonSet()
	rows := ui.phaseEditorRows()
	if len(rows) == 0 {
		ui.screen = screenScenarioBuilder
		return
	}
	ui.phase.Cursor = clampInt(ui.phase.Cursor, 0, len(rows)-1)
	active := rows[ui.phase.Cursor]

	if active.Kind == phaseRowNewDays {
		for ch := rl.GetCharPressed(); ch > 0; ch = rl.GetCharPressed() {
			if ch >= '0' && ch <= '9' {
				ui.phase.NewDays += string(rune(ch))
			}
		}
		if rl.IsKeyPressed(rl.KeyBackspace) && len(ui.phase.NewDays) > 0 {
			ui.phase.NewDays = ui.phase.NewDays[:len(ui.phase.NewDays)-1]
		}
	}

	if rl.IsKeyPressed(rl.KeyEscape) {
		ui.screen = screenScenarioBuilder
		return
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		ui.phase.Cursor = wrapIndex(ui.phase.Cursor+1, len(rows))
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		ui.phase.Cursor = wrapIndex(ui.phase.Cursor-1, len(rows))
	}

	if rl.IsKeyPressed(rl.KeyLeft) {
		active = rows[ui.phase.Cursor]
		if active.Kind == phaseRowNewSeason {
			seasons := builderSeasonOptions()
			ui.phase.NewSeasonIdx = wrapIndex(ui.phase.NewSeasonIdx-1, len(seasons))
		}
	}
	if rl.IsKeyPressed(rl.KeyRight) {
		active = rows[ui.phase.Cursor]
		if active.Kind == phaseRowNewSeason {
			seasons := builderSeasonOptions()
			ui.phase.NewSeasonIdx = wrapIndex(ui.phase.NewSeasonIdx+1, len(seasons))
		}
	}

	if !rl.IsKeyPressed(rl.KeyEnter) {
		return
	}

	rows = ui.phaseEditorRows()
	ui.phase.Cursor = clampInt(ui.phase.Cursor, 0, len(rows)-1)
	active = rows[ui.phase.Cursor]
	if !active.Active {
		return
	}

	setIdx := ui.scenarioDefaultSeasonSetIndex()
	if setIdx < 0 {
		ui.sb.Status = "No season set available."
		ui.screen = screenScenarioBuilder
		return
	}
	phases := &ui.sb.Scenario.SeasonSets[setIdx].Phases

	switch active.Kind {
	case phaseRowNewSeason:
		seasons := builderSeasonOptions()
		ui.phase.NewSeasonIdx = wrapIndex(ui.phase.NewSeasonIdx+1, len(seasons))
	case phaseRowAddPhase:
		if !ui.phase.Adding {
			ui.phase.Adding = true
			if strings.TrimSpace(ui.phase.NewDays) == "" {
				ui.phase.NewDays = "7"
			}
			ui.phase.Cursor = 0
			return
		}
		if len(*phases) >= maxScenarioPhases {
			ui.sb.Status = fmt.Sprintf("Maximum phases reached (%d).", maxScenarioPhases)
			return
		}
		if len(*phases) > 0 && (*phases)[len(*phases)-1].Days == 0 {
			ui.sb.Status = "Current final phase has 0 days. Set it before adding another."
			return
		}
		days, err := strconv.Atoi(strings.TrimSpace(ui.phase.NewDays))
		if err != nil || days < 0 {
			ui.sb.Status = "Days must be a number >= 0."
			return
		}
		seasons := builderSeasonOptions()
		*phases = append(*phases, game.SeasonPhase{
			Season: seasons[ui.phase.NewSeasonIdx],
			Days:   days,
		})
		ui.phase.Adding = false
		ui.phase.NewDays = "7"
		ui.phase.Cursor = 0
		ui.sb.Status = "Added phase."
	case phaseRowRemoveLast:
		if len(*phases) <= 1 {
			ui.sb.Status = "At least one phase is required."
			return
		}
		*phases = append([]game.SeasonPhase(nil), (*phases)[:len(*phases)-1]...)
		ui.sb.Status = "Removed last phase."
	case phaseRowBack:
		ui.screen = screenScenarioBuilder
	}
}

func (ui *gameUI) drawPhaseEditor() {
	ui.ensureScenarioSeasonSet()
	rows := ui.phaseEditorRows()
	if len(rows) == 0 {
		panel := rl.NewRectangle(20, 20, float32(ui.width-40), float32(ui.height-40))
		drawPanel(panel, "Phase Builder")
		drawWrappedText("No phase rows available.", panel, 60, 22, colorWarn)
		return
	}
	ui.phase.Cursor = clampInt(ui.phase.Cursor, 0, len(rows)-1)

	left := rl.NewRectangle(20, 20, float32(ui.width)*0.34, float32(ui.height-40))
	right := rl.NewRectangle(left.X+left.Width+16, 20, float32(ui.width)-left.Width-56, float32(ui.height-40))
	drawPanel(left, "Phase Builder")
	drawPanel(right, "Phase Timeline")

	y := int32(left.Y) + 56
	for i, row := range rows {
		if i == ui.phase.Cursor {
			rl.DrawRectangle(int32(left.X)+10, y-6, int32(left.Width)-20, 32, rl.Fade(colorAccent, 0.2))
		}
		clr := colorText
		if !row.Active {
			clr = colorDim
		}
		drawText(row.Label, int32(left.X)+18, y, 20, clr)
		if strings.TrimSpace(row.Value) != "" {
			drawText(row.Value, int32(left.X)+220, y, 20, colorAccent)
		}
		y += 40
	}
	drawText("Up/Down move  Left/Right season  Enter select  Esc back", int32(left.X)+14, int32(left.Y+left.Height)-30, 17, colorDim)

	setIdx := ui.scenarioDefaultSeasonSetIndex()
	lines := []string{
		fmt.Sprintf("Scenario: %s", ui.sb.Scenario.Name),
		fmt.Sprintf("Season Profile: %s", ui.sb.Scenario.DefaultSeasonSetID),
		"",
		"Phases:",
	}
	if setIdx >= 0 {
		for i, phase := range ui.sb.Scenario.SeasonSets[setIdx].Phases {
			lines = append(lines, fmt.Sprintf("%d. %s (%d day(s))", i+1, builderSeasonLabel(phase.Season), phase.Days))
		}
	}
	lines = append(lines,
		"",
		"Notes:",
		"When adding a phase, season and days appear on the left.",
		"Set days to 0 only for the final phase.",
	)
	drawLines(right, 44, 20, lines, colorText)

	if strings.TrimSpace(ui.sb.Status) != "" {
		drawText(ui.sb.Status, int32(left.X)+14, int32(left.Y+left.Height)-52, 17, colorWarn)
	}
}
