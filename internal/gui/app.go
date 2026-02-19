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
	"github.com/appengine-ltd/survive-it/internal/parser"
	"github.com/appengine-ltd/survive-it/internal/update"
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
	screenStatsBuilder
	screenPlayerConfig
	screenKitPicker
	screenScenarioBuilder
	screenPhaseEditor
	screenOptions
	screenLoad
	screenRun
	screenRunPlayers
	screenRunCommandLibrary
	screenRunInventory
)

type menuAction int

const (
	actionStart menuAction = iota
	actionLoad
	actionScenarioBuilder
	actionOptions
	actionInstallUpdate
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
	IssuedKit     []game.KitItem
	IssuedCustom  bool
}

type temperatureUnit int

const (
	tempUnitC temperatureUnit = iota
	tempUnitF
)

type optionsState struct {
	Cursor     int
	TempUnit   temperatureUnit
	GameSounds bool
}

type playerConfigState struct {
	Cursor      int
	PlayerIndex int
	EditingName bool
	NameBuffer  string
	Players     []game.PlayerConfig
	ReturnTo    screen
}

type statsBuilderState struct {
	Cursor      int
	PlayerIndex int
	Editing     bool
	EditBuffer  string
	ReturnTo    screen
}

type phaseEditorState struct {
	Cursor       int
	Adding       bool
	NewSeasonIdx int
	NewDays      string
}

type runPlayersState struct {
	Cursor int
}

type runInventoryState struct {
	PlayerIndex int
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

type updateResult struct {
	status    string
	available bool
	err       error
	apply     bool
}

type gameUI struct {
	cfg AppConfig

	width  int32
	height int32
	quit   bool

	screen screen

	menuCursor int

	setup           setupState
	opts            optionsState
	pick            scenarioPickerState
	sbuild          statsBuilderState
	pcfg            playerConfigState
	kit             kitPickerState
	sb              scenarioBuilderState
	phase           phaseEditorState
	load            loadState
	rplay           runPlayersState
	rinv            runInventoryState
	customScenarios []game.Scenario

	run         *game.RunState
	runMessages []string
	runInput    string
	status      string
	runFocus    int
	lastEntity  string

	cmdParser      *parser.Parser
	commandSink    CommandSink
	intentQueue    *intentQueue
	pendingClarify *parser.ClarifyQuestion
	pendingOptions []parser.Intent

	updateAvailable      bool
	updateBusy           bool
	updateStatus         string
	menuNeedsUpdateCheck bool
	updateResultCh       chan updateResult

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
	intentQ := newIntentQueue(64)
	ui := &gameUI{
		cfg:          cfg,
		width:        1366,
		height:       768,
		screen:       screenMenu,
		autoDayHours: 2,
		cmdParser:    parser.New(),
		intentQueue:  intentQ,
		commandSink:  intentQ,
		setup: setupState{
			ModeIndex:   0,
			PlayerCount: 1,
			RunDays:     365,
		},
		opts: optionsState{
			TempUnit:   tempUnitC,
			GameSounds: true,
		},
		updateResultCh: make(chan updateResult, 4),
	}
	if !cfg.NoUpdate {
		ui.menuNeedsUpdateCheck = true
	}
	custom, _ := loadCustomScenarios(defaultCustomScenariosFile)
	ui.customScenarios = custom
	game.SetExternalScenarios(custom)
	ui.syncScenarioSelection()
	ui.ensureSetupPlayers()
	ui.lastTick = time.Now()
	return ui
}

func (ui *gameUI) Run() error {
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagMsaa4xHint)
	rl.InitWindow(ui.width, ui.height, "survive-it")
	rl.SetExitKey(0)
	rl.SetTargetFPS(60)
	// Slightly soften text edges on scaled/default font rendering.
	defaultFont := rl.GetFontDefault()
	rl.SetTextureFilter(defaultFont.Texture, rl.FilterBilinear)

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

	rl.CloseWindow()
	return nil
}

func (ui *gameUI) update(delta time.Duration) {
	ui.pollUpdateResult()

	switch ui.screen {
	case screenMenu:
		ui.updateMenu()
	case screenSetup:
		ui.updateSetup()
	case screenScenarioPicker:
		ui.updateScenarioPicker()
	case screenStatsBuilder:
		ui.updateStatsBuilder()
	case screenPlayerConfig:
		ui.updatePlayerConfig()
	case screenKitPicker:
		ui.updateKitPicker()
	case screenScenarioBuilder:
		ui.updateScenarioBuilder()
	case screenPhaseEditor:
		ui.updatePhaseEditor()
	case screenOptions:
		ui.updateOptions()
	case screenLoad:
		ui.updateLoad()
	case screenRun:
		ui.updateRun(delta)
	case screenRunPlayers:
		ui.updateRunPlayers()
	case screenRunCommandLibrary:
		ui.updateRunCommandLibrary()
	case screenRunInventory:
		ui.updateRunInventory()
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
	case screenStatsBuilder:
		ui.drawStatsBuilder()
	case screenPlayerConfig:
		ui.drawPlayerConfig()
	case screenKitPicker:
		ui.drawKitPicker()
	case screenScenarioBuilder:
		ui.drawScenarioBuilder()
	case screenPhaseEditor:
		ui.drawPhaseEditor()
	case screenOptions:
		ui.drawOptions()
	case screenLoad:
		ui.drawLoad()
	case screenRun:
		ui.drawRun()
	case screenRunPlayers:
		ui.drawRunPlayers()
	case screenRunCommandLibrary:
		ui.drawRunCommandLibrary()
	case screenRunInventory:
		ui.drawRunInventory()
	}
}

func (ui *gameUI) updateMenu() {
	if ui.menuNeedsUpdateCheck {
		ui.menuNeedsUpdateCheck = false
		ui.triggerUpdateCheck()
	}

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
			ui.setup = setupState{
				ModeIndex:    ui.setup.ModeIndex,
				PlayerCount:  defaultPlayerCountForMode(ui.selectedMode()),
				RunDays:      ui.selectedScenario().DefaultDays,
				IssuedKit:    nil,
				IssuedCustom: false,
			}
			ui.syncScenarioSelection()
			ui.setup.RunDays = ui.selectedScenario().DefaultDays
			ui.ensureSetupPlayers()
			ui.status = ""
			ui.screen = screenSetup
		case actionLoad:
			ui.openLoad(false)
		case actionScenarioBuilder:
			ui.openScenarioBuilder()
		case actionOptions:
			ui.screen = screenOptions
		case actionInstallUpdate:
			ui.triggerApplyUpdate()
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
	drawTextCentered(fmt.Sprintf("v%s (%s) %s", ui.cfg.Version, ui.cfg.Commit, ui.cfg.BuildDate), titleRect, 42, 18, colorDim)
	if ui.updateBusy {
		status := strings.TrimSpace(ui.updateStatus)
		if status == "" {
			status = "Checking for updates..."
		}
		drawTextCentered(status, titleRect, 72, 17, colorText)
	} else if ui.updateAvailable && strings.TrimSpace(ui.updateStatus) != "" {
		drawTextCentered(ui.updateStatus, titleRect, 72, 17, colorAccent)
	}

	items := ui.menuItems()
	menuHeight := float32(150 + len(items)*72)
	if menuHeight < 420 {
		menuHeight = 420
	}
	maxHeight := float32(ui.height) - 220
	if menuHeight > maxHeight {
		menuHeight = maxHeight
	}
	menuRect := rl.NewRectangle(float32(ui.width/2-230), 185, 460, menuHeight)
	drawPanel(menuRect, "Main Menu")
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
	items := []menuItem{
		{Label: "Play Game", Action: actionStart},
		{Label: "Load Game", Action: actionLoad},
		{Label: "Scenario Builder", Action: actionScenarioBuilder},
		{Label: "Options", Action: actionOptions},
	}
	if ui.updateAvailable && !ui.cfg.NoUpdate {
		items = append(items, menuItem{Label: "Install Update", Action: actionInstallUpdate})
	}
	items = append(items, menuItem{Label: "Quit", Action: actionQuit})
	return items
}

func (ui *gameUI) enterMenu() {
	ui.screen = screenMenu
	ui.menuNeedsUpdateCheck = !ui.cfg.NoUpdate
}

func (ui *gameUI) triggerUpdateCheck() {
	if ui.cfg.NoUpdate || ui.updateBusy {
		return
	}
	ui.updateBusy = true
	ui.updateStatus = "Checking for updates..."
	currentVersion := ui.cfg.Version
	go func() {
		res, err := update.Check(update.CheckParams{
			CurrentVersion: currentVersion,
		})
		available := strings.HasPrefix(res, "Update available:")
		ui.updateResultCh <- updateResult{
			status:    res,
			available: available,
			err:       err,
		}
	}()
}

func (ui *gameUI) triggerApplyUpdate() {
	if ui.cfg.NoUpdate || ui.updateBusy {
		return
	}
	ui.updateBusy = true
	ui.updateStatus = "Downloading update..."
	currentVersion := ui.cfg.Version
	go func() {
		res, err := update.Apply(currentVersion)
		available := strings.HasPrefix(res, "Update available:")
		ui.updateResultCh <- updateResult{
			status:    res,
			available: available,
			err:       err,
			apply:     true,
		}
	}()
}

