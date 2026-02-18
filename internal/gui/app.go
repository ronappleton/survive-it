package gui

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/appengine-ltd/survive-it/internal/game"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type AppConfig struct {
	Version   string
	Commit    string
	BuildDate string
	NoUpdate  bool
}

type App struct {
	cfg AppConfig
}

func NewApp(cfg AppConfig) *App {
	return &App{cfg: cfg}
}

type screen int

const (
	screenMenu screen = iota
	screenSetup
	screenScenarioPicker
	screenPlayerConfig
	screenKitPicker
	screenScenarioBuilder
	screenLoad
	screenRun
)

type menuAction int

const (
	actionStart menuAction = iota
	actionLoad
	actionScenarioBuilder
	actionQuit
)

type menuItem struct {
	Label  string
	Action menuAction
}

type setupState struct {
	Cursor        int
	ModeIndex     int
	ScenarioIndex int
	PlayerCount   int
	RunDays       int
}

type playerConfigState struct {
	Cursor      int
	PlayerIndex int
	EditingName bool
	NameBuffer  string
	Players     []game.PlayerConfig
}

type scenarioPickerState struct {
	Cursor int
}

type loadState struct {
	Cursor      int
	ReturnToRun bool
	Entries     []saveEntry
}

type saveEntry struct {
	Path  string
	Saved savedRun
}

type savedRun struct {
	FormatVersion int           `json:"format_version"`
	SavedAt       time.Time     `json:"saved_at"`
	Run           game.RunState `json:"run"`
}

type gameUI struct {
	cfg AppConfig

	width  int32
	height int32
	quit   bool

	screen screen

	menuCursor int

	setup           setupState
	pick            scenarioPickerState
	pcfg            playerConfigState
	kit             kitPickerState
	sb              scenarioBuilderState
	load            loadState
	customScenarios []game.Scenario

	run         *game.RunState
	runMessages []string
	runInput    string
	status      string
	runFocus    int

	lastTick     time.Time
	runPlayedFor time.Duration
	autoDayHours int
}

var (
	colorBG     = rl.NewColor(8, 12, 18, 255)
	colorPanel  = rl.NewColor(14, 24, 35, 255)
	colorBorder = rl.NewColor(25, 200, 120, 255)
	colorText   = rl.NewColor(175, 245, 195, 255)
	colorDim    = rl.NewColor(108, 165, 124, 255)
	colorAccent = rl.NewColor(60, 255, 145, 255)
	colorWarn   = rl.NewColor(255, 198, 96, 255)
)

func (a *App) Run() error {
	ui := newGameUI(a.cfg)
	return ui.Run()
}

func newGameUI(cfg AppConfig) *gameUI {
	ui := &gameUI{
		cfg:          cfg,
		width:        1024,
		height:       768,
		screen:       screenMenu,
		autoDayHours: 2,
		setup: setupState{
			ModeIndex:   0,
			PlayerCount: 1,
			RunDays:     30,
		},
	}
	custom, _ := loadCustomScenarios(defaultCustomScenariosFile)
	ui.customScenarios = custom
	game.SetExternalScenarios(custom)
	ui.syncScenarioSelection()
	ui.lastTick = time.Now()
	return ui
}

func (ui *gameUI) Run() error {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(ui.width, ui.height, "survive-it")
	defer rl.CloseWindow()
	rl.SetExitKey(0)
	rl.SetTargetFPS(60)

	for !ui.quit && !rl.WindowShouldClose() {
		now := time.Now()
		delta := now.Sub(ui.lastTick)
		if delta < 0 {
			delta = 0
		}
		ui.lastTick = now

		ui.width = int32(rl.GetScreenWidth())
		ui.height = int32(rl.GetScreenHeight())

		ui.update(delta)

		rl.BeginDrawing()
		rl.ClearBackground(colorBG)
		ui.draw()
		rl.EndDrawing()
	}

	return nil
}

func (ui *gameUI) update(delta time.Duration) {
	switch ui.screen {
	case screenMenu:
		ui.updateMenu()
	case screenSetup:
		ui.updateSetup()
	case screenScenarioPicker:
		ui.updateScenarioPicker()
	case screenPlayerConfig:
		ui.updatePlayerConfig()
	case screenKitPicker:
		ui.updateKitPicker()
	case screenScenarioBuilder:
		ui.updateScenarioBuilder()
	case screenLoad:
		ui.updateLoad()
	case screenRun:
		ui.updateRun(delta)
	}
}

func (ui *gameUI) draw() {
	switch ui.screen {
	case screenMenu:
		ui.drawMenu()
	case screenSetup:
		ui.drawSetup()
	case screenScenarioPicker:
		ui.drawScenarioPicker()
	case screenPlayerConfig:
		ui.drawPlayerConfig()
	case screenKitPicker:
		ui.drawKitPicker()
	case screenScenarioBuilder:
		ui.drawScenarioBuilder()
	case screenLoad:
		ui.drawLoad()
	case screenRun:
		ui.drawRun()
	}
}

func (ui *gameUI) updateMenu() {
	items := ui.menuItems()
	if rl.IsKeyPressed(rl.KeyDown) {
		ui.menuCursor = wrapIndex(ui.menuCursor+1, len(items))
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		ui.menuCursor = wrapIndex(ui.menuCursor-1, len(items))
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		switch items[ui.menuCursor].Action {
		case actionStart:
			ui.setup = setupState{ModeIndex: ui.setup.ModeIndex, PlayerCount: defaultPlayerCountForMode(ui.selectedMode()), RunDays: 30}
			ui.syncScenarioSelection()
			ui.status = ""
			ui.screen = screenSetup
		case actionLoad:
			ui.openLoad(false)
		case actionScenarioBuilder:
			ui.openScenarioBuilder()
		case actionQuit:
			ui.quit = true
		}
	}
	if rl.IsKeyPressed(rl.KeyQ) {
		ui.quit = true
	}
}

func (ui *gameUI) drawMenu() {
	titleRect := rl.NewRectangle(20, 20, float32(ui.width-40), 120)
	drawPanel(titleRect, "SURVIVE IT")
	drawTextCentered("raylib-go Edition", titleRect, 34, 40, colorAccent)
	drawTextCentered(fmt.Sprintf("v%s (%s) %s", ui.cfg.Version, ui.cfg.Commit, ui.cfg.BuildDate), titleRect, 62, 18, colorDim)

	menuRect := rl.NewRectangle(float32(ui.width/2-230), 185, 460, 380)
	drawPanel(menuRect, "Main Menu")
	items := ui.menuItems()
	for i, item := range items {
		y := int32(menuRect.Y) + 70 + int32(i*72)
		r := rl.NewRectangle(menuRect.X+36, float32(y), menuRect.Width-72, 52)
		if i == ui.menuCursor {
			rl.DrawRectangleRounded(r, 0.3, 8, rl.Fade(colorAccent, 0.2))
			rl.DrawRectangleRoundedLinesEx(r, 0.3, 8, 2, colorAccent)
			rl.DrawText(item.Label, int32(r.X)+18, y+14, 28, colorAccent)
		} else {
			rl.DrawRectangleRounded(r, 0.3, 8, rl.Fade(colorPanel, 0.7))
			rl.DrawRectangleRoundedLinesEx(r, 0.3, 8, 1.5, colorBorder)
			rl.DrawText(item.Label, int32(r.X)+18, y+14, 28, colorText)
		}
	}

	hintRect := rl.NewRectangle(20, float32(ui.height-64), float32(ui.width-40), 40)
	drawTextCentered("Up/Down to move, Enter to select, Q to quit", hintRect, 8, 18, colorDim)
}

