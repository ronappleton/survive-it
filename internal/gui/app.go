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
	screenProfiles
	screenScenarioPicker
	screenStatsBuilder
	screenPlayerConfig
	screenKitPicker
	screenScenarioBuilder
	screenPhaseEditor
	screenOptions
	screenAISettings
	screenLoad
	screenRun
	screenRunMap
	screenRunPlayers
	screenRunCommandLibrary
	screenRunInventory
)

type menuAction int

const (
	actionStart menuAction = iota
	actionLoad
	actionProfiles
	actionScenarioBuilder
	actionAI
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

type profilesState struct {
	Cursor     int
	EditingNew bool
	EditingID  string
	NameBuffer string
	Status     string
	ReturnTo   screen
}

type runSkillSnapshot struct {
	Hunting      int
	Fishing      int
	Foraging     int
	Trapping     int
	Firecraft    int
	Sheltercraft int
	Cooking      int
	Navigation   int
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
	profilesUI      profilesState
	ai              aiSettingsState
	customScenarios []game.Scenario

	run         *game.RunState
	runMessages []string
	runInput    string
	status      string
	runFocus    int
	lastEntity  string

	skillBaseline      map[int]runSkillSnapshot
	skillBaselineDay   int
	skillBaselineBlock string

	cmdParser     *parser.Parser
	commandSink   CommandSink
	intentQueue   *intentQueue
	pendingIntent *parser.PendingIntent

	updateAvailable      bool
	updateBusy           bool
	updateStatus         string
	menuNeedsUpdateCheck bool
	updateResultCh       chan updateResult

	lastTick     time.Time
	runPlayedFor time.Duration
	autoDayHours int

	profiles          []playerProfile
	selectedProfileID string
	runProfileID      string
}

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
		sb: scenarioBuilderState{
			EditingRow: -1,
		},
		skillBaselineDay: -1,
		updateResultCh:   make(chan updateResult, 4),
	}
	if !cfg.NoUpdate {
		ui.menuNeedsUpdateCheck = true
	}
	custom, _ := loadCustomScenarios(defaultCustomScenariosFile)
	ui.customScenarios = custom
	game.SetExternalScenarios(custom)
	ui.syncScenarioSelection()
	ui.ensureSetupPlayers()
	ui.initPlayerProfiles()
	ui.applySelectedProfileToSetupPrimary()
	ui.ensureSetupPlayers()
	ui.lastTick = time.Now()
	return ui
}

func (ui *gameUI) Run() error {
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagMsaa4xHint)
	rl.InitWindow(ui.width, ui.height, "survive-it")
	initTypography()
	defer shutdownTypography()
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
	case screenProfiles:
		ui.updateProfiles()
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
	case screenAISettings:
		ui.updateAISettings()
	case screenLoad:
		ui.updateLoad()
	case screenRun:
		ui.updateRun(delta)
	case screenRunMap:
		ui.updateRunMap()
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
	case screenProfiles:
		ui.drawProfiles()
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
	case screenAISettings:
		ui.drawAISettings()
	case screenLoad:
		ui.drawLoad()
	case screenRun:
		ui.drawRun()
	case screenRunMap:
		ui.drawRunMap()
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
			ui.applySelectedProfileToSetupPrimary()
			ui.ensureSetupPlayers()
			ui.status = ""
			ui.screen = screenSetup
		case actionLoad:
			ui.openLoad(false)
		case actionProfiles:
			ui.openProfilesScreen(screenMenu)
		case actionScenarioBuilder:
			ui.openScenarioBuilder()
		case actionAI:
			ui.openAISettings()
		case actionOptions:
			ui.screen = screenOptions
		case actionInstallUpdate:
			ui.triggerApplyUpdate()
		case actionQuit:
			ui.quit = true
		}
	}
	if ShiftPressedKey(rl.KeyQ) {
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
			drawText(item.Label, int32(r.X)+18, y+14, 28, colorAccent)
		} else {
			rl.DrawRectangleRounded(r, 0.3, 8, rl.Fade(colorPanel, 0.7))
			rl.DrawRectangleRoundedLinesEx(r, 0.3, 8, 1.5, colorBorder)
			drawText(item.Label, int32(r.X)+18, y+14, 28, colorText)
		}
	}

	hintRect := rl.NewRectangle(20, float32(ui.height-64), float32(ui.width-40), 40)
	drawTextCentered("Up/Down to move, Enter to select, Shift+Q to quit", hintRect, 8, typeScale.Small, colorDim)
}