func (ui *gameUI) pollUpdateResult() {
	for {
		select {
		case result := <-ui.updateResultCh:
			ui.updateBusy = false
			if result.err != nil {
				ui.updateStatus = "Update failed: " + result.err.Error()
				if !result.apply {
					ui.updateAvailable = false
				}
				continue
			}
			ui.updateStatus = strings.TrimSpace(result.status)
			ui.updateAvailable = result.available
			if ui.updateAvailable {
				ui.updateStatus = strings.Replace(ui.updateStatus, "Run update to install.", "Select Install Update from menu.", 1)
			}
			if result.apply && ui.updateStatus == "" {
				ui.updateStatus = "Update applied."
			}
		default:
			return
		}
	}
}

func (ui *gameUI) updateOptions() {
	if rl.IsKeyPressed(rl.KeyEscape) {
		ui.enterMenu()
		return
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		ui.opts.Cursor = wrapIndex(ui.opts.Cursor+1, 4)
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		ui.opts.Cursor = wrapIndex(ui.opts.Cursor-1, 4)
	}
	if rl.IsKeyPressed(rl.KeyLeft) {
		ui.adjustOptions(-1)
	}
	if rl.IsKeyPressed(rl.KeyRight) {
		ui.adjustOptions(1)
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		switch ui.opts.Cursor {
		case 0, 1, 2:
			ui.adjustOptions(1)
		case 3:
			ui.enterMenu()
		}
	}
}

func (ui *gameUI) adjustOptions(delta int) {
	switch ui.opts.Cursor {
	case 0:
		if ui.opts.TempUnit == tempUnitC {
			ui.opts.TempUnit = tempUnitF
		} else {
			ui.opts.TempUnit = tempUnitC
		}
	case 1:
		ui.autoDayHours = clampInt(ui.autoDayHours+delta, 1, 24)
	case 2:
		ui.opts.GameSounds = !ui.opts.GameSounds
	}
}

func (ui *gameUI) drawOptions() {
	left := rl.NewRectangle(20, 20, float32(ui.width)*0.45, float32(ui.height-40))
	right := rl.NewRectangle(left.X+left.Width+16, 20, float32(ui.width)-left.Width-56, float32(ui.height-40))
	drawPanel(left, "Options")
	drawPanel(right, "Preview")

	rows := []struct {
		label string
		value string
	}{
		{"Temperature Unit", temperatureUnitLabel(ui.opts.TempUnit)},
		{"Game Hours Per Day", fmt.Sprintf("%d", ui.autoDayHours)},
		{"Game Sounds", map[bool]string{true: "On", false: "Off"}[ui.opts.GameSounds]},
		{"Back", "Enter"},
	}
	for i, row := range rows {
		y := int32(left.Y) + 62 + int32(i*56)
		if i == ui.opts.Cursor {
			rl.DrawRectangle(int32(left.X)+16, y-8, int32(left.Width)-32, 42, rl.Fade(colorAccent, 0.2))
		}
		rl.DrawText(row.label, int32(left.X)+26, y, 24, colorText)
		rl.DrawText(row.value, int32(left.X)+286, y, 24, colorAccent)
	}
	rl.DrawText("Left/Right adjust values, Enter select, Esc back", int32(left.X)+22, int32(left.Y+left.Height)-38, 18, colorDim)

	exampleC := 7
	exampleF := celsiusToFahrenheit(exampleC)
	lines := []string{
		"Current Settings",
		"",
		"Temperature Unit: " + temperatureUnitLabel(ui.opts.TempUnit),
		"Game Hours Per Day: " + fmt.Sprintf("%d", ui.autoDayHours),
		"Game Sounds: " + map[bool]string{true: "On", false: "Off"}[ui.opts.GameSounds],
		"",
		"Temperature Example:",
		fmt.Sprintf("%dC = %dF", exampleC, exampleF),
		"Displayed as: " + ui.formatTemperature(exampleC),
		"",
		"Game sounds controls short ambient cues",
		"(rain start, weather changes, alerts).",
	}
	drawLines(right, 46, 22, lines, colorText)
}

func (ui *gameUI) updateSetup() {
	ui.ensureSetupPlayers()
	if rl.IsKeyPressed(rl.KeyEscape) {
		ui.enterMenu()
		return
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		ui.setup.Cursor = wrapIndex(ui.setup.Cursor+1, 8)
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		ui.setup.Cursor = wrapIndex(ui.setup.Cursor-1, 8)
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
			ui.openStatsBuilder(screenSetup)
		case 5:
			ui.preparePlayerConfig()
			ui.pcfg.ReturnTo = screenSetup
			ui.screen = screenPlayerConfig
		case 6:
			ui.startRunFromConfig()
		case 7:
			ui.enterMenu()
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
		ui.setup.IssuedCustom = false
		ui.syncScenarioSelection()
		ui.setup.RunDays = ui.selectedScenario().DefaultDays
	case 1:
		if len(scenarios) == 0 {
			return
		}
		ui.setup.ScenarioIndex = wrapIndex(ui.setup.ScenarioIndex+delta, len(scenarios))
		if !ui.setup.IssuedCustom {
			ui.setup.IssuedKit = recommendedIssuedKitForScenario(ui.selectedMode(), ui.selectedScenario())
		}
		ui.setup.RunDays = ui.selectedScenario().DefaultDays
	case 2:
		ui.setup.PlayerCount = clampInt(ui.setup.PlayerCount+delta, 1, 8)
	case 3:
		ui.setup.RunDays = clampInt(ui.setup.RunDays+delta*5, 1, 365)
	}
	ui.ensureSetupPlayers()
}

func (ui *gameUI) drawSetup() {
	left := rl.NewRectangle(20, 20, float32(ui.width)*0.56, float32(ui.height-40))
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
		{"Configure Player Stats", "Enter"},
		{"Configure Players", "Enter"},
		{"Start Game", "Enter"},
		{"Back", "Enter"},
	}

	leftColX := int32(left.X) + 28
	rightColX := int32(left.X) + int32(left.Width*0.62)
	rightColMax := int(left.Width) - int(rightColX-int32(left.X)) - 24
	for i, row := range rows {
		y := int32(left.Y) + 58 + int32(i*54)
		if i == ui.setup.Cursor {
			rl.DrawRectangle(int32(left.X)+18, y-8, int32(left.Width)-36, 42, rl.Fade(colorAccent, 0.2))
		}
		rl.DrawText(row.label, leftColX, y, 24, colorText)
		value := row.value
		if len(value) > rightColMax/9 {
			value = value[:maxInt(1, rightColMax/9-3)] + "..."
		}
		rl.DrawText(value, rightColX, y, 24, colorAccent)
	}
	rl.DrawText("Left/Right change   Enter select/open", int32(left.X)+26, int32(left.Y+left.Height)-38, 18, colorDim)

	s := ui.selectedScenario()
	drawWrappedText("Name: "+s.Name, right, 30, 25, colorAccent)
	drawWrappedText("Location: "+safeText(s.Location), right, 64, 22, colorText)
	drawWrappedText("Biome: "+s.Biome, right, 96, 22, colorText)
	drawWrappedText("Description: "+safeText(s.Description), right, 132, 20, colorText)
	drawWrappedText("Daunting: "+safeText(s.Daunting), right, 258, 20, colorWarn)
	drawWrappedText("Motivation: "+safeText(s.Motivation), right, 390, 20, colorAccent)
	tr := game.TemperatureRangeForBiome(s.Biome)
	drawWrappedText("Temperature Range: "+ui.formatTemperatureRange(tr.MinC, tr.MaxC), right, 522, 20, colorText)
	wildlife := game.WildlifeForBiome(s.Biome)
	drawWrappedText("Wildlife: "+strings.Join(wildlife, ", "), right, 558, 20, colorDim)
	drawWrappedText("Issued kit is auto-selected when game starts.", right, 598, 20, colorText)
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
		if !ui.setup.IssuedCustom {
			ui.setup.IssuedKit = recommendedIssuedKitForScenario(ui.selectedMode(), ui.selectedScenario())
		}
		ui.ensureSetupPlayers()
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
	drawWrappedText("Location: "+safeText(sel.Location), right, 64, 22, colorText)
	drawWrappedText("Biome: "+sel.Biome, right, 96, 22, colorText)
	drawWrappedText("Description: "+safeText(sel.Description), right, 128, 20, colorText)
	drawWrappedText("Daunting: "+safeText(sel.Daunting), right, 252, 20, colorWarn)
	drawWrappedText("Motivation: "+safeText(sel.Motivation), right, 382, 20, colorAccent)
	drawWrappedText("Enter to select, Esc back", right, int32(right.Height)-38, 19, colorDim)
}