func (ui *gameUI) menuItems() []menuItem {
	return []menuItem{
		{Label: "Start Run", Action: actionStart},
		{Label: "Load Run", Action: actionLoad},
		{Label: "Scenario Builder", Action: actionScenarioBuilder},
		{Label: "Quit", Action: actionQuit},
	}
}

func (ui *gameUI) updateSetup() {
	if rl.IsKeyPressed(rl.KeyEscape) {
		ui.screen = screenMenu
		return
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		ui.setup.Cursor = wrapIndex(ui.setup.Cursor+1, 7)
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		ui.setup.Cursor = wrapIndex(ui.setup.Cursor-1, 7)
	}

	if rl.IsKeyPressed(rl.KeyLeft) {
		ui.adjustSetup(-1)
	}
	if rl.IsKeyPressed(rl.KeyRight) {
		ui.adjustSetup(1)
	}

	if rl.IsKeyPressed(rl.KeyEnter) {
		switch ui.setup.Cursor {
		case 1:
			ui.pick.Cursor = ui.setup.ScenarioIndex
			ui.screen = screenScenarioPicker
		case 4:
			ui.preparePlayerConfig()
			ui.screen = screenPlayerConfig
		case 5:
			ui.openScenarioBuilder()
		case 6:
			ui.screen = screenMenu
		}
	}
}

func (ui *gameUI) adjustSetup(delta int) {
	scenarios := ui.activeScenarios()
	switch ui.setup.Cursor {
	case 0:
		modes := modeOptions()
		ui.setup.ModeIndex = wrapIndex(ui.setup.ModeIndex+delta, len(modes))
		ui.setup.PlayerCount = defaultPlayerCountForMode(ui.selectedMode())
		ui.syncScenarioSelection()
	case 1:
		if len(scenarios) == 0 {
			return
		}
		ui.setup.ScenarioIndex = wrapIndex(ui.setup.ScenarioIndex+delta, len(scenarios))
	case 2:
		ui.setup.PlayerCount = clampInt(ui.setup.PlayerCount+delta, 1, 8)
	case 3:
		ui.setup.RunDays = clampInt(ui.setup.RunDays+delta*3, 1, 300)
	}
}

func (ui *gameUI) drawSetup() {
	left := rl.NewRectangle(20, 20, float32(ui.width)*0.42, float32(ui.height-40))
	right := rl.NewRectangle(left.X+left.Width+20, 20, float32(ui.width)-left.Width-60, float32(ui.height-40))
	drawPanel(left, "New Run Setup")
	drawPanel(right, "Scenario Preview")

	rows := []struct {
		label string
		value string
	}{
		{"Mode", modeLabel(ui.selectedMode())},
		{"Scenario", ui.selectedScenario().Name},
		{"Players", fmt.Sprintf("%d", ui.setup.PlayerCount)},
		{"Run Length (days)", fmt.Sprintf("%d", ui.setup.RunDays)},
		{"Configure Players", "Enter"},
		{"Scenario Builder", "Enter"},
		{"Back", "Enter"},
	}

	for i, row := range rows {
		y := int32(left.Y) + 58 + int32(i*54)
		if i == ui.setup.Cursor {
			rl.DrawRectangle(int32(left.X)+18, y-8, int32(left.Width)-36, 42, rl.Fade(colorAccent, 0.2))
		}
		rl.DrawText(row.label, int32(left.X)+28, y, 24, colorText)
		rl.DrawText(row.value, int32(left.X)+280, y, 24, colorAccent)
	}
	rl.DrawText("Left/Right change   Enter select/open", int32(left.X)+26, int32(left.Y+left.Height)-38, 18, colorDim)

	s := ui.selectedScenario()
	drawWrappedText("Name: "+s.Name, right, 30, 25, colorAccent)
	drawWrappedText("Biome: "+s.Biome, right, 64, 22, colorText)
	drawWrappedText("Description: "+safeText(s.Description), right, 108, 20, colorText)
	drawWrappedText("Daunting: "+safeText(s.Daunting), right, 232, 20, colorWarn)
	drawWrappedText("Motivation: "+safeText(s.Motivation), right, 366, 20, colorAccent)
	tr := game.TemperatureRangeForBiome(s.Biome)
	drawWrappedText(fmt.Sprintf("Temperature Range: %dC to %dC", tr.MinC, tr.MaxC), right, 496, 20, colorText)
	wildlife := game.WildlifeForBiome(s.Biome)
	drawWrappedText("Wildlife: "+strings.Join(wildlife, ", "), right, 540, 20, colorDim)
}

func (ui *gameUI) updateScenarioPicker() {
	scenarios := ui.activeScenarios()
	if len(scenarios) == 0 {
		ui.screen = screenSetup
		return
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		ui.screen = screenSetup
		return
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		ui.pick.Cursor = wrapIndex(ui.pick.Cursor+1, len(scenarios))
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		ui.pick.Cursor = wrapIndex(ui.pick.Cursor-1, len(scenarios))
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		ui.setup.ScenarioIndex = ui.pick.Cursor
		ui.screen = screenSetup
	}
}

func (ui *gameUI) drawScenarioPicker() {
	left := rl.NewRectangle(20, 20, float32(ui.width)*0.35, float32(ui.height-40))
	right := rl.NewRectangle(left.X+left.Width+20, 20, float32(ui.width)-left.Width-60, float32(ui.height-40))
	drawPanel(left, "Scenario List")
	drawPanel(right, "Details")

	scenarios := ui.activeScenarios()
	for i, scenario := range scenarios {
		y := int32(left.Y) + 56 + int32(i*34)
		if y > int32(left.Y+left.Height)-40 {
			break
		}
		clr := colorText
		if i == ui.pick.Cursor {
			clr = colorAccent
			rl.DrawRectangle(int32(left.X)+12, y-6, int32(left.Width)-24, 28, rl.Fade(colorAccent, 0.18))
		}
		rl.DrawText(scenario.Name, int32(left.X)+22, y, 22, clr)
	}

	sel := scenarios[clampInt(ui.pick.Cursor, 0, len(scenarios)-1)]
	drawWrappedText("Name: "+sel.Name, right, 30, 25, colorAccent)
	drawWrappedText("Biome: "+sel.Biome, right, 64, 22, colorText)
	drawWrappedText("Description: "+safeText(sel.Description), right, 106, 20, colorText)
	drawWrappedText("Daunting: "+safeText(sel.Daunting), right, 230, 20, colorWarn)
	drawWrappedText("Motivation: "+safeText(sel.Motivation), right, 360, 20, colorAccent)
	drawWrappedText("Enter to select, Esc back", right, int32(right.Height)-38, 19, colorDim)
}