func (ui *gameUI) menuItems() []menuItem {
	items := []menuItem{
		{Label: "Play Game", Action: actionStart},
		{Label: "Load Game", Action: actionLoad},
		{Label: "Player Profiles", Action: actionProfiles},
		{Label: "Scenario Builder", Action: actionScenarioBuilder},
		{Label: "AI", Action: actionAI},
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
		drawText(row.label, int32(left.X)+26, y, 24, colorText)
		drawText(row.value, int32(left.X)+286, y, 24, colorAccent)
	}
	drawText("Left/Right adjust values, Enter select, Esc back", int32(left.X)+22, int32(left.Y+left.Height)-38, 18, colorDim)

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
		ui.setup.Cursor = wrapIndex(ui.setup.Cursor+1, 9)
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		ui.setup.Cursor = wrapIndex(ui.setup.Cursor-1, 9)
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
		case 2:
			ui.openProfilesScreen(screenSetup)
		case 5:
			ui.openStatsBuilder(screenSetup)
		case 6:
			ui.preparePlayerConfig()
			ui.pcfg.ReturnTo = screenSetup
			ui.screen = screenPlayerConfig
		case 7:
			ui.startRunFromConfig()
		case 8:
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
		if len(ui.profiles) == 0 {
			return
		}
		current := ui.selectedProfileIndex()
		next := wrapIndex(current+delta, len(ui.profiles))
		ui.selectProfileByIndex(next)
		ui.applySelectedProfileToSetupPrimary()
	case 3:
		ui.setup.PlayerCount = clampInt(ui.setup.PlayerCount+delta, 1, 8)
	case 4:
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
		{"Player Profile", ui.selectedProfileSummary()},
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
		drawText(row.label, leftColX, y, 24, colorText)
		value := row.value
		if len(value) > rightColMax/9 {
			value = value[:maxInt(1, rightColMax/9-3)] + "..."
		}
		drawText(value, rightColX, y, 24, colorAccent)
	}
	drawText("Left/Right change   Enter select/open", int32(left.X)+26, int32(left.Y+left.Height)-38, 18, colorDim)

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
		drawText(scenario.Name, int32(left.X)+22, y, 22, clr)
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
		drawText(row.label, int32(left.X)+20, y, 20, colorText)
		drawText(truncateForUI(row.value, 38), int32(left.X)+250, y, 20, colorAccent)
	}
	drawText("Up/Down move  Left/Right adjust  Enter edit/select", int32(left.X)+16, int32(left.Y+left.Height)-38, 18, colorDim)

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
		drawText("Editing (Enter apply, Esc cancel)", int32(r.X)+12, int32(r.Y)+10, 18, colorAccent)
		drawText(truncateForUI(ui.sbuild.EditBuffer, 72)+"_", int32(r.X)+12, int32(r.Y)+42, 24, colorText)
	}
}

func (ui *gameUI) preparePlayerConfig() {
	ui.ensureSetupPlayers()
	if len(ui.pcfg.Players) != ui.setup.PlayerCount {
		ui.pcfg.Players = make([]game.PlayerConfig, ui.setup.PlayerCount)
		for i := range ui.pcfg.Players {
			ui.pcfg.Players[i] = defaultPlayerConfig(i, ui.selectedMode())
		}
		ui.applySelectedProfileToSetupPrimary()
		ui.ensureSetupPlayers()
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
		drawText(row.label, int32(left.X)+18, y, 22, colorText)
		drawText(truncateForUI(row.value, 40), int32(left.X)+290, y, 22, colorAccent)
	}

	if ui.pcfg.EditingName {
		r := rl.NewRectangle(left.X+20, left.Y+left.Height-104, left.Width-40, 72)
		rl.DrawRectangleRounded(r, 0.2, 8, rl.Fade(colorPanel, 0.95))
		rl.DrawRectangleRoundedLinesEx(r, 0.2, 8, 2, colorAccent)
		drawText("Editing Name", int32(r.X)+12, int32(r.Y)+10, 18, colorAccent)
		drawText(truncateForUI(ui.pcfg.NameBuffer, 52)+"_", int32(r.X)+12, int32(r.Y)+34, 24, colorText)
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
		"Reset personal kit in Kit Picker with Shift+R.",
	}
	drawLines(right, 44, 21, detailLines, colorText)
}

func (ui *gameUI) startRunFromConfig() {
	ui.ensureSetupPlayers()
	ui.applySelectedProfileToSetupPrimary()
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
	ui.skillBaseline = nil
	ui.skillBaselineDay = -1
	ui.skillBaselineBlock = ""
	ui.pendingIntent = nil
	ui.runProfileID = ui.selectedProfileID
	ui.status = ""
	ui.appendRunMessage("Run started")
	ui.appendRunMessage(fmt.Sprintf("Mode: %s | Scenario: %s | Players: %d", modeLabel(run.Config.Mode), run.Scenario.Name, len(run.Players)))
	ui.appendRunMessage(fmt.Sprintf("Issued kit assigned: %s", kitSummary(issuedKit, 0)))
	ui.screen = screenRun
	ui.syncRunPlayedForToMetabolism()
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
	if ShiftPressedKey(rl.KeyR) {
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
		ui.skillBaseline = nil
		ui.skillBaselineDay = -1
		ui.skillBaselineBlock = ""
		ui.pendingIntent = nil
		ui.runProfileID = ui.selectedProfileID
		ui.status = ""
		ui.runMessages = nil
		ui.appendRunMessage("Loaded " + filepath.Base(entry.Path))
		ui.screen = screenRun
		ui.syncRunPlayedForToMetabolism()
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
		drawText(filepath.Base(entry.Path), int32(left.X)+20, y, 20, colorText)
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
		"Shift+R to refresh",
		"Esc back",
	}
	drawLines(right, 48, 22, lines, colorText)
}