func (ui *gameUI) openStatsBuilder(returnTo screen) {
	ui.preparePlayerConfig()
	ui.sbuild = statsBuilderState{
		Cursor:      0,
		PlayerIndex: 0,
		Editing:     false,
		EditBuffer:  "",
		ReturnTo:    returnTo,
	}
	ui.screen = screenStatsBuilder
}

func (ui *gameUI) updateStatsBuilder() {
	if len(ui.pcfg.Players) == 0 {
		ui.preparePlayerConfig()
	}
	if ui.sbuild.PlayerIndex < 0 || ui.sbuild.PlayerIndex >= len(ui.pcfg.Players) {
		ui.sbuild.PlayerIndex = 0
	}
	if ui.sbuild.Editing {
		captureTextInput(&ui.sbuild.EditBuffer, 180)
		if rl.IsKeyPressed(rl.KeyEnter) {
			player := &ui.pcfg.Players[ui.sbuild.PlayerIndex]
			switch ui.sbuild.Cursor {
			case 1:
				name := strings.TrimSpace(ui.sbuild.EditBuffer)
				if name != "" {
					player.Name = name
				}
			case 10:
				player.Traits = mergeTraitModifiers(player.Traits, parseTraitModifiers(ui.sbuild.EditBuffer, true))
			case 11:
				player.Traits = mergeTraitModifiers(player.Traits, parseTraitModifiers(ui.sbuild.EditBuffer, false))
			}
			ui.sbuild.Editing = false
		}
		if rl.IsKeyPressed(rl.KeyEscape) {
			ui.sbuild.Editing = false
		}
		return
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		if ui.sbuild.ReturnTo == 0 {
			ui.sbuild.ReturnTo = screenSetup
		}
		ui.screen = ui.sbuild.ReturnTo
		return
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		ui.sbuild.Cursor = wrapIndex(ui.sbuild.Cursor+1, 13)
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		ui.sbuild.Cursor = wrapIndex(ui.sbuild.Cursor-1, 13)
	}
	player := &ui.pcfg.Players[ui.sbuild.PlayerIndex]
	if rl.IsKeyPressed(rl.KeyLeft) {
		ui.adjustStatsBuilder(player, -1)
	}
	if rl.IsKeyPressed(rl.KeyRight) {
		ui.adjustStatsBuilder(player, 1)
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		switch ui.sbuild.Cursor {
		case 1:
			ui.sbuild.Editing = true
			ui.sbuild.EditBuffer = player.Name
			return
		case 10:
			ui.sbuild.Editing = true
			ui.sbuild.EditBuffer = formatTraitsList(player.Traits, true)
			return
		case 11:
			ui.sbuild.Editing = true
			ui.sbuild.EditBuffer = formatTraitsList(player.Traits, false)
			return
		case 12:
			if ui.sbuild.ReturnTo == 0 {
				ui.sbuild.ReturnTo = screenSetup
			}
			ui.screen = ui.sbuild.ReturnTo
			return
		}
		ui.adjustStatsBuilder(player, 1)
	}
}

func (ui *gameUI) adjustStatsBuilder(p *game.PlayerConfig, delta int) {
	switch ui.sbuild.Cursor {
	case 0:
		ui.sbuild.PlayerIndex = wrapIndex(ui.sbuild.PlayerIndex+delta, len(ui.pcfg.Players))
	case 1:
		// Name row is text-edit only.
	case 2:
		p.Strength = clampInt(p.Strength+delta, -3, 3)
		p.Endurance = p.Strength
	case 3:
		p.MentalStrength = clampInt(p.MentalStrength+delta, -3, 3)
		p.Mental = p.MentalStrength
	case 4:
		p.Agility = clampInt(p.Agility+delta, -3, 3)
		p.Bushcraft = p.Agility
	case 5:
		p.Hunting = clampInt(p.Hunting+delta*2, 0, 100)
	case 6:
		p.Fishing = clampInt(p.Fishing+delta*2, 0, 100)
	case 7:
		p.Foraging = clampInt(p.Foraging+delta*2, 0, 100)
	case 8:
		p.Crafting = clampInt(p.Crafting+delta*2, 0, 100)
	case 9:
		p.Gathering = clampInt(p.Gathering+delta*2, 0, 100)
	}
}

func (ui *gameUI) drawStatsBuilder() {
	if len(ui.pcfg.Players) == 0 {
		drawWrappedText("No players configured.", rl.NewRectangle(20, 20, float32(ui.width-40), float32(ui.height-40)), 50, 22, colorWarn)
		return
	}
	if ui.sbuild.PlayerIndex < 0 || ui.sbuild.PlayerIndex >= len(ui.pcfg.Players) {
		ui.sbuild.PlayerIndex = 0
	}
	p := ui.pcfg.Players[ui.sbuild.PlayerIndex]

	left := rl.NewRectangle(20, 20, float32(ui.width)*0.45, float32(ui.height-40))
	right := rl.NewRectangle(left.X+left.Width+16, 20, float32(ui.width)-left.Width-56, float32(ui.height-40))
	drawPanel(left, "Stats Builder")
	drawPanel(right, "Player Stats")

	rows := []struct {
		label string
		value string
	}{
		{"Player", playerSlotLabel(ui.sbuild.PlayerIndex, len(ui.pcfg.Players))},
		{"Name", playerNameWithYou(p, ui.sbuild.PlayerIndex)},
		{"Strength", fmt.Sprintf("%+d", p.Strength)},
		{"Mental Strength", fmt.Sprintf("%+d", p.MentalStrength)},
		{"Agility", fmt.Sprintf("%+d", p.Agility)},
		{"Hunting", fmt.Sprintf("%d", p.Hunting)},
		{"Fishing", fmt.Sprintf("%d", p.Fishing)},
		{"Foraging", fmt.Sprintf("%d", p.Foraging)},
		{"Crafting", fmt.Sprintf("%d", p.Crafting)},
		{"Gathering", fmt.Sprintf("%d", p.Gathering)},
		{"Positive Traits", formatTraitsList(p.Traits, true)},
		{"Negative Traits", formatTraitsList(p.Traits, false)},
		{"Back", "Enter"},
	}
	for i, row := range rows {
		y := int32(left.Y) + 52 + int32(i*46)
		if i == ui.sbuild.Cursor {
			rl.DrawRectangle(int32(left.X)+10, y-6, int32(left.Width)-20, 34, rl.Fade(colorAccent, 0.2))
		}
		rl.DrawText(row.label, int32(left.X)+20, y, 20, colorText)
		rl.DrawText(truncateForUI(row.value, 38), int32(left.X)+250, y, 20, colorAccent)
	}
	rl.DrawText("Up/Down move  Left/Right adjust  Enter edit/select", int32(left.X)+16, int32(left.Y+left.Height)-38, 18, colorDim)

	detail := []string{
		fmt.Sprintf("Mode: %s", modeLabel(ui.selectedMode())),
		fmt.Sprintf("Scenario: %s", ui.selectedScenario().Name),
		fmt.Sprintf("Player Slot: %s", playerSlotLabel(ui.sbuild.PlayerIndex, len(ui.pcfg.Players))),
		fmt.Sprintf("Name: %s", playerNameWithYou(p, ui.sbuild.PlayerIndex)),
		"",
		fmt.Sprintf("Strength %+d | Mental %+d | Agility %+d", p.Strength, p.MentalStrength, p.Agility),
		fmt.Sprintf("Hunt %d | Fish %d | Forage %d", p.Hunting, p.Fishing, p.Foraging),
		fmt.Sprintf("Craft %d | Gather %d", p.Crafting, p.Gathering),
		"",
		"Positive Traits:",
		"  " + truncateForUI(formatTraitsList(p.Traits, true), 64),
		"Negative Traits:",
		"  " + truncateForUI(formatTraitsList(p.Traits, false), 64),
		"",
		"Trait format examples:",
		"  Calm+2, Focused+1",
		"  Clumsy-1, Hotheaded-2",
		"",
		"Skills grow with play effort and success.",
	}
	drawLines(right, 44, 21, detail, colorText)

	if ui.sbuild.Editing {
		r := rl.NewRectangle(left.X+20, left.Y+left.Height-120, left.Width-40, 88)
		rl.DrawRectangleRounded(r, 0.2, 8, rl.Fade(colorPanel, 0.95))
		rl.DrawRectangleRoundedLinesEx(r, 0.2, 8, 2, colorAccent)
		rl.DrawText("Editing (Enter apply, Esc cancel)", int32(r.X)+12, int32(r.Y)+10, 18, colorAccent)
		rl.DrawText(truncateForUI(ui.sbuild.EditBuffer, 72)+"_", int32(r.X)+12, int32(r.Y)+42, 24, colorText)
	}
}

func (ui *gameUI) preparePlayerConfig() {
	ui.ensureSetupPlayers()
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
	ui.pcfg.ReturnTo = screenSetup
	ui.status = ""
}