func (ui *gameUI) preparePlayerConfig() {
	if len(ui.pcfg.Players) != ui.setup.PlayerCount {
		ui.pcfg.Players = make([]game.PlayerConfig, ui.setup.PlayerCount)
		for i := range ui.pcfg.Players {
			ui.pcfg.Players[i] = defaultPlayerConfig(i, ui.selectedMode())
		}
	}
	ui.pcfg.Cursor = 0
	ui.pcfg.PlayerIndex = 0
	ui.pcfg.EditingName = false
	ui.pcfg.NameBuffer = ""
	ui.status = ""
}

func (ui *gameUI) updatePlayerConfig() {
	if len(ui.pcfg.Players) == 0 {
		ui.preparePlayerConfig()
	}
	if ui.pcfg.PlayerIndex < 0 {
		ui.pcfg.PlayerIndex = 0
	}
	if ui.pcfg.PlayerIndex >= len(ui.pcfg.Players) {
		ui.pcfg.PlayerIndex = len(ui.pcfg.Players) - 1
	}
	player := &ui.pcfg.Players[ui.pcfg.PlayerIndex]

	if ui.pcfg.EditingName {
		captureTextInput(&ui.pcfg.NameBuffer, 28)
		if rl.IsKeyPressed(rl.KeyEnter) {
			name := strings.TrimSpace(ui.pcfg.NameBuffer)
			if name != "" {
				player.Name = name
			}
			ui.pcfg.EditingName = false
		}
		if rl.IsKeyPressed(rl.KeyEscape) {
			ui.pcfg.EditingName = false
		}
		return
	}

	if rl.IsKeyPressed(rl.KeyEscape) {
		ui.screen = screenSetup
		return
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		ui.pcfg.Cursor = wrapIndex(ui.pcfg.Cursor+1, 14)
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		ui.pcfg.Cursor = wrapIndex(ui.pcfg.Cursor-1, 14)
	}

	if rl.IsKeyPressed(rl.KeyLeft) {
		ui.adjustPlayerConfig(player, -1)
	}
	if rl.IsKeyPressed(rl.KeyRight) {
		ui.adjustPlayerConfig(player, 1)
	}

	if rl.IsKeyPressed(rl.KeyEnter) {
		switch ui.pcfg.Cursor {
		case 1:
			ui.pcfg.EditingName = true
			ui.pcfg.NameBuffer = player.Name
		case 10:
			ui.openKitPicker()
		case 11:
			ui.pcfg.PlayerIndex = wrapIndex(ui.pcfg.PlayerIndex+1, len(ui.pcfg.Players))
		case 12:
			ui.startRunFromConfig()
		case 13:
			ui.screen = screenSetup
		}
	}
}

func (ui *gameUI) adjustPlayerConfig(p *game.PlayerConfig, delta int) {
	switch ui.pcfg.Cursor {
	case 0:
		ui.pcfg.PlayerIndex = wrapIndex(ui.pcfg.PlayerIndex+delta, len(ui.pcfg.Players))
	case 2:
		sexes := []game.Sex{game.SexMale, game.SexFemale, game.SexNonBinary, game.SexOther}
		i := indexOfSex(sexes, p.Sex)
		p.Sex = sexes[wrapIndex(i+delta, len(sexes))]
	case 3:
		types := []game.BodyType{game.BodyTypeNeutral, game.BodyTypeMale, game.BodyTypeFemale}
		i := indexOfBodyType(types, p.BodyType)
		p.BodyType = types[wrapIndex(i+delta, len(types))]
	case 4:
		p.WeightKg = clampInt(p.WeightKg+delta, 40, 180)
	case 5:
		p.HeightFt = clampInt(p.HeightFt+delta, 4, 7)
	case 6:
		p.HeightIn = clampInt(p.HeightIn+delta, 0, 11)
	case 7:
		p.Endurance = clampInt(p.Endurance+delta, -3, 3)
	case 8:
		p.Bushcraft = clampInt(p.Bushcraft+delta, -3, 3)
	case 9:
		p.Mental = clampInt(p.Mental+delta, -3, 3)
	}
}

func (ui *gameUI) drawPlayerConfig() {
	left := rl.NewRectangle(20, 20, float32(ui.width)*0.34, float32(ui.height-40))
	mid := rl.NewRectangle(left.X+left.Width+16, 20, float32(ui.width)*0.29, float32(ui.height-40))
	right := rl.NewRectangle(mid.X+mid.Width+16, 20, float32(ui.width)-left.Width-mid.Width-68, float32(ui.height-40))
	drawPanel(left, "Player Config")
	drawPanel(mid, "Player Details")
	drawPanel(right, "Player Preview")

	if len(ui.pcfg.Players) == 0 {
		drawWrappedText("No players configured.", left, 50, 22, colorWarn)
		return
	}
	p := ui.pcfg.Players[ui.pcfg.PlayerIndex]

	rows := []struct {
		label string
		value string
	}{
		{"Player", fmt.Sprintf("%d / %d", ui.pcfg.PlayerIndex+1, len(ui.pcfg.Players))},
		{"Name", p.Name},
		{"Sex", string(p.Sex)},
		{"Body Type", string(p.BodyType)},
		{"Weight (kg)", fmt.Sprintf("%d", p.WeightKg)},
		{"Height (ft)", fmt.Sprintf("%d", p.HeightFt)},
		{"Height (in)", fmt.Sprintf("%d", p.HeightIn)},
		{"Endurance", fmt.Sprintf("%+d", p.Endurance)},
		{"Bushcraft", fmt.Sprintf("%+d", p.Bushcraft)},
		{"Mental", fmt.Sprintf("%+d", p.Mental)},
		{"Kit Picker", "Enter"},
		{"Next Player", "Enter"},
		{"Start Run", "Enter"},
		{"Back", "Enter"},
	}
	for i, row := range rows {
		y := int32(left.Y) + 58 + int32(i*42)
		if i == ui.pcfg.Cursor {
			rl.DrawRectangle(int32(left.X)+10, y-6, int32(left.Width)-20, 34, rl.Fade(colorAccent, 0.2))
		}
		rl.DrawText(row.label, int32(left.X)+20, y, 20, colorText)
		rl.DrawText(row.value, int32(left.X)+220, y, 20, colorAccent)
	}

	if ui.pcfg.EditingName {
		r := rl.NewRectangle(left.X+20, left.Y+right.Height-104, left.Width-40, 72)
		rl.DrawRectangleRounded(r, 0.2, 8, rl.Fade(colorPanel, 0.95))
		rl.DrawRectangleRoundedLinesEx(r, 0.2, 8, 2, colorAccent)
		rl.DrawText("Editing Name", int32(r.X)+12, int32(r.Y)+10, 18, colorAccent)
		rl.DrawText(ui.pcfg.NameBuffer+"_", int32(r.X)+12, int32(r.Y)+34, 24, colorText)
	}

	detailLines := []string{
		fmt.Sprintf("Name: %s", p.Name),
		fmt.Sprintf("Sex: %s", p.Sex),
		fmt.Sprintf("Body Type: %s", p.BodyType),
		fmt.Sprintf("Height: %d ft %d in", p.HeightFt, p.HeightIn),
		fmt.Sprintf("Weight: %d kg", p.WeightKg),
		"",
		fmt.Sprintf("Endurance: %+d", p.Endurance),
		fmt.Sprintf("Bushcraft: %+d", p.Bushcraft),
		fmt.Sprintf("Mental: %+d", p.Mental),
		fmt.Sprintf("Kit: %d / %d selected", len(p.Kit), maxInt(1, p.KitLimit)),
		"",
		"Controls:",
		"Up/Down move rows",
		"Left/Right adjust values",
		"Enter select",
		"Esc back",
	}
	drawLines(mid, 44, 21, detailLines, colorText)

	drawPlayerPreview(right, p)
}