func (ui *gameUI) updateRun(delta time.Duration) {
	if ui.run == nil {
		ui.enterMenu()
		return
	}
	ui.run.EnsureTopology()
	ui.runFocus = 0
	if HotkeysEnabled(ui) && ShiftPressedKey(rl.KeyP) {
		ui.rplay.Cursor = 0
		ui.screen = screenRunPlayers
		return
	}
	if HotkeysEnabled(ui) && ShiftPressedKey(rl.KeyH) {
		ui.screen = screenRunCommandLibrary
		return
	}
	if HotkeysEnabled(ui) && ShiftPressedKey(rl.KeyI) {
		ui.rinv.PlayerIndex = 0
		ui.screen = screenRunInventory
		return
	}
	if HotkeysEnabled(ui) && ShiftPressedKey(rl.KeyM) {
		ui.screen = screenRunMap
		return
	}
	if ui.handlePendingIntentHotkeys() {
		return
	}
	ui.processIntentQueue()
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
		ui.leaveRunToMenu()
		return
	}
	if HotkeysEnabled(ui) && ShiftPressedKey(rl.KeyS) {
		path := savePathForSlot(1)
		if err := saveRunToFile(path, *ui.run); err != nil {
			ui.status = "Save failed: " + err.Error()
		} else {
			ui.status = "Saved to " + path
			ui.appendRunMessage(ui.status)
		}
	}
	if HotkeysEnabled(ui) && ShiftPressedKey(rl.KeyL) {
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
	ui.run.EnsureTopology()
	layout := runScreenLayout(ui.width, ui.height)
	drawPanel(layout.TopRect, "Run Status")
	drawPanel(layout.LogRect, "Message Log")
	drawPanel(layout.MiniMapRect, "Minimap")
	drawPanel(layout.InputRect, "")

	focus := game.PlayerState{}
	if len(ui.run.Players) > 0 {
		focus = ui.run.Players[0]
	}

	season := "unknown"
	if s, ok := ui.run.CurrentSeason(); ok {
		season = string(s)
	}
	weather := game.WeatherLabel(ui.run.Weather.Type)
	nextIn := ui.autoDayDuration() - ui.runPlayedFor
	if nextIn < 0 {
		nextIn = 0
	}
	posX, posY := ui.run.CurrentMapPosition()
	clockLine := fmt.Sprintf("Clock %s | Auto-Next %s | Position (%d,%d) | Time Block %s", formatClockFromHours(ui.run.ClockHours), formatClockDuration(nextIn), posX, posY, ui.run.CurrentTimeBlock())
	clockW := measureText(clockLine, typeScale.Small)
	clockX := int32(layout.TopRect.X+layout.TopRect.Width) - int32(spaceM) - int32(clockW)
	minClockX := int32(layout.TopRect.X) + 220
	if clockX < minClockX {
		clockX = minClockX
	}
	drawText(clockLine, clockX, int32(layout.TopRect.Y)+int32(spaceS), typeScale.Small, colorText)

	header := fmt.Sprintf("Mode: %s | Scenario: %s | Day: %d | Season: %s | Weather: %s | Temp: %s", modeLabel(ui.run.Config.Mode), ui.run.Scenario.Name, ui.run.Day, season, weather, ui.formatTemperature(ui.run.Weather.TemperatureC))
	drawText(header, int32(layout.TopRect.X)+14, int32(layout.TopRect.Y)+40, typeScale.Body, colorAccent)

	barInset := float32(14)
	barGap := float32(runLayoutGap)
	row1Y := layout.TopRect.Y + 88
	row2Y := row1Y + 36
	barLeft := layout.TopRect.X + barInset
	barRight := layout.TopRect.X + layout.TopRect.Width*0.58
	if barRight > layout.TopRect.X+layout.TopRect.Width-spaceM {
		barRight = layout.TopRect.X + layout.TopRect.Width - spaceM
	}
	if barRight < barLeft+300 {
		barRight = barLeft + 300
	}
	row1W := max(float32(90), (barRight-barLeft-barGap*3)/4)
	row2W := max(float32(90), (barRight-barLeft-barGap*2)/3)
	condition := runConditionScore(focus)
	drawRunStatBar(rl.NewRectangle(barLeft, row1Y, row1W, 8), "Condition", condition, false)
	drawRunStatBar(rl.NewRectangle(barLeft+(row1W+barGap)*1, row1Y, row1W, 8), "Energy", focus.Energy, false)
	drawRunStatBar(rl.NewRectangle(barLeft+(row1W+barGap)*2, row1Y, row1W, 8), "Hydration", focus.Hydration, false)
	drawRunStatBar(rl.NewRectangle(barLeft+(row1W+barGap)*3, row1Y, row1W, 8), "Morale", focus.Morale, false)
	drawRunStatBar(rl.NewRectangle(barLeft, row2Y, row2W, 8), "Hunger", focus.Hunger, true)
	drawRunStatBar(rl.NewRectangle(barLeft+(row2W+barGap)*1, row2Y, row2W, 8), "Thirst", focus.Thirst, true)
	drawRunStatBar(rl.NewRectangle(barLeft+(row2W+barGap)*2, row2Y, row2W, 8), "Fatigue", focus.Fatigue, true)

	ui.ensureRunSkillBaseline()
	skillValues := skillSnapshotFromPlayer(focus)
	skillDelta := ui.runSkillDelta(focus)
	skillsX := barRight + barGap + 4
	skillsY := row1Y - 2
	drawText("Skills", int32(skillsX), int32(skillsY), typeScale.Small, colorAccent)
	skillRows := []struct {
		Name  string
		Value int
		Delta int
	}{
		{Name: "Hunting", Value: skillValues.Hunting, Delta: skillDelta.Hunting},
		{Name: "Fishing", Value: skillValues.Fishing, Delta: skillDelta.Fishing},
		{Name: "Foraging", Value: skillValues.Foraging, Delta: skillDelta.Foraging},
		{Name: "Trapping", Value: skillValues.Trapping, Delta: skillDelta.Trapping},
		{Name: "Firecraft", Value: skillValues.Firecraft, Delta: skillDelta.Firecraft},
		{Name: "Shelter", Value: skillValues.Sheltercraft, Delta: skillDelta.Sheltercraft},
		{Name: "Cooking", Value: skillValues.Cooking, Delta: skillDelta.Cooking},
		{Name: "Navigation", Value: skillValues.Navigation, Delta: skillDelta.Navigation},
	}
	col1X := int32(skillsX)
	col2X := int32(skillsX + max(float32(170), layout.TopRect.Width*0.18))
	rowY := int32(skillsY) + 20
	for i, row := range skillRows {
		colX := col1X
		if i%2 == 1 {
			colX = col2X
		}
		if i%2 == 0 && i > 0 {
			rowY += 20
		}
		clr := colorText
		if row.Delta > 0 {
			clr = colorAccent
		}
		drawText(fmt.Sprintf("%s %d (%s)", row.Name, row.Value, formatSignedDelta(row.Delta)), colX, rowY, typeScale.Small, clr)
	}

	if len(ui.run.Players) > 1 {
		parts := make([]string, 0, len(ui.run.Players)-1)
		for _, p := range ui.run.Players[1:] {
			task := strings.TrimSpace(p.CurrentTask)
			if task == "" {
				task = "Idle"
			}
			parts = append(parts, fmt.Sprintf("%s: %s", p.Name, task))
		}
		team := strings.Join(parts, " | ")
		drawText("Team "+truncateForUI(team, 95), int32(layout.TopRect.X)+14, int32(layout.TopRect.Y+layout.TopRect.Height)-24, typeScale.Small, colorDim)
	}

	drawRunMessageLog(layout.LogRect, ui.runMessages)
	ui.drawMiniMap(layout.MiniMapRect, false)

	cmdHint := "Shortcuts: Shift+M map  Shift+P players  Shift+H help  Shift+I inventory  Shift+S save  Shift+L load"
	textY := int32(layout.InputRect.Y) + 18
	if ui.pendingIntent != nil {
		clarify := ui.formatPendingIntentLine()
		lines := wrapText(clarify, 16, int32(layout.InputRect.Width)-28)
		for _, line := range lines {
			drawText(line, int32(layout.InputRect.X)+14, textY, 16, colorWarn)
			textY += 18
		}
	}
	drawText(cmdHint, int32(layout.InputRect.X)+14, textY, 17, colorDim)
	inputY := textY + 22
	in := strings.TrimSpace(ui.runInput)
	if in == "" {
		drawText("> ", int32(layout.InputRect.X)+14, inputY, 24, colorText)
	} else {
		drawText("> "+ui.runInput+"_", int32(layout.InputRect.X)+14, inputY, 24, colorAccent)
	}
	if strings.TrimSpace(ui.status) != "" {
		statusX := int32(layout.InputRect.X + layout.InputRect.Width*0.5)
		drawText(ui.status, statusX, inputY, 20, colorWarn)
	}
}