func (ui *gameUI) updatePlayerConfig() {
	ui.ensureSetupPlayers()
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
		if ui.pcfg.ReturnTo == 0 {
			ui.pcfg.ReturnTo = screenSetup
		}
		ui.screen = ui.pcfg.ReturnTo
		return
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		ui.pcfg.Cursor = wrapIndex(ui.pcfg.Cursor+1, 4)
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		ui.pcfg.Cursor = wrapIndex(ui.pcfg.Cursor-1, 4)
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
		case 2:
			ui.openPersonalKitPicker(screenPlayerConfig)
		case 3:
			if ui.pcfg.ReturnTo == 0 {
				ui.pcfg.ReturnTo = screenSetup
			}
			ui.screen = ui.pcfg.ReturnTo
		}
	}
}

func (ui *gameUI) adjustPlayerConfig(p *game.PlayerConfig, delta int) {
	switch ui.pcfg.Cursor {
	case 0:
		ui.pcfg.PlayerIndex = wrapIndex(ui.pcfg.PlayerIndex+delta, len(ui.pcfg.Players))
	}
}

func (ui *gameUI) drawPlayerConfig() {
	left := rl.NewRectangle(20, 20, float32(ui.width)*0.42, float32(ui.height-40))
	right := rl.NewRectangle(left.X+left.Width+16, 20, float32(ui.width)-left.Width-56, float32(ui.height-40))
	drawPanel(left, "Player Config")
	drawPanel(right, "Player Details")

	if len(ui.pcfg.Players) == 0 {
		drawWrappedText("No players configured.", left, 50, 22, colorWarn)
		return
	}
	p := ui.pcfg.Players[ui.pcfg.PlayerIndex]

	rows := []struct {
		label string
		value string
	}{
		{"Player", playerSlotLabel(ui.pcfg.PlayerIndex, len(ui.pcfg.Players))},
		{"Name", playerNameWithYou(p, ui.pcfg.PlayerIndex)},
		{"Open Personal Kit Picker", "Enter"},
		{"Back", "Enter"},
	}
	for i, row := range rows {
		y := int32(left.Y) + 56 + int32(i*52)
		if i == ui.pcfg.Cursor {
			rl.DrawRectangle(int32(left.X)+10, y-6, int32(left.Width)-20, 36, rl.Fade(colorAccent, 0.2))
		}
		rl.DrawText(row.label, int32(left.X)+18, y, 22, colorText)
		rl.DrawText(truncateForUI(row.value, 40), int32(left.X)+290, y, 22, colorAccent)
	}

	if ui.pcfg.EditingName {
		r := rl.NewRectangle(left.X+20, left.Y+left.Height-104, left.Width-40, 72)
		rl.DrawRectangleRounded(r, 0.2, 8, rl.Fade(colorPanel, 0.95))
		rl.DrawRectangleRoundedLinesEx(r, 0.2, 8, 2, colorAccent)
		rl.DrawText("Editing Name", int32(r.X)+12, int32(r.Y)+10, 18, colorAccent)
		rl.DrawText(truncateForUI(ui.pcfg.NameBuffer, 52)+"_", int32(r.X)+12, int32(r.Y)+34, 24, colorText)
	}

	detailLines := []string{
		fmt.Sprintf("Mode: %s", modeLabel(ui.selectedMode())),
		fmt.Sprintf("Scenario: %s", ui.selectedScenario().Name),
		fmt.Sprintf("Player: %s", playerSlotLabel(ui.pcfg.PlayerIndex, len(ui.pcfg.Players))),
		fmt.Sprintf("Name: %s", playerNameWithYou(p, ui.pcfg.PlayerIndex)),
		"",
		"Personal Kit:",
		"  " + kitSummary(p.Kit, maxInt(1, p.KitLimit)),
		"",
		"Issued kit is assigned at game start",
		"based on scenario conditions.",
		"",
		"Controls:",
		"Up/Down move rows",
		"Left/Right switch player",
		"Enter select",
		"Esc back",
		"",
		"Reset personal kit in Kit Picker with R.",
	}
	drawLines(right, 44, 21, detailLines, colorText)
}

func (ui *gameUI) startRunFromConfig() {
	ui.ensureSetupPlayers()
	issuedKit := runtimeIssuedKit(ui.selectedMode(), ui.selectedScenario(), ui.pcfg.Players)
	ui.setup.IssuedKit = issuedKit
	cfg := game.RunConfig{
		Mode:        ui.selectedMode(),
		ScenarioID:  ui.selectedScenario().ID,
		PlayerCount: len(ui.pcfg.Players),
		Players:     append([]game.PlayerConfig(nil), ui.pcfg.Players...),
		IssuedKit:   append([]game.KitItem(nil), issuedKit...),
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
	ui.lastEntity = ""
	ui.pendingClarify = nil
	ui.pendingOptions = nil
	ui.status = ""
	ui.appendRunMessage("Run started")
	ui.appendRunMessage(fmt.Sprintf("Mode: %s | Scenario: %s | Players: %d", modeLabel(run.Config.Mode), run.Scenario.Name, len(run.Players)))
	ui.appendRunMessage(fmt.Sprintf("Issued kit assigned: %s", kitSummary(issuedKit, 0)))
	ui.screen = screenRun
}

func (ui *gameUI) updateLoad() {
	if rl.IsKeyPressed(rl.KeyEscape) {
		if ui.load.ReturnToRun && ui.run != nil {
			ui.screen = screenRun
		} else {
			ui.enterMenu()
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
		ui.lastEntity = ""
		ui.pendingClarify = nil
		ui.pendingOptions = nil
		ui.status = ""
		ui.runMessages = nil
		ui.appendRunMessage("Loaded " + filepath.Base(entry.Path))
		ui.screen = screenRun
	}
}

func (ui *gameUI) drawLoad() {
	left := rl.NewRectangle(20, 20, float32(ui.width)*0.35, float32(ui.height-40))
	right := rl.NewRectangle(left.X+left.Width+16, 20, float32(ui.width)-left.Width-56, float32(ui.height-40))
	drawPanel(left, "Load Game")
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
		"Temp: " + ui.formatTemperature(weather.TemperatureC),
		"",
		"Enter to load",
		"R to refresh",
		"Esc back",
	}
	drawLines(right, 48, 22, lines, colorText)
}

func (ui *gameUI) updateRun(delta time.Duration) {
	if ui.run == nil {
		ui.enterMenu()
		return
	}
	shiftDown := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
	if shiftDown && rl.IsKeyPressed(rl.KeyP) {
		ui.rplay.Cursor = ui.runFocus
		ui.screen = screenRunPlayers
		return
	}
	if shiftDown && rl.IsKeyPressed(rl.KeyH) {
		ui.screen = screenRunCommandLibrary
		return
	}
	if shiftDown && rl.IsKeyPressed(rl.KeyI) {
		ui.rinv.PlayerIndex = ui.runFocus
		ui.screen = screenRunInventory
		return
	}
	if ui.handleClarifyHotkeys() {
		return
	}
	ui.processIntentQueue()
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
			ui.appendRunMessage(fmt.Sprintf("Day %d started | Weather: %s | Temp: %s", ui.run.Day, weather, ui.formatTemperature(ui.run.Weather.TemperatureC)))
		}
	}

	if rl.IsKeyPressed(rl.KeyEscape) {
		ui.enterMenu()
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
	header := fmt.Sprintf("SURVIVE IT!   Mode: %s   Scenario: %s   Day: %d   Season: %s   Weather: %s   Temp: %s", modeLabel(ui.run.Config.Mode), ui.run.Scenario.Name, ui.run.Day, season, weather, ui.formatTemperature(ui.run.Weather.TemperatureC))
	rl.DrawText(header, int32(top.X)+14, int32(top.Y)+38, 21, colorAccent)
	nextIn := ui.autoDayDuration() - ui.runPlayedFor
	if nextIn < 0 {
		nextIn = 0
	}
	rl.DrawText(fmt.Sprintf("Game Time: Day %d  Clock %s  Auto-Next: %s", ui.run.Day, formatClockFromHours(ui.run.ClockHours), formatClockDuration(nextIn)), int32(top.X)+14, int32(top.Y)+66, 19, colorText)

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

	cmdHint := "Commands: next | save | load | menu | hunt ... | trap ... | gut ... | cook ... | eat ... | go ...   Shift+P players  Shift+H help  Shift+I inventory  Esc menu"
	textY := int32(bottom.Y) + 18
	if ui.pendingClarify != nil {
		clarify := ui.formatClarifyLine()
		lines := wrapText(clarify, 16, int32(bottom.Width)-28)
		for _, line := range lines {
			rl.DrawText(line, int32(bottom.X)+14, textY, 16, colorWarn)
			textY += 18
		}
	}
	rl.DrawText(cmdHint, int32(bottom.X)+14, textY, 17, colorDim)
	inputY := textY + 22
	in := strings.TrimSpace(ui.runInput)
	if in == "" {
		rl.DrawText("> ", int32(bottom.X)+14, inputY, 24, colorText)
	} else {
		rl.DrawText("> "+ui.runInput+"_", int32(bottom.X)+14, inputY, 24, colorAccent)
	}
	if strings.TrimSpace(ui.status) != "" {
		statusX := int32(bottom.X + bottom.Width*0.45)
		rl.DrawText(ui.status, statusX, inputY, 20, colorWarn)
	}
}