func (ui *gameUI) startRunFromConfig() {
	cfg := game.RunConfig{
		Mode:        ui.selectedMode(),
		ScenarioID:  ui.selectedScenario().ID,
		PlayerCount: len(ui.pcfg.Players),
		Players:     append([]game.PlayerConfig(nil), ui.pcfg.Players...),
		RunLength:   game.RunLength{Days: ui.setup.RunDays},
		Seed:        time.Now().UnixNano(),
	}
	run, err := game.NewRunState(cfg)
	if err != nil {
		ui.status = "Failed to start: " + err.Error()
		return
	}
	ui.run = &run
	ui.runMessages = nil
	ui.runPlayedFor = 0
	ui.runFocus = 0
	ui.runInput = ""
	ui.status = ""
	ui.appendRunMessage("Run started")
	ui.appendRunMessage(fmt.Sprintf("Mode: %s | Scenario: %s | Players: %d", modeLabel(run.Config.Mode), run.Scenario.Name, len(run.Players)))
	ui.screen = screenRun
}

func (ui *gameUI) updateLoad() {
	if rl.IsKeyPressed(rl.KeyEscape) {
		if ui.load.ReturnToRun && ui.run != nil {
			ui.screen = screenRun
		} else {
			ui.screen = screenMenu
		}
		return
	}
	if rl.IsKeyPressed(rl.KeyR) {
		ui.openLoad(ui.load.ReturnToRun)
		return
	}
	if len(ui.load.Entries) == 0 {
		return
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		ui.load.Cursor = wrapIndex(ui.load.Cursor+1, len(ui.load.Entries))
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		ui.load.Cursor = wrapIndex(ui.load.Cursor-1, len(ui.load.Entries))
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		entry := ui.load.Entries[ui.load.Cursor]
		r := entry.Saved.Run
		r.EnsureWeather()
		r.EnsurePlayerRuntimeStats()
		ui.run = &r
		ui.runPlayedFor = 0
		ui.runFocus = 0
		ui.status = ""
		ui.runMessages = nil
		ui.appendRunMessage("Loaded " + filepath.Base(entry.Path))
		ui.screen = screenRun
	}
}

func (ui *gameUI) drawLoad() {
	left := rl.NewRectangle(20, 20, float32(ui.width)*0.35, float32(ui.height-40))
	right := rl.NewRectangle(left.X+left.Width+16, 20, float32(ui.width)-left.Width-56, float32(ui.height-40))
	drawPanel(left, "Load Run")
	drawPanel(right, "Save Details")

	if len(ui.load.Entries) == 0 {
		drawWrappedText("No save files found. Use save command in run screen first.", left, 60, 22, colorWarn)
		drawWrappedText("Esc back", left, int32(left.Height)-36, 20, colorDim)
		return
	}

	for i, entry := range ui.load.Entries {
		y := int32(left.Y) + 54 + int32(i*42)
		if y > int32(left.Y+left.Height)-40 {
			break
		}
		if i == ui.load.Cursor {
			rl.DrawRectangle(int32(left.X)+10, y-6, int32(left.Width)-20, 32, rl.Fade(colorAccent, 0.2))
		}
		rl.DrawText(filepath.Base(entry.Path), int32(left.X)+20, y, 20, colorText)
	}

	sel := ui.load.Entries[ui.load.Cursor]
	run := sel.Saved.Run
	weather := run.Weather
	weatherLabel := game.WeatherLabel(weather.Type)
	lines := []string{
		"File: " + filepath.Base(sel.Path),
		"Saved: " + sel.Saved.SavedAt.Local().Format("2006-01-02 15:04:05"),
		"",
		"Mode: " + modeLabel(run.Config.Mode),
		"Scenario: " + run.Scenario.Name,
		fmt.Sprintf("Day: %d", run.Day),
		fmt.Sprintf("Players: %d", len(run.Players)),
		fmt.Sprintf("Weather: %s", weatherLabel),
		fmt.Sprintf("Temp: %dC", weather.TemperatureC),
		"",
		"Enter to load",
		"R to refresh",
		"Esc back",
	}
	drawLines(right, 48, 22, lines, colorText)
}