func skillSnapshotFromPlayer(player game.PlayerState) runSkillSnapshot {
	return runSkillSnapshot{
		Hunting:      clampInt(player.Hunting, 0, 100),
		Fishing:      clampInt(player.Fishing, 0, 100),
		Foraging:     clampInt(player.Foraging, 0, 100),
		Trapping:     clampInt(player.Trapping, 0, 100),
		Firecraft:    clampInt(player.Firecraft, 0, 100),
		Sheltercraft: clampInt(player.Sheltercraft, 0, 100),
		Cooking:      clampInt(player.Cooking, 0, 100),
		Navigation:   clampInt(player.Navigation, 0, 100),
	}
}

func (ui *gameUI) ensureRunSkillBaseline() {
	if ui == nil || ui.run == nil {
		return
	}
	block := string(ui.run.CurrentTimeBlock())
	if ui.skillBaseline != nil && ui.skillBaselineDay == ui.run.Day && ui.skillBaselineBlock == block {
		return
	}
	ui.skillBaseline = make(map[int]runSkillSnapshot, len(ui.run.Players))
	for _, player := range ui.run.Players {
		ui.skillBaseline[player.ID] = skillSnapshotFromPlayer(player)
	}
	ui.skillBaselineDay = ui.run.Day
	ui.skillBaselineBlock = block
}

func (ui *gameUI) runSkillDelta(player game.PlayerState) runSkillSnapshot {
	current := skillSnapshotFromPlayer(player)
	if ui == nil {
		return runSkillSnapshot{}
	}
	if ui.skillBaseline == nil {
		return runSkillSnapshot{}
	}
	base, ok := ui.skillBaseline[player.ID]
	if !ok {
		ui.skillBaseline[player.ID] = current
		return runSkillSnapshot{}
	}
	return runSkillSnapshot{
		Hunting:      current.Hunting - base.Hunting,
		Fishing:      current.Fishing - base.Fishing,
		Foraging:     current.Foraging - base.Foraging,
		Trapping:     current.Trapping - base.Trapping,
		Firecraft:    current.Firecraft - base.Firecraft,
		Sheltercraft: current.Sheltercraft - base.Sheltercraft,
		Cooking:      current.Cooking - base.Cooking,
		Navigation:   current.Navigation - base.Navigation,
	}
}

func formatSignedDelta(v int) string {
	if v >= 0 {
		return fmt.Sprintf("+%d", v)
	}
	return fmt.Sprintf("%d", v)
}