func (ui *gameUI) updateRunPlayers() {
	if ui.run == nil || len(ui.run.Players) == 0 {
		ui.screen = screenRun
		return
	}
	if ui.rplay.Cursor < 0 || ui.rplay.Cursor >= len(ui.run.Players) {
		ui.rplay.Cursor = 0
	}
	shiftDown := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
	if rl.IsKeyPressed(rl.KeyEscape) {
		ui.screen = screenRun
		return
	}
	if shiftDown && rl.IsKeyPressed(rl.KeyH) {
		ui.screen = screenRunCommandLibrary
		return
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		ui.rplay.Cursor = wrapIndex(ui.rplay.Cursor+1, len(ui.run.Players))
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		ui.rplay.Cursor = wrapIndex(ui.rplay.Cursor-1, len(ui.run.Players))
	}
}

func (ui *gameUI) drawRunPlayers() {
	if ui.run == nil || len(ui.run.Players) == 0 {
		area := rl.NewRectangle(20, 20, float32(ui.width-40), float32(ui.height-40))
		drawPanel(area, "Run Players")
		drawWrappedText("No active players.", area, 60, 24, colorWarn)
		return
	}

	left := rl.NewRectangle(20, 20, float32(ui.width)*0.37, float32(ui.height-40))
	right := rl.NewRectangle(left.X+left.Width+16, 20, float32(ui.width)-left.Width-56, float32(ui.height-40))
	drawPanel(left, "Run Players")
	drawPanel(right, "Player Detail")

	y := int32(left.Y) + 52
	for i, p := range ui.run.Players {
		if y > int32(left.Y+left.Height)-42 {
			break
		}
		if i == ui.rplay.Cursor {
			rl.DrawRectangle(int32(left.X)+10, y-6, int32(left.Width)-20, 56, rl.Fade(colorAccent, 0.18))
		}
		rl.DrawText(fmt.Sprintf("%d. %s", p.ID, p.Name), int32(left.X)+18, y, 20, colorAccent)
		y += 22
		rl.DrawText(fmt.Sprintf("E:%d  H2O:%d  M:%d  Hu:%d  Th:%d  Fa:%d  Ail:%d", p.Energy, p.Hydration, p.Morale, p.Hunger, p.Thirst, p.Fatigue, len(p.Ailments)),
			int32(left.X)+20, y, 17, colorText)
		y += 34
	}
	rl.DrawText("Up/Down move  Shift+H command library  Shift+I inventory  Esc back", int32(left.X)+14, int32(left.Y+left.Height)-30, 17, colorDim)

	sel := ui.run.Players[clampInt(ui.rplay.Cursor, 0, len(ui.run.Players)-1)]
	needs := game.DailyNutritionNeedsForPlayer(sel)

	lines := []string{
		fmt.Sprintf("%s (Player %d/%d)", sel.Name, sel.ID, len(ui.run.Players)),
		fmt.Sprintf("Sex: %s  Body: %s", sel.Sex, sel.BodyType),
		fmt.Sprintf("Height: %d ft %d in  Weight: %d kg", sel.HeightFt, sel.HeightIn, sel.WeightKg),
		fmt.Sprintf("Mods  End:%+d  Bush:%+d  Ment:%+d", sel.Endurance, sel.Bushcraft, sel.Mental),
		"",
		fmt.Sprintf("Energy:%d  Hydration:%d  Morale:%d", sel.Energy, sel.Hydration, sel.Morale),
		fmt.Sprintf("Reserves: %dkcal  %dgP  %dgF  %dgS", sel.CaloriesReserveKcal, sel.ProteinReserveG, sel.FatReserveG, sel.SugarReserveG),
		fmt.Sprintf("Needs/day: %dkcal  %dgP  %dgF  %dgS", needs.CaloriesKcal, needs.ProteinG, needs.FatG, needs.SugarG),
		fmt.Sprintf("Nutrition: %dkcal  %dgP  %dgF  %dgS", sel.Nutrition.CaloriesKcal, sel.Nutrition.ProteinG, sel.Nutrition.FatG, sel.Nutrition.SugarG),
		fmt.Sprintf("Deficit Streaks  Nutrition:%d  Dehydration:%d", sel.NutritionDeficitDays, sel.DehydrationDays),
	}
	drawLines(right, 42, 18, lines, colorText)

	barY := right.Y + 250
	barGap := float32(10)
	barW := right.Width - 30
	drawRunStatBar(rl.NewRectangle(right.X+14, barY, barW, 16), "Hunger", sel.Hunger, true)
	drawRunStatBar(rl.NewRectangle(right.X+14, barY+26+barGap, barW, 16), "Thirst", sel.Thirst, true)
	drawRunStatBar(rl.NewRectangle(right.X+14, barY+52+barGap*2, barW, 16), "Fatigue", sel.Fatigue, true)

	ailments := "None"
	if len(sel.Ailments) > 0 {
		parts := make([]string, 0, len(sel.Ailments))
		for _, ail := range sel.Ailments {
			name := ail.Name
			if strings.TrimSpace(name) == "" {
				name = string(ail.Type)
			}
			parts = append(parts, fmt.Sprintf("%s (%dd)", name, ail.DaysRemaining))
		}
		ailments = strings.Join(parts, ", ")
	}
	personalKit := "(none)"
	if len(sel.Kit) > 0 {
		parts := make([]string, 0, len(sel.Kit))
		for _, item := range sel.Kit {
			parts = append(parts, string(item))
		}
		personalKit = strings.Join(parts, ", ")
	}
	issuedKit := "(none)"
	if len(ui.run.Config.IssuedKit) > 0 {
		parts := make([]string, 0, len(ui.run.Config.IssuedKit))
		for _, item := range ui.run.Config.IssuedKit {
			parts = append(parts, string(item))
		}
		issuedKit = strings.Join(parts, ", ")
	}
	drawWrappedText("Ailments: "+ailments, right, 360, 18, colorWarn)
	drawWrappedText("Personal Kit: "+personalKit, right, 430, 18, colorText)
	drawWrappedText("Issued Kit: "+issuedKit, right, 500, 18, colorDim)
	drawWrappedText(fmt.Sprintf("Try: actions p%d   |   hunt fish p%d 300", sel.ID, sel.ID), right, int32(right.Height)-44, 17, colorDim)
}

func (ui *gameUI) updateRunCommandLibrary() {
	if rl.IsKeyPressed(rl.KeyEscape) {
		ui.screen = screenRun
		return
	}
	shiftDown := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
	if shiftDown && rl.IsKeyPressed(rl.KeyI) {
		ui.rinv.PlayerIndex = ui.runFocus
		ui.screen = screenRunInventory
		return
	}
}

func (ui *gameUI) drawRunCommandLibrary() {
	panel := rl.NewRectangle(20, 20, float32(ui.width-40), float32(ui.height-40))
	left := rl.NewRectangle(panel.X+8, panel.Y+38, panel.Width*0.5-14, panel.Height-46)
	right := rl.NewRectangle(left.X+left.Width+12, panel.Y+38, panel.Width-left.Width-26, panel.Height-46)
	drawPanel(panel, "Command Library")
	drawPanel(left, "Run Commands")
	drawPanel(right, "Actions & Shortcuts")

	leftLines := []string{
		"Core:",
		"next",
		"save",
		"load",
		"menu",
		"",
		"Hunting:",
		"hunt land [raw] [liver] [p#] [grams]",
		"hunt fish [raw] [liver] [p#] [grams]",
		"hunt air [raw] [liver] [p#] [grams]",
		"forage [roots|berries|fruits|vegetables|any] [p#] [grams]",
		"",
		"Camp Systems:",
		"trees",
		"wood gather|dry|stock",
		"resources",
		"plants",
		"collect <resource|any> [qty] [p#]",
		"bark strip [tree|any] [qty] [p#]",
		"inventory camp|personal|stash|take|add|drop",
		"trap list|set|status|check",
		"gut <carcass> [kg] [p#]",
		"cook <raw_meat> [kg] [p#]",
		"preserve <smoke|dry|salt> <meat> [kg] [p#]",
		"eat <food_item> [grams|kg] [p#]",
		"go <n|s|e|w> [km] [p#]",
		"fire status|methods|prep|ember|ignite|build|tend|out",
		"shelter list|build|status",
		"craft list|make|inventory",
	}
	rightLines := []string{
		"Equipment Actions:",
		"actions [p#]",
		"use <item> <action> [p#]",
		"Example: use paracord tie sticks together p1",
		"Example: use first aid kit treat wound p1",
		"Example: use rations eat p1",
		"",
		"Shortcuts:",
		"Shift+P  open player UX",
		"Shift+H  open command library",
		"Shift+I  open inventory UX",
		"S         save",
		"L         load",
		"Esc       back/menu",
		"",
		"Tip:",
		"Use 'help' or 'commands' in the run input",
		"to print the full command reference to history.",
	}
	drawLines(left, 44, 17, leftLines, colorText)
	drawLines(right, 44, 17, rightLines, colorText)
}