func (ui *gameUI) updateRun(delta time.Duration) {
	if ui.run == nil {
		ui.screen = screenMenu
		return
	}
	if len(ui.run.Players) > 0 {
		if rl.IsKeyPressed(rl.KeyTab) || rl.IsKeyPressed(rl.KeyRightBracket) {
			ui.runFocus = wrapIndex(ui.runFocus+1, len(ui.run.Players))
		}
		if rl.IsKeyPressed(rl.KeyLeftBracket) {
			ui.runFocus = wrapIndex(ui.runFocus-1, len(ui.run.Players))
		}
	} else {
		ui.runFocus = 0
	}
	ui.runPlayedFor += delta
	dayDuration := ui.autoDayDuration()
	ui.run.ApplyRealtimeMetabolism(ui.runPlayedFor, dayDuration)
	for ui.runPlayedFor >= dayDuration {
		prevDay := ui.run.Day
		ui.run.AdvanceDay()
		ui.runPlayedFor -= dayDuration
		ui.run.ApplyRealtimeMetabolism(ui.runPlayedFor, dayDuration)
		if ui.run.Day != prevDay {
			weather := game.WeatherLabel(ui.run.Weather.Type)
			ui.appendRunMessage(fmt.Sprintf("Day %d started | Weather: %s | Temp: %dC", ui.run.Day, weather, ui.run.Weather.TemperatureC))
		}
	}

	if rl.IsKeyPressed(rl.KeyEscape) {
		ui.screen = screenMenu
		return
	}
	if rl.IsKeyPressed(rl.KeyS) {
		path := savePathForSlot(1)
		if err := saveRunToFile(path, *ui.run); err != nil {
			ui.status = "Save failed: " + err.Error()
		} else {
			ui.status = "Saved to " + path
			ui.appendRunMessage(ui.status)
		}
	}
	if rl.IsKeyPressed(rl.KeyL) {
		ui.openLoad(true)
		return
	}

	captureTextInput(&ui.runInput, 180)
	if rl.IsKeyPressed(rl.KeyBackspace) && len(ui.runInput) == 0 {
		// noop, already handled by captureTextInput.
	}

	if rl.IsKeyPressed(rl.KeyEnter) {
		ui.submitRunInput()
	}
}

func (ui *gameUI) drawRun() {
	if ui.run == nil {
		return
	}
	outer := rl.NewRectangle(20, 20, float32(ui.width-40), float32(ui.height-40))
	topH := float32(194)
	bottomH := float32(86)
	if outer.Height < 470 {
		topH = 164
		bottomH = 76
	}
	middleTop := outer.Y + topH
	bottomTop := outer.Y + outer.Height - bottomH
	if bottomTop-middleTop < 180 {
		bottomTop = middleTop + 180
	}
	splitX := outer.X + outer.Width*0.66
	top := rl.NewRectangle(outer.X, outer.Y, outer.Width, topH)
	left := rl.NewRectangle(outer.X, middleTop, splitX-outer.X, bottomTop-middleTop)
	right := rl.NewRectangle(splitX, middleTop, outer.X+outer.Width-splitX, bottomTop-middleTop)
	bottom := rl.NewRectangle(outer.X, bottomTop, outer.Width, outer.Y+outer.Height-bottomTop)

	rl.DrawRectangleRec(outer, rl.Fade(colorPanel, 0.92))
	rl.DrawRectangleRec(top, rl.Fade(colorPanel, 0.98))
	rl.DrawRectangleRec(left, rl.Fade(colorPanel, 0.92))
	rl.DrawRectangleRec(right, rl.Fade(colorPanel, 0.9))
	rl.DrawRectangleRec(bottom, rl.Fade(colorPanel, 0.98))

	drawUILine(outer.X, outer.Y, outer.X+outer.Width, outer.Y, 2, colorBorder)
	drawUILine(outer.X, outer.Y+outer.Height, outer.X+outer.Width, outer.Y+outer.Height, 2, colorBorder)
	drawUILine(outer.X, outer.Y, outer.X, outer.Y+outer.Height, 2, colorBorder)
	drawUILine(outer.X+outer.Width, outer.Y, outer.X+outer.Width, outer.Y+outer.Height, 2, colorBorder)
	drawUILine(outer.X, middleTop, outer.X+outer.Width, middleTop, 2, rl.Fade(colorBorder, 0.85))
	drawUILine(outer.X, bottomTop, outer.X+outer.Width, bottomTop, 2, rl.Fade(colorBorder, 0.85))
	drawUILine(splitX, middleTop, splitX, bottomTop, 2, rl.Fade(colorBorder, 0.8))

	rl.DrawText("RUN STATUS", int32(top.X)+12, int32(top.Y)+8, 20, colorAccent)
	rl.DrawText("MESSAGE HISTORY", int32(left.X)+12, int32(left.Y)+8, 18, colorAccent)
	rl.DrawText("PLAYERS", int32(right.X)+12, int32(right.Y)+8, 18, colorAccent)
	rl.DrawText("COMMAND INPUT", int32(bottom.X)+12, int32(bottom.Y)+8, 18, colorAccent)

	season := "unknown"
	if s, ok := ui.run.CurrentSeason(); ok {
		season = string(s)
	}
	weather := game.WeatherLabel(ui.run.Weather.Type)
	header := fmt.Sprintf("SURVIVE IT!   Mode: %s   Scenario: %s   Day: %d   Season: %s   Weather: %s   Temp: %dC", modeLabel(ui.run.Config.Mode), ui.run.Scenario.Name, ui.run.Day, season, weather, ui.run.Weather.TemperatureC)
	rl.DrawText(header, int32(top.X)+14, int32(top.Y)+38, 21, colorAccent)
	nextIn := ui.autoDayDuration() - ui.runPlayedFor
	if nextIn < 0 {
		nextIn = 0
	}
	rl.DrawText(fmt.Sprintf("Game Time: Day %d %s   Next Day In: %s", ui.run.Day, formatClockDuration(ui.runPlayedFor), formatClockDuration(nextIn)), int32(top.X)+14, int32(top.Y)+66, 19, colorText)

	focus := game.PlayerState{}
	if len(ui.run.Players) > 0 {
		ui.runFocus = clampInt(ui.runFocus, 0, len(ui.run.Players)-1)
		focus = ui.run.Players[ui.runFocus]
		rl.DrawText(fmt.Sprintf("Focus Player: %s (%d/%d)   TAB or [ ] switch", focus.Name, ui.runFocus+1, len(ui.run.Players)), int32(top.X)+14, int32(top.Y)+90, 18, colorDim)
	}

	barInset := float32(14)
	barGap := float32(12)
	row1Y := top.Y + 124
	row2Y := row1Y + 34
	row1W := (top.Width - barInset*2 - barGap*3) / 4
	row2W := (top.Width - barInset*2 - barGap*2) / 3
	condition := runConditionScore(focus)
	drawRunStatBar(rl.NewRectangle(top.X+barInset, row1Y, row1W, 18), "Condition", condition, false)
	drawRunStatBar(rl.NewRectangle(top.X+barInset+(row1W+barGap)*1, row1Y, row1W, 18), "Energy", focus.Energy, false)
	drawRunStatBar(rl.NewRectangle(top.X+barInset+(row1W+barGap)*2, row1Y, row1W, 18), "Hydration", focus.Hydration, false)
	drawRunStatBar(rl.NewRectangle(top.X+barInset+(row1W+barGap)*3, row1Y, row1W, 18), "Morale", focus.Morale, false)
	drawRunStatBar(rl.NewRectangle(top.X+barInset, row2Y, row2W, 18), "Hunger", focus.Hunger, true)
	drawRunStatBar(rl.NewRectangle(top.X+barInset+(row2W+barGap)*1, row2Y, row2W, 18), "Thirst", focus.Thirst, true)
	drawRunStatBar(rl.NewRectangle(top.X+barInset+(row2W+barGap)*2, row2Y, row2W, 18), "Fatigue", focus.Fatigue, true)

	lineY := int32(left.Y) + 38
	maxMessageLines := int((left.Height - 50) / 23)
	if maxMessageLines < 4 {
		maxMessageLines = 4
	}
	start := len(ui.runMessages) - maxMessageLines
	if start < 0 {
		start = 0
	}
	for i := start; i < len(ui.runMessages); i++ {
		if lineY > int32(left.Y+left.Height)-26 {
			break
		}
		rl.DrawText(ui.runMessages[i], int32(left.X)+12, lineY, 18, colorText)
		lineY += 23
	}

	py := int32(right.Y) + 36
	for i := range ui.run.Players {
		p := ui.run.Players[i]
		if i == ui.runFocus {
			rl.DrawRectangle(int32(right.X)+6, py-4, int32(right.Width)-12, 54, rl.Fade(colorAccent, 0.16))
		}
		name := fmt.Sprintf("P%d %s", p.ID, p.Name)
		rl.DrawText(name, int32(right.X)+12, py, 19, colorAccent)
		py += 22
		overview := fmt.Sprintf("Cond:%d  E:%d  H2O:%d  M:%d", runConditionScore(p), p.Energy, p.Hydration, p.Morale)
		rl.DrawText(overview, int32(right.X)+16, py, 17, colorText)
		py += 20
		effects := fmt.Sprintf("Hu:%d  Th:%d  Fa:%d  Ailments:%d", p.Hunger, p.Thirst, p.Fatigue, len(p.Ailments))
		rl.DrawText(effects, int32(right.X)+16, py, 17, colorDim)
		py += 22
		if py > int32(right.Y+right.Height)-44 {
			break
		}
	}

	cmdHint := "Commands: next | save | load | menu | hunt ... | fire ... | craft ...   Hotkeys: S save  L load  Esc menu"
	rl.DrawText(cmdHint, int32(bottom.X)+14, int32(bottom.Y)+34, 17, colorDim)
	in := strings.TrimSpace(ui.runInput)
	if in == "" {
		rl.DrawText("> ", int32(bottom.X)+14, int32(bottom.Y)+56, 24, colorText)
	} else {
		rl.DrawText("> "+ui.runInput+"_", int32(bottom.X)+14, int32(bottom.Y)+56, 24, colorAccent)
	}
	if strings.TrimSpace(ui.status) != "" {
		statusX := int32(bottom.X + bottom.Width*0.45)
		rl.DrawText(ui.status, statusX, int32(bottom.Y)+56, 20, colorWarn)
	}
}