func (ui *gameUI) updateRunPlayers() {
	if ui.run == nil || len(ui.run.Players) == 0 {
		ui.screen = screenRun
		return
	}
	if ui.rplay.Cursor < 0 || ui.rplay.Cursor >= len(ui.run.Players) {
		ui.rplay.Cursor = 0
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		ui.screen = screenRun
		return
	}
	if ShiftPressedKey(rl.KeyH) {
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
		drawText(fmt.Sprintf("%d. %s", p.ID, p.Name), int32(left.X)+18, y, 20, colorAccent)
		y += 22
		drawText(fmt.Sprintf("E:%d  H2O:%d  M:%d  Hu:%d  Th:%d  Fa:%d  Ail:%d", p.Energy, p.Hydration, p.Morale, p.Hunger, p.Thirst, p.Fatigue, len(p.Ailments)),
			int32(left.X)+20, y, 17, colorText)
		y += 34
	}
	drawText("Up/Down move  Shift+H command library  Shift+I inventory  Esc back", int32(left.X)+14, int32(left.Y+left.Height)-30, 17, colorDim)

	sel := ui.run.Players[clampInt(ui.rplay.Cursor, 0, len(ui.run.Players)-1)]
	needs := game.DailyNutritionNeedsForPlayer(sel)

	lines := []string{
		fmt.Sprintf("%s (Player %d/%d)", sel.Name, sel.ID, len(ui.run.Players)),
		fmt.Sprintf("Task: %s", safeText(sel.CurrentTask)),
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
	barW := clampFloat32(right.Width*0.2, 120, 210)
	drawRunStatBar(rl.NewRectangle(right.X+14, barY, barW, 8), "Hunger", sel.Hunger, true)
	drawRunStatBar(rl.NewRectangle(right.X+14, barY+26+barGap, barW, 8), "Thirst", sel.Thirst, true)
	drawRunStatBar(rl.NewRectangle(right.X+14, barY+52+barGap*2, barW, 8), "Fatigue", sel.Fatigue, true)

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
	drawWrappedText(fmt.Sprintf("Try: actions p%d   |   hunt fish p%d", sel.ID, sel.ID), right, int32(right.Height)-44, 17, colorDim)
}

func (ui *gameUI) updateRunCommandLibrary() {
	if rl.IsKeyPressed(rl.KeyEscape) {
		ui.screen = screenRun
		return
	}
	if ShiftPressedKey(rl.KeyI) {
		ui.rinv.PlayerIndex = 0
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
		"look [left|right|front|back]",
		"look closer at <plants|trees|insects|water>",
		"next",
		"save",
		"load",
		"menu",
		"",
		"Hunting:",
		"hunt land|fish|air [p#]",
		"fish [p#]",
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
		"ask <player> <task>",
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
		"Shift+M  open topology map",
		"Shift+S  save",
		"Shift+L  load",
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
	if ShiftPressedKey(rl.KeyH) {
		ui.screen = screenRunCommandLibrary
		return
	}
	if ShiftPressedKey(rl.KeyP) {
		ui.rplay.Cursor = clampInt(ui.rinv.PlayerIndex, 0, len(ui.run.Players)-1)
		ui.screen = screenRunPlayers
		return
	}
	if len(ui.run.Players) == 0 {
		ui.rinv.PlayerIndex = 0
		return
	}
	ui.rinv.PlayerIndex = clampInt(ui.rinv.PlayerIndex, 0, len(ui.run.Players)-1)
	if rl.IsKeyPressed(rl.KeyDown) || ShiftPressedKey(rl.KeyTab) {
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
		fmt.Sprintf("Skills  Hunt:%d Fish:%d Forage:%d Trap:%d Fire:%d", player.Hunting, player.Fishing, player.Foraging, player.Trapping, player.Firecraft),
		fmt.Sprintf("         Build:%d Cook:%d Nav:%d Craft:%d Gather:%d", player.Sheltercraft, player.Cooking, player.Navigation, player.Crafting, player.Gathering),
		"",
		"Up/Down or Shift+Tab cycle players",
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
		ui.syncRunPlayedForToMetabolism()
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
		ui.leaveRunToMenu()
		ui.updateLastEntityFromIntent(intent, true)
		return
	}

	res := ui.run.ExecuteRunCommand(command)
	if res.Handled {
		ui.syncRunPlayedForToMetabolism()
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
		player := ui.run.Players[0]
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

func resolveTypedPendingOptionInput(raw string, options []parser.Intent) (parser.Intent, bool) {
	n := strings.TrimSpace(strings.ToLower(raw))
	if n == "" {
		return parser.Intent{}, false
	}
	if v, err := strconv.Atoi(n); err == nil {
		idx := v - 1
		if idx >= 0 && idx < len(options) {
			return options[idx], true
		}
	}
	for _, option := range options {
		cmd := parser.IntentToCommandString(option)
		if n == cmd || strings.HasPrefix(cmd, n) {
			return option, true
		}
	}
	return parser.Intent{}, false
}

func (ui *gameUI) handlePendingIntentHotkeys() bool {
	if ui.pendingIntent == nil || len(ui.pendingIntent.Options) == 0 {
		return false
	}
	idx := pressedClarifyIndex()
	if idx < 0 {
		return false
	}
	if idx >= len(ui.pendingIntent.Options) {
		ui.status = "No such option."
		return true
	}
	choice := ui.pendingIntent.Options[idx]
	if ui.commandSink != nil {
		ui.commandSink.EnqueueIntent(choice)
	}
	ui.appendRunMessage(fmt.Sprintf("Clarified: %s", parser.IntentToCommandString(choice)))
	ui.clearPendingIntent()
	ui.status = ""
	return true
}

func (ui *gameUI) formatPendingIntentLine() string {
	if ui.pendingIntent == nil {
		return ""
	}
	parts := make([]string, 0, len(ui.pendingIntent.Options)+1)
	parts = append(parts, ui.pendingIntent.Prompt)
	for i, option := range ui.pendingIntent.Options {
		parts = append(parts, fmt.Sprintf("%d) %s", i+1, parser.IntentToCommandString(option)))
	}
	return strings.Join(parts, "  ")
}

func (ui *gameUI) setPendingIntent(p parser.PendingIntent) {
	copyPending := parser.PendingIntent{
		OriginalKind:  p.OriginalKind,
		OriginalVerb:  strings.ToLower(strings.TrimSpace(p.OriginalVerb)),
		FilledArgs:    append([]string(nil), p.FilledArgs...),
		MissingFields: append([]string(nil), p.MissingFields...),
		Prompt:        strings.TrimSpace(p.Prompt),
		Options:       append([]parser.Intent(nil), p.Options...),
	}
	ui.pendingIntent = &copyPending
	if copyPending.Prompt != "" {
		ui.appendRunMessage(copyPending.Prompt)
	}
	ui.status = ""
}

func (ui *gameUI) clearPendingIntent() {
	ui.pendingIntent = nil
}

func (ui *gameUI) repromptPendingIntent(prompt string, options []parser.Intent) {
	if ui.pendingIntent == nil {
		return
	}
	ui.pendingIntent.Prompt = strings.TrimSpace(prompt)
	if options != nil {
		ui.pendingIntent.Options = append([]parser.Intent(nil), options...)
	}
	if ui.pendingIntent.Prompt != "" {
		ui.appendRunMessage(ui.pendingIntent.Prompt)
		ui.status = ui.pendingIntent.Prompt
	}
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

	if ui.pendingIntent != nil {
		if strings.EqualFold(commandRaw, "cancel") {
			ui.clearPendingIntent()
			ui.status = "Cancelled."
			ui.appendRunMessage("Cancelled.")
			return
		}
		selected, ok := ui.resolvePendingIntentAnswer(commandRaw)
		if !ok {
			return
		}
		if ui.commandSink != nil {
			ui.commandSink.EnqueueIntent(selected)
		}
		ui.clearPendingIntent()
		ui.status = ""
		return
	}

	ctx := ui.buildParseContext()
	intent := ui.cmdParser.Parse(ctx, commandRaw)
	if pending, ok := ui.pendingIntentFromParsedIntent(intent, commandRaw); ok {
		ui.setPendingIntent(*pending)
		return
	}

	if ui.commandSink != nil {
		ui.commandSink.EnqueueIntent(intent)
		return
	}
	ui.status = "Command queue unavailable."
}

func (ui *gameUI) pendingIntentFromParsedIntent(intent parser.Intent, raw string) (*parser.PendingIntent, bool) {
	if pending, ok := pendingGoDistanceFromIntent(intent, raw); ok {
		return pending, true
	}
	if pending, ok := ui.pendingCraftItemFromIntent(intent); ok {
		return pending, true
	}
	if intent.Clarify == nil {
		return nil, false
	}
	return &parser.PendingIntent{
		OriginalKind:  intent.Kind,
		OriginalVerb:  intent.Verb,
		FilledArgs:    append([]string(nil), intent.Args...),
		MissingFields: parseMissingFieldsFromPrompt(intent.Verb, intent.Clarify.Prompt, intent.Clarify.Options),
		Prompt:        intent.Clarify.Prompt,
		Options:       append([]parser.Intent(nil), intent.Clarify.Options...),
	}, true
}

func pendingGoDistanceFromIntent(intent parser.Intent, raw string) (*parser.PendingIntent, bool) {
	if strings.ToLower(strings.TrimSpace(intent.Verb)) != "go" {
		return nil, false
	}
	if intent.Clarify != nil {
		return nil, false
	}
	if intent.Quantity != nil {
		return nil, false
	}
	direction := ""
	for _, arg := range intent.Args {
		if d := mapDirectionToken(arg); d != "" {
			direction = d
			break
		}
	}
	if direction == "" {
		return nil, false
	}
	if len(intent.Args) >= 2 {
		for _, arg := range intent.Args[1:] {
			if _, ok := game.ParseTravelDistanceInput(arg); ok {
				return nil, false
			}
		}
	}
	playerID := findPlayerToken(raw)
	filled := []string{direction}
	if playerID > 0 {
		filled = append(filled, fmt.Sprintf("p%d", playerID))
	}
	return &parser.PendingIntent{
		OriginalKind:  parser.Command,
		OriginalVerb:  "go",
		FilledArgs:    filled,
		MissingFields: []string{"distance"},
		Prompt:        "How far? (e.g. '500m', '3km', or '5 tiles')",
	}, true
}

func (ui *gameUI) pendingCraftItemFromIntent(intent parser.Intent) (*parser.PendingIntent, bool) {
	if strings.ToLower(strings.TrimSpace(intent.Verb)) != "craft" {
		return nil, false
	}
	args := make([]string, 0, len(intent.Args))
	for _, arg := range intent.Args {
		n := strings.ToLower(strings.TrimSpace(arg))
		if n != "" {
			args = append(args, n)
		}
	}
	if len(args) > 1 {
		return nil, false
	}
	if len(args) == 1 && args[0] != "make" {
		return nil, false
	}
	options := ui.craftPendingOptionsForRun()
	prompt := "What do you want to craft? Choose a number or type an item id."
	if len(options) == 0 {
		prompt = "What do you want to craft? Use 'craft list' for available ids."
	}
	return &parser.PendingIntent{
		OriginalKind:  parser.Command,
		OriginalVerb:  "craft",
		FilledArgs:    []string{"make"},
		MissingFields: []string{"craft_item"},
		Prompt:        prompt,
		Options:       options,
	}, true
}

func (ui *gameUI) resolvePendingIntentAnswer(raw string) (parser.Intent, bool) {
	if ui == nil || ui.pendingIntent == nil {
		return parser.Intent{}, false
	}
	pending := ui.pendingIntent
	if len(pending.Options) > 0 {
		if selected, ok := resolveTypedPendingOptionInput(raw, pending.Options); ok {
			return selected, true
		}
	}
	switch strings.ToLower(strings.TrimSpace(pending.OriginalVerb)) {
	case "go":
		if pendingMissingField(pending, "distance") {
			return ui.resolvePendingGoDistanceAnswer(raw)
		}
	case "craft":
		if pendingMissingField(pending, "craft_item") {
			return ui.resolvePendingCraftItemAnswer(raw)
		}
	}

	ui.repromptPendingIntent("Please answer the pending prompt or type 'cancel'.", nil)
	return parser.Intent{}, false
}

func (ui *gameUI) resolvePendingGoDistanceAnswer(raw string) (parser.Intent, bool) {
	if ui == nil || ui.pendingIntent == nil {
		return parser.Intent{}, false
	}
	pending := ui.pendingIntent
	if len(pending.FilledArgs) == 0 {
		ui.repromptPendingIntent("Direction is missing; re-enter movement command.", nil)
		return parser.Intent{}, false
	}

	km, ok := game.ParseTravelDistanceInput(raw)
	if !ok || km <= 0 {
		ui.repromptPendingIntent("Distance required (e.g. 500m, 3km, 5 tiles).", nil)
		return parser.Intent{}, false
	}
	qraw := strings.ToLower(strings.TrimSpace(raw))
	if qraw == "" {
		ui.repromptPendingIntent("Distance required (e.g. 500m, 3km, 5 tiles).", nil)
		return parser.Intent{}, false
	}
	args := append([]string(nil), pending.FilledArgs...)
	base := "go " + strings.Join(args, " ")
	return parser.Intent{
		Raw:        strings.TrimSpace(base + " " + qraw),
		Normalised: strings.TrimSpace(base + " " + qraw),
		Kind:       parser.Command,
		Verb:       "go",
		Args:       args,
		Quantity:   &parser.Quantity{Raw: qraw, N: int(math.Round(km * 1000)), Unit: "distance"},
		Confidence: 0.98,
	}, true
}

func (ui *gameUI) resolvePendingCraftItemAnswer(raw string) (parser.Intent, bool) {
	if ui == nil || ui.pendingIntent == nil {
		return parser.Intent{}, false
	}
	pending := ui.pendingIntent
	if selected, ok := resolveTypedPendingOptionInput(raw, pending.Options); ok {
		return selected, true
	}

	input := normalizeCraftAnswerInput(raw)
	if input == "" {
		ui.repromptPendingIntent("Choose a craft item by number or item id (or type 'cancel').", nil)
		return parser.Intent{}, false
	}

	options := game.CraftablesForBiome(ui.run.Scenario.Biome)
	if len(options) == 0 {
		ui.repromptPendingIntent("No craftables available here.", nil)
		return parser.Intent{}, false
	}

	sort.Slice(options, func(i, j int) bool {
		return strings.ToLower(options[i].Name) < strings.ToLower(options[j].Name)
	})

	exact := make([]game.CraftableSpec, 0, 2)
	prefix := make([]game.CraftableSpec, 0, 4)
	contains := make([]game.CraftableSpec, 0, 4)
	for _, spec := range options {
		id := strings.ToLower(strings.TrimSpace(spec.ID))
		name := strings.ToLower(strings.TrimSpace(spec.Name))
		switch {
		case input == id || input == name:
			exact = append(exact, spec)
		case strings.HasPrefix(id, input) || strings.HasPrefix(name, input):
			prefix = append(prefix, spec)
		case len(input) >= 3 && (strings.Contains(id, input) || strings.Contains(name, input)):
			contains = append(contains, spec)
		}
	}

	matches := exact
	if len(matches) == 0 {
		matches = prefix
	}
	if len(matches) == 0 {
		matches = contains
	}
	if len(matches) == 0 {
		ui.repromptPendingIntent("Unknown craft item. Type an item id or choose a numbered option.", ui.craftPendingOptionsForRun())
		return parser.Intent{}, false
	}
	if len(matches) > 1 {
		ui.repromptPendingIntent("Multiple craft items match. Choose one by number.", intentsFromCraftables(matches, 6))
		return parser.Intent{}, false
	}

	args := append([]string(nil), pending.FilledArgs...)
	args = append(args, strings.ToLower(strings.TrimSpace(matches[0].ID)))
	base := strings.TrimSpace("craft " + strings.Join(args, " "))
	return parser.Intent{
		Raw:        base,
		Normalised: base,
		Kind:       parser.Command,
		Verb:       "craft",
		Args:       args,
		Confidence: 0.93,
	}, true
}

func (ui *gameUI) craftPendingOptionsForRun() []parser.Intent {
	if ui == nil || ui.run == nil {
		return nil
	}
	options := game.CraftablesForBiome(ui.run.Scenario.Biome)
	return intentsFromCraftables(options, 6)
}

func intentsFromCraftables(items []game.CraftableSpec, limit int) []parser.Intent {
	if len(items) == 0 || limit <= 0 {
		return nil
	}
	options := append([]game.CraftableSpec(nil), items...)
	sort.Slice(options, func(i, j int) bool {
		return strings.ToLower(options[i].Name) < strings.ToLower(options[j].Name)
	})
	if len(options) > limit {
		options = options[:limit]
	}
	out := make([]parser.Intent, 0, len(options))
	for _, spec := range options {
		id := strings.ToLower(strings.TrimSpace(spec.ID))
		if id == "" {
			continue
		}
		out = append(out, parser.Intent{
			Kind:       parser.Command,
			Verb:       "craft",
			Args:       []string{"make", id},
			Confidence: 0.9,
		})
	}
	return out
}

func normalizeCraftAnswerInput(raw string) string {
	n := strings.ToLower(strings.TrimSpace(raw))
	n = strings.TrimPrefix(n, "craft make ")
	n = strings.TrimPrefix(n, "craft ")
	n = strings.TrimPrefix(n, "make ")
	return strings.TrimSpace(n)
}

func parseMissingFieldsFromPrompt(verb, prompt string, options []parser.Intent) []string {
	v := strings.ToLower(strings.TrimSpace(verb))
	if v == "craft" {
		if strings.Contains(strings.ToLower(prompt), "needs at least") || strings.Contains(strings.ToLower(prompt), "what do you want to craft") {
			return []string{"craft_item"}
		}
	}
	if v == "go" && strings.Contains(strings.ToLower(prompt), "how far") {
		return []string{"distance"}
	}

	n := strings.ToLower(strings.TrimSpace(prompt))
	switch {
	case strings.Contains(n, "which direction"):
		return []string{"direction"}
	case strings.Contains(n, "needs at least"):
		return []string{"argument"}
	case strings.Contains(n, "what should i"):
		return []string{"entity"}
	case len(options) > 0:
		return []string{"selection"}
	default:
		return []string{"clarification"}
	}
}

func pendingMissingField(pending *parser.PendingIntent, want string) bool {
	if pending == nil {
		return false
	}
	want = strings.ToLower(strings.TrimSpace(want))
	for _, field := range pending.MissingFields {
		if strings.ToLower(strings.TrimSpace(field)) == want {
			return true
		}
	}
	return false
}

func mapDirectionToken(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
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

func findPlayerToken(raw string) int {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(raw)))
	for _, field := range fields {
		if len(field) < 2 || field[0] != 'p' {
			continue
		}
		if v, err := strconv.Atoi(strings.TrimPrefix(field, "p")); err == nil && v > 0 {
			return v
		}
	}
	return 1
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
	DrawPanel(rect, title, false)
}

func drawTextCentered(text string, rect rl.Rectangle, yOffset int32, fontSize int32, clr rl.Color) {
	width := measureText(text, fontSize)
	x := int32(rect.X + (rect.Width-float32(width))/2)
	drawText(text, x, int32(rect.Y)+yOffset, fontSize, clr)
}

func drawWrappedText(text string, rect rl.Rectangle, y int32, size int32, clr rl.Color) {
	maxWidth := int32(rect.Width - spaceM*2)
	lines := wrapText(text, size, maxWidth)
	lineStep := textLineHeight(size)
	for i, line := range lines {
		drawText(line, int32(rect.X+spaceM), int32(rect.Y)+y+int32(i)*lineStep, size, clr)
	}
}

func drawLines(rect rl.Rectangle, y int32, size int32, lines []string, clr rl.Color) {
	lineStep := textLineHeight(size)
	for i, line := range lines {
		drawText(line, int32(rect.X+spaceM), int32(rect.Y)+y+int32(i)*lineStep, size, clr)
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
		if measureText(candidate, size) <= maxWidth {
			current = candidate
			continue
		}
		lines = append(lines, current)
		current = word
	}
	lines = append(lines, current)
	return lines
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
	if !scenarioNeedsIssuedKit(biome) {
		return nil
	}
	kit := make([]game.KitItem, 0, 4)

	switch {
	case strings.Contains(biome, "winter"), strings.Contains(biome, "arctic"), strings.Contains(biome, "tundra"), strings.Contains(biome, "cold"):
		kit = append(kit, game.KitFerroRod, game.KitThermalLayer, game.KitHatchet, game.KitCookingPot)
	case strings.Contains(biome, "desert"), strings.Contains(biome, "arid"):
		kit = append(kit, game.KitCanteen, game.KitWaterFilter, game.KitPurificationTablets, game.KitHatchet)
	default:
		return nil
	}

	targetCount := 2
	if mode == game.ModeNakedAndAfraid {
		targetCount = 1
	}
	if mode == game.ModeNakedAndAfraidXL {
		targetCount = 1
	}

	allowed := issuedKitOptionsForMode(mode)
	filtered := filterKitItemsToAllowed(kit, allowed)
	filtered = uniqueKitItemsList(filtered)
	return firstNKitItems(filtered, targetCount)
}

func scenarioNeedsIssuedKit(biome string) bool {
	return strings.Contains(biome, "winter") ||
		strings.Contains(biome, "arctic") ||
		strings.Contains(biome, "tundra") ||
		strings.Contains(biome, "cold") ||
		strings.Contains(biome, "desert") ||
		strings.Contains(biome, "arid")
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
	recommended := uniqueKitItemsList(recommendedIssuedKitForScenario(mode, scenario))
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
		if hasKitItem(filtered, item) {
			continue
		}
		filtered = append(filtered, item)
	}
	return uniqueKitItemsList(filtered)
}

func uniqueKitItemsList(items []game.KitItem) []game.KitItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]game.KitItem, 0, len(items))
	for _, item := range items {
		if hasKitItem(out, item) {
			continue
		}
		out = append(out, item)
	}
	return out
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
		player.Trapping = clampInt(player.Trapping, 0, 100)
		player.Firecraft = clampInt(player.Firecraft, 0, 100)
		player.Sheltercraft = clampInt(player.Sheltercraft, 0, 100)
		player.Cooking = clampInt(player.Cooking, 0, 100)
		player.Navigation = clampInt(player.Navigation, 0, 100)
		if strings.TrimSpace(player.CurrentTask) == "" {
			player.CurrentTask = "Idle"
		}
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
		Trapping:       10,
		Firecraft:      10,
		Sheltercraft:   10,
		Cooking:        10,
		Navigation:     10,
		CurrentTask:    "Idle",
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

func clampFloat32(v float32, min float32, max float32) float32 {
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

func (ui *gameUI) syncRunPlayedForToMetabolism() {
	if ui == nil || ui.run == nil {
		return
	}
	dayDuration := ui.autoDayDuration()
	if dayDuration <= 0 {
		return
	}
	progress := ui.run.MetabolismProgress
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	ui.runPlayedFor = time.Duration(float64(dayDuration) * progress)
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
	thresholds := TelemetryThresholds{Warning: 35, Danger: 20, Inverted: false}
	if danger {
		thresholds = TelemetryThresholds{Warning: 40, Danger: 70, Inverted: true}
	}
	DrawTelemetryBar(label, value, rect, thresholds)
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