func (ui *gameUI) updateRunInventory() {
	if ui.run == nil {
		ui.screen = screenRun
		return
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		ui.screen = screenRun
		return
	}
	shiftDown := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
	if shiftDown && rl.IsKeyPressed(rl.KeyH) {
		ui.screen = screenRunCommandLibrary
		return
	}
	if shiftDown && rl.IsKeyPressed(rl.KeyP) {
		ui.rplay.Cursor = clampInt(ui.rinv.PlayerIndex, 0, len(ui.run.Players)-1)
		ui.screen = screenRunPlayers
		return
	}
	if len(ui.run.Players) == 0 {
		ui.rinv.PlayerIndex = 0
		return
	}
	ui.rinv.PlayerIndex = clampInt(ui.rinv.PlayerIndex, 0, len(ui.run.Players)-1)
	if rl.IsKeyPressed(rl.KeyDown) || rl.IsKeyPressed(rl.KeyTab) {
		ui.rinv.PlayerIndex = wrapIndex(ui.rinv.PlayerIndex+1, len(ui.run.Players))
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		ui.rinv.PlayerIndex = wrapIndex(ui.rinv.PlayerIndex-1, len(ui.run.Players))
	}
}

func (ui *gameUI) drawRunInventory() {
	panel := rl.NewRectangle(20, 20, float32(ui.width-40), float32(ui.height-40))
	left := rl.NewRectangle(panel.X+8, panel.Y+38, panel.Width*0.52-14, panel.Height-46)
	right := rl.NewRectangle(left.X+left.Width+12, panel.Y+38, panel.Width-left.Width-26, panel.Height-46)
	drawPanel(panel, "Run Inventory")
	drawPanel(left, "Camp / Shelter")
	drawPanel(right, "Personal Carry")

	campLines := []string{
		fmt.Sprintf("Day %d  Clock %s", ui.run.Day, formatClockFromHours(ui.run.ClockHours)),
		ui.run.CampInventorySummary(),
		"",
		"Traps:",
	}
	if len(ui.run.PlacedTraps) == 0 {
		campLines = append(campLines, "none")
	} else {
		for i, trap := range ui.run.PlacedTraps {
			state := "armed"
			if !trap.Armed {
				state = "awaiting check"
			}
			pending := ""
			if trap.PendingCatchKg > 0 {
				pending = fmt.Sprintf(" catch %.2fkg %s", trap.PendingCatchKg, trap.PendingCatchType)
			}
			campLines = append(campLines, fmt.Sprintf("#%d %s (%s) cond:%d%% %s%s", i+1, trap.Name, trap.Quality, trap.Condition, state, pending))
		}
	}
	campLines = append(campLines, "", "Quick commands:", "inventory camp", "inventory personal p#", "inventory stash/take ...", "trap check")
	drawLines(left, 44, 18, campLines, colorText)

	if len(ui.run.Players) == 0 {
		drawWrappedText("No players.", right, 44, 18, colorWarn)
		return
	}
	idx := clampInt(ui.rinv.PlayerIndex, 0, len(ui.run.Players)-1)
	player := ui.run.Players[idx]
	personal := ui.run.PersonalInventorySummary(player.ID)
	crafted := "none"
	if len(ui.run.CraftedItems) > 0 {
		crafted = strings.Join(ui.run.CraftedItems, ", ")
	}
	rightLines := []string{
		fmt.Sprintf("Player %d/%d: %s", idx+1, len(ui.run.Players), player.Name),
		personal,
		"",
		"Crafted items (recipes unlocked):",
		crafted,
		"",
		fmt.Sprintf("Carry stats  Str:%+d End:%+d Agi:%+d", player.Strength, player.Endurance, player.Agility),
		fmt.Sprintf("Skills  Craft:%d Gather:%d Hunt:%d Fish:%d Forage:%d", player.Crafting, player.Gathering, player.Hunting, player.Fishing, player.Foraging),
		"",
		"Up/Down cycle players",
		"Shift+P player detail",
		"Shift+H command library",
		"Esc back",
	}
	drawLines(right, 44, 18, rightLines, colorText)
}

func (ui *gameUI) processIntentQueue() {
	if ui.intentQueue == nil {
		return
	}
	for {
		intent, ok := ui.intentQueue.Dequeue()
		if !ok {
			return
		}
		ui.executeIntent(intent)
	}
}

func (ui *gameUI) executeIntent(intent parser.Intent) {
	if ui.run == nil {
		return
	}
	command := parser.IntentToCommandString(intent)
	if command == "" {
		ui.status = "No command to execute."
		return
	}
	verb := strings.ToLower(strings.TrimSpace(intent.Verb))
	prevDay := ui.run.Day
	prevClock := ui.run.ClockHours

	switch verb {
	case "next":
		ui.run.AdvanceDay()
		ui.runPlayedFor = 0
		ui.status = ""
	case "save":
		path := savePathForSlot(1)
		if err := saveRunToFile(path, *ui.run); err != nil {
			ui.status = "Save failed: " + err.Error()
			return
		}
		ui.status = "Saved to " + path
		ui.appendRunMessage(ui.status)
		ui.updateLastEntityFromIntent(intent, true)
		return
	case "load":
		ui.openLoad(true)
		ui.updateLastEntityFromIntent(intent, true)
		return
	case "menu", "back":
		ui.enterMenu()
		ui.updateLastEntityFromIntent(intent, true)
		return
	}

	if strings.HasPrefix(command, "hunt") || strings.HasPrefix(command, "catch") {
		ui.handleHuntCommand(command)
		ui.updateLastEntityFromIntent(intent, true)
		return
	}

	res := ui.run.ExecuteRunCommand(command)
	if res.Handled {
		ui.status = ""
		ui.appendRunMessage(res.Message)
		if res.HoursAdvanced > 0 {
			ui.appendRunMessage(fmt.Sprintf("Action time +%.1fh | Clock %s -> %s", res.HoursAdvanced, formatClockFromHours(prevClock), formatClockFromHours(ui.run.ClockHours)))
		}
		if ui.run.Day != prevDay {
			weather := game.WeatherLabel(ui.run.Weather.Type)
			ui.appendRunMessage(fmt.Sprintf("Day %d started | Weather: %s | Temp: %s", ui.run.Day, weather, ui.formatTemperature(ui.run.Weather.TemperatureC)))
		}
		ui.updateLastEntityFromIntent(intent, true)
		return
	}
	ui.status = "Unknown command"
	ui.updateLastEntityFromIntent(intent, false)
}

func (ui *gameUI) updateLastEntityFromIntent(intent parser.Intent, handled bool) {
	if !handled || len(intent.Args) == 0 {
		return
	}
	candidate := strings.TrimSpace(strings.ToLower(intent.Args[0]))
	if candidate == "" {
		return
	}
	if candidate == "north" || candidate == "south" || candidate == "east" || candidate == "west" ||
		candidate == "n" || candidate == "s" || candidate == "e" || candidate == "w" {
		return
	}
	if _, err := strconv.Atoi(candidate); err == nil {
		return
	}
	ui.lastEntity = candidate
}

func (ui *gameUI) buildParseContext() parser.ParseContext {
	ctx := parser.ParseContext{
		KnownDirections: []string{"north", "south", "east", "west", "n", "s", "e", "w"},
		LastEntity:      ui.lastEntity,
	}
	if ui.run == nil {
		return ctx
	}
	seenInv := map[string]bool{}
	addInv := func(v string) {
		v = strings.TrimSpace(strings.ToLower(v))
		if v == "" || seenInv[v] {
			return
		}
		seenInv[v] = true
		ctx.Inventory = append(ctx.Inventory, v)
	}
	seenNear := map[string]bool{}
	addNear := func(v string) {
		v = strings.TrimSpace(strings.ToLower(v))
		if v == "" || seenNear[v] {
			return
		}
		seenNear[v] = true
		ctx.Nearby = append(ctx.Nearby, v)
	}

	if len(ui.run.Players) > 0 {
		idx := clampInt(ui.runFocus, 0, len(ui.run.Players)-1)
		player := ui.run.Players[idx]
		for _, item := range player.PersonalItems {
			addInv(item.ID)
		}
		for _, item := range player.Kit {
			addInv(string(item))
		}
	}
	for _, item := range ui.run.Config.IssuedKit {
		addInv(string(item))
	}
	for _, item := range ui.run.CampInventory {
		addInv(item.ID)
	}
	for _, stock := range ui.run.ResourceStock {
		addNear(stock.ID)
	}
	for _, resource := range game.ResourcesForBiome(ui.run.Scenario.Biome) {
		addNear(resource.ID)
	}
	for _, tree := range game.TreesForBiome(ui.run.Scenario.Biome) {
		addNear(tree.ID)
		addNear(tree.Name)
	}
	for _, trap := range game.TrapsForBiome(ui.run.Scenario.Biome) {
		addNear(trap.ID)
		addNear(trap.Name)
	}

	return ctx
}