func (ui *gameUI) submitRunInput() {
	if ui.run == nil {
		return
	}
	command := strings.TrimSpace(strings.ToLower(ui.runInput))
	ui.runInput = ""
	if command == "" {
		ui.status = "Enter a command."
		return
	}

	switch command {
	case "next":
		prev := ui.run.Day
		ui.run.AdvanceDay()
		ui.runPlayedFor = 0
		if ui.run.Day != prev {
			ui.appendRunMessage(fmt.Sprintf("Day %d started", ui.run.Day))
		}
		ui.status = ""
		return
	case "save":
		path := savePathForSlot(1)
		if err := saveRunToFile(path, *ui.run); err != nil {
			ui.status = "Save failed: " + err.Error()
			return
		}
		ui.status = "Saved to " + path
		ui.appendRunMessage(ui.status)
		return
	case "load":
		ui.openLoad(true)
		return
	case "menu", "back":
		ui.screen = screenMenu
		return
	}

	if strings.HasPrefix(command, "hunt") || strings.HasPrefix(command, "catch") {
		ui.handleHuntCommand(command)
		return
	}

	res := ui.run.ExecuteRunCommand(command)
	if res.Handled {
		ui.status = ""
		ui.appendRunMessage(res.Message)
		return
	}
	ui.status = "Unknown command"
}

func (ui *gameUI) handleHuntCommand(command string) {
	if ui.run == nil {
		return
	}
	fields := strings.Fields(command)
	domain := game.AnimalDomainLand
	choice := game.MealChoice{PortionGrams: 300, Cooked: true, EatLiver: false}
	playerID := 1
	for _, field := range fields[1:] {
		switch field {
		case "land":
			domain = game.AnimalDomainLand
		case "fish":
			domain = game.AnimalDomainWater
		case "air", "bird":
			domain = game.AnimalDomainAir
		case "raw":
			choice.Cooked = false
		case "liver":
			choice.EatLiver = true
		default:
			if strings.HasPrefix(field, "p") {
				if id, err := strconv.Atoi(strings.TrimPrefix(field, "p")); err == nil && id > 0 {
					playerID = id
				}
				continue
			}
			if grams, err := strconv.Atoi(field); err == nil && grams > 0 {
				choice.PortionGrams = grams
			}
		}
	}

	catch, outcome, err := ui.run.CatchAndConsume(playerID, domain, choice)
	if err != nil {
		ui.status = "Hunt failed: " + err.Error()
		return
	}
	prep := "cooked"
	if !choice.Cooked {
		prep = "raw"
	}
	msg := fmt.Sprintf("P%d ate %dg %s (%s, %dg caught): +%dE +%dH2O +%dM | %dkcal %dgP %dgF %dgS",
		outcome.PlayerID,
		outcome.PortionGrams,
		catch.Animal.Name,
		prep,
		catch.WeightGrams,
		outcome.EnergyDelta,
		outcome.HydrationDelta,
		outcome.MoraleDelta,
		outcome.Nutrition.CaloriesKcal,
		outcome.Nutrition.ProteinG,
		outcome.Nutrition.FatG,
		outcome.Nutrition.SugarG,
	)
	if len(outcome.DiseaseEvents) > 0 {
		parts := make([]string, 0, len(outcome.DiseaseEvents))
		for _, event := range outcome.DiseaseEvents {
			parts = append(parts, event.Ailment.Name)
		}
		msg += " | illness risk triggered: " + strings.Join(parts, ", ")
	}
	ui.status = ""
	ui.appendRunMessage(msg)
}

func (ui *gameUI) openLoad(returnToRun bool) {
	entries, err := loadSaves()
	if err != nil {
		ui.status = "Load failed: " + err.Error()
		entries = nil
	}
	ui.load = loadState{Cursor: 0, ReturnToRun: returnToRun, Entries: entries}
	ui.screen = screenLoad
}

func loadSaves() ([]saveEntry, error) {
	matches, err := filepath.Glob("survive-it-save-*.json")
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	entries := make([]saveEntry, 0, len(matches))
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var payload savedRun
		if err := json.Unmarshal(data, &payload); err != nil {
			continue
		}
		entries = append(entries, saveEntry{Path: path, Saved: payload})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Saved.SavedAt.After(entries[j].Saved.SavedAt)
	})
	return entries, nil
}