func (ui *gameUI) handleClarifyHotkeys() bool {
	if ui.pendingClarify == nil || len(ui.pendingOptions) == 0 {
		return false
	}
	idx := pressedClarifyIndex()
	if idx < 0 {
		return false
	}
	if idx >= len(ui.pendingOptions) {
		ui.status = "No such clarify option."
		return true
	}
	ui.selectClarifyOption(idx)
	return true
}

func pressedClarifyIndex() int {
	switch {
	case rl.IsKeyPressed(rl.KeyOne):
		return 0
	case rl.IsKeyPressed(rl.KeyTwo):
		return 1
	case rl.IsKeyPressed(rl.KeyThree):
		return 2
	case rl.IsKeyPressed(rl.KeyFour):
		return 3
	case rl.IsKeyPressed(rl.KeyFive):
		return 4
	case rl.IsKeyPressed(rl.KeySix):
		return 5
	case rl.IsKeyPressed(rl.KeySeven):
		return 6
	case rl.IsKeyPressed(rl.KeyEight):
		return 7
	case rl.IsKeyPressed(rl.KeyNine):
		return 8
	default:
		return -1
	}
}

func (ui *gameUI) selectClarifyOption(index int) {
	if index < 0 || index >= len(ui.pendingOptions) {
		return
	}
	choice := ui.pendingOptions[index]
	if ui.commandSink != nil {
		ui.commandSink.EnqueueIntent(choice)
	}
	ui.pendingClarify = nil
	ui.pendingOptions = nil
	ui.status = ""
	ui.appendRunMessage(fmt.Sprintf("Clarified: %s", parser.IntentToCommandString(choice)))
}

func (ui *gameUI) resolveTypedClarifyInput(raw string) (parser.Intent, bool) {
	n := strings.TrimSpace(strings.ToLower(raw))
	if n == "" {
		return parser.Intent{}, false
	}
	if v, err := strconv.Atoi(n); err == nil {
		idx := v - 1
		if idx >= 0 && idx < len(ui.pendingOptions) {
			return ui.pendingOptions[idx], true
		}
	}
	for _, option := range ui.pendingOptions {
		cmd := parser.IntentToCommandString(option)
		if n == cmd || strings.HasPrefix(cmd, n) {
			return option, true
		}
	}
	return parser.Intent{}, false
}

func (ui *gameUI) formatClarifyLine() string {
	if ui.pendingClarify == nil {
		return ""
	}
	parts := make([]string, 0, len(ui.pendingOptions)+1)
	parts = append(parts, ui.pendingClarify.Prompt)
	for i, option := range ui.pendingOptions {
		parts = append(parts, fmt.Sprintf("%d) %s", i+1, parser.IntentToCommandString(option)))
	}
	return strings.Join(parts, "  ")
}