func saveRunToFile(path string, run game.RunState) error {
	payload := savedRun{FormatVersion: 1, SavedAt: time.Now().UTC(), Run: run}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func savePathForSlot(slot int) string {
	if slot < 1 {
		slot = 1
	}
	return fmt.Sprintf("survive-it-save-%d.json", slot)
}

func (ui *gameUI) appendRunMessage(message string) {
	line := strings.TrimSpace(message)
	if line == "" {
		return
	}
	formatted := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), line)
	ui.runMessages = append(ui.runMessages, formatted)
	if len(ui.runMessages) > 260 {
		ui.runMessages = append([]string(nil), ui.runMessages[len(ui.runMessages)-260:]...)
	}
}

func drawPanel(rect rl.Rectangle, title string) {
	rl.DrawRectangleRounded(rect, 0.04, 8, colorPanel)
	rl.DrawRectangleRoundedLinesEx(rect, 0.04, 8, 2, colorBorder)
	rl.DrawText(title, int32(rect.X)+12, int32(rect.Y)+8, 20, colorAccent)
}

func drawTextCentered(text string, rect rl.Rectangle, yOffset int32, fontSize int32, clr rl.Color) {
	width := rl.MeasureText(text, fontSize)
	x := int32(rect.X + (rect.Width-float32(width))/2)
	rl.DrawText(text, x, int32(rect.Y)+yOffset, fontSize, clr)
}

func drawWrappedText(text string, rect rl.Rectangle, y int32, size int32, clr rl.Color) {
	maxWidth := int32(rect.Width) - 26
	lines := wrapText(text, size, maxWidth)
	for i, line := range lines {
		rl.DrawText(line, int32(rect.X)+14, int32(rect.Y)+y+int32(i)*(size+6), size, clr)
	}
}

func drawLines(rect rl.Rectangle, y int32, size int32, lines []string, clr rl.Color) {
	for i, line := range lines {
		rl.DrawText(line, int32(rect.X)+14, int32(rect.Y)+y+int32(i)*(size+6), size, clr)
	}
}

func wrapText(text string, size int32, maxWidth int32) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}
	lines := make([]string, 0, 8)
	current := words[0]
	for _, word := range words[1:] {
		candidate := current + " " + word
		if int32(rl.MeasureText(candidate, size)) <= maxWidth {
			current = candidate
			continue
		}
		lines = append(lines, current)
		current = word
	}
	lines = append(lines, current)
	return lines
}

func drawPlayerPreview(rect rl.Rectangle, p game.PlayerConfig) {
	cx := rect.X + rect.Width/2
	groundY := rect.Y + rect.Height - 70

	heightInches := p.HeightFt*12 + p.HeightIn
	if heightInches <= 0 {
		heightInches = 70
	}
	heightScale := float32(heightInches) / 70.0
	if heightScale < 0.76 {
		heightScale = 0.76
	}
	if heightScale > 1.25 {
		heightScale = 1.25
	}

	weightScale := float32(p.WeightKg) / 75.0
	if weightScale < 0.65 {
		weightScale = 0.65
	}
	if weightScale > 1.45 {
		weightScale = 1.45
	}

	bodyHeight := 190.0 * heightScale
	headRadius := 18.0 * (0.9 + (heightScale-1.0)*0.3)
	shoulderW := 46.0 * weightScale
	hipW := 34.0 * weightScale
	torsoH := 72.0 * heightScale
	legH := 78.0 * heightScale
	armH := 60.0 * heightScale

	switch p.BodyType {
	case game.BodyTypeMale:
		shoulderW *= 1.12
		hipW *= 0.95
	case game.BodyTypeFemale:
		shoulderW *= 0.95
		hipW *= 1.12
	}

	topY := groundY - float32(bodyHeight)
	headY := topY + float32(headRadius)
	torsoY := headY + float32(headRadius) + 8

	// Back card
	card := rl.NewRectangle(rect.X+20, rect.Y+40, rect.Width-40, rect.Height-100)
	rl.DrawRectangleRounded(card, 0.08, 8, rl.Fade(colorBorder, 0.08))
	rl.DrawRectangleRoundedLinesEx(card, 0.08, 8, 1.2, rl.Fade(colorBorder, 0.55))

	// Head
	rl.DrawCircle(int32(cx), int32(headY), float32(headRadius), rl.NewColor(46, 220, 120, 255))
	// Neck
	rl.DrawRectangle(int32(cx)-6, int32(headY+float32(headRadius)-1), 12, 10, rl.NewColor(35, 192, 104, 255))
	// Torso
	torso := rl.NewRectangle(cx-float32(shoulderW)/2, torsoY, float32(shoulderW), float32(torsoH))
	rl.DrawRectangleRounded(torso, 0.28, 8, rl.NewColor(28, 174, 95, 255))

	// Arms
	armW := float32(10 * weightScale)
	if armW < 8 {
		armW = 8
	}
	leftArm := rl.NewRectangle(torso.X-armW+2, torsoY+8, armW, float32(armH))
	rightArm := rl.NewRectangle(torso.X+torso.Width-2, torsoY+8, armW, float32(armH))
	rl.DrawRectangleRounded(leftArm, 0.45, 6, rl.NewColor(23, 155, 84, 255))
	rl.DrawRectangleRounded(rightArm, 0.45, 6, rl.NewColor(23, 155, 84, 255))

	// Hips
	hips := rl.NewRectangle(cx-float32(hipW)/2, torsoY+float32(torsoH)-4, float32(hipW), 16)
	rl.DrawRectangleRounded(hips, 0.3, 6, rl.NewColor(21, 145, 79, 255))

	// Legs
	legGap := float32(8)
	legW := float32(10 * weightScale)
	if legW < 8 {
		legW = 8
	}
	leftLeg := rl.NewRectangle(cx-legGap-legW, hips.Y+12, legW, float32(legH))
	rightLeg := rl.NewRectangle(cx+legGap, hips.Y+12, legW, float32(legH))
	rl.DrawRectangleRounded(leftLeg, 0.36, 6, rl.NewColor(19, 132, 72, 255))
	rl.DrawRectangleRounded(rightLeg, 0.36, 6, rl.NewColor(19, 132, 72, 255))

	// Feet shadow
	rl.DrawEllipse(int32(cx)-16, int32(groundY)+6, 22, 6, rl.Fade(colorBorder, 0.4))
	rl.DrawEllipse(int32(cx)+16, int32(groundY)+6, 22, 6, rl.Fade(colorBorder, 0.4))

	rl.DrawText("Preview reacts to height, weight, and body type", int32(rect.X)+20, int32(rect.Y+rect.Height)-44, 18, colorDim)
}