func (ui *gameUI) submitRunInput() {
	if ui.run == nil {
		return
	}
	commandRaw := strings.TrimSpace(ui.runInput)
	ui.runInput = ""
	if commandRaw == "" {
		ui.status = "Enter a command."
		return
	}

	if ui.pendingClarify != nil {
		if len(ui.pendingOptions) == 0 {
			ui.pendingClarify = nil
		}
	}
	if ui.pendingClarify != nil {
		selected, ok := ui.resolveTypedClarifyInput(commandRaw)
		if !ok {
			ui.status = "Pick a clarify option by number or command text."
			return
		}
		if ui.commandSink != nil {
			ui.commandSink.EnqueueIntent(selected)
		}
		ui.pendingClarify = nil
		ui.pendingOptions = nil
		ui.status = ""
		return
	}

	ctx := ui.buildParseContext()
	intent := ui.cmdParser.Parse(ctx, commandRaw)
	if intent.Clarify != nil {
		ui.pendingClarify = intent.Clarify
		ui.pendingOptions = append([]parser.Intent(nil), intent.Clarify.Options...)
		ui.status = ""
		return
	}
	if ui.commandSink != nil {
		ui.commandSink.EnqueueIntent(intent)
		return
	}
	ui.status = "Command queue unavailable."
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

func maxKitLimitForMode(mode game.GameMode) int {
	return defaultKitLimitForMode(mode)
}

func issuedKitOptionsForMode(mode game.GameMode) []game.KitItem {
	switch mode {
	case game.ModeNakedAndAfraid:
		return []game.KitItem{
			game.KitSixInchKnife,
			game.KitMachete,
			game.KitParacord50ft,
			game.KitFerroRod,
			game.KitFirePlunger,
			game.KitMagnifyingLens,
			game.KitCookingPot,
			game.KitCanteen,
			game.KitWaterFilter,
			game.KitPurificationTablets,
			game.KitFishingLineHooks,
			game.KitSnareWire,
			game.KitTarp,
			game.KitMosquitoNet,
			game.KitInsectRepellent,
		}
	case game.ModeNakedAndAfraidXL:
		return []game.KitItem{
			game.KitSixInchKnife,
			game.KitMachete,
			game.KitHatchet,
			game.KitParacord50ft,
			game.KitFerroRod,
			game.KitFirePlunger,
			game.KitMagnifyingLens,
			game.KitCookingPot,
			game.KitMetalCup,
			game.KitCanteen,
			game.KitWaterFilter,
			game.KitPurificationTablets,
			game.KitFishingLineHooks,
			game.KitGillNet,
			game.KitSnareWire,
			game.KitBowArrows,
			game.KitTarp,
			game.KitMosquitoNet,
			game.KitInsectRepellent,
		}
	default:
		return []game.KitItem{
			game.KitHatchet,
			game.KitSixInchKnife,
			game.KitFoldingSaw,
			game.KitParacord50ft,
			game.KitFerroRod,
			game.KitFirePlunger,
			game.KitMagnifyingLens,
			game.KitCookingPot,
			game.KitCanteen,
			game.KitWaterFilter,
			game.KitPurificationTablets,
			game.KitFishingLineHooks,
			game.KitGillNet,
			game.KitSnareWire,
			game.KitBowArrows,
			game.KitTarp,
			game.KitSleepingBag,
			game.KitMosquitoNet,
		}
	}
}

func recommendedIssuedKitForScenario(mode game.GameMode, scenario game.Scenario) []game.KitItem {
	biome := strings.ToLower(strings.TrimSpace(scenario.Biome))
	kit := make([]game.KitItem, 0, 6)

	switch {
	case strings.Contains(biome, "winter"), strings.Contains(biome, "arctic"), strings.Contains(biome, "tundra"), strings.Contains(biome, "cold"):
		kit = append(kit, game.KitFerroRod, game.KitThermalLayer, game.KitHatchet, game.KitCookingPot)
	case strings.Contains(biome, "wet"), strings.Contains(biome, "swamp"), strings.Contains(biome, "jungle"), strings.Contains(biome, "rainforest"):
		kit = append(kit, game.KitParacord50ft, game.KitTarp, game.KitFerroRod, game.KitMosquitoNet)
	case strings.Contains(biome, "desert"), strings.Contains(biome, "savanna"), strings.Contains(biome, "dry"):
		kit = append(kit, game.KitCanteen, game.KitWaterFilter, game.KitHatchet, game.KitParacord50ft)
	default:
		kit = append(kit, game.KitHatchet, game.KitParacord50ft, game.KitFerroRod, game.KitCookingPot)
	}

	targetCount := 4
	if mode == game.ModeNakedAndAfraid {
		targetCount = 2
	}
	if mode == game.ModeNakedAndAfraidXL {
		targetCount = 3
	}

	allowed := issuedKitOptionsForMode(mode)
	filtered := filterKitItemsToAllowed(kit, allowed)
	if len(filtered) == 0 {
		filtered = append([]game.KitItem(nil), allowed...)
	}
	return firstNKitItems(filtered, targetCount)
}

func firstNKitItems(items []game.KitItem, n int) []game.KitItem {
	if n <= 0 || len(items) == 0 {
		return nil
	}
	if len(items) <= n {
		return append([]game.KitItem(nil), items...)
	}
	return append([]game.KitItem(nil), items[:n]...)
}

func filterKitItemsToAllowed(items []game.KitItem, allowed []game.KitItem) []game.KitItem {
	if len(items) == 0 || len(allowed) == 0 {
		return nil
	}
	allowedSet := make(map[game.KitItem]struct{}, len(allowed))
	for _, item := range allowed {
		allowedSet[item] = struct{}{}
	}

	out := make([]game.KitItem, 0, len(items))
	for _, item := range items {
		if _, ok := allowedSet[item]; !ok {
			continue
		}
		if hasKitItem(out, item) {
			continue
		}
		out = append(out, item)
	}
	return out
}

func runtimeIssuedKit(mode game.GameMode, scenario game.Scenario, players []game.PlayerConfig) []game.KitItem {
	recommended := recommendedIssuedKitForScenario(mode, scenario)
	if len(recommended) == 0 {
		return nil
	}
	personal := make(map[game.KitItem]struct{})
	for _, player := range players {
		for _, item := range player.Kit {
			personal[item] = struct{}{}
		}
	}
	filtered := make([]game.KitItem, 0, len(recommended))
	for _, item := range recommended {
		if _, exists := personal[item]; exists {
			continue
		}
		filtered = append(filtered, item)
	}
	if len(filtered) == 0 {
		return recommended
	}
	return filtered
}

func (ui *gameUI) ensureSetupPlayers() {
	mode := ui.selectedMode()
	ui.setup.PlayerCount = clampInt(ui.setup.PlayerCount, 1, 8)

	if len(ui.pcfg.Players) < ui.setup.PlayerCount {
		for len(ui.pcfg.Players) < ui.setup.PlayerCount {
			ui.pcfg.Players = append(ui.pcfg.Players, defaultPlayerConfig(len(ui.pcfg.Players), mode))
		}
	}
	if len(ui.pcfg.Players) > ui.setup.PlayerCount {
		ui.pcfg.Players = append([]game.PlayerConfig(nil), ui.pcfg.Players[:ui.setup.PlayerCount]...)
	}

	defaultLimit := defaultKitLimitForMode(mode)
	maxLimit := maxKitLimitForMode(mode)
	for i := range ui.pcfg.Players {
		player := &ui.pcfg.Players[i]
		if strings.TrimSpace(player.Name) == "" {
			player.Name = defaultPlayerConfig(i, mode).Name
		}
		if i == 0 && strings.TrimSpace(player.Name) == "" {
			player.Name = "You"
		}
		player.Strength = clampInt(player.Strength, -3, 3)
		player.MentalStrength = clampInt(player.MentalStrength, -3, 3)
		player.Agility = clampInt(player.Agility, -3, 3)
		player.Hunting = clampInt(player.Hunting, 0, 100)
		player.Fishing = clampInt(player.Fishing, 0, 100)
		player.Foraging = clampInt(player.Foraging, 0, 100)
		player.Crafting = clampInt(player.Crafting, 0, 100)
		player.Gathering = clampInt(player.Gathering, 0, 100)
		player.Endurance = player.Strength
		player.Mental = player.MentalStrength
		player.Bushcraft = player.Agility
		if player.KitLimit <= 0 {
			player.KitLimit = defaultLimit
		}
		if player.KitLimit > maxLimit {
			player.KitLimit = maxLimit
		}
		if len(player.Kit) > player.KitLimit {
			player.Kit = append([]game.KitItem(nil), player.Kit[:player.KitLimit]...)
		}
	}

	if !ui.setup.IssuedCustom {
		ui.setup.IssuedKit = recommendedIssuedKitForScenario(mode, ui.selectedScenario())
	} else {
		ui.setup.IssuedKit = filterKitItemsToAllowed(ui.setup.IssuedKit, issuedKitOptionsForMode(mode))
	}

	if len(ui.pcfg.Players) == 0 {
		ui.pcfg.PlayerIndex = 0
		ui.sbuild.PlayerIndex = 0
		return
	}
	ui.pcfg.PlayerIndex = clampInt(ui.pcfg.PlayerIndex, 0, len(ui.pcfg.Players)-1)
	ui.sbuild.PlayerIndex = clampInt(ui.sbuild.PlayerIndex, 0, len(ui.pcfg.Players)-1)
}

func (ui *gameUI) resetIssuedKitRecommendations() {
	ui.setup.IssuedCustom = false
	ui.setup.IssuedKit = recommendedIssuedKitForScenario(ui.selectedMode(), ui.selectedScenario())
}

func defaultPlayerConfig(i int, mode game.GameMode) game.PlayerConfig {
	defaultNames := []string{"Sophia", "Daniel", "Mia", "Jack", "Luna", "Leo", "Avery", "Harper"}
	name := fmt.Sprintf("Player %d", i+1)
	if i == 0 {
		name = "You"
	}
	if i < len(defaultNames) {
		name = defaultNames[i]
	}
	if i == 0 {
		name = "You"
	}
	return game.PlayerConfig{
		Name:           name,
		Sex:            game.SexOther,
		BodyType:       game.BodyTypeNeutral,
		WeightKg:       75,
		HeightFt:       5,
		HeightIn:       10,
		Endurance:      0,
		Bushcraft:      0,
		Mental:         0,
		Strength:       0,
		MentalStrength: 0,
		Agility:        0,
		Hunting:        10,
		Fishing:        10,
		Foraging:       10,
		Crafting:       10,
		Gathering:      10,
		Traits:         nil,
		KitLimit:       defaultKitLimitForMode(mode),
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

func truncateForUI(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func playerSlotLabel(idx, total int) string {
	if idx == 0 {
		return fmt.Sprintf("%d / %d (YOU)", idx+1, total)
	}
	return fmt.Sprintf("%d / %d", idx+1, total)
}

func playerNameWithYou(p game.PlayerConfig, idx int) string {
	name := strings.TrimSpace(p.Name)
	if name == "" {
		name = fmt.Sprintf("Player %d", idx+1)
	}
	if idx == 0 {
		return name + " (YOU)"
	}
	return name
}

func parseTraitModifiers(input string, positive bool) []game.TraitModifier {
	parts := strings.Split(input, ",")
	out := make([]game.TraitModifier, 0, len(parts))
	for _, part := range parts {
		raw := strings.TrimSpace(part)
		if raw == "" {
			continue
		}
		mod := 1
		name := raw
		for i := len(raw) - 1; i >= 0; i-- {
			if raw[i] == '+' || raw[i] == '-' {
				if i == len(raw)-1 {
					break
				}
				value, err := strconv.Atoi(strings.TrimSpace(raw[i:]))
				if err != nil {
					break
				}
				mod = value
				name = strings.TrimSpace(raw[:i])
				break
			}
		}
		if name == "" {
			continue
		}
		if positive && mod < 0 {
			mod = -mod
		}
		if !positive && mod > 0 {
			mod = -mod
		}
		out = append(out, game.TraitModifier{
			Name:     name,
			Modifier: mod,
			Positive: positive,
		})
	}
	return out
}

func mergeTraitModifiers(existing []game.TraitModifier, incoming []game.TraitModifier) []game.TraitModifier {
	out := make([]game.TraitModifier, 0, len(existing)+len(incoming))
	for _, trait := range existing {
		if strings.TrimSpace(trait.Name) == "" {
			continue
		}
		out = append(out, trait)
	}
	for _, trait := range incoming {
		if strings.TrimSpace(trait.Name) == "" {
			continue
		}
		replaced := false
		for i := range out {
			if strings.EqualFold(strings.TrimSpace(out[i].Name), strings.TrimSpace(trait.Name)) && out[i].Positive == trait.Positive {
				out[i] = trait
				replaced = true
				break
			}
		}
		if !replaced {
			out = append(out, trait)
		}
	}
	return out
}

func formatTraitsList(traits []game.TraitModifier, positive bool) string {
	filtered := make([]string, 0, len(traits))
	for _, trait := range traits {
		if positive && trait.Modifier <= 0 {
			continue
		}
		if !positive && trait.Modifier >= 0 {
			continue
		}
		filtered = append(filtered, fmt.Sprintf("%s%+d", strings.TrimSpace(trait.Name), trait.Modifier))
	}
	if len(filtered) == 0 {
		return "none"
	}
	return strings.Join(filtered, ", ")
}

func kitSummary(items []game.KitItem, limit int) string {
	if len(items) == 0 {
		if limit > 0 {
			return fmt.Sprintf("0/%d (none)", limit)
		}
		return "none"
	}

	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, string(item))
	}
	preview := strings.Join(parts, ", ")
	if len(parts) > 3 {
		preview = strings.Join(parts[:3], ", ") + fmt.Sprintf(" (+%d more)", len(parts)-3)
	}
	if limit > 0 {
		return fmt.Sprintf("%d/%d %s", len(items), limit, preview)
	}
	return fmt.Sprintf("%d %s", len(items), preview)
}

func temperatureUnitLabel(unit temperatureUnit) string {
	if unit == tempUnitF {
		return "Fahrenheit"
	}
	return "Celsius"
}

func celsiusToFahrenheit(c int) int {
	return int(math.Round(float64(c)*9.0/5.0 + 32.0))
}

func (ui *gameUI) formatTemperature(c int) string {
	if ui.opts.TempUnit == tempUnitF {
		return fmt.Sprintf("%dF", celsiusToFahrenheit(c))
	}
	return fmt.Sprintf("%dC", c)
}

func (ui *gameUI) formatTemperatureRange(minC int, maxC int) string {
	if ui.opts.TempUnit == tempUnitF {
		return fmt.Sprintf("%dF to %dF", celsiusToFahrenheit(minC), celsiusToFahrenheit(maxC))
	}
	return fmt.Sprintf("%dC to %dC", minC, maxC)
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

func formatClockFromHours(hours float64) string {
	if hours < 0 {
		hours = 0
	}
	whole := int(math.Floor(hours)) % 24
	minutes := int(math.Round((hours - math.Floor(hours)) * 60))
	if minutes >= 60 {
		minutes -= 60
		whole = (whole + 1) % 24
	}
	return fmt.Sprintf("%02d:%02d", whole, minutes)
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