func captureTextInput(target *string, maxLen int) {
	for ch := rl.GetCharPressed(); ch > 0; ch = rl.GetCharPressed() {
		if ch >= 32 && ch <= 126 && len(*target) < maxLen {
			*target += string(rune(ch))
		}
	}
	if rl.IsKeyPressed(rl.KeyBackspace) && len(*target) > 0 {
		*target = (*target)[:len(*target)-1]
	}
}

func modeOptions() []game.GameMode {
	return []game.GameMode{game.ModeAlone, game.ModeNakedAndAfraid, game.ModeNakedAndAfraidXL}
}

func (ui *gameUI) selectedMode() game.GameMode {
	modes := modeOptions()
	if ui.setup.ModeIndex < 0 || ui.setup.ModeIndex >= len(modes) {
		ui.setup.ModeIndex = 0
	}
	return modes[ui.setup.ModeIndex]
}

func scenariosForMode(mode game.GameMode) []game.Scenario {
	return scenariosForModeWithCustom(mode, nil)
}

func scenariosForModeWithCustom(mode game.GameMode, custom []game.Scenario) []game.Scenario {
	all := append([]game.Scenario{}, game.BuiltInScenarios()...)
	all = append(all, custom...)
	out := make([]game.Scenario, 0, len(all))
	for _, scenario := range all {
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

func (ui *gameUI) activeScenarios() []game.Scenario {
	return scenariosForModeWithCustom(ui.selectedMode(), ui.customScenarios)
}

func (ui *gameUI) selectedScenario() game.Scenario {
	scenarios := ui.activeScenarios()
	if len(scenarios) == 0 {
		return game.Scenario{Name: "No Scenarios", Biome: "unknown"}
	}
	ui.setup.ScenarioIndex = clampInt(ui.setup.ScenarioIndex, 0, len(scenarios)-1)
	return scenarios[ui.setup.ScenarioIndex]
}

func (ui *gameUI) syncScenarioSelection() {
	scenarios := ui.activeScenarios()
	if len(scenarios) == 0 {
		ui.setup.ScenarioIndex = 0
		return
	}
	ui.setup.ScenarioIndex = clampInt(ui.setup.ScenarioIndex, 0, len(scenarios)-1)
}

func defaultPlayerCountForMode(mode game.GameMode) int {
	if mode == game.ModeAlone {
		return 1
	}
	if mode == game.ModeNakedAndAfraidXL {
		return 4
	}
	return 2
}

func defaultKitLimitForMode(mode game.GameMode) int {
	switch mode {
	case game.ModeAlone:
		return 10
	case game.ModeNakedAndAfraidXL, game.ModeNakedAndAfraid:
		return 1
	default:
		return 1
	}
}

func defaultPlayerConfig(i int, mode game.GameMode) game.PlayerConfig {
	defaultNames := []string{"Sophia", "Daniel", "Mia", "Jack", "Luna", "Leo", "Avery", "Harper"}
	name := fmt.Sprintf("Player %d", i+1)
	if i < len(defaultNames) {
		name = defaultNames[i]
	}
	return game.PlayerConfig{
		Name:      name,
		Sex:       game.SexOther,
		BodyType:  game.BodyTypeNeutral,
		WeightKg:  75,
		HeightFt:  5,
		HeightIn:  10,
		Endurance: 0,
		Bushcraft: 0,
		Mental:    0,
		KitLimit:  defaultKitLimitForMode(mode),
	}
}

func modeLabel(mode game.GameMode) string {
	switch mode {
	case game.ModeAlone:
		return "Alone"
	case game.ModeNakedAndAfraid:
		return "Naked and Afraid"
	case game.ModeNakedAndAfraidXL:
		return "Naked and Afraid XL"
	default:
		return string(mode)
	}
}

func indexOfSex(items []game.Sex, target game.Sex) int {
	for i := range items {
		if items[i] == target {
			return i
		}
	}
	return 0
}

func indexOfBodyType(items []game.BodyType, target game.BodyType) int {
	for i := range items {
		if items[i] == target {
			return i
		}
	}
	return 0
}

func wrapIndex(i int, size int) int {
	if size <= 0 {
		return 0
	}
	for i < 0 {
		i += size
	}
	for i >= size {
		i -= size
	}
	return i
}

func clampInt(v int, min int, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func safeText(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func (ui *gameUI) autoDayDuration() time.Duration {
	h := ui.autoDayHours
	if h < 1 {
		h = 2
	}
	return time.Duration(h) * time.Hour
}

func formatClockDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(math.Round(d.Seconds()))
	hours := total / 3600
	minutes := (total % 3600) / 60
	seconds := total % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func drawUILine(x1 float32, y1 float32, x2 float32, y2 float32, thickness float32, clr rl.Color) {
	rl.DrawLineEx(rl.Vector2{X: x1, Y: y1}, rl.Vector2{X: x2, Y: y2}, thickness, clr)
}

func drawRunStatBar(rect rl.Rectangle, label string, value int, danger bool) {
	v := clampInt(value, 0, 100)
	fillWidth := (rect.Width - 2) * float32(v) / 100
	if fillWidth < 0 {
		fillWidth = 0
	}
	rl.DrawText(fmt.Sprintf("%s %d%%", label, v), int32(rect.X)+2, int32(rect.Y)-16, 15, colorText)
	rl.DrawRectangleRec(rect, rl.NewColor(8, 16, 12, 255))
	if fillWidth > 0 {
		fill := rl.NewRectangle(rect.X+1, rect.Y+1, fillWidth, rect.Height-2)
		rl.DrawRectangleRec(fill, runBarColor(v, danger))
	}
	rl.DrawRectangleLinesEx(rect, 1.4, rl.Fade(colorBorder, 0.75))
}

func runBarColor(value int, danger bool) rl.Color {
	v := clampInt(value, 0, 100)
	if danger {
		if v >= 70 {
			return rl.NewColor(242, 84, 84, 230)
		}
		if v >= 40 {
			return rl.NewColor(255, 190, 92, 235)
		}
		return rl.NewColor(60, 236, 136, 230)
	}
	if v >= 70 {
		return rl.NewColor(60, 236, 136, 230)
	}
	if v >= 40 {
		return rl.NewColor(255, 190, 92, 235)
	}
	return rl.NewColor(242, 84, 84, 230)
}

func runConditionScore(player game.PlayerState) int {
	condition := 0
	condition += player.Energy
	condition += player.Hydration
	condition += player.Morale
	condition += 100 - clampInt(player.Hunger, 0, 100)
	condition += 100 - clampInt(player.Thirst, 0, 100)
	condition += 100 - clampInt(player.Fatigue, 0, 100)
	condition /= 6
	condition -= len(player.Ailments) * 5
	return clampInt(condition, 0, 100)
}
