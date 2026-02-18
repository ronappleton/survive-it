package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/appengine-ltd/survive-it/internal/game"
	"github.com/appengine-ltd/survive-it/internal/update"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
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

type screen int

const (
	screenMenu screen = iota
	screenSetup
	screenScenarioPicker
	screenStatsBuilder
	screenPlayerConfig
	screenKitPicker
	screenLoadRun
	screenScenarioBuilder
	screenBuilderScenarioPicker
	screenOptions
	screenPhaseEditor
	screenRun
)

func NewApp(cfg AppConfig) *App {
	return &App{cfg: cfg}
}

func (a *App) Run() error {
	m := newMenuModel(a.cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// --- Styles (retro green) ---
var (
	green       = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	brightGreen = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	dimGreen    = lipgloss.NewStyle().Foreground(lipgloss.Color("22"))
	border      = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	dosBox      = lipgloss.Border{
		Top:          "-",
		Bottom:       "-",
		Left:         "|",
		Right:        "|",
		TopLeft:      "+",
		TopRight:     "+",
		BottomLeft:   "+",
		BottomRight:  "+",
		MiddleLeft:   "|",
		MiddleRight:  "|",
		Middle:       "+",
		MiddleTop:    "+",
		MiddleBottom: "+",
	}
	saveFilePattern = regexp.MustCompile(`^survive-it-save-[a-zA-Z0-9._-]+\.json$`)
)

// --- Menu model ---

type menuItem int

const (
	itemStart menuItem = iota
	itemLoadRun
	itemScenarioBuilder
	itemOptions
	itemInstallUpdate
	itemQuit
)

type menuEntry struct {
	label  string
	action menuItem
}

const (
	defaultCustomScenariosFile = "survive-it-scenarios.json"
	saveSlotCount              = 3
	maxSaveFileBytes           = 2 << 20
	maxScenarioFileBytes       = 2 << 20
)

type menuModel struct {
	w, h   int
	cfg    AppConfig
	idx    int
	screen screen
	setup  setupState
	pick   scenarioPickerState
	bpick  builderScenarioPickerState
	sbuild statsBuilderState
	pcfg   playerConfigState
	kit    kitPickerState
	load   loadRunState
	build  scenarioBuilderState
	phase  phaseEditorState
	opts   optionsState

	run             *game.RunState
	runInput        string
	activeSaveSlot  int
	customScenarios []customScenarioRecord
	loadReturn      screen
	status          string
	busy            bool
	updateAvailable bool
	updateStatus    string
	err             string

	lastTickAt   time.Time
	runPlayedFor time.Duration
}

func newMenuModel(cfg AppConfig) menuModel {
	customScenarios, _ := loadCustomScenarios(defaultCustomScenariosFile)

	return menuModel{
		cfg:             cfg,
		idx:             0,
		setup:           newSetupState(),
		pick:            newScenarioPickerState(),
		bpick:           newBuilderScenarioPickerState(),
		sbuild:          newStatsBuilderState(),
		pcfg:            newPlayerConfigState(),
		kit:             newKitPickerState(),
		load:            newLoadRunState(),
		build:           newScenarioBuilderState(),
		phase:           newPhaseEditorState(),
		opts:            newOptionsState(),
		activeSaveSlot:  1,
		customScenarios: customScenarios,
	}
}

func (m menuModel) Init() tea.Cmd {
	// Approximate 1024x768 using typical terminal cells at ~8x16 px.
	resizeCmd := resizeTerminalBestEffort(128, 48)
	if m.cfg.NoUpdate {
		return tea.Batch(resizeCmd, clockTickCmd())
	}
	return tea.Batch(resizeCmd, checkUpdateCmd(m.cfg.Version, true), clockTickCmd())
}

type updateResultMsg struct {
	status    string
	err       error
	available bool
	auto      bool
}

type clockTickMsg struct {
	at time.Time
}

type savedRun struct {
	FormatVersion int           `json:"format_version"`
	SavedAt       time.Time     `json:"saved_at"`
	Run           game.RunState `json:"run"`
}

type customScenarioRecord struct {
	Scenario      game.Scenario `json:"scenario"`
	PreferredMode game.GameMode `json:"preferred_mode"`
	Notes         string        `json:"notes,omitempty"`
}

type customScenarioLibrary struct {
	FormatVersion int                    `json:"format_version"`
	Custom        []customScenarioRecord `json:"custom,omitempty"`
	Scenarios     []game.Scenario        `json:"scenarios,omitempty"` // legacy v1 support
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.screen == screenRun {
			return m.updateRun(msg)
		}
		if m.screen == screenSetup {
			return m.updateSetup(msg)
		}
		if m.screen == screenScenarioPicker {
			return m.updateScenarioPicker(msg)
		}
		if m.screen == screenStatsBuilder {
			return m.updateStatsBuilder(msg)
		}
		if m.screen == screenPlayerConfig {
			return m.updatePlayerConfig(msg)
		}
		if m.screen == screenKitPicker {
			return m.updateKitPicker(msg)
		}
		if m.screen == screenLoadRun {
			return m.updateLoadRun(msg)
		}
		if m.screen == screenScenarioBuilder {
			return m.updateScenarioBuilder(msg)
		}
		if m.screen == screenBuilderScenarioPicker {
			return m.updateBuilderScenarioPicker(msg)
		}
		if m.screen == screenOptions {
			return m.updateOptions(msg)
		}
		if m.screen == screenPhaseEditor {
			return m.updatePhaseEditor(msg)
		}

		return m.updateMenu(msg)
	case updateResultMsg:
		m.busy = false
		if msg.err != nil {
			if !msg.auto {
				m.status = fmt.Sprintf("Update check failed: %v", msg.err)
			}
			m.updateAvailable = false
			m.updateStatus = ""

			return m, nil
		}
		m.updateAvailable = msg.available
		m.updateStatus = msg.status
		if !msg.auto {
			m.status = msg.status
		} else if msg.available {
			m.status = ""
		}

		return m, nil
	case clockTickMsg:
		if m.lastTickAt.IsZero() {
			m.lastTickAt = msg.at
			return m, clockTickCmd()
		}

		delta := msg.at.Sub(m.lastTickAt)
		m.lastTickAt = msg.at
		if delta < 0 {
			delta = 0
		}

		if m.screen == screenRun && m.run != nil {
			m.runPlayedFor += delta
			dayDuration := m.autoDayDuration()
			for m.runPlayedFor >= dayDuration {
				m.advanceRunDay()
				m.runPlayedFor -= dayDuration
				out := m.run.EvaluateRun()
				if out.Status != game.RunOutcomeOngoing {
					break
				}
			}
		}

		return m, clockTickCmd()
	case tea.WindowSizeMsg:
		m.w = msg.Width
		m.h = msg.Height

		return m, nil
	}

	return m, nil
}

func (m menuModel) View() string {
	if m.screen == screenSetup {
		return m.viewSetup()
	}
	if m.screen == screenScenarioPicker {
		return m.viewScenarioPicker()
	}
	if m.screen == screenStatsBuilder {
		return m.viewStatsBuilder()
	}
	if m.screen == screenPlayerConfig {
		return m.viewPlayerConfig()
	}
	if m.screen == screenKitPicker {
		return m.viewKitPicker()
	}
	if m.screen == screenLoadRun {
		return m.viewLoadRun()
	}
	if m.screen == screenScenarioBuilder {
		return m.viewScenarioBuilder()
	}
	if m.screen == screenBuilderScenarioPicker {
		return m.viewBuilderScenarioPicker()
	}
	if m.screen == screenOptions {
		return m.viewOptions()
	}
	if m.screen == screenPhaseEditor {
		return m.viewPhaseEditor()
	}
	if m.screen == screenRun {
		return m.viewRun()
	}
	return m.viewMenu()
}

func (m menuModel) updateRun(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "Q", "esc":
		next, cmd := m.returnToMainMenu()
		return next, cmd

	case "N":
		m.advanceRunDay()
		m.runPlayedFor = 0
		return m, nil
	case "S":
		if m.run == nil {
			m.status = "No active run to save."
			return m, nil
		}
		path := savePathForSlot(m.activeSaveSlot)
		if err := saveRunToFile(path, *m.run); err != nil {
			m.status = fmt.Sprintf("Save failed: %v", err)
			return m, nil
		}
		m.status = fmt.Sprintf("Saved run to slot %d (%s)", m.activeSaveSlot, path)
		return m, nil
	case "L":
		m = m.openLoadRun(screenRun)
		return m, nil
	case "enter":
		return m.submitRunInput()
	case "backspace":
		if len(m.runInput) > 0 {
			m.runInput = m.runInput[:len(m.runInput)-1]
		}
		return m, nil
	case "ctrl+h":
		if len(m.runInput) > 0 {
			m.runInput = m.runInput[:len(m.runInput)-1]
		}
		return m, nil
	}

	if len(msg.Runes) > 0 {
		m.runInput += string(msg.Runes)
		return m, nil
	}

	return m, nil
}

func (m *menuModel) advanceRunDay() {
	if m.run == nil {
		m.status = "No active run."
		return
	}

	m.run.AdvanceDay()
	out := m.run.EvaluateRun()
	if out.Status == game.RunOutcomeCritical {
		m.status = fmt.Sprintf("CRITICAL: %v", out.CriticalPlayerIDs)
		return
	}
	if out.Status == game.RunOutcomeCompleted {
		m.status = "COMPLETED"
		return
	}
	m.status = ""
}

func (m menuModel) submitRunInput() (tea.Model, tea.Cmd) {
	command := strings.TrimSpace(strings.ToLower(m.runInput))
	m.runInput = ""

	switch command {
	case "":
		m.status = "Enter a command or use Shift+ shortcuts."
		return m, nil
	case "next":
		m.advanceRunDay()
		m.runPlayedFor = 0
		return m, nil
	case "save":
		if m.run == nil {
			m.status = "No active run to save."
			return m, nil
		}
		path := savePathForSlot(m.activeSaveSlot)
		if err := saveRunToFile(path, *m.run); err != nil {
			m.status = fmt.Sprintf("Save failed: %v", err)
			return m, nil
		}
		m.status = fmt.Sprintf("Saved run to slot %d", m.activeSaveSlot)
		return m, nil
	case "load":
		m = m.openLoadRun(screenRun)
		return m, nil
	case "menu", "back":
		next, cmd := m.returnToMainMenu()
		return next, cmd
	default:
		if strings.HasPrefix(command, "hunt") || strings.HasPrefix(command, "catch") {
			return m.handleHuntCommand(command)
		}
		if m.run != nil {
			res := m.run.ExecuteRunCommand(command)
			if res.Handled {
				m.status = res.Message
				return m, nil
			}
		}
		m.status = "Unknown command. Try: next, save, load, menu, hunt land|fish|air, actions, use <item> <action>."
		return m, nil
	}
}

func (m menuModel) handleHuntCommand(command string) (tea.Model, tea.Cmd) {
	if m.run == nil {
		m.status = "No active run."
		return m, nil
	}

	fields := strings.Fields(command)
	domain := game.AnimalDomainLand
	choice := game.MealChoice{
		PortionGrams: 300,
		Cooked:       true,
		EatLiver:     false,
	}
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
				if parsed, err := strconv.Atoi(strings.TrimPrefix(field, "p")); err == nil && parsed > 0 {
					playerID = parsed
				}
				continue
			}
			if grams, err := strconv.Atoi(field); err == nil && grams > 0 {
				choice.PortionGrams = grams
			}
		}
	}

	catch, outcome, err := m.run.CatchAndConsume(playerID, domain, choice)
	if err != nil {
		m.status = fmt.Sprintf("Hunt failed: %v", err)
		return m, nil
	}

	prep := "cooked"
	if !choice.Cooked {
		prep = "raw"
	}
	msg := fmt.Sprintf("P%d ate %dg %s (%s, %dg caught): +%dE +%dH2O +%dM | %dkcal %dgP %dgF",
		outcome.PlayerID, outcome.PortionGrams, catch.Animal.Name, prep, catch.WeightGrams,
		outcome.EnergyDelta, outcome.HydrationDelta, outcome.MoraleDelta,
		outcome.Nutrition.CaloriesKcal, outcome.Nutrition.ProteinG, outcome.Nutrition.FatG)
	if len(outcome.DiseaseEvents) > 0 {
		names := make([]string, 0, len(outcome.DiseaseEvents))
		for _, event := range outcome.DiseaseEvents {
			names = append(names, event.Ailment.Name)
		}
		msg += " | illness risk triggered: " + strings.Join(names, ", ")
	}
	m.status = msg
	return m, nil
}

func menuItems(m menuModel) []menuEntry {
	items := []menuEntry{
		{label: "Start", action: itemStart},
		{label: "Load Run", action: itemLoadRun},
		{label: "Scenario Builder", action: itemScenarioBuilder},
		{label: "Options", action: itemOptions},
	}
	if m.updateAvailable && !m.cfg.NoUpdate {
		items = append(items, menuEntry{label: "Install Update", action: itemInstallUpdate})
	}
	items = append(items, menuEntry{label: "Quit", action: itemQuit})
	return items
}

func (m menuModel) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := menuItems(m)
	itemCount := len(items)
	if itemCount == 0 {
		return m, nil
	}
	if m.idx < 0 {
		m.idx = 0
	}
	if m.idx >= itemCount {
		m.idx = itemCount - 1
	}

	switch msg.String() {
	case "ctrl+c", "Q", "q":
		return m, tea.Quit
	case "up", "k":
		m.idx = wrapIndex(m.idx, -1, itemCount)
		return m, nil
	case "down", "j":
		m.idx = wrapIndex(m.idx, 1, itemCount)
		return m, nil
	case "enter":
		switch items[m.idx].action {
		case itemStart:
			m.setup = newSetupState()
			m.pcfg = newPlayerConfigState()
			m.kit = newKitPickerState()
			m = m.ensureSetupScenarioSelection()
			m = m.ensureSetupPlayers()
			m.screen = screenSetup
			m.status = ""
			return m, nil
		case itemLoadRun:
			m = m.openLoadRun(screenMenu)
			return m, nil
		case itemScenarioBuilder:
			m.build = newScenarioBuilderState()
			m.build.playerCountIdx = m.setup.playerCountIdx
			m.screen = screenScenarioBuilder
			return m, nil
		case itemOptions:
			m.screen = screenOptions
			return m, nil
		case itemInstallUpdate:
			if m.busy {
				return m, nil
			}
			m.busy = true
			m.status = "Downloading update…"
			return m, applyUpdateCmd(m.cfg.Version)
		case itemQuit:
			return m, tea.Quit
		}
	}

	return m, nil
}

func temperatureUnitLabel(unit temperatureUnit) string {
	switch unit {
	case tempUnitFahrenheit:
		return "Fahrenheit (F)"
	default:
		return "Celsius (C)"
	}
}

func (m menuModel) normalizeOptionsState() menuModel {
	units := temperatureUnits()
	if m.opts.tempUnitIdx < 0 || m.opts.tempUnitIdx >= len(units) {
		m.opts.tempUnitIdx = 0
	}
	m.opts.tempUnit = units[m.opts.tempUnitIdx]

	dayOptions := dayDurationHoursOptions()
	if m.opts.dayHoursIdx < 0 || m.opts.dayHoursIdx >= len(dayOptions) {
		m.opts.dayHoursIdx = 0
	}
	m.opts.dayHours = dayOptions[m.opts.dayHoursIdx]
	if m.opts.dayHours < 1 {
		m.opts.dayHours = 2
	}
	if m.opts.cursor < 0 || m.opts.cursor > 2 {
		m.opts.cursor = 0
	}
	return m
}

func (m menuModel) adjustOptionsChoice(delta int) menuModel {
	m = m.normalizeOptionsState()
	switch m.opts.cursor {
	case 0:
		m.opts.tempUnitIdx = wrapIndex(m.opts.tempUnitIdx, delta, len(temperatureUnits()))
	case 1:
		m.opts.dayHoursIdx = wrapIndex(m.opts.dayHoursIdx, delta, len(dayDurationHoursOptions()))
	}
	return m.normalizeOptionsState()
}

func (m menuModel) updateOptions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m = m.normalizeOptionsState()

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "Q", "q", "esc":
		next, cmd := m.returnToMainMenu()
		return next, cmd
	case "up", "k":
		m.opts.cursor = wrapIndex(m.opts.cursor, -1, 3)
		return m, nil
	case "down", "j":
		m.opts.cursor = wrapIndex(m.opts.cursor, 1, 3)
		return m, nil
	case "left":
		m = m.adjustOptionsChoice(-1)
		return m, nil
	case "right":
		m = m.adjustOptionsChoice(1)
		return m, nil
	case "enter":
		if m.opts.cursor == 2 {
			next, cmd := m.returnToMainMenu()
			return next, cmd
		}
		m = m.adjustOptionsChoice(1)
		return m, nil
	}

	return m, nil
}

func (m menuModel) viewOptions() string {
	m = m.normalizeOptionsState()
	rows := []struct {
		label string
		value string
	}{
		{label: "Temperature Unit", value: temperatureUnitLabel(m.opts.tempUnit)},
		{label: "Auto Day Length", value: fmt.Sprintf("%d hour(s)", m.opts.dayHours)},
		{label: "Back", value: ""},
	}

	totalWidth := m.w
	if totalWidth < 90 {
		totalWidth = 90
	}
	contentHeight := m.h - 8
	if contentHeight < 16 {
		contentHeight = 16
	}
	listWidth := totalWidth / 3
	if listWidth < 30 {
		listWidth = 30
	}
	detailWidth := totalWidth - listWidth - 4
	if detailWidth < 44 {
		detailWidth = 44
	}

	var list strings.Builder
	for i, row := range rows {
		cursor := "  "
		style := green
		if i == m.opts.cursor {
			cursor = "> "
			style = brightGreen
		}
		if row.value == "" {
			list.WriteString(cursor + style.Render(row.label) + "\n")
			continue
		}
		list.WriteString(cursor + style.Render(fmt.Sprintf("%-20s %s", row.label+":", row.value)) + "\n")
	}

	detail := strings.Join([]string{
		brightGreen.Render("Gameplay Options"),
		green.Render(fmt.Sprintf("Temperature Display: %s", temperatureUnitLabel(m.opts.tempUnit))),
		green.Render(fmt.Sprintf("Auto Day Length: %d hour(s)", m.opts.dayHours)),
		"",
		dimGreen.Render("Temperature controls display only for now."),
		dimGreen.Render("During a run, each full Auto Day Length advances one game day automatically."),
	}, "\n")

	listPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(listWidth).
		Height(contentHeight).
		Render(list.String())
	detailPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(detailWidth).
		Height(contentHeight).
		Render(detail)

	var b strings.Builder
	b.WriteString(dosTitle("Options") + "\n")
	b.WriteString(dimGreen.Render("Configure display and pacing defaults.") + "\n\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane))
	b.WriteString("\n" + dimGreen.Render("↑/↓ move  ←/→ change  Enter select  Shift+Q back") + "\n")
	return b.String()
}

type setupScenarioOption struct {
	scenario game.Scenario
	label    string
}

type scenarioPickerState struct {
	cursor int
}

type temperatureUnit string

const (
	tempUnitCelsius    temperatureUnit = "celsius"
	tempUnitFahrenheit temperatureUnit = "fahrenheit"
)

type optionsState struct {
	cursor      int
	tempUnitIdx int
	dayHoursIdx int
	tempUnit    temperatureUnit
	dayHours    int
}

type builderScenarioPickerState struct {
	cursor   int
	returnTo screen
}

type builderScenarioOption struct {
	label     string
	scenario  game.Scenario
	mode      game.GameMode
	isNew     bool
	isCustom  bool
	customIdx int
}

type statsBuilderState struct {
	cursor    int
	playerIdx int
	returnTo  screen
}

type playerConfigState struct {
	cursor    int
	playerIdx int
	returnTo  screen
}

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

type kitPickerState struct {
	categoryIdx int
	itemIdx     int
	target      kitPickerTarget
	focus       kitPickerFocus
	returnTo    screen
}

type saveSlotMeta struct {
	Slot      int
	Path      string
	Exists    bool
	Summary   string
	ErrDetail string
	Run       *game.RunState
	SavedAt   time.Time
}

type loadRunState struct {
	cursor  int
	entries []saveSlotMeta
}

type scenarioBuilderState struct {
	cursor         int
	selectedIdx    int // 0 = new, 1..N = existing custom scenario index+1
	selectedLabel  string
	sourceID       game.ScenarioID
	sourceName     string
	sourceIsCustom bool
	name           string
	modeIdx        int
	playerCountIdx int
	biomeIdx       int
	wildlifeText   string
	useCustomDays  bool
	defaultDaysIdx int
	customDays     string
	seasonSetID    string
	phases         []phaseBuilderPhase
}

type phaseBuilderPhase struct {
	seasonIdx int
	days      string
}

type phaseEditorState struct {
	cursor       int
	adding       bool
	newSeasonIdx int
	newDays      string
}

type phaseEditorRowKind int

const (
	phaseRowNewSeason phaseEditorRowKind = iota
	phaseRowNewDays
	phaseRowAddPhase
	phaseRowRemoveLast
	phaseRowBack
)

type phaseEditorRow struct {
	label  string
	value  string
	kind   phaseEditorRowKind
	active bool
}

type scenarioBuilderRowKind int

const (
	builderRowScenario scenarioBuilderRowKind = iota
	builderRowName
	builderRowMode
	builderRowPlayerCount
	builderRowBiome
	builderRowWildlife
	builderRowDaysMode
	builderRowDaysPreset
	builderRowDaysCustom
	builderRowSeasonProfileID
	builderRowEditPhases
	builderRowPlayerEditor
	builderRowSave
	builderRowDelete
	builderRowCancel
)

const maxBuilderPhases = 12

type scenarioBuilderRow struct {
	label    string
	value    string
	kind     scenarioBuilderRowKind
	phaseIdx int
	active   bool
}

type statsBuilderRowKind int

const (
	statsRowPlayer statsBuilderRowKind = iota
	statsRowSex
	statsRowBodyType
	statsRowWeightKg
	statsRowHeightFt
	statsRowHeightIn
	statsRowEndurance
	statsRowBushcraft
	statsRowMental
	statsRowBack
)

type statsBuilderRow struct {
	label  string
	value  string
	kind   statsBuilderRowKind
	active bool
}

type playerConfigRowKind int

const (
	playerRowPlayer playerConfigRowKind = iota
	playerRowName
	playerRowSex
	playerRowBodyType
	playerRowWeightKg
	playerRowHeightFt
	playerRowHeightIn
	playerRowEndurance
	playerRowBushcraft
	playerRowMental
	playerRowKitLimit
	playerRowPersonalKit
	playerRowEditPersonalKit
	playerRowResetPersonalKit
	playerRowEditIssuedKit
	playerRowResetIssuedKit
	playerRowBack
)

type playerConfigRow struct {
	label  string
	value  string
	kind   playerConfigRowKind
	active bool
}

type runLengthOption struct {
	label     string
	openEnded bool
	days      int
}

type setupState struct {
	cursor         int
	modeIdx        int
	scenarioID     game.ScenarioID
	playerCountIdx int
	runLengthIdx   int
	players        []game.PlayerConfig
	issuedKit      []game.KitItem
	issuedCustom   bool
}

func newSetupState() setupState {
	s := setupState{
		cursor:         0,
		modeIdx:        0,
		scenarioID:     "",
		playerCountIdx: 0, // Alone defaults to 1 player
		runLengthIdx:   0,
	}
	s.players = make([]game.PlayerConfig, setupPlayerCounts()[s.playerCountIdx])
	for i := range s.players {
		s.players[i] = defaultPlayerConfig(game.ModeAlone)
	}
	return s
}

func newScenarioPickerState() scenarioPickerState {
	return scenarioPickerState{cursor: 0}
}

func temperatureUnits() []temperatureUnit {
	return []temperatureUnit{tempUnitCelsius, tempUnitFahrenheit}
}

func dayDurationHoursOptions() []int {
	return []int{1, 2, 3, 4, 6, 8, 12}
}

func newOptionsState() optionsState {
	dayOptions := dayDurationHoursOptions()
	defaultDayHours := 2
	defaultDayIdx := 0
	for i, hours := range dayOptions {
		if hours == defaultDayHours {
			defaultDayIdx = i
			break
		}
	}
	return optionsState{
		cursor:      0,
		tempUnitIdx: 0,
		dayHoursIdx: defaultDayIdx,
		tempUnit:    tempUnitCelsius,
		dayHours:    defaultDayHours,
	}
}

func newBuilderScenarioPickerState() builderScenarioPickerState {
	return builderScenarioPickerState{
		cursor:   0,
		returnTo: screenScenarioBuilder,
	}
}

func newStatsBuilderState() statsBuilderState {
	return statsBuilderState{
		cursor:    0,
		playerIdx: 0,
		returnTo:  screenSetup,
	}
}

func newPlayerConfigState() playerConfigState {
	return playerConfigState{
		cursor:    0,
		playerIdx: 0,
		returnTo:  screenSetup,
	}
}

func newKitPickerState() kitPickerState {
	return kitPickerState{
		categoryIdx: 0,
		itemIdx:     0,
		target:      kitTargetPersonal,
		focus:       kitFocusCategories,
		returnTo:    screenPlayerConfig,
	}
}

func newLoadRunState() loadRunState {
	return loadRunState{
		cursor:  0,
		entries: nil,
	}
}

func newScenarioBuilderState() scenarioBuilderState {
	defaultBiome := builderBiomes()[0]
	return scenarioBuilderState{
		cursor:         0,
		selectedIdx:    0,
		selectedLabel:  "New Scenario",
		sourceID:       "",
		sourceName:     "",
		sourceIsCustom: false,
		name:           "",
		modeIdx:        0,
		playerCountIdx: 0,
		biomeIdx:       0,
		wildlifeText:   strings.Join(game.WildlifeForBiome(defaultBiome), ", "),
		useCustomDays:  false,
		defaultDaysIdx: 1,
		customDays:     "60",
		seasonSetID:    "custom_profile",
		phases: []phaseBuilderPhase{
			{seasonIdx: 0, days: "14"},
			{seasonIdx: 1, days: "0"},
		},
	}
}

func newPhaseEditorState() phaseEditorState {
	return phaseEditorState{
		cursor:       0,
		adding:       false,
		newSeasonIdx: 0,
		newDays:      "7",
	}
}

func setupModes() []game.GameMode {
	return []game.GameMode{
		game.ModeAlone,
		game.ModeNakedAndAfraid,
		game.ModeNakedAndAfraidXL,
	}
}

func (m menuModel) setupScenarioOptionsForMode(mode game.GameMode) []setupScenarioOption {
	options := make([]setupScenarioOption, 0)
	for _, scenario := range m.availableScenarios() {
		if !scenarioSupportsMode(scenario, mode) {
			continue
		}
		label := scenario.Name
		if isCustomScenarioID(scenario.ID) {
			label = scenario.Name + " (Custom)"
		}
		options = append(options, setupScenarioOption{
			scenario: scenario,
			label:    label,
		})
	}
	return options
}

func scenarioSupportsMode(s game.Scenario, mode game.GameMode) bool {
	if len(s.SupportedModes) == 0 {
		return true
	}
	for _, allowed := range s.SupportedModes {
		if allowed == mode {
			return true
		}
	}
	return false
}

func setupPlayerCounts() []int {
	return []int{1, 2, 3, 4, 5, 6, 7, 8}
}

func setupRunLengths() []runLengthOption {
	return []runLengthOption{
		{label: "7 days", days: 7},
		{label: "14 days", days: 14},
		{label: "30 days", days: 30},
		{label: "60 days", days: 60},
		{label: "Open ended", openEnded: true},
	}
}

func maxKitLimitForMode(mode game.GameMode) int {
	switch mode {
	case game.ModeAlone:
		return 10
	case game.ModeNakedAndAfraid:
		return 1
	case game.ModeNakedAndAfraidXL:
		return 1
	default:
		return 1
	}
}

func defaultKitLimitForMode(mode game.GameMode) int {
	return maxKitLimitForMode(mode)
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
	default: // Alone
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

func playerConfigSuggestedNames() []string {
	return []string{
		"Sophia", "Daniel", "Maya", "Ethan", "Harper",
		"Riley", "Quinn", "Rowan", "Kai", "Avery",
		"Emma", "Olivia", "Jack", "Tom", "Freya",
		"Jordan", "Morgan", "Taylor", "Alex", "Noah",
	}
}

func defaultPlayerNameForIndex(idx int) string {
	names := playerConfigSuggestedNames()
	if len(names) == 0 {
		return fmt.Sprintf("Player %d", idx+1)
	}
	base := names[idx%len(names)]
	if idx >= len(names) {
		cycle := idx/len(names) + 1
		return fmt.Sprintf("%s %d", base, cycle)
	}
	return base
}

func playerConfigSexes() []game.Sex {
	return []game.Sex{
		game.SexMale,
		game.SexFemale,
		game.SexNonBinary,
		game.SexOther,
	}
}

func playerConfigBodyTypes() []game.BodyType {
	return []game.BodyType{
		game.BodyTypeNeutral,
		game.BodyTypeMale,
		game.BodyTypeFemale,
	}
}

func defaultPlayerConfig(mode game.GameMode) game.PlayerConfig {
	return game.PlayerConfig{
		Sex:       game.SexOther,
		BodyType:  game.BodyTypeNeutral,
		WeightKg:  75,
		HeightFt:  5,
		HeightIn:  10,
		Endurance: 0,
		Bushcraft: 0,
		Mental:    0,
		KitLimit:  defaultKitLimitForMode(mode),
		Kit:       []game.KitItem{},
	}
}

func recommendedIssuedKitForScenario(mode game.GameMode, scenario game.Scenario) []game.KitItem {
	biome := strings.ToLower(scenario.Biome)
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
	filtered := make([]game.KitItem, 0, len(items))
	for _, item := range items {
		if _, ok := allowedSet[item]; !ok {
			continue
		}
		if hasKitItem(filtered, item) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func (m menuModel) ensureSetupPlayers() menuModel {
	mode := m.setupMode()
	count := setupPlayerCounts()[m.setup.playerCountIdx]

	if len(m.setup.players) < count {
		for len(m.setup.players) < count {
			m.setup.players = append(m.setup.players, defaultPlayerConfig(mode))
		}
	}
	if len(m.setup.players) > count {
		m.setup.players = m.setup.players[:count]
	}

	defaultLimit := defaultKitLimitForMode(mode)
	maxLimit := maxKitLimitForMode(mode)
	for i := range m.setup.players {
		if strings.TrimSpace(m.setup.players[i].Name) == "" {
			m.setup.players[i].Name = defaultPlayerNameForIndex(i)
		}
		if m.setup.players[i].KitLimit <= 0 {
			m.setup.players[i].KitLimit = defaultLimit
		}
		if m.setup.players[i].KitLimit > maxLimit {
			m.setup.players[i].KitLimit = maxLimit
		}
		if len(m.setup.players[i].Kit) > m.setup.players[i].KitLimit {
			m.setup.players[i].Kit = append([]game.KitItem(nil), m.setup.players[i].Kit[:m.setup.players[i].KitLimit]...)
		}
	}

	if scenario, found := m.selectedSetupScenario(); found {
		if !m.setup.issuedCustom {
			m.setup.issuedKit = recommendedIssuedKitForScenario(mode, scenario)
		}
	}

	if len(m.setup.players) > 0 && m.pcfg.playerIdx >= len(m.setup.players) {
		m.pcfg.playerIdx = len(m.setup.players) - 1
	}
	if len(m.setup.players) > 0 && m.sbuild.playerIdx >= len(m.setup.players) {
		m.sbuild.playerIdx = len(m.setup.players) - 1
	}

	return m
}

func (m menuModel) selectedSetupScenario() (game.Scenario, bool) {
	options := m.setupScenarioOptionsForMode(m.setupMode())
	if len(options) == 0 {
		return game.Scenario{}, false
	}
	for _, option := range options {
		if option.scenario.ID == m.setup.scenarioID {
			return option.scenario, true
		}
	}
	return options[0].scenario, true
}

func builderBiomes() []string {
	return []string{"Forest", "Jungle", "Arctic", "Coast", "Mountain", "Desert"}
}

func builderDefaultDays() []int {
	return []int{7, 14, 30, 60, 90}
}

func builderSeasonOptions() []game.SeasonID {
	return []game.SeasonID{game.SeasonAutumn, game.SeasonWinter, game.SeasonWet}
}

func builderSeasonLabel(id game.SeasonID) string {
	switch id {
	case game.SeasonAutumn:
		return "Autumn"
	case game.SeasonWinter:
		return "Winter"
	case game.SeasonWet:
		return "Wet"
	default:
		return string(id)
	}
}

func scenarioDefaultMode(s game.Scenario) game.GameMode {
	if len(s.SupportedModes) > 0 {
		return s.SupportedModes[0]
	}
	return game.ModeAlone
}

func selectedModeIndex(mode game.GameMode) int {
	modes := setupModes()
	for i, m := range modes {
		if m == mode {
			return i
		}
	}
	return 0
}

func selectedBiomeIndex(biome string) int {
	biomes := builderBiomes()
	for i, b := range biomes {
		if strings.EqualFold(biome, b) {
			return i
		}
	}
	return 0
}

func sanitizeSeasonSetID(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return "custom_profile"
	}
	replacer := strings.NewReplacer(
		" ", "_",
		"-", "_",
		".", "_",
		"/", "_",
		"\\", "_",
	)
	value = replacer.Replace(value)
	value = strings.Trim(value, "_")
	if value == "" {
		return "custom_profile"
	}
	return value
}

func wrapIndex(current, delta, size int) int {
	next := current + delta
	for next < 0 {
		next += size
	}
	return next % size
}

func dosRule(width int) string {
	if width < 40 {
		width = 40
	}
	return strings.Repeat("-", width)
}

func dosTitle(title string) string {
	return brightGreen.Render("[" + strings.ToUpper(title) + "]")
}

func (m menuModel) returnToMainMenu() (menuModel, tea.Cmd) {
	m.screen = screenMenu
	if m.cfg.NoUpdate || m.busy {
		return m, nil
	}
	m.busy = true
	return m, checkUpdateCmd(m.cfg.Version, true)
}

func (m menuModel) updateSetup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	const rowCount = 9 // mode, scenario, players, run length, stats, player/kit, issued kit, start, cancel
	m = m.ensureSetupScenarioSelection()
	m = m.ensureSetupPlayers()

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "Q", "q", "esc":
		next, cmd := m.returnToMainMenu()
		return next, cmd
	case "up", "k":
		m.setup.cursor = wrapIndex(m.setup.cursor, -1, rowCount)
		return m, nil
	case "down", "j":
		m.setup.cursor = wrapIndex(m.setup.cursor, 1, rowCount)
		return m, nil
	case "left":
		if m.setup.cursor == 1 {
			m = m.openScenarioPicker()
			return m, nil
		}
		if m.setup.cursor == 4 {
			m.sbuild.returnTo = screenSetup
			m.sbuild.playerIdx = 0
			m.screen = screenStatsBuilder
			return m, nil
		}
		if m.setup.cursor == 5 {
			m.pcfg.returnTo = screenSetup
			m.pcfg.playerIdx = 0
			m.screen = screenPlayerConfig
			return m, nil
		}
		if m.setup.cursor == 6 {
			m = m.openIssuedKitPicker(screenSetup)
			return m, nil
		}
		m = m.adjustSetupChoice(-1)
		return m, nil
	case "right":
		if m.setup.cursor == 1 {
			m = m.openScenarioPicker()
			return m, nil
		}
		if m.setup.cursor == 4 {
			m.sbuild.returnTo = screenSetup
			m.sbuild.playerIdx = 0
			m.screen = screenStatsBuilder
			return m, nil
		}
		if m.setup.cursor == 5 {
			m.pcfg.returnTo = screenSetup
			m.pcfg.playerIdx = 0
			m.screen = screenPlayerConfig
			return m, nil
		}
		if m.setup.cursor == 6 {
			m = m.openIssuedKitPicker(screenSetup)
			return m, nil
		}
		m = m.adjustSetupChoice(1)
		return m, nil
	case "enter":
		switch m.setup.cursor {
		case 1:
			m = m.openScenarioPicker()
			return m, nil
		case 4:
			m.sbuild.returnTo = screenSetup
			m.sbuild.playerIdx = 0
			m.screen = screenStatsBuilder
			return m, nil
		case 5:
			m.pcfg.returnTo = screenSetup
			m.pcfg.playerIdx = 0
			m.screen = screenPlayerConfig
			return m, nil
		case 6:
			m = m.openIssuedKitPicker(screenSetup)
			return m, nil
		case 7:
			return m.startRunFromSetup()
		case 8:
			next, cmd := m.returnToMainMenu()
			return next, cmd
		default:
			m = m.adjustSetupChoice(1)
			return m, nil
		}
	}

	return m, nil
}

func (m menuModel) adjustSetupChoice(delta int) menuModel {
	switch m.setup.cursor {
	case 0:
		m.setup.modeIdx = wrapIndex(m.setup.modeIdx, delta, len(setupModes()))
		m = m.ensureSetupScenarioSelection()
	case 1:
		options := m.setupScenarioOptionsForMode(m.setupMode())
		if len(options) == 0 {
			m.setup.scenarioID = ""
			return m
		}
		idx := 0
		for i, option := range options {
			if option.scenario.ID == m.setup.scenarioID {
				idx = i
				break
			}
		}
		idx = wrapIndex(idx, delta, len(options))
		m.setup.scenarioID = options[idx].scenario.ID
		mode, found := m.preferredModeForScenario(options[idx].scenario.ID)
		if found {
			m.setup.modeIdx = selectedModeIndex(mode)
		}
	case 2:
		m.setup.playerCountIdx = wrapIndex(m.setup.playerCountIdx, delta, len(setupPlayerCounts()))
	case 3:
		m.setup.runLengthIdx = wrapIndex(m.setup.runLengthIdx, delta, len(setupRunLengths()))
	}

	return m.ensureSetupPlayers()
}

func (m menuModel) startRunFromSetup() (tea.Model, tea.Cmd) {
	m = m.ensureSetupScenarioSelection()
	m = m.ensureSetupPlayers()
	modes := setupModes()
	playerCounts := setupPlayerCounts()
	runLengths := setupRunLengths()
	mode := modes[m.setup.modeIdx]
	scenarioID := m.setup.scenarioID
	if scenarioID == "" {
		options := m.setupScenarioOptionsForMode(mode)
		if len(options) == 0 {
			m.status = fmt.Sprintf("No scenarios available for mode %s.", modeLabel(mode))
			return m, nil
		}
		scenarioID = options[0].scenario.ID
	}

	runLength := runLengths[m.setup.runLengthIdx]
	players := append([]game.PlayerConfig(nil), m.setup.players...)
	for i := range players {
		players[i].Kit = append([]game.KitItem(nil), players[i].Kit...)
	}
	cfg := game.RunConfig{
		Mode:        mode,
		ScenarioID:  scenarioID,
		PlayerCount: playerCounts[m.setup.playerCountIdx],
		RunLength: game.RunLength{
			OpenEnded: runLength.openEnded,
			Days:      runLength.days,
		},
		Seed:      0,
		Players:   players,
		IssuedKit: append([]game.KitItem(nil), m.setup.issuedKit...),
	}

	state, err := newRunStateWithScenarios(cfg, m.availableScenarios())
	if err != nil {
		m.status = fmt.Sprintf("Failed to start: %v", err)
		return m, nil
	}

	m.run = &state
	m.runPlayedFor = 0
	m.screen = screenRun
	m.status = ""
	return m, nil
}

func (m menuModel) viewSetup() string {
	m = m.ensureSetupScenarioSelection()
	m = m.ensureSetupPlayers()
	modes := setupModes()
	scenarios := m.setupScenarioOptionsForMode(m.setupMode())
	playerCounts := setupPlayerCounts()
	runLengths := setupRunLengths()
	scenarioLabel := "No scenarios available"
	if selected, found := findSetupScenarioByID(scenarios, m.setup.scenarioID); found {
		scenarioLabel = selected.label
	}
	readyPlayers := 0
	for _, p := range m.setup.players {
		if len(p.Kit) > 0 {
			readyPlayers++
		}
	}

	rows := []struct {
		label string
		value string
	}{
		{label: "Mode", value: modeLabel(modes[m.setup.modeIdx])},
		{label: "Scenario", value: scenarioLabel},
		{label: "Players", value: fmt.Sprintf("%d", playerCounts[m.setup.playerCountIdx])},
		{label: "Run Length", value: runLengths[m.setup.runLengthIdx].label},
		{label: "Configure Player Stats", value: ""},
		{label: "Configure Players & Kit", value: ""},
		{label: "Configure Issued Kit", value: kitSummary(m.setup.issuedKit, 0)},
		{label: "Start Run", value: ""},
		{label: "Cancel", value: ""},
	}

	totalWidth := m.w
	if totalWidth < 100 {
		totalWidth = 100
	}
	contentHeight := m.h - 8
	if contentHeight < 16 {
		contentHeight = 16
	}
	listWidth := totalWidth / 3
	if listWidth < 34 {
		listWidth = 34
	}
	detailWidth := totalWidth - listWidth - 4
	if detailWidth < 52 {
		detailWidth = 52
	}

	var list strings.Builder
	for i, row := range rows {
		cursor := "  "
		lineStyle := green
		if i == m.setup.cursor {
			cursor = "> "
			lineStyle = brightGreen
		}

		if row.value == "" {
			list.WriteString(cursor + lineStyle.Render(row.label) + "\n")
			continue
		}

		list.WriteString(cursor + lineStyle.Render(fmt.Sprintf("%-22s %s", row.label+":", row.value)) + "\n")
	}

	detail := m.setupDetailText(rows[m.setup.cursor].label, scenarioLabel, readyPlayers, len(m.setup.players))
	listPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(listWidth).
		Height(contentHeight).
		Render(list.String())
	detailPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(detailWidth).
		Height(contentHeight).
		Render(detail)

	var b strings.Builder
	b.WriteString(dosTitle("New Run Wizard") + "\n")
	b.WriteString(dimGreen.Render("Step through setup screens, then start run.") + "\n")
	b.WriteString(dimGreen.Render(fmt.Sprintf("Player kits configured: %d/%d", readyPlayers, len(m.setup.players))) + "\n\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane))
	b.WriteString("\n" + dimGreen.Render("↑/↓ move  ←/→ change values  Enter open/confirm  Shift+Q back") + "\n")
	b.WriteString(dimGreen.Render("Natural flow: Scenario -> Player Stats -> Player Kit -> Issued Kit -> Start Run") + "\n")
	if m.status != "" {
		b.WriteString(green.Render(m.status) + "\n")
	}

	return b.String()
}

func (m menuModel) setupDetailText(activeLabel, scenarioLabel string, readyPlayers, totalPlayers int) string {
	lines := []string{
		brightGreen.Render("Run Setup Summary"),
		green.Render(fmt.Sprintf("Mode: %s", modeLabel(m.setupMode()))),
		green.Render(fmt.Sprintf("Scenario: %s", scenarioLabel)),
		green.Render(fmt.Sprintf("Players Ready: %d/%d", readyPlayers, totalPlayers)),
		green.Render(fmt.Sprintf("Issued Kit: %s", kitSummary(m.setup.issuedKit, 0))),
		"",
		brightGreen.Render("Active Step"),
		green.Render(activeLabel),
		"",
		dimGreen.Render("Enter on Scenario opens detailed selector."),
		dimGreen.Render("Configure Player Stats opens Stats Builder."),
		dimGreen.Render("Configure Players & Kit opens Player Editor."),
		dimGreen.Render("Configure Issued Kit opens Kit Picker."),
	}
	return strings.Join(lines, "\n")
}

func (m menuModel) setupMode() game.GameMode {
	modes := setupModes()
	if m.setup.modeIdx < 0 || m.setup.modeIdx >= len(modes) {
		return modes[0]
	}
	return modes[m.setup.modeIdx]
}

func (m menuModel) ensureSetupScenarioSelection() menuModel {
	options := m.setupScenarioOptionsForMode(m.setupMode())
	if len(options) == 0 {
		m.setup.scenarioID = ""
		return m
	}
	for _, option := range options {
		if option.scenario.ID == m.setup.scenarioID {
			return m
		}
	}
	m.setup.scenarioID = options[0].scenario.ID
	return m
}

func findSetupScenarioByID(options []setupScenarioOption, id game.ScenarioID) (setupScenarioOption, bool) {
	for _, option := range options {
		if option.scenario.ID == id {
			return option, true
		}
	}
	return setupScenarioOption{}, false
}

func (m menuModel) openScenarioPicker() menuModel {
	options := m.setupScenarioOptionsForMode(m.setupMode())
	if len(options) == 0 {
		m.status = fmt.Sprintf("No scenarios available for mode %s.", modeLabel(m.setupMode()))
		return m
	}

	cursor := 0
	for i, option := range options {
		if option.scenario.ID == m.setup.scenarioID {
			cursor = i
			break
		}
	}

	m.pick = scenarioPickerState{cursor: cursor}
	m.screen = screenScenarioPicker
	return m
}

func (m menuModel) updateScenarioPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	options := m.setupScenarioOptionsForMode(m.setupMode())
	if len(options) == 0 {
		m.screen = screenSetup
		m.status = fmt.Sprintf("No scenarios available for mode %s.", modeLabel(m.setupMode()))
		return m, nil
	}

	if m.pick.cursor < 0 || m.pick.cursor >= len(options) {
		m.pick.cursor = 0
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "Q", "q", "esc":
		if m.pcfg.returnTo == 0 {
			m.pcfg.returnTo = screenSetup
		}
		m.screen = m.pcfg.returnTo
		return m, nil
	case "up", "k":
		m.pick.cursor = wrapIndex(m.pick.cursor, -1, len(options))
		return m, nil
	case "down", "j":
		m.pick.cursor = wrapIndex(m.pick.cursor, 1, len(options))
		return m, nil
	case "enter":
		m.setup.scenarioID = options[m.pick.cursor].scenario.ID
		m.screen = screenSetup
		return m, nil
	}

	return m, nil
}

func (m menuModel) viewScenarioPicker() string {
	options := m.setupScenarioOptionsForMode(m.setupMode())
	if len(options) == 0 {
		return brightGreen.Render("No scenarios available for this mode.")
	}

	if m.pick.cursor < 0 || m.pick.cursor >= len(options) {
		m.pick.cursor = 0
	}
	selected := options[m.pick.cursor].scenario

	totalWidth := m.w
	if totalWidth < 90 {
		totalWidth = 90
	}
	contentHeight := m.h - 8
	if contentHeight < 16 {
		contentHeight = 16
	}

	listWidth := totalWidth / 3
	if listWidth < 28 {
		listWidth = 28
	}
	detailWidth := totalWidth - listWidth - 4
	if detailWidth < 40 {
		detailWidth = 40
	}

	var list strings.Builder
	for i, option := range options {
		line := option.label
		if i == m.pick.cursor {
			list.WriteString(brightGreen.Render("> " + line))
		} else {
			list.WriteString(green.Render("  " + line))
		}
		list.WriteString("\n")
	}

	detailText := m.scenarioDetailText(selected)
	listPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(listWidth).
		Height(contentHeight).
		Render(list.String())
	detailPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(detailWidth).
		Height(contentHeight).
		Render(detailText)

	var b strings.Builder
	b.WriteString(dosTitle("Scenario Select") + "\n")
	b.WriteString(dimGreen.Render("Mode: "+modeLabel(m.setupMode())+"  |  ↑/↓ browse, Enter select, Shift+Q back") + "\n\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane))
	return b.String()
}

func (m menuModel) updateBuilderScenarioPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	options := m.builderScenarioOptions()
	if len(options) == 0 {
		m.screen = screenScenarioBuilder
		m.status = "No scenarios available."
		return m, nil
	}
	if m.bpick.cursor < 0 || m.bpick.cursor >= len(options) {
		m.bpick.cursor = 0
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "Q", "q", "esc":
		if m.bpick.returnTo == 0 {
			m.bpick.returnTo = screenScenarioBuilder
		}
		m.screen = m.bpick.returnTo
		return m, nil
	case "up", "k":
		m.bpick.cursor = wrapIndex(m.bpick.cursor, -1, len(options))
		return m, nil
	case "down", "j":
		m.bpick.cursor = wrapIndex(m.bpick.cursor, 1, len(options))
		return m, nil
	case "enter":
		selected := options[m.bpick.cursor]
		m = m.applyBuilderScenarioOption(selected)
		if m.bpick.returnTo == 0 {
			m.bpick.returnTo = screenScenarioBuilder
		}
		m.screen = m.bpick.returnTo
		return m, nil
	}

	return m, nil
}

func (m menuModel) viewBuilderScenarioPicker() string {
	options := m.builderScenarioOptions()
	if len(options) == 0 {
		return brightGreen.Render("No scenarios available.")
	}
	if m.bpick.cursor < 0 || m.bpick.cursor >= len(options) {
		m.bpick.cursor = 0
	}

	selected := options[m.bpick.cursor]

	totalWidth := m.w
	if totalWidth < 100 {
		totalWidth = 100
	}
	contentHeight := m.h - 8
	if contentHeight < 16 {
		contentHeight = 16
	}
	listWidth := totalWidth / 3
	if listWidth < 34 {
		listWidth = 34
	}
	detailWidth := totalWidth - listWidth - 4
	if detailWidth < 52 {
		detailWidth = 52
	}

	var list strings.Builder
	for i, option := range options {
		cursor := "  "
		lineStyle := green
		if i == m.bpick.cursor {
			cursor = "> "
			lineStyle = brightGreen
		}
		list.WriteString(cursor + lineStyle.Render(option.label) + "\n")
	}

	detailText := ""
	if selected.isNew {
		detailText = strings.Join([]string{
			brightGreen.Render("New Scenario"),
			green.Render("Start with a blank custom scenario draft."),
			"",
			dimGreen.Render("Use this option to create a scenario from scratch."),
		}, "\n")
	} else {
		detailText = m.scenarioDetailText(selected.scenario)
	}

	listPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(listWidth).
		Height(contentHeight).
		Render(list.String())
	detailPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(detailWidth).
		Height(contentHeight).
		Render(detailText)

	var b strings.Builder
	b.WriteString(dosTitle("Builder Scenario Select") + "\n")
	b.WriteString(dimGreen.Render("Left: all scenarios. Right: scenario details.") + "\n\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane))
	b.WriteString("\n" + dimGreen.Render("↑/↓ move  Enter select  Shift+Q back") + "\n")
	return b.String()
}

func (m menuModel) normalizeStatsBuilderState() menuModel {
	if len(m.setup.players) > 0 {
		if m.sbuild.playerIdx < 0 || m.sbuild.playerIdx >= len(m.setup.players) {
			m.sbuild.playerIdx = 0
		}
	} else {
		m.sbuild.playerIdx = 0
	}

	rows := m.statsBuilderRows()
	if len(rows) > 0 {
		if m.sbuild.cursor < 0 || m.sbuild.cursor >= len(rows) {
			m.sbuild.cursor = 0
		}
	} else {
		m.sbuild.cursor = 0
	}

	return m
}

func (m menuModel) statsBuilderRows() []statsBuilderRow {
	if len(m.setup.players) == 0 {
		return nil
	}
	playerIdx := m.sbuild.playerIdx
	if playerIdx < 0 || playerIdx >= len(m.setup.players) {
		playerIdx = 0
	}
	p := m.setup.players[playerIdx]

	return []statsBuilderRow{
		{label: "Player", value: playerEditorSlotLabel(playerIdx, len(m.setup.players)), kind: statsRowPlayer, active: true},
		{label: "Sex", value: string(p.Sex), kind: statsRowSex, active: true},
		{label: "Body Type", value: string(p.BodyType), kind: statsRowBodyType, active: true},
		{label: "Weight (kg)", value: fmt.Sprintf("%d", p.WeightKg), kind: statsRowWeightKg, active: true},
		{label: "Height (ft)", value: fmt.Sprintf("%d", p.HeightFt), kind: statsRowHeightFt, active: true},
		{label: "Height (in)", value: fmt.Sprintf("%d", p.HeightIn), kind: statsRowHeightIn, active: true},
		{label: "Endurance Modifier", value: fmt.Sprintf("%+d", p.Endurance), kind: statsRowEndurance, active: true},
		{label: "Bushcraft Modifier", value: fmt.Sprintf("%+d", p.Bushcraft), kind: statsRowBushcraft, active: true},
		{label: "Mental Modifier", value: fmt.Sprintf("%+d", p.Mental), kind: statsRowMental, active: true},
		{label: "Back To Wizard", value: "", kind: statsRowBack, active: true},
	}
}

func statsBuilderRowSupportsCycle(kind statsBuilderRowKind) bool {
	switch kind {
	case statsRowPlayer, statsRowSex, statsRowBodyType, statsRowWeightKg, statsRowHeightFt,
		statsRowHeightIn, statsRowEndurance, statsRowBushcraft, statsRowMental:
		return true
	default:
		return false
	}
}

func (m menuModel) adjustStatsBuilderChoice(delta int) menuModel {
	if len(m.setup.players) == 0 {
		return m
	}
	m = m.normalizeStatsBuilderState()
	rows := m.statsBuilderRows()
	if len(rows) == 0 || m.sbuild.cursor < 0 || m.sbuild.cursor >= len(rows) {
		return m
	}
	row := rows[m.sbuild.cursor]
	if !row.active {
		return m
	}

	p := &m.setup.players[m.sbuild.playerIdx]
	switch row.kind {
	case statsRowPlayer:
		m.sbuild.playerIdx = wrapIndex(m.sbuild.playerIdx, delta, len(m.setup.players))
	case statsRowSex:
		options := playerConfigSexes()
		idx := indexOfSex(options, p.Sex)
		p.Sex = options[wrapIndex(idx, delta, len(options))]
	case statsRowBodyType:
		options := playerConfigBodyTypes()
		idx := indexOfBodyType(options, p.BodyType)
		p.BodyType = options[wrapIndex(idx, delta, len(options))]
	case statsRowWeightKg:
		p.WeightKg = clampInt(p.WeightKg+delta, 35, 220)
	case statsRowHeightFt:
		p.HeightFt = clampInt(p.HeightFt+delta, 4, 7)
	case statsRowHeightIn:
		p.HeightIn = clampInt(p.HeightIn+delta, 0, 11)
	case statsRowEndurance:
		p.Endurance = clampInt(p.Endurance+delta, -3, 3)
	case statsRowBushcraft:
		p.Bushcraft = clampInt(p.Bushcraft+delta, -3, 3)
	case statsRowMental:
		p.Mental = clampInt(p.Mental+delta, -3, 3)
	}
	return m
}

func (m menuModel) updateStatsBuilder(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m = m.ensureSetupScenarioSelection()
	m = m.ensureSetupPlayers()
	m = m.normalizeStatsBuilderState()
	rows := m.statsBuilderRows()
	if len(rows) == 0 {
		m.screen = screenSetup
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "Q", "q", "esc":
		if m.sbuild.returnTo == 0 {
			m.sbuild.returnTo = screenSetup
		}
		m.screen = m.sbuild.returnTo
		return m, nil
	case "up", "k":
		m.sbuild.cursor = wrapIndex(m.sbuild.cursor, -1, len(rows))
		return m, nil
	case "down", "j":
		m.sbuild.cursor = wrapIndex(m.sbuild.cursor, 1, len(rows))
		return m, nil
	case "left":
		m = m.adjustStatsBuilderChoice(-1)
		return m, nil
	case "right":
		m = m.adjustStatsBuilderChoice(1)
		return m, nil
	case "enter":
		row := rows[m.sbuild.cursor]
		if !row.active {
			return m, nil
		}
		if row.kind == statsRowBack {
			if m.sbuild.returnTo == 0 {
				m.sbuild.returnTo = screenSetup
			}
			m.screen = m.sbuild.returnTo
			return m, nil
		}
		if statsBuilderRowSupportsCycle(row.kind) {
			m = m.adjustStatsBuilderChoice(1)
		}
		return m, nil
	}

	return m, nil
}

func (m menuModel) viewStatsBuilder() string {
	m = m.ensureSetupScenarioSelection()
	m = m.ensureSetupPlayers()
	m = m.normalizeStatsBuilderState()
	rows := m.statsBuilderRows()
	if len(rows) == 0 {
		return brightGreen.Render("No players available. Return to wizard and choose player count.")
	}

	scenarioLabel := "Unknown scenario"
	if selected, found := m.selectedSetupScenario(); found {
		scenarioLabel = selected.Name
	}

	playerCount := len(m.setup.players)
	activePlayer := m.sbuild.playerIdx + 1
	current := m.setup.players[m.sbuild.playerIdx]

	totalWidth := m.w
	if totalWidth < 100 {
		totalWidth = 100
	}
	contentHeight := m.h - 8
	if contentHeight < 16 {
		contentHeight = 16
	}
	listWidth := totalWidth / 3
	if listWidth < 34 {
		listWidth = 34
	}
	detailWidth := totalWidth - listWidth - 4
	if detailWidth < 52 {
		detailWidth = 52
	}

	var list strings.Builder
	for i, row := range rows {
		cursor := "  "
		lineStyle := green
		if i == m.sbuild.cursor {
			cursor = "> "
			lineStyle = brightGreen
		}
		if !row.active {
			lineStyle = dimGreen
		}
		if row.value == "" {
			list.WriteString(cursor + lineStyle.Render(row.label) + "\n")
			continue
		}
		list.WriteString(cursor + lineStyle.Render(fmt.Sprintf("%-22s %s", row.label+":", row.value)) + "\n")
	}

	detail := strings.Join([]string{
		brightGreen.Render("Player Stats"),
		green.Render(fmt.Sprintf("Mode: %s", modeLabel(m.setupMode()))),
		green.Render(fmt.Sprintf("Scenario: %s", scenarioLabel)),
		green.Render(fmt.Sprintf("Player Slot: %d/%d", activePlayer, playerCount)),
		green.Render(fmt.Sprintf("Sex: %s  |  Body: %s", current.Sex, current.BodyType)),
		green.Render(fmt.Sprintf("Height: %d ft %d in  |  Weight: %d kg", current.HeightFt, current.HeightIn, current.WeightKg)),
		green.Render(fmt.Sprintf("Modifiers  End:%+d  Bush:%+d  Ment:%+d", current.Endurance, current.Bushcraft, current.Mental)),
		"",
		brightGreen.Render("Notes"),
		dimGreen.Render("Use ←/→ to adjust selected stat."),
		dimGreen.Render("Use Player row to switch between players."),
	}, "\n")

	listPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(listWidth).
		Height(contentHeight).
		Render(list.String())
	detailPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(detailWidth).
		Height(contentHeight).
		Render(detail)

	var b strings.Builder
	b.WriteString(dosTitle("Stats Builder") + "\n")
	b.WriteString(dimGreen.Render("Configure player stats for this run.") + "\n\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane))
	b.WriteString("\n" + dimGreen.Render("↑/↓ move  ←/→ change values  Enter cycle/select  Shift+Q back") + "\n")
	if m.status != "" {
		b.WriteString(green.Render(m.status) + "\n")
	}
	return b.String()
}

func (m menuModel) scenarioDetailText(s game.Scenario) string {
	desc := s.Description
	if strings.TrimSpace(desc) == "" {
		desc = "A survival challenge scenario."
	}
	daunting := s.Daunting
	if strings.TrimSpace(daunting) == "" {
		daunting = "Conditions may deteriorate quickly if you lose momentum."
	}
	motivation := s.Motivation
	if strings.TrimSpace(motivation) == "" {
		motivation = "Stay disciplined and complete the challenge."
	}

	var b strings.Builder
	wildlife := scenarioWildlife(s)
	insects := game.InsectsForBiome(s.Biome)
	tempRange := game.TemperatureRangeForBiome(s.Biome)
	b.WriteString(brightGreen.Render(s.Name) + "\n")
	b.WriteString(green.Render(fmt.Sprintf("Biome: %s", s.Biome)) + "\n")
	b.WriteString(green.Render(fmt.Sprintf("Temperature Range: %s", m.formatTemperatureRange(tempRange))) + "\n")
	b.WriteString(green.Render(fmt.Sprintf("Default Days: %d", s.DefaultDays)) + "\n\n")
	b.WriteString(brightGreen.Render("Wildlife") + "\n")
	b.WriteString(green.Render(strings.Join(wildlife, ", ")) + "\n")
	b.WriteString(brightGreen.Render("Insects (Auto)") + "\n")
	b.WriteString(green.Render(strings.Join(insects, ", ")) + "\n\n")
	b.WriteString(dimGreen.Render(desc) + "\n\n")
	b.WriteString(brightGreen.Render("Daunting") + "\n")
	b.WriteString(green.Render(daunting) + "\n\n")
	b.WriteString(brightGreen.Render("Motivation") + "\n")
	b.WriteString(green.Render(motivation))
	return b.String()
}

func scenarioWildlife(s game.Scenario) []string {
	if len(s.Wildlife) > 0 {
		return append([]string(nil), s.Wildlife...)
	}
	return game.WildlifeForBiome(s.Biome)
}

func (m menuModel) openPersonalKitPicker(returnTo screen) menuModel {
	m = m.ensureSetupPlayers()
	m = m.normalizePlayerConfigState()
	m.kit = newKitPickerState()
	m.kit.target = kitTargetPersonal
	m.kit.returnTo = returnTo
	m.kit.categoryIdx = 0
	m.kit.itemIdx = 0
	m.kit.focus = kitFocusCategories
	if len(m.setup.players) > 0 {
		playerKit := m.setup.players[m.pcfg.playerIdx].Kit
		if len(playerKit) > 0 {
			categories := m.kitPickerCategories()
			for _, item := range playerKit {
				if catIdx, itemIdx, ok := findKitPickerPosition(categories, item); ok {
					m.kit.categoryIdx = catIdx
					m.kit.itemIdx = itemIdx
					break
				}
			}
		}
	}
	m = m.normalizeKitPickerState(m.kitPickerCategories())
	m.screen = screenKitPicker
	return m
}

func (m menuModel) openIssuedKitPicker(returnTo screen) menuModel {
	m = m.ensureSetupPlayers()
	m.kit = newKitPickerState()
	m.kit.target = kitTargetIssued
	m.kit.returnTo = returnTo
	m.kit.categoryIdx = 0
	m.kit.itemIdx = 0
	m.kit.focus = kitFocusCategories
	if len(m.setup.issuedKit) > 0 {
		categories := m.kitPickerCategories()
		for _, item := range m.setup.issuedKit {
			if catIdx, itemIdx, ok := findKitPickerPosition(categories, item); ok {
				m.kit.categoryIdx = catIdx
				m.kit.itemIdx = itemIdx
				break
			}
		}
	}
	m = m.normalizeKitPickerState(m.kitPickerCategories())
	m.screen = screenKitPicker
	return m
}

func (m menuModel) kitPickerItems() []game.KitItem {
	if m.kit.target == kitTargetIssued {
		items := append([]game.KitItem(nil), issuedKitOptionsForMode(m.setupMode())...)
		for _, selected := range m.setup.issuedKit {
			if hasKitItem(items, selected) {
				continue
			}
			items = append(items, selected)
		}
		return items
	}
	return game.AllKitItems()
}

type kitCategory struct {
	label string
	items []game.KitItem
}

func (m menuModel) kitPickerCategories() []kitCategory {
	return categorizeKitItems(m.kitPickerItems())
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
		itemsForLabel := buckets[label]
		if len(itemsForLabel) == 0 {
			continue
		}
		out = append(out, kitCategory{
			label: label,
			items: itemsForLabel,
		})
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

func findKitPickerPosition(categories []kitCategory, item game.KitItem) (int, int, bool) {
	for catIdx, category := range categories {
		for itemIdx, candidate := range category.items {
			if candidate == item {
				return catIdx, itemIdx, true
			}
		}
	}
	return 0, 0, false
}

func (m menuModel) normalizeKitPickerState(categories []kitCategory) menuModel {
	if len(categories) == 0 {
		m.kit.categoryIdx = 0
		m.kit.itemIdx = 0
		m.kit.focus = kitFocusCategories
		return m
	}
	if m.kit.categoryIdx < 0 || m.kit.categoryIdx >= len(categories) {
		m.kit.categoryIdx = 0
	}
	currentItems := categories[m.kit.categoryIdx].items
	if len(currentItems) == 0 {
		m.kit.itemIdx = 0
		m.kit.focus = kitFocusCategories
		return m
	}
	if m.kit.itemIdx < 0 || m.kit.itemIdx >= len(currentItems) {
		m.kit.itemIdx = 0
	}
	if m.kit.focus != kitFocusCategories && m.kit.focus != kitFocusItems {
		m.kit.focus = kitFocusCategories
	}
	return m
}

func (m menuModel) updateKitPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	categories := m.kitPickerCategories()
	if len(categories) == 0 {
		m.screen = m.kit.returnTo
		m.status = "No kit items available."
		return m, nil
	}
	m = m.normalizeKitPickerState(categories)
	items := categories[m.kit.categoryIdx].items

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "Q", "q", "esc":
		m.screen = m.kit.returnTo
		return m, nil
	case "left", "h":
		m.kit.focus = kitFocusCategories
		return m, nil
	case "right", "l":
		if len(items) > 0 {
			m.kit.focus = kitFocusItems
		}
		return m, nil
	case "up", "k":
		if m.kit.focus == kitFocusCategories {
			m.kit.categoryIdx = wrapIndex(m.kit.categoryIdx, -1, len(categories))
			m = m.normalizeKitPickerState(categories)
		} else if len(items) > 0 {
			m.kit.itemIdx = wrapIndex(m.kit.itemIdx, -1, len(items))
		}
		return m, nil
	case "down", "j":
		if m.kit.focus == kitFocusCategories {
			m.kit.categoryIdx = wrapIndex(m.kit.categoryIdx, 1, len(categories))
			m = m.normalizeKitPickerState(categories)
		} else if len(items) > 0 {
			m.kit.itemIdx = wrapIndex(m.kit.itemIdx, 1, len(items))
		}
		return m, nil
	case "enter":
		if m.kit.focus == kitFocusCategories {
			if len(items) == 0 {
				m.status = "No items in this category."
				return m, nil
			}
			m.kit.focus = kitFocusItems
			return m, nil
		}
		if len(items) == 0 {
			return m, nil
		}
		m = m.toggleKitPickerSelection(items[m.kit.itemIdx])
		return m, nil
	case " ":
		if m.kit.focus == kitFocusItems && len(items) > 0 {
			m = m.toggleKitPickerSelection(items[m.kit.itemIdx])
		}
		return m, nil
	case "R", "r":
		if m.kit.target == kitTargetIssued {
			m = m.resetIssuedKitRecommendations()
			categories = m.kitPickerCategories()
			m = m.normalizeKitPickerState(categories)
			return m, nil
		}
	}

	return m, nil
}

func (m menuModel) toggleKitPickerSelection(item game.KitItem) menuModel {
	switch m.kit.target {
	case kitTargetPersonal:
		if len(m.setup.players) == 0 {
			return m
		}
		m = m.normalizePlayerConfigState()
		p := &m.setup.players[m.pcfg.playerIdx]
		if idx := indexOfKitItem(p.Kit, item); idx >= 0 {
			p.Kit = removeKitItemAt(p.Kit, idx)
			m.status = ""
			return m
		}
		if len(p.Kit) >= p.KitLimit {
			m.status = fmt.Sprintf("Limit reached for %s (%d item max).", playerDisplayName(*p), p.KitLimit)
			return m
		}
		p.Kit = append(p.Kit, item)
		m.status = ""
	case kitTargetIssued:
		if idx := indexOfKitItem(m.setup.issuedKit, item); idx >= 0 {
			m.setup.issuedKit = removeKitItemAt(m.setup.issuedKit, idx)
			m.setup.issuedCustom = true
			m.status = ""
			return m
		}
		m.setup.issuedKit = append(m.setup.issuedKit, item)
		m.setup.issuedCustom = true
		m.status = ""
	}
	return m
}

func playerDisplayName(p game.PlayerConfig) string {
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return "player"
	}
	return name
}

func playerEditorSlotLabel(idx, total int) string {
	if idx == 0 {
		return fmt.Sprintf("%d / %d (YOU)", idx+1, total)
	}
	return fmt.Sprintf("%d / %d", idx+1, total)
}

func (m menuModel) viewKitPicker() string {
	categories := m.kitPickerCategories()
	if len(categories) == 0 {
		return brightGreen.Render("No kit items available.")
	}
	m = m.normalizeKitPickerState(categories)

	title := "KIT PICKER"
	sub := "Toggle with Enter"
	selectedCount := 0
	limit := 0
	selected := map[game.KitItem]bool{}

	if m.kit.target == kitTargetPersonal && len(m.setup.players) > 0 {
		m = m.normalizePlayerConfigState()
		p := m.setup.players[m.pcfg.playerIdx]
		title = fmt.Sprintf("KIT PICKER :: PLAYER %d", m.pcfg.playerIdx+1)
		sub = fmt.Sprintf("Name: %s  |  Limit: %d", playerDisplayName(p), p.KitLimit)
		limit = p.KitLimit
		selectedCount = len(p.Kit)
		for _, item := range p.Kit {
			selected[item] = true
		}
	} else {
		title = "KIT PICKER :: ISSUED"
		if m.setup.issuedCustom {
			sub = "Scenario-adjusted custom issued kit"
		} else {
			sub = "Scenario recommendation active"
		}
		selectedCount = len(m.setup.issuedKit)
		for _, item := range m.setup.issuedKit {
			selected[item] = true
		}
	}

	currentCategory := categories[m.kit.categoryIdx]
	currentItems := currentCategory.items

	totalWidth := m.w - 2
	if totalWidth < 108 {
		totalWidth = 108
	}
	contentHeight := m.h - 8
	if contentHeight < 16 {
		contentHeight = 16
	}
	categoryPaneTotal := totalWidth / 4
	if categoryPaneTotal < 24 {
		categoryPaneTotal = 24
	}
	if categoryPaneTotal > 36 {
		categoryPaneTotal = 36
	}
	itemPaneTotal := totalWidth / 3
	if itemPaneTotal < 32 {
		itemPaneTotal = 32
	}
	if itemPaneTotal > 52 {
		itemPaneTotal = 52
	}
	detailPaneTotal := totalWidth - categoryPaneTotal - itemPaneTotal
	if detailPaneTotal < 40 {
		detailPaneTotal = 40
		itemPaneTotal = totalWidth - categoryPaneTotal - detailPaneTotal
		if itemPaneTotal < 30 {
			itemPaneTotal = 30
			categoryPaneTotal = totalWidth - itemPaneTotal - detailPaneTotal
		}
	}
	categoryWidth := categoryPaneTotal - 4
	if categoryWidth < 18 {
		categoryWidth = 18
	}
	itemWidth := itemPaneTotal - 4
	if itemWidth < 24 {
		itemWidth = 24
	}
	detailWidth := detailPaneTotal - 4
	if detailWidth < 30 {
		detailWidth = 30
	}

	categoryLines := make([]string, 0, len(categories))
	for i, category := range categories {
		cursor := "  "
		line := truncateForPane(fmt.Sprintf("%s (%d)", category.label, len(category.items)), categoryWidth-3)
		lineStyle := green
		if i == m.kit.categoryIdx {
			cursor = "> "
			if m.kit.focus == kitFocusCategories {
				lineStyle = brightGreen
			}
		}
		categoryLines = append(categoryLines, cursor+lineStyle.Render(line))
	}
	categoryLines = kitPickerViewportLines(categoryLines, m.kit.categoryIdx, contentHeight)

	itemLines := make([]string, 0, len(currentItems))
	for i, item := range currentItems {
		cursor := "  "
		line := fmt.Sprintf("[ ] %s", item)
		if selected[item] {
			line = fmt.Sprintf("[*] %s", item)
		}
		line = truncateForPane(line, itemWidth-3)
		lineStyle := green
		if i == m.kit.itemIdx {
			cursor = "> "
			if m.kit.focus == kitFocusItems {
				lineStyle = brightGreen
			}
		}
		itemLines = append(itemLines, cursor+lineStyle.Render(line))
	}
	if len(itemLines) == 0 {
		itemLines = []string{dimGreen.Render("No items in category")}
	} else {
		itemLines = kitPickerViewportLines(itemLines, m.kit.itemIdx, contentHeight)
	}

	selectedList := make([]string, 0, selectedCount)
	for _, category := range categories {
		for _, item := range category.items {
			if selected[item] {
				selectedList = append(selectedList, string(item))
			}
		}
	}
	selectedLines := []string{"- none"}
	if len(selectedList) > 0 {
		selectedLines = make([]string, 0, len(selectedList))
		for _, item := range selectedList {
			selectedLines = append(selectedLines, "- "+item)
		}
	}
	const selectedPreviewMax = 8
	selectedPreview := selectedLines
	if len(selectedPreview) > selectedPreviewMax {
		remaining := len(selectedPreview) - selectedPreviewMax
		selectedPreview = append(selectedPreview[:selectedPreviewMax], fmt.Sprintf("... +%d more", remaining))
	}

	activeItemLabel := "<none>"
	activeFlavor := "Select a category, then choose an item."
	if len(currentItems) > 0 {
		activeItem := currentItems[m.kit.itemIdx]
		activeItemLabel = string(activeItem)
		activeFlavor = kitItemFlavorText(activeItem)
	}
	var detail strings.Builder
	detail.WriteString(brightGreen.Render("Selected Items") + "\n")
	if limit > 0 {
		detail.WriteString(green.Render(fmt.Sprintf("Count: %d/%d", selectedCount, limit)) + "\n\n")
	} else {
		detail.WriteString(green.Render(fmt.Sprintf("Count: %d", selectedCount)) + "\n\n")
	}
	detail.WriteString(green.Render(strings.Join(selectedPreview, "\n")) + "\n\n")
	detail.WriteString(brightGreen.Render("Focused Category") + "\n")
	detail.WriteString(green.Render(currentCategory.label) + "\n\n")
	detail.WriteString(brightGreen.Render("Focused Item") + "\n")
	detail.WriteString(green.Render(activeItemLabel) + "\n")
	detail.WriteString(dimGreen.Render(activeFlavor) + "\n\n")
	detail.WriteString(dimGreen.Render("Enter on category opens item list.") + "\n")
	detail.WriteString(dimGreen.Render("Enter/Space on item toggles selection.") + "\n")
	if m.kit.target == kitTargetIssued {
		detail.WriteString(dimGreen.Render("Shift+R resets issued kit to recommendation.") + "\n")
	}

	var b strings.Builder
	b.WriteString(dosTitle(title) + "\n")
	if limit > 0 {
		b.WriteString(dimGreen.Render(fmt.Sprintf("%s  |  Selected: %d/%d", sub, selectedCount, limit)) + "\n")
	} else {
		b.WriteString(dimGreen.Render(fmt.Sprintf("%s  |  Selected: %d", sub, selectedCount)) + "\n")
	}
	b.WriteString(border.Render(dosRule(totalWidth)) + "\n\n")
	categoryPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(categoryWidth).
		Height(contentHeight).
		Render(strings.Join(categoryLines, "\n"))
	itemPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(itemWidth).
		Height(contentHeight).
		Render(strings.Join(itemLines, "\n"))
	detailPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(detailWidth).
		Height(contentHeight).
		Render(detail.String())
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, categoryPane, itemPane, detailPane))
	b.WriteString("\n" + border.Render(dosRule(totalWidth)) + "\n")
	if m.kit.target == kitTargetIssued {
		b.WriteString(dimGreen.Render("↑/↓ move  ←/→ switch pane  Enter select/toggle  R reset recommended  Shift+Q back") + "\n")
	} else {
		b.WriteString(dimGreen.Render("↑/↓ move  ←/→ switch pane  Enter select/toggle  Shift+Q back") + "\n")
	}
	if m.status != "" {
		b.WriteString(green.Render(m.status) + "\n")
	}
	return b.String()
}

func truncateForPane(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if len(text) <= width {
		return text
	}
	if width <= 3 {
		return text[:width]
	}
	return text[:width-3] + "..."
}

func kitPickerViewportLines(lines []string, cursor, height int) []string {
	if height <= 0 {
		return []string{}
	}
	if len(lines) <= height {
		return lines
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= len(lines) {
		cursor = len(lines) - 1
	}
	start := cursor - height/2
	if start < 0 {
		start = 0
	}
	maxStart := len(lines) - height
	if start > maxStart {
		start = maxStart
	}
	return lines[start : start+height]
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func kitItemFlavorText(item game.KitItem) string {
	name := strings.ToLower(string(item))
	switch {
	case strings.Contains(name, "knife"), strings.Contains(name, "hatchet"), strings.Contains(name, "saw"), strings.Contains(name, "machete"):
		return "Tool item: useful for shelter building, carving, and camp maintenance."
	case strings.Contains(name, "water"), strings.Contains(name, "canteen"), strings.Contains(name, "purification"):
		return "Water item: improves hydration safety and access."
	case strings.Contains(name, "ferro"), strings.Contains(name, "fire"), strings.Contains(name, "magnifying"):
		return "Fire item: improves fire-start options and heat reliability."
	case strings.Contains(name, "tarp"), strings.Contains(name, "blanket"), strings.Contains(name, "sleep"), strings.Contains(name, "thermal"):
		return "Shelter item: helps with exposure management and recovery."
	case strings.Contains(name, "fishing"), strings.Contains(name, "snare"), strings.Contains(name, "bow"):
		return "Food item: supports hunting or fishing strategy."
	case strings.Contains(name, "first aid"), strings.Contains(name, "insect"), strings.Contains(name, "net"):
		return "Protection item: helps reduce injury/insect burden."
	default:
		return "General survival item: evaluate role against scenario conditions."
	}
}

func (m menuModel) updatePlayerConfig(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m = m.ensureSetupScenarioSelection()
	m = m.ensureSetupPlayers()
	m = m.normalizePlayerConfigState()
	rows := m.playerConfigRows()
	if len(rows) == 0 {
		m.screen = screenSetup
		return m, nil
	}
	if m.pcfg.cursor < 0 || m.pcfg.cursor >= len(rows) {
		m.pcfg.cursor = 0
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "Q", "q", "esc":
		m.screen = screenSetup
		return m, nil
	case "up", "k":
		m.pcfg.cursor = wrapIndex(m.pcfg.cursor, -1, len(rows))
		return m, nil
	case "down", "j":
		m.pcfg.cursor = wrapIndex(m.pcfg.cursor, 1, len(rows))
		return m, nil
	case "left":
		m = m.adjustPlayerConfigChoice(-1)
		return m, nil
	case "right":
		m = m.adjustPlayerConfigChoice(1)
		return m, nil
	case "backspace", "ctrl+h":
		m = m.backspacePlayerConfigText()
		return m, nil
	case "enter":
		row := rows[m.pcfg.cursor]
		if !row.active {
			return m, nil
		}
		switch row.kind {
		case playerRowEditPersonalKit:
			m = m.openPersonalKitPicker(screenPlayerConfig)
			return m, nil
		case playerRowResetPersonalKit:
			m = m.resetPersonalKitSelection()
		case playerRowEditIssuedKit:
			m = m.openIssuedKitPicker(screenPlayerConfig)
			return m, nil
		case playerRowResetIssuedKit:
			m = m.resetIssuedKitRecommendations()
		case playerRowBack:
			if m.pcfg.returnTo == 0 {
				m.pcfg.returnTo = screenSetup
			}
			m.screen = m.pcfg.returnTo
			return m, nil
		default:
			if m.playerConfigRowSupportsCycle(row.kind) {
				m = m.adjustPlayerConfigChoice(1)
			}
		}
		return m, nil
	}

	if len(msg.Runes) > 0 {
		m = m.appendPlayerConfigText(msg.Runes)
		return m, nil
	}

	return m, nil
}

func (m menuModel) viewPlayerConfig() string {
	m = m.ensureSetupScenarioSelection()
	m = m.ensureSetupPlayers()
	m = m.normalizePlayerConfigState()
	rows := m.playerConfigRows()
	if len(rows) == 0 {
		return brightGreen.Render("No players available. Return to wizard and choose player count.")
	}
	if m.pcfg.cursor < 0 || m.pcfg.cursor >= len(rows) {
		m.pcfg.cursor = 0
	}

	scenarioLabel := "Unknown scenario"
	if selected, found := m.selectedSetupScenario(); found {
		scenarioLabel = selected.Name
	}

	playerCount := len(m.setup.players)
	activePlayer := m.pcfg.playerIdx + 1
	current := m.setup.players[m.pcfg.playerIdx]

	totalWidth := m.w
	if totalWidth < 128 {
		totalWidth = 128
	}
	contentHeight := m.h - 9
	if contentHeight < 16 {
		contentHeight = 16
	}
	listWidth := totalWidth / 3
	if listWidth < 34 {
		listWidth = 34
	}
	statsWidth := totalWidth / 3
	if statsWidth < 38 {
		statsWidth = 38
	}
	asciiWidth := totalWidth - listWidth - statsWidth - 6
	if asciiWidth < 34 {
		asciiWidth = 34
	}

	var list strings.Builder
	for i, row := range rows {
		cursor := "  "
		lineStyle := green
		if i == m.pcfg.cursor {
			cursor = "> "
			lineStyle = brightGreen
		}
		if !row.active {
			lineStyle = dimGreen
		}

		if row.value == "" {
			list.WriteString(cursor + lineStyle.Render(row.label) + "\n")
			continue
		}

		value := row.value
		if row.kind == playerRowName && strings.TrimSpace(value) == "" {
			value = "<auto/random or type custom name>"
		}
		list.WriteString(cursor + lineStyle.Render(fmt.Sprintf("%-24s %s", row.label+":", value)) + "\n")
	}

	detail := m.playerEditorStatsText(current, scenarioLabel, activePlayer, playerCount)
	ascii := m.playerEditorAsciiText(current, asciiWidth-4, contentHeight-4)
	listPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(listWidth).
		Height(contentHeight).
		Render(list.String())
	statsPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(statsWidth).
		Height(contentHeight).
		Render(detail)
	asciiPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(asciiWidth).
		Height(contentHeight).
		Render(ascii)

	var b strings.Builder
	b.WriteString(dosTitle("Player Editor") + "\n")
	b.WriteString(dimGreen.Render(fmt.Sprintf("Mode: %s  |  Scenario: %s  |  Player %d/%d", modeLabel(m.setupMode()), scenarioLabel, activePlayer, playerCount)) + "\n")
	b.WriteString("\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listPane, statsPane, asciiPane))
	b.WriteString("\n" + dimGreen.Render("↑/↓ move  ←/→ change values  Enter open/select  type for Name  Shift+Q back") + "\n")
	b.WriteString(dimGreen.Render("Use Open Kit Picker rows for full kit selection screens.") + "\n")
	if m.status != "" {
		b.WriteString(green.Render(m.status) + "\n")
	}

	return b.String()
}

func (m menuModel) playerEditorStatsText(p game.PlayerConfig, scenarioLabel string, activePlayer, playerCount int) string {
	name := strings.TrimSpace(p.Name)
	if name == "" {
		name = fmt.Sprintf("Player %d", activePlayer)
	}
	editing := fmt.Sprintf("Editing: TEAMMATE %d", activePlayer-1)
	if activePlayer == 1 {
		editing = "Editing: YOU"
	}
	renderKitList := func(items []game.KitItem) string {
		if len(items) == 0 {
			return green.Render("- none")
		}
		lines := make([]string, 0, len(items))
		for _, item := range items {
			lines = append(lines, green.Render("- "+string(item)))
		}
		return strings.Join(lines, "\n")
	}
	issuedModeLine := green.Render(fmt.Sprintf("Issued options follow %s rules.", modeLabel(m.setupMode())))

	return strings.Join([]string{
		brightGreen.Render(name),
		green.Render(fmt.Sprintf("Player Slot: %d/%d", activePlayer, playerCount)),
		brightGreen.Render(editing),
		green.Render(fmt.Sprintf("Game Mode: %s", modeLabel(m.setupMode()))),
		green.Render(fmt.Sprintf("Scenario: %s", scenarioLabel)),
		green.Render(fmt.Sprintf("Sex: %s  |  Body: %s", p.Sex, p.BodyType)),
		green.Render(fmt.Sprintf("Height: %d ft %d in  |  Weight: %d kg", p.HeightFt, p.HeightIn, p.WeightKg)),
		green.Render(fmt.Sprintf("Mods  End:%+d  Bush:%+d  Ment:%+d", p.Endurance, p.Bushcraft, p.Mental)),
		green.Render(fmt.Sprintf("Kit  %d/%d selected", len(p.Kit), p.KitLimit)),
		"",
		brightGreen.Render("Selected Personal Kit"),
		renderKitList(p.Kit),
		"",
		brightGreen.Render("Selected Issued Kit"),
		renderKitList(m.setup.issuedKit),
		issuedModeLine,
		"",
		brightGreen.Render("Player Notes"),
		dimGreen.Render(fmt.Sprintf("Series limit: up to %d personal kit item(s) per player.", maxKitLimitForMode(m.setupMode()))),
		dimGreen.Render("Kit limit is constrained by selected game mode."),
	}, "\n")
}

func (m menuModel) playerEditorAsciiText(p game.PlayerConfig, widthChars, heightRows int) string {
	return renderPlayerBodyANSI(p, widthChars, heightRows)
}

func playerASCIIArt(p game.PlayerConfig) string {
	heightScore := (p.HeightFt-4)*12 + p.HeightIn
	switch {
	case heightScore < 14:
		heightScore = 14
	case heightScore > 48:
		heightScore = 48
	}

	weight := p.WeightKg
	if weight < 45 {
		weight = 45
	}
	if weight > 140 {
		weight = 140
	}

	bodyWidth := 3
	switch {
	case weight >= 105:
		bodyWidth = 8
	case weight >= 90:
		bodyWidth = 7
	case weight >= 75:
		bodyWidth = 6
	case weight >= 60:
		bodyWidth = 5
	}

	switch p.BodyType {
	case game.BodyTypeMale:
		bodyWidth++
	case game.BodyTypeFemale:
		if bodyWidth > 3 {
			bodyWidth--
		}
	}
	if bodyWidth < 3 {
		bodyWidth = 3
	}
	if bodyWidth > 10 {
		bodyWidth = 10
	}

	legRows := 3
	switch {
	case heightScore >= 38:
		legRows = 7
	case heightScore >= 32:
		legRows = 6
	case heightScore >= 26:
		legRows = 5
	}
	shoulderWidth := bodyWidth + 4
	figureWidth := maxInt(22, shoulderWidth+12)
	torsoFill := "#"
	if p.BodyType == game.BodyTypeFemale {
		torsoFill = "*"
	}
	if p.BodyType == game.BodyTypeNeutral {
		torsoFill = "+"
	}

	center := func(text string) string {
		padding := (figureWidth - len(text)) / 2
		if padding < 0 {
			padding = 0
		}
		return strings.Repeat(" ", padding) + text
	}

	armLine := "/" + strings.Repeat(" ", shoulderWidth+2) + "\\"
	shoulderLine := "." + strings.Repeat("_", shoulderWidth) + "."
	torsoLine := "| " + strings.Repeat(torsoFill, bodyWidth) + " |"
	waistLine := "|" + strings.Repeat("_", bodyWidth+2) + "|"
	legGap := maxInt(2, bodyWidth-1)
	shin := maxInt(1, bodyWidth/3)

	var b strings.Builder
	b.WriteString(center(" ___ ") + "\n")
	b.WriteString(center("/o o\\") + "\n")
	b.WriteString(center("\\_^_/") + "\n")
	b.WriteString(center(shoulderLine) + "\n")
	b.WriteString(center(armLine) + "\n")
	for i := 0; i < 2; i++ {
		b.WriteString(center(torsoLine) + "\n")
	}
	b.WriteString(center(waistLine) + "\n")
	for i := 0; i < legRows; i++ {
		spread := legGap + i/2
		b.WriteString(center("/"+strings.Repeat(" ", shin)+"\\"+strings.Repeat(" ", spread)+"/"+strings.Repeat(" ", shin)+"\\") + "\n")
	}
	b.WriteString(center("/_"+strings.Repeat("_", shin)+"\\"+strings.Repeat(" ", legGap+legRows/2)+"/"+strings.Repeat("_", shin)+"_\\") + "\n")
	return b.String()
}

func (m menuModel) playerConfigRows() []playerConfigRow {
	if len(m.setup.players) == 0 {
		return nil
	}

	allKit := game.AllKitItems()
	playerIdx := m.pcfg.playerIdx
	if playerIdx < 0 || playerIdx >= len(m.setup.players) {
		playerIdx = 0
	}

	p := m.setup.players[playerIdx]

	return []playerConfigRow{
		{label: "Player", value: playerEditorSlotLabel(playerIdx, len(m.setup.players)), kind: playerRowPlayer, active: true},
		{label: "Name", value: p.Name, kind: playerRowName, active: true},
		{label: "Sex", value: string(p.Sex), kind: playerRowSex, active: true},
		{label: "Body Type", value: string(p.BodyType), kind: playerRowBodyType, active: true},
		{label: "Weight (kg)", value: fmt.Sprintf("%d", p.WeightKg), kind: playerRowWeightKg, active: true},
		{label: "Height (ft)", value: fmt.Sprintf("%d", p.HeightFt), kind: playerRowHeightFt, active: true},
		{label: "Height (in)", value: fmt.Sprintf("%d", p.HeightIn), kind: playerRowHeightIn, active: true},
		{label: "Endurance Modifier", value: fmt.Sprintf("%+d", p.Endurance), kind: playerRowEndurance, active: true},
		{label: "Bushcraft Modifier", value: fmt.Sprintf("%+d", p.Bushcraft), kind: playerRowBushcraft, active: true},
		{label: "Mental Modifier", value: fmt.Sprintf("%+d", p.Mental), kind: playerRowMental, active: true},
		{label: "Kit Limit", value: fmt.Sprintf("%d", p.KitLimit), kind: playerRowKitLimit, active: true},
		{label: "Open Personal Kit Picker", value: "", kind: playerRowEditPersonalKit, active: len(allKit) > 0},
		{label: "Reset Personal Kit", value: "", kind: playerRowResetPersonalKit, active: len(p.Kit) > 0},
		{label: "Open Issued Kit Picker", value: "", kind: playerRowEditIssuedKit, active: len(issuedKitOptionsForMode(m.setupMode())) > 0},
		{label: "Reset Issued", value: "", kind: playerRowResetIssuedKit, active: true},
		{label: "Back To Wizard", value: "", kind: playerRowBack, active: true},
	}
}

func (m menuModel) playerConfigRowSupportsCycle(kind playerConfigRowKind) bool {
	switch kind {
	case playerRowPlayer, playerRowSex, playerRowBodyType, playerRowWeightKg,
		playerRowHeightFt, playerRowHeightIn, playerRowEndurance, playerRowBushcraft,
		playerRowMental, playerRowKitLimit:
		return true
	default:
		return false
	}
}

func (m menuModel) adjustPlayerConfigChoice(delta int) menuModel {
	if len(m.setup.players) == 0 {
		return m
	}
	m = m.normalizePlayerConfigState()
	rows := m.playerConfigRows()
	if len(rows) == 0 || m.pcfg.cursor < 0 || m.pcfg.cursor >= len(rows) {
		return m
	}
	row := rows[m.pcfg.cursor]
	if !row.active {
		return m
	}

	p := &m.setup.players[m.pcfg.playerIdx]
	switch row.kind {
	case playerRowPlayer:
		m.pcfg.playerIdx = wrapIndex(m.pcfg.playerIdx, delta, len(m.setup.players))
	case playerRowSex:
		options := playerConfigSexes()
		idx := indexOfSex(options, p.Sex)
		p.Sex = options[wrapIndex(idx, delta, len(options))]
	case playerRowBodyType:
		options := playerConfigBodyTypes()
		idx := indexOfBodyType(options, p.BodyType)
		p.BodyType = options[wrapIndex(idx, delta, len(options))]
	case playerRowWeightKg:
		p.WeightKg = clampInt(p.WeightKg+delta, 35, 220)
	case playerRowHeightFt:
		p.HeightFt = clampInt(p.HeightFt+delta, 4, 7)
	case playerRowHeightIn:
		p.HeightIn = clampInt(p.HeightIn+delta, 0, 11)
	case playerRowEndurance:
		p.Endurance = clampInt(p.Endurance+delta, -3, 3)
	case playerRowBushcraft:
		p.Bushcraft = clampInt(p.Bushcraft+delta, -3, 3)
	case playerRowMental:
		p.Mental = clampInt(p.Mental+delta, -3, 3)
	case playerRowKitLimit:
		maxLimit := maxKitLimitForMode(m.setupMode())
		p.KitLimit = clampInt(p.KitLimit+delta, 1, maxLimit)
		if len(p.Kit) > p.KitLimit {
			p.Kit = append([]game.KitItem(nil), p.Kit[:p.KitLimit]...)
		}
	}

	return m
}

func (m menuModel) appendPlayerConfigText(runes []rune) menuModel {
	if len(m.setup.players) == 0 {
		return m
	}
	m = m.normalizePlayerConfigState()
	rows := m.playerConfigRows()
	if len(rows) == 0 || m.pcfg.cursor < 0 || m.pcfg.cursor >= len(rows) {
		return m
	}
	row := rows[m.pcfg.cursor]
	if row.kind != playerRowName {
		return m
	}

	p := &m.setup.players[m.pcfg.playerIdx]
	for _, r := range runes {
		if r >= 32 && r <= 126 {
			if len(p.Name) >= 32 {
				break
			}
			p.Name += string(r)
		}
	}
	return m
}

func (m menuModel) backspacePlayerConfigText() menuModel {
	if len(m.setup.players) == 0 {
		return m
	}
	m = m.normalizePlayerConfigState()
	rows := m.playerConfigRows()
	if len(rows) == 0 || m.pcfg.cursor < 0 || m.pcfg.cursor >= len(rows) {
		return m
	}
	row := rows[m.pcfg.cursor]
	if row.kind != playerRowName {
		return m
	}

	p := &m.setup.players[m.pcfg.playerIdx]
	if len(p.Name) > 0 {
		p.Name = p.Name[:len(p.Name)-1]
	}
	return m
}

func (m menuModel) resetPersonalKitSelection() menuModel {
	if len(m.setup.players) == 0 {
		return m
	}
	m = m.normalizePlayerConfigState()
	p := &m.setup.players[m.pcfg.playerIdx]
	if len(p.Kit) == 0 {
		return m
	}
	p.Kit = nil
	m.status = "Personal kit reset."
	return m
}

func (m menuModel) resetIssuedKitRecommendations() menuModel {
	scenario, found := m.selectedSetupScenario()
	if !found {
		m.status = "No scenario selected."
		return m
	}
	m.setup.issuedCustom = false
	m.setup.issuedKit = recommendedIssuedKitForScenario(m.setupMode(), scenario)
	m.status = "Issued kit reset to scenario recommendations."
	return m
}

func (m menuModel) normalizePlayerConfigState() menuModel {
	if len(m.setup.players) > 0 {
		if m.pcfg.playerIdx < 0 || m.pcfg.playerIdx >= len(m.setup.players) {
			m.pcfg.playerIdx = 0
		}
	} else {
		m.pcfg.playerIdx = 0
	}

	return m
}

func hasKitItem(items []game.KitItem, target game.KitItem) bool {
	return indexOfKitItem(items, target) >= 0
}

func indexOfKitItem(items []game.KitItem, target game.KitItem) int {
	for i, item := range items {
		if item == target {
			return i
		}
	}
	return -1
}

func removeKitItemAt(items []game.KitItem, idx int) []game.KitItem {
	if idx < 0 || idx >= len(items) {
		return items
	}
	return append(append([]game.KitItem{}, items[:idx]...), items[idx+1:]...)
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

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func indexOfSex(values []game.Sex, target game.Sex) int {
	for i, value := range values {
		if value == target {
			return i
		}
	}
	return 0
}

func indexOfBodyType(values []game.BodyType, target game.BodyType) int {
	for i, value := range values {
		if value == target {
			return i
		}
	}
	return 0
}

func (m menuModel) availableScenarios() []game.Scenario {
	scenarios := append([]game.Scenario{}, game.BuiltInScenarios()...)
	for _, record := range m.customScenarios {
		scenario := record.Scenario
		if len(scenario.SupportedModes) == 0 {
			scenario.SupportedModes = []game.GameMode{record.PreferredMode}
		}
		scenarios = append(scenarios, scenario)
	}
	return scenarios
}

func (m menuModel) preferredModeForScenario(id game.ScenarioID) (game.GameMode, bool) {
	for _, record := range m.customScenarios {
		if record.Scenario.ID == id {
			return record.PreferredMode, true
		}
	}
	return "", false
}

func isCustomScenarioID(id game.ScenarioID) bool {
	return strings.HasPrefix(string(id), "custom_")
}

func (m menuModel) updateLoadRun(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	rowCount := len(m.load.entries) + 1 // files + cancel
	if rowCount <= 1 {
		rowCount = 1
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "Q", "q", "esc":
		if m.loadReturn == 0 {
			next, cmd := m.returnToMainMenu()
			return next, cmd
		}
		m.screen = m.loadReturn
		return m, nil
	case "up", "k":
		m.load.cursor = wrapIndex(m.load.cursor, -1, rowCount)
		return m, nil
	case "down", "j":
		m.load.cursor = wrapIndex(m.load.cursor, 1, rowCount)
		return m, nil
	case "enter":
		if m.load.cursor >= len(m.load.entries) {
			if m.loadReturn == 0 {
				next, cmd := m.returnToMainMenu()
				return next, cmd
			}
			m.screen = m.loadReturn
			return m, nil
		}

		selected := m.load.entries[m.load.cursor]
		state, err := loadRunFromFile(selected.Path, m.availableScenarios())
		if err != nil {
			m.status = fmt.Sprintf("Load failed: %v", err)
			return m, nil
		}

		m.run = &state
		m.runPlayedFor = 0
		if slot, ok := slotNumberFromPath(selected.Path); ok {
			m.activeSaveSlot = slot
		}
		m.screen = screenRun
		m.status = fmt.Sprintf("Loaded %s", selected.Path)
		return m, nil
	}

	return m, nil
}

func (m menuModel) openLoadRun(returnTo screen) menuModel {
	m.load = newLoadRunState()
	m.loadReturn = returnTo
	m.load.entries = loadRunEntries()
	m.screen = screenLoadRun
	if len(m.load.entries) == 0 {
		m.status = "No saves found. Use Shift+S during a run to create one."
	}
	return m
}

func loadRunEntries() []saveSlotMeta {
	dirEntries, err := os.ReadDir(".")
	if err != nil {
		return nil
	}
	paths := make([]string, 0, len(dirEntries))
	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if saveFilePattern.MatchString(name) {
			paths = append(paths, name)
		}
	}
	sort.Strings(paths)

	entries := make([]saveSlotMeta, 0, len(paths))
	for _, path := range paths {
		meta := loadRunMetadata(path)
		entries = append(entries, meta)
	}
	return entries
}

func loadRunMetadata(path string) saveSlotMeta {
	meta := saveSlotMeta{
		Path: path,
	}
	if slot, ok := slotNumberFromPath(path); ok {
		meta.Slot = slot
	}

	data, err := readDataFile(path, maxSaveFileBytes)
	if err != nil {
		meta.Exists = false
		if errors.Is(err, os.ErrNotExist) {
			meta.Summary = "Missing"
		} else {
			meta.Summary = "Unreadable save"
			meta.ErrDetail = err.Error()
		}
		return meta
	}

	var payload savedRun
	if err := json.Unmarshal(data, &payload); err != nil {
		meta.Exists = true
		meta.Summary = "Unreadable save"
		meta.ErrDetail = err.Error()
		return meta
	}

	meta.Exists = true
	meta.SavedAt = payload.SavedAt
	meta.Run = &payload.Run
	meta.Summary = fmt.Sprintf("%s | Day %d | %s",
		payload.Run.Scenario.Name,
		payload.Run.Day,
		payload.SavedAt.Local().Format("2006-01-02 15:04"),
	)
	return meta
}

func slotNumberFromPath(path string) (int, bool) {
	name := strings.TrimSpace(path)
	if !strings.HasPrefix(name, "survive-it-save-") || !strings.HasSuffix(name, ".json") {
		return 0, false
	}
	raw := strings.TrimPrefix(name, "survive-it-save-")
	raw = strings.TrimSuffix(raw, ".json")
	slot, err := strconv.Atoi(raw)
	if err != nil || slot < 1 {
		return 0, false
	}
	return slot, true
}

func (m menuModel) viewLoadRun() string {
	if m.load.cursor < 0 || m.load.cursor > len(m.load.entries) {
		m.load.cursor = 0
	}

	totalWidth := m.w
	if totalWidth < 100 {
		totalWidth = 100
	}
	contentHeight := m.h - 8
	if contentHeight < 16 {
		contentHeight = 16
	}
	listWidth := totalWidth / 3
	if listWidth < 34 {
		listWidth = 34
	}
	detailWidth := totalWidth - listWidth - 4
	if detailWidth < 52 {
		detailWidth = 52
	}

	var list strings.Builder
	if len(m.load.entries) == 0 {
		list.WriteString(dimGreen.Render("  No save files found.") + "\n")
	}
	for i, entry := range m.load.entries {
		cursor := "  "
		style := green
		if i == m.load.cursor {
			cursor = "> "
			style = brightGreen
		}
		label := entry.Path
		if entry.Slot > 0 {
			label = fmt.Sprintf("Slot %d (%s)", entry.Slot, entry.Path)
		}
		list.WriteString(cursor + style.Render(label) + "\n")
	}
	cancelCursor := "  "
	cancelStyle := green
	if m.load.cursor == len(m.load.entries) {
		cancelCursor = "> "
		cancelStyle = brightGreen
	}
	list.WriteString("\n" + cancelCursor + cancelStyle.Render("Cancel") + "\n")

	detail := "Select a save to inspect details."
	if len(m.load.entries) > 0 && m.load.cursor < len(m.load.entries) {
		detail = m.loadDetailText(m.load.entries[m.load.cursor])
	}

	listPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(listWidth).
		Height(contentHeight).
		Render(list.String())
	detailPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(detailWidth).
		Height(contentHeight).
		Render(detail)

	var b strings.Builder
	b.WriteString(dosTitle("Load Run") + "\n")
	b.WriteString(dimGreen.Render("Left: save files. Right: run overview and scenario details.") + "\n\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane))
	b.WriteString("\n" + dimGreen.Render("↑/↓ move  Enter load  Shift+Q back") + "\n")
	if m.status != "" {
		b.WriteString(green.Render(m.status) + "\n")
	}
	return b.String()
}

func (m menuModel) loadDetailText(meta saveSlotMeta) string {
	if !meta.Exists {
		if meta.ErrDetail != "" {
			return brightGreen.Render(meta.Path) + "\n\n" + dimGreen.Render("Unreadable save: "+meta.ErrDetail)
		}
		return brightGreen.Render(meta.Path) + "\n\n" + dimGreen.Render("Save file missing.")
	}
	if meta.Run == nil {
		return brightGreen.Render(meta.Path) + "\n\n" + dimGreen.Render("No run metadata.")
	}
	run := meta.Run
	mode := modeLabel(run.Config.Mode)
	season, ok := run.CurrentSeason()
	seasonLabel := "unknown"
	if ok {
		seasonLabel = string(season)
	}

	var players strings.Builder
	for _, p := range run.Players {
		players.WriteString(fmt.Sprintf("- %s (%s, %s) E:%d H:%d M:%d\n",
			p.Name, p.Sex, p.BodyType, p.Energy, p.Hydration, p.Morale))
	}
	if len(run.Players) == 0 {
		players.WriteString("- none\n")
	}

	scDesc := strings.TrimSpace(run.Scenario.Description)
	if scDesc == "" {
		scDesc = "No description."
	}
	daunting := strings.TrimSpace(run.Scenario.Daunting)
	if daunting == "" {
		daunting = "No daunting note."
	}

	return strings.Join([]string{
		brightGreen.Render(run.Scenario.Name),
		green.Render(fmt.Sprintf("Saved: %s", meta.SavedAt.Local().Format("2006-01-02 15:04"))),
		green.Render(fmt.Sprintf("Mode: %s  |  Day: %d  |  Season: %s", mode, run.Day, seasonLabel)),
		green.Render(fmt.Sprintf("Days Run So Far: %d", run.Day)),
		"",
		brightGreen.Render("Players"),
		green.Render(players.String()),
		brightGreen.Render("Scenario"),
		dimGreen.Render(scDesc),
		"",
		brightGreen.Render("Daunting"),
		green.Render(daunting),
	}, "\n")
}

func (m menuModel) builderScenarioOptions() []builderScenarioOption {
	options := []builderScenarioOption{
		{
			label:     "New Scenario",
			isNew:     true,
			isCustom:  false,
			customIdx: -1,
			mode:      game.ModeAlone,
		},
	}

	for _, scenario := range game.BuiltInScenarios() {
		options = append(options, builderScenarioOption{
			label:     scenario.Name + " (Built-in)",
			scenario:  scenario,
			mode:      scenarioDefaultMode(scenario),
			isNew:     false,
			isCustom:  false,
			customIdx: -1,
		})
	}

	for i, record := range m.customScenarios {
		mode := record.PreferredMode
		if mode == "" {
			mode = scenarioDefaultMode(record.Scenario)
		}
		options = append(options, builderScenarioOption{
			label:     record.Scenario.Name + " (Custom)",
			scenario:  record.Scenario,
			mode:      mode,
			isNew:     false,
			isCustom:  true,
			customIdx: i,
		})
	}

	return options
}

func (m menuModel) openBuilderScenarioPicker() menuModel {
	options := m.builderScenarioOptions()
	if len(options) == 0 {
		m.status = "No scenarios available."
		return m
	}

	cursor := 0
	if m.build.sourceID != "" {
		for i, option := range options {
			if option.isNew {
				continue
			}
			if option.scenario.ID == m.build.sourceID && option.isCustom == m.build.sourceIsCustom {
				cursor = i
				break
			}
		}
	}

	m.bpick = newBuilderScenarioPickerState()
	m.bpick.cursor = cursor
	m.bpick.returnTo = screenScenarioBuilder
	m.screen = screenBuilderScenarioPicker
	return m
}

func (m menuModel) applyBuilderScenarioOption(option builderScenarioOption) menuModel {
	cursor := m.build.cursor
	playerCountIdx := m.build.playerCountIdx
	if playerCountIdx < 0 || playerCountIdx >= len(setupPlayerCounts()) {
		playerCountIdx = 0
	}

	loaded := newScenarioBuilderState()
	loaded.cursor = cursor
	loaded.playerCountIdx = playerCountIdx

	if option.isNew {
		m.build = loaded
		return m
	}

	loaded.selectedLabel = option.label
	loaded.sourceID = option.scenario.ID
	loaded.sourceName = option.scenario.Name
	loaded.sourceIsCustom = option.isCustom
	if option.isCustom {
		loaded.selectedIdx = option.customIdx + 1
	}

	loaded.name = option.scenario.Name
	loaded.modeIdx = selectedModeIndex(option.mode)
	loaded.biomeIdx = selectedBiomeIndex(option.scenario.Biome)
	if len(option.scenario.Wildlife) > 0 {
		loaded.wildlifeText = strings.Join(option.scenario.Wildlife, ", ")
	} else {
		loaded.wildlifeText = strings.Join(game.WildlifeForBiome(option.scenario.Biome), ", ")
	}

	if idx := indexOfInt(builderDefaultDays(), option.scenario.DefaultDays); idx >= 0 {
		loaded.useCustomDays = false
		loaded.defaultDaysIdx = idx
		loaded.customDays = fmt.Sprintf("%d", option.scenario.DefaultDays)
	} else {
		loaded.useCustomDays = true
		loaded.customDays = fmt.Sprintf("%d", option.scenario.DefaultDays)
	}

	loaded.seasonSetID = sanitizeSeasonSetID(string(option.scenario.DefaultSeasonSetID))
	if loaded.seasonSetID == "" {
		loaded.seasonSetID = "custom_profile"
	}

	set := game.SeasonSet{}
	found := false
	for _, seasonSet := range option.scenario.SeasonSets {
		if seasonSet.ID == option.scenario.DefaultSeasonSetID {
			set = seasonSet
			found = true
			break
		}
	}
	if !found && len(option.scenario.SeasonSets) > 0 {
		set = option.scenario.SeasonSets[0]
	}

	loaded.phases = make([]phaseBuilderPhase, 0, maxBuilderPhases)
	for _, phase := range set.Phases {
		if len(loaded.phases) >= maxBuilderPhases {
			break
		}
		loaded.phases = append(loaded.phases, phaseBuilderPhase{
			seasonIdx: selectedSeasonIndex(phase.Season),
			days:      fmt.Sprintf("%d", phase.Days),
		})
	}
	if len(loaded.phases) == 0 {
		loaded.phases = []phaseBuilderPhase{{seasonIdx: 0, days: "0"}}
	}

	m.build = loaded
	return m
}

func (m menuModel) updateScenarioBuilder(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	rows := m.scenarioBuilderRows()
	if len(rows) == 0 {
		return m, nil
	}
	if m.build.cursor < 0 || m.build.cursor >= len(rows) {
		m.build.cursor = len(rows) - 1
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "Q", "q", "esc":
		next, cmd := m.returnToMainMenu()
		return next, cmd
	case "up", "k":
		m.build.cursor = wrapIndex(m.build.cursor, -1, len(rows))
		return m, nil
	case "down", "j":
		m.build.cursor = wrapIndex(m.build.cursor, 1, len(rows))
		return m, nil
	case "left":
		m = m.adjustScenarioBuilderChoice(-1)
		return m, nil
	case "right":
		m = m.adjustScenarioBuilderChoice(1)
		return m, nil
	case "backspace", "ctrl+h":
		m = m.backspaceScenarioBuilderText()
		return m, nil
	case "enter":
		row := rows[m.build.cursor]
		if !row.active {
			return m, nil
		}

		switch row.kind {
		case builderRowScenario:
			m = m.openBuilderScenarioPicker()
			return m, nil
		case builderRowSave:
			return m.saveScenarioFromBuilder()
		case builderRowDelete:
			return m.deleteScenarioFromBuilder()
		case builderRowEditPhases:
			m.phase = newPhaseEditorState()
			m.screen = screenPhaseEditor
			return m, nil
		case builderRowPlayerEditor:
			m.setup.playerCountIdx = m.build.playerCountIdx
			m = m.ensureSetupPlayers()
			m.pcfg.returnTo = screenScenarioBuilder
			m.pcfg.playerIdx = 0
			m.status = "Player editor: you can set your own kit and every player kit."
			m.screen = screenPlayerConfig
			return m, nil
		case builderRowCancel:
			next, cmd := m.returnToMainMenu()
			return next, cmd
		default:
			if m.scenarioBuilderRowSupportsCycle(row) {
				m = m.adjustScenarioBuilderChoice(1)
			}
			return m, nil
		}
	}

	if len(msg.Runes) > 0 {
		m = m.appendScenarioBuilderText(msg.Runes)
		return m, nil
	}

	return m, nil
}

func (m menuModel) adjustScenarioBuilderChoice(delta int) menuModel {
	rows := m.scenarioBuilderRows()
	if len(rows) == 0 {
		return m
	}
	if m.build.cursor < 0 || m.build.cursor >= len(rows) {
		return m
	}

	row := rows[m.build.cursor]
	if !row.active {
		return m
	}

	switch row.kind {
	case builderRowMode:
		m.build.modeIdx = wrapIndex(m.build.modeIdx, delta, len(setupModes()))
	case builderRowPlayerCount:
		m.build.playerCountIdx = wrapIndex(m.build.playerCountIdx, delta, len(setupPlayerCounts()))
		m.setup.playerCountIdx = m.build.playerCountIdx
		m = m.ensureSetupPlayers()
	case builderRowBiome:
		m.build.biomeIdx = wrapIndex(m.build.biomeIdx, delta, len(builderBiomes()))
		m.build.wildlifeText = strings.Join(game.WildlifeForBiome(builderBiomes()[m.build.biomeIdx]), ", ")
	case builderRowWildlife:
		m.build.wildlifeText = strings.Join(game.WildlifeForBiome(builderBiomes()[m.build.biomeIdx]), ", ")
	case builderRowDaysMode:
		m.build.useCustomDays = !m.build.useCustomDays
	case builderRowDaysPreset:
		m.build.defaultDaysIdx = wrapIndex(m.build.defaultDaysIdx, delta, len(builderDefaultDays()))
	}

	return m
}

func (m menuModel) saveScenarioFromBuilder() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.build.name)
	if name == "" {
		m.status = "Scenario name is required."
		return m, nil
	}

	defaultDays := builderDefaultDays()[m.build.defaultDaysIdx]
	if m.build.useCustomDays {
		v, err := parseNonNegativeInt(m.build.customDays)
		if err != nil || v < 1 {
			m.status = "Custom default days must be a number >= 1."
			return m, nil
		}
		defaultDays = v
	}

	phaseCount := len(m.build.phases)
	if phaseCount < 1 {
		m.status = "At least one season phase is required."
		return m, nil
	}
	seasonOptions := builderSeasonOptions()
	phases := make([]game.SeasonPhase, 0, phaseCount)
	for i := 0; i < phaseCount; i++ {
		days, err := parseNonNegativeInt(m.build.phases[i].days)
		if err != nil {
			m.status = fmt.Sprintf("Phase %d days must be a number >= 0.", i+1)
			return m, nil
		}
		if i < phaseCount-1 && days == 0 {
			m.status = fmt.Sprintf("Phase %d cannot be 0 unless it is the last phase.", i+1)
			return m, nil
		}
		phases = append(phases, game.SeasonPhase{
			Season: seasonOptions[m.build.phases[i].seasonIdx],
			Days:   days,
		})
	}

	seasonSetID := game.SeasonSetID(sanitizeSeasonSetID(m.build.seasonSetID))
	if seasonSetID == "" {
		seasonSetID = game.SeasonSetID("custom_profile")
	}

	editingBuiltIn := m.build.sourceID != "" && !m.build.sourceIsCustom
	if editingBuiltIn && strings.EqualFold(strings.TrimSpace(name), strings.TrimSpace(m.build.sourceName)) {
		m.status = "When editing a built-in scenario, save with a new scenario name."
		return m, nil
	}

	editingCustom := m.build.sourceIsCustom && m.build.sourceID != ""
	scenarioID := game.ScenarioID("")
	if editingCustom {
		scenarioID = m.build.sourceID
	} else {
		scenarioID = makeCustomScenarioID(name, m.availableScenarios())
	}
	biome := builderBiomes()[m.build.biomeIdx]
	wildlife := normalizedCSVList(m.build.wildlifeText)
	if len(wildlife) == 0 {
		wildlife = game.WildlifeForBiome(biome)
	}

	scenario := game.Scenario{
		ID:          scenarioID,
		Name:        name,
		Biome:       biome,
		Wildlife:    wildlife,
		Description: "Custom scenario crafted in the scenario editor.",
		Daunting:    "Custom conditions can become severe without careful planning.",
		Motivation:  "Adapt, persist, and complete your own survival blueprint.",
		SupportedModes: []game.GameMode{
			setupModes()[m.build.modeIdx],
		},
		DefaultDays: defaultDays,
		IssuedKit:   game.IssuedKit{},
		SeasonSets: []game.SeasonSet{
			{
				ID:     seasonSetID,
				Phases: phases,
			},
		},
		DefaultSeasonSetID: seasonSetID,
	}

	record := customScenarioRecord{
		Scenario:      scenario,
		PreferredMode: setupModes()[m.build.modeIdx],
	}

	updated := append([]customScenarioRecord{}, m.customScenarios...)
	savedCustomIdx := -1
	if editingCustom {
		for i := range updated {
			if updated[i].Scenario.ID == scenarioID {
				updated[i] = record
				savedCustomIdx = i
				break
			}
		}
		if savedCustomIdx == -1 {
			updated = append(updated, record)
			savedCustomIdx = len(updated) - 1
		}
	} else {
		updated = append(updated, record)
		savedCustomIdx = len(updated) - 1
	}

	if err := saveCustomScenarios(defaultCustomScenariosFile, updated); err != nil {
		m.status = fmt.Sprintf("Failed to save scenario: %v", err)
		return m, nil
	}

	m.customScenarios = updated
	m.status = fmt.Sprintf("Saved scenario: %s", scenario.Name)
	m = m.loadScenarioBuilderSelection(savedCustomIdx + 1)
	return m, nil
}

func (m menuModel) deleteScenarioFromBuilder() (tea.Model, tea.Cmd) {
	if !m.build.sourceIsCustom || m.build.sourceID == "" {
		m.status = "Only custom scenarios can be deleted."
		return m, nil
	}

	idx := -1
	for i, record := range m.customScenarios {
		if record.Scenario.ID == m.build.sourceID {
			idx = i
			break
		}
	}
	if idx == -1 {
		m.status = "Selected custom scenario could not be found."
		return m, nil
	}
	updated := append([]customScenarioRecord{}, m.customScenarios[:idx]...)
	updated = append(updated, m.customScenarios[idx+1:]...)

	if err := saveCustomScenarios(defaultCustomScenariosFile, updated); err != nil {
		m.status = fmt.Sprintf("Failed to delete scenario: %v", err)
		return m, nil
	}

	m.customScenarios = updated
	m.build = newScenarioBuilderState()
	m.status = "Scenario deleted."
	return m, nil
}

func (m menuModel) scenarioBuilderRowSupportsCycle(row scenarioBuilderRow) bool {
	switch row.kind {
	case builderRowMode, builderRowPlayerCount, builderRowBiome, builderRowWildlife, builderRowDaysMode, builderRowDaysPreset:
		return true
	default:
		return false
	}
}

func (m menuModel) scenarioBuilderRows() []scenarioBuilderRow {
	selectedLabel := strings.TrimSpace(m.build.selectedLabel)
	if selectedLabel == "" {
		selectedLabel = "New Scenario"
	}
	if m.build.playerCountIdx < 0 || m.build.playerCountIdx >= len(setupPlayerCounts()) {
		m.build.playerCountIdx = 0
	}

	rows := []scenarioBuilderRow{
		{label: "Scenario", value: selectedLabel, kind: builderRowScenario, active: true},
		{label: "Name", value: m.build.name, kind: builderRowName, active: true},
		{label: "Game Mode", value: modeLabel(setupModes()[m.build.modeIdx]), kind: builderRowMode, active: true},
		{label: "Player Count", value: fmt.Sprintf("%d", setupPlayerCounts()[m.build.playerCountIdx]), kind: builderRowPlayerCount, active: true},
		{label: "Biome", value: builderBiomes()[m.build.biomeIdx], kind: builderRowBiome, active: true},
		{label: "Wildlife", value: m.build.wildlifeText, kind: builderRowWildlife, active: true},
		{label: "Days Mode", value: map[bool]string{true: "Custom", false: "Preset"}[m.build.useCustomDays], kind: builderRowDaysMode, active: true},
		{label: "Days Preset", value: fmt.Sprintf("%d", builderDefaultDays()[m.build.defaultDaysIdx]), kind: builderRowDaysPreset, active: !m.build.useCustomDays},
		{label: "Days Custom", value: m.build.customDays, kind: builderRowDaysCustom, active: m.build.useCustomDays},
		{label: "Season Profile ID", value: m.build.seasonSetID, kind: builderRowSeasonProfileID, active: true},
		{label: "Phase Builder", value: fmt.Sprintf("%d phase(s)", len(m.build.phases)), kind: builderRowEditPhases, active: true},
		{label: "Player Editor", value: "Open player configuration", kind: builderRowPlayerEditor, active: true},
		{label: "Save Scenario", value: "", kind: builderRowSave, active: true},
		{label: "Delete Scenario", value: "", kind: builderRowDelete, active: m.build.sourceIsCustom},
		{label: "Cancel", value: "", kind: builderRowCancel, active: true},
	}

	return rows
}

func (m menuModel) appendScenarioBuilderText(runes []rune) menuModel {
	rows := m.scenarioBuilderRows()
	if len(rows) == 0 || m.build.cursor < 0 || m.build.cursor >= len(rows) {
		return m
	}

	row := rows[m.build.cursor]
	if !row.active {
		return m
	}

	text := string(runes)
	switch row.kind {
	case builderRowName:
		m.build.name += text
	case builderRowWildlife:
		for _, r := range runes {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' || r == ',' || r == '-' {
				m.build.wildlifeText += string(r)
			}
		}
	case builderRowDaysCustom:
		for _, r := range runes {
			if r >= '0' && r <= '9' {
				m.build.customDays += string(r)
			}
		}
	case builderRowSeasonProfileID:
		for _, r := range runes {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
				m.build.seasonSetID += string(r)
			}
		}
	}

	return m
}

func (m menuModel) backspaceScenarioBuilderText() menuModel {
	rows := m.scenarioBuilderRows()
	if len(rows) == 0 || m.build.cursor < 0 || m.build.cursor >= len(rows) {
		return m
	}
	row := rows[m.build.cursor]

	backspace := func(s string) string {
		if len(s) == 0 {
			return s
		}
		return s[:len(s)-1]
	}

	switch row.kind {
	case builderRowName:
		m.build.name = backspace(m.build.name)
	case builderRowWildlife:
		m.build.wildlifeText = backspace(m.build.wildlifeText)
	case builderRowDaysCustom:
		m.build.customDays = backspace(m.build.customDays)
	case builderRowSeasonProfileID:
		m.build.seasonSetID = backspace(m.build.seasonSetID)
	}

	return m
}

func (m menuModel) loadScenarioBuilderSelection(selected int) menuModel {
	if selected == 0 {
		return m.applyBuilderScenarioOption(builderScenarioOption{
			label:     "New Scenario",
			isNew:     true,
			isCustom:  false,
			customIdx: -1,
			mode:      game.ModeAlone,
		})
	}
	idx := selected - 1
	if idx < 0 || idx >= len(m.customScenarios) {
		return m.applyBuilderScenarioOption(builderScenarioOption{
			label:     "New Scenario",
			isNew:     true,
			isCustom:  false,
			customIdx: -1,
			mode:      game.ModeAlone,
		})
	}
	record := m.customScenarios[idx]
	mode := record.PreferredMode
	if mode == "" {
		mode = scenarioDefaultMode(record.Scenario)
	}
	return m.applyBuilderScenarioOption(builderScenarioOption{
		label:     record.Scenario.Name + " (Custom)",
		scenario:  record.Scenario,
		mode:      mode,
		isNew:     false,
		isCustom:  true,
		customIdx: idx,
	})
}

func (m menuModel) viewScenarioBuilder() string {
	rows := m.scenarioBuilderRows()
	if len(rows) == 0 {
		return brightGreen.Render("No scenario builder rows available.")
	}
	if m.build.cursor < 0 || m.build.cursor >= len(rows) {
		m.build.cursor = 0
	}

	totalWidth := m.w
	if totalWidth < 100 {
		totalWidth = 100
	}
	contentHeight := m.h - 8
	if contentHeight < 16 {
		contentHeight = 16
	}
	listWidth := totalWidth / 3
	if listWidth < 34 {
		listWidth = 34
	}
	detailWidth := totalWidth - listWidth - 4
	if detailWidth < 52 {
		detailWidth = 52
	}

	var list strings.Builder
	for i, row := range rows {
		cursor := "  "
		lineStyle := green
		if i == m.build.cursor {
			cursor = "> "
			lineStyle = brightGreen
		}
		if !row.active {
			lineStyle = dimGreen
		}
		if row.value == "" {
			list.WriteString(cursor + lineStyle.Render(row.label) + "\n")
			continue
		}

		value := row.value
		if row.kind == builderRowName && strings.TrimSpace(value) == "" {
			value = "<type name>"
		}
		if row.kind == builderRowWildlife && strings.TrimSpace(value) == "" {
			value = "<comma separated animals>"
		}
		if row.kind == builderRowDaysCustom && strings.TrimSpace(value) == "" {
			value = "<type days>"
		}
		if row.kind == builderRowSeasonProfileID && strings.TrimSpace(value) == "" {
			value = "<profile_id>"
		}
		list.WriteString(cursor + lineStyle.Render(fmt.Sprintf("%-18s %s", row.label+":", value)) + "\n")
	}

	detail := m.scenarioBuilderDetailText(rows[m.build.cursor])
	listPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(listWidth).
		Height(contentHeight).
		Render(list.String())
	detailPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(detailWidth).
		Height(contentHeight).
		Render(detail)

	var b strings.Builder
	b.WriteString(dosTitle("Scenario Builder / Editor") + "\n")
	b.WriteString(dimGreen.Render("Left: editable fields. Right: context and preview.") + "\n\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane))
	b.WriteString("\n" + dimGreen.Render("↑/↓ move  ←/→ change  Enter open/select  type in text fields  Shift+Q back") + "\n")
	if m.status != "" {
		b.WriteString(green.Render(m.status) + "\n")
	}

	return b.String()
}

func (m menuModel) scenarioBuilderDetailText(row scenarioBuilderRow) string {
	phasePreview := make([]string, 0, len(m.build.phases))
	seasonOptions := builderSeasonOptions()
	for i, phase := range m.build.phases {
		season := seasonOptions[clampInt(phase.seasonIdx, 0, len(seasonOptions)-1)]
		phasePreview = append(phasePreview, fmt.Sprintf("%d. %s (%s day(s))", i+1, builderSeasonLabel(season), phase.days))
	}
	if len(phasePreview) == 0 {
		phasePreview = append(phasePreview, "No phases configured.")
	}

	notes := []string{
		brightGreen.Render("Field Notes"),
		dimGreen.Render("Use this screen for core scenario properties."),
		dimGreen.Render("Use Phase Builder for seasonal phase timeline."),
		dimGreen.Render("Use Player Editor to configure participants."),
	}

	switch row.kind {
	case builderRowEditPhases:
		notes = append(notes,
			"",
			brightGreen.Render("Phase Builder"),
			green.Render("Open a dedicated two-column phase editor."),
		)
	case builderRowPlayerEditor:
		notes = append(notes,
			"",
			brightGreen.Render("Player Editor"),
			green.Render("Configure player names, stats, and kit selections."),
		)
	case builderRowSave:
		notes = append(notes,
			"",
			brightGreen.Render("Save"),
			green.Render("Persists scenario into survive-it-scenarios.json."),
		)
	}
	wildlifePreview := normalizedCSVList(m.build.wildlifeText)
	if len(wildlifePreview) == 0 {
		wildlifePreview = game.WildlifeForBiome(builderBiomes()[m.build.biomeIdx])
	}

	return strings.Join([]string{
		brightGreen.Render("Current Scenario Draft"),
		green.Render(fmt.Sprintf("Name: %s", strings.TrimSpace(m.build.name))),
		green.Render(fmt.Sprintf("Mode: %s", modeLabel(setupModes()[m.build.modeIdx]))),
		green.Render(fmt.Sprintf("Player Count: %d", setupPlayerCounts()[m.build.playerCountIdx])),
		green.Render(fmt.Sprintf("Biome: %s", builderBiomes()[m.build.biomeIdx])),
		green.Render(fmt.Sprintf("Temperature Range: %s", m.formatTemperatureRange(game.TemperatureRangeForBiome(builderBiomes()[m.build.biomeIdx])))),
		green.Render(fmt.Sprintf("Wildlife: %s", strings.Join(wildlifePreview, ", "))),
		green.Render(fmt.Sprintf("Insects (Auto): %s", strings.Join(game.InsectsForBiome(builderBiomes()[m.build.biomeIdx]), ", "))),
		green.Render(fmt.Sprintf("Default Days: %s", map[bool]string{true: m.build.customDays, false: fmt.Sprintf("%d", builderDefaultDays()[m.build.defaultDaysIdx])}[m.build.useCustomDays])),
		green.Render(fmt.Sprintf("Season Profile ID: %s", strings.TrimSpace(m.build.seasonSetID))),
		"",
		brightGreen.Render("Phase Timeline"),
		green.Render(strings.Join(phasePreview, "\n")),
		"",
		strings.Join(notes, "\n"),
		"",
		brightGreen.Render("Run Setup"),
		dimGreen.Render("Preset or custom scenarios both support full player/kit editing."),
	}, "\n")
}

func (m menuModel) updatePhaseEditor(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	rows := m.phaseEditorRows()
	if len(rows) == 0 {
		m.screen = screenScenarioBuilder
		return m, nil
	}
	if m.phase.cursor < 0 || m.phase.cursor >= len(rows) {
		m.phase.cursor = 0
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "Q", "q", "esc":
		m.screen = screenScenarioBuilder
		return m, nil
	case "up", "k":
		m.phase.cursor = wrapIndex(m.phase.cursor, -1, len(rows))
		return m, nil
	case "down", "j":
		m.phase.cursor = wrapIndex(m.phase.cursor, 1, len(rows))
		return m, nil
	case "left":
		row := rows[m.phase.cursor]
		if row.kind == phaseRowNewSeason {
			m.phase.newSeasonIdx = wrapIndex(m.phase.newSeasonIdx, -1, len(builderSeasonOptions()))
		}
		return m, nil
	case "right":
		row := rows[m.phase.cursor]
		if row.kind == phaseRowNewSeason {
			m.phase.newSeasonIdx = wrapIndex(m.phase.newSeasonIdx, 1, len(builderSeasonOptions()))
		}
		return m, nil
	case "backspace", "ctrl+h":
		row := rows[m.phase.cursor]
		if row.kind == phaseRowNewDays && len(m.phase.newDays) > 0 {
			m.phase.newDays = m.phase.newDays[:len(m.phase.newDays)-1]
		}
		return m, nil
	case "enter":
		row := rows[m.phase.cursor]
		switch row.kind {
		case phaseRowNewSeason:
			m.phase.newSeasonIdx = wrapIndex(m.phase.newSeasonIdx, 1, len(builderSeasonOptions()))
			return m, nil
		case phaseRowAddPhase:
			if !m.phase.adding {
				m.phase.adding = true
				if strings.TrimSpace(m.phase.newDays) == "" {
					m.phase.newDays = "7"
				}
				m.phase.cursor = 0
				return m, nil
			}
			if len(m.build.phases) >= maxBuilderPhases {
				m.status = fmt.Sprintf("Maximum phases reached (%d).", maxBuilderPhases)
				return m, nil
			}
			days, err := parseNonNegativeInt(strings.TrimSpace(m.phase.newDays))
			if err != nil {
				m.status = "Days must be a number >= 0."
				return m, nil
			}
			if len(m.build.phases) > 0 && days == 0 {
				m.status = "0 days is only valid for the last phase."
				return m, nil
			}
			m.build.phases = append(m.build.phases, phaseBuilderPhase{
				seasonIdx: m.phase.newSeasonIdx,
				days:      fmt.Sprintf("%d", days),
			})
			m.phase.adding = false
			m.phase.newDays = "7"
			m.phase.cursor = 0
			m.status = "Added phase."
			return m, nil
		case phaseRowRemoveLast:
			if len(m.build.phases) <= 1 {
				m.status = "At least one phase is required."
				return m, nil
			}
			m.build.phases = m.build.phases[:len(m.build.phases)-1]
			m.status = "Removed last phase."
			return m, nil
		case phaseRowBack:
			m.screen = screenScenarioBuilder
			return m, nil
		}
	}

	if len(msg.Runes) > 0 {
		row := rows[m.phase.cursor]
		if row.kind == phaseRowNewDays {
			for _, r := range msg.Runes {
				if r >= '0' && r <= '9' {
					m.phase.newDays += string(r)
				}
			}
		}
		return m, nil
	}

	return m, nil
}

func (m menuModel) viewPhaseEditor() string {
	rows := m.phaseEditorRows()
	if len(rows) == 0 {
		return brightGreen.Render("No phase editor options available.")
	}
	if m.phase.cursor < 0 || m.phase.cursor >= len(rows) {
		m.phase.cursor = 0
	}
	totalWidth := m.w
	if totalWidth < 100 {
		totalWidth = 100
	}
	contentHeight := m.h - 8
	if contentHeight < 16 {
		contentHeight = 16
	}
	listWidth := totalWidth / 3
	if listWidth < 34 {
		listWidth = 34
	}
	detailWidth := totalWidth - listWidth - 4
	if detailWidth < 52 {
		detailWidth = 52
	}

	var list strings.Builder
	for i, row := range rows {
		cursor := "  "
		style := green
		if i == m.phase.cursor {
			cursor = "> "
			style = brightGreen
		}
		if !row.active {
			style = dimGreen
		}
		if row.value == "" {
			list.WriteString(cursor + style.Render(row.label) + "\n")
			continue
		}
		value := row.value
		if row.kind == phaseRowNewDays && strings.TrimSpace(value) == "" {
			value = "<type days>"
		}
		list.WriteString(cursor + style.Render(fmt.Sprintf("%-18s %s", row.label+":", value)) + "\n")
	}

	detail := m.phaseEditorTimelineText()
	listPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(listWidth).
		Height(contentHeight).
		Render(list.String())
	detailPane := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(detailWidth).
		Height(contentHeight).
		Render(detail)

	var b strings.Builder
	b.WriteString(dosTitle("Phase Builder") + "\n")
	b.WriteString(dimGreen.Render("Left: phase actions and add-phase fields. Right: phase timeline.") + "\n\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane))
	b.WriteString("\n" + dimGreen.Render("↑/↓ move  ←/→ season  type days  Enter select  Shift+Q back") + "\n")
	if m.status != "" {
		b.WriteString(green.Render(m.status) + "\n")
	}
	return b.String()
}

func (m menuModel) phaseEditorRows() []phaseEditorRow {
	seasonOptions := builderSeasonOptions()
	if len(seasonOptions) == 0 {
		return nil
	}
	if m.phase.newSeasonIdx < 0 || m.phase.newSeasonIdx >= len(seasonOptions) {
		m.phase.newSeasonIdx = 0
	}

	rows := make([]phaseEditorRow, 0, 5)
	if m.phase.adding {
		rows = append(rows,
			phaseEditorRow{
				label:  "New Phase Season",
				value:  builderSeasonLabel(seasonOptions[m.phase.newSeasonIdx]),
				kind:   phaseRowNewSeason,
				active: true,
			},
			phaseEditorRow{
				label:  "New Phase Days",
				value:  m.phase.newDays,
				kind:   phaseRowNewDays,
				active: true,
			},
		)
	}

	addLabel := "Add Phase"
	if m.phase.adding {
		addLabel = "Confirm Add Phase"
	}
	rows = append(rows,
		phaseEditorRow{label: addLabel, value: "", kind: phaseRowAddPhase, active: true},
		phaseEditorRow{label: "Remove Last Phase", value: "", kind: phaseRowRemoveLast, active: len(m.build.phases) > 1},
		phaseEditorRow{label: "Back", value: "", kind: phaseRowBack, active: true},
	)

	return rows
}

func (m menuModel) phaseEditorTimelineText() string {
	seasonOptions := builderSeasonOptions()
	lines := []string{
		brightGreen.Render("Phase Timeline"),
	}

	if len(m.build.phases) == 0 {
		lines = append(lines, dimGreen.Render("No phases configured."))
	} else {
		for i, phase := range m.build.phases {
			season := seasonOptions[clampInt(phase.seasonIdx, 0, len(seasonOptions)-1)]
			line := fmt.Sprintf("%d. %-8s  %s day(s)", i+1, builderSeasonLabel(season), phase.days)
			lines = append(lines, green.Render(line))
		}
	}

	lines = append(lines,
		"",
		brightGreen.Render("Notes"),
		dimGreen.Render("When adding a phase, season and days appear at top-left."),
		dimGreen.Render("Set days to 0 only for the final phase."),
	)

	return strings.Join(lines, "\n")
}

func (m menuModel) viewRun() string {
	totalHeight := m.h
	if totalHeight < 20 {
		totalHeight = 20
	}

	totalWidth := m.w
	if totalWidth < 60 {
		totalWidth = 60
	}

	headerRows := 5
	controlsRows := 3
	inputRows := 4
	bodyRows := totalHeight - headerRows - controlsRows - inputRows
	if bodyRows < 8 {
		bodyRows = 8
	}

	// Lipgloss height applies to the content area; borders add 2 rows.
	contentHeight := func(totalRows int) int {
		h := totalRows - 2
		if h < 1 {
			return 1
		}
		return h
	}

	paneStyle := lipgloss.NewStyle().
		Border(dosBox).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(totalWidth - 4)

	header := paneStyle.Copy().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Height(contentHeight(headerRows)).
		Render(m.headerText())
	body := paneStyle.Copy().
		Foreground(lipgloss.Color("10")).
		Height(contentHeight(bodyRows)).
		Render(m.bodyText())
	inputPanel := paneStyle.Copy().
		Foreground(lipgloss.Color("#FAFAFA")).
		Height(contentHeight(inputRows)).
		Render(m.footerText())

	return lipgloss.JoinVertical(lipgloss.Left, header, body, m.controlsLine(totalWidth), inputPanel)
}

func (m menuModel) viewMenu() string {
	items := menuItems(m)
	if len(items) == 0 {
		return brightGreen.Render("No menu items available.")
	}
	if m.idx < 0 {
		m.idx = 0
	}
	if m.idx >= len(items) {
		m.idx = len(items) - 1
	}

	contentWidth := 42
	if m.w > 0 {
		maxAllowed := m.w - 8
		if maxAllowed < contentWidth {
			contentWidth = maxAllowed
		}
		if contentWidth < 30 {
			contentWidth = 30
		}
	}

	buttonStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Padding(0, 1).
		Border(dosBox).
		Align(lipgloss.Center)

	var buttons strings.Builder
	for i, it := range items {
		label := strings.ToUpper(it.label)
		style := buttonStyle.Copy().BorderForeground(lipgloss.Color("2")).Foreground(lipgloss.Color("2"))
		if i == m.idx {
			style = style.BorderForeground(lipgloss.Color("10")).Foreground(lipgloss.Color("10")).Bold(true)
		}
		buttons.WriteString(style.Render(label) + "\n")
	}

	var header strings.Builder
	header.WriteString(dosTitle("Survive It") + "\n")
	header.WriteString(dimGreen.Render("DOS Survival Terminal") + "\n")
	header.WriteString(dimGreen.Render(fmt.Sprintf("v%s  (%s)  %s", m.cfg.Version, m.cfg.Commit, m.cfg.BuildDate)) + "\n")
	if m.busy {
		busyLine := strings.TrimSpace(m.status)
		if busyLine == "" {
			busyLine = "Checking for updates..."
		}
		header.WriteString(green.Render(busyLine) + "\n")
	} else if m.updateAvailable {
		header.WriteString(brightGreen.Render(m.updateStatus) + "\n")
		header.WriteString(dimGreen.Render("Select INSTALL UPDATE from the menu.") + "\n")
	}

	main := lipgloss.JoinVertical(
		lipgloss.Center,
		header.String(),
		border.Render(dosRule(contentWidth)),
		"",
		buttons.String(),
		border.Render(dosRule(contentWidth)),
		dimGreen.Render("↑/↓ move  Enter select  Shift+Q quit"),
	)

	return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, main)
}

func checkUpdateCmd(currentVersion string, auto bool) tea.Cmd {
	return func() tea.Msg {
		// Tiny delay so the UI visibly switches to a busy state.
		time.Sleep(150 * time.Millisecond)

		res, err := update.Check(update.CheckParams{
			CurrentVersion: currentVersion,
		})
		if err != nil {
			return updateResultMsg{err: err, auto: auto}
		}
		return updateResultMsg{
			status:    res,
			available: strings.HasPrefix(res, "Update available:"),
			auto:      auto,
		}
	}
}

func clockTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return clockTickMsg{at: t}
	})
}

func applyUpdateCmd(currentVersion string) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(150 * time.Millisecond)

		res, err := update.Apply(currentVersion)
		if err != nil {
			return updateResultMsg{err: err}
		}
		return updateResultMsg{status: res}
	}
}

func convertToFahrenheit(c int) int {
	return int(float64(c)*9.0/5.0 + 32.0)
}

func (m menuModel) formatTemperature(celsius int) string {
	if m.opts.tempUnit == tempUnitFahrenheit {
		return fmt.Sprintf("%dF", convertToFahrenheit(celsius))
	}
	return fmt.Sprintf("%dC", celsius)
}

func (m menuModel) formatTemperatureRange(r game.TemperatureRange) string {
	if m.opts.tempUnit == tempUnitFahrenheit {
		return fmt.Sprintf("%dF to %dF", convertToFahrenheit(r.MinC), convertToFahrenheit(r.MaxC))
	}
	return fmt.Sprintf("%dC to %dC", r.MinC, r.MaxC)
}

func (m menuModel) currentRunWeather() game.WeatherState {
	if m.run == nil {
		return game.WeatherState{}
	}
	m.run.EnsureWeather()
	return m.run.Weather
}

func (m menuModel) currentRunTemperatureC() int {
	return m.currentRunWeather().TemperatureC
}

func (m menuModel) headerText() string {
	season, ok := m.run.CurrentSeason()
	seasonStr := "unknown"
	if ok {
		seasonStr = string(season)
	}
	weather := m.currentRunWeather()
	weatherLabel := game.WeatherLabel(weather.Type)
	if weather.StreakDays > 1 {
		weatherLabel = fmt.Sprintf("%s x%d", weatherLabel, weather.StreakDays)
	}
	temp := m.formatTemperature(m.currentRunTemperatureC())
	gameTime := m.gameTimeLabel()
	nextDayIn := m.nextDayCountdownLabel()

	var b strings.Builder
	b.WriteString(lipgloss.JoinVertical(
		lipgloss.Left,
		brightGreen.Render("SURVIVE IT!"),
		dimGreen.Render(fmt.Sprintf("Mode: %s  |  Scenario: %s  |  Season: %s  |  Day: %d  |  Weather: %s  |  Temp: %s",
			modeLabel(m.run.Config.Mode), m.run.Scenario.Name, seasonStr, m.run.Day, weatherLabel, temp)),
		dimGreen.Render(fmt.Sprintf("Game Time: %s  |  Next Day In: %s", gameTime, nextDayIn)),
	))
	return b.String()
}

func (m menuModel) autoDayDuration() time.Duration {
	if m.opts.dayHours < 1 {
		return 2 * time.Hour
	}
	return time.Duration(m.opts.dayHours) * time.Hour
}

func (m menuModel) gameTimeLabel() string {
	if m.run == nil {
		return "Day 0 00:00:00"
	}
	return fmt.Sprintf("Day %d %s", m.run.Day, formatClockDuration(m.runPlayedFor))
}

func (m menuModel) nextDayCountdownLabel() string {
	remaining := m.autoDayDuration() - m.runPlayedFor
	if remaining < 0 {
		remaining = 0
	}
	return formatClockDuration(remaining)
}

func formatClockDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSeconds := int(d / time.Second)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
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

func (m menuModel) bodyText() string {
	if len(m.run.Players) == 0 {
		return dimGreen.Render("No players loaded.")
	}

	rows := make([][]string, 0, len(m.run.Players))
	for _, p := range m.run.Players {
		rows = append(rows, []string{
			p.Name,
			string(p.Sex),
			string(p.BodyType),
			fmt.Sprintf("%d", p.Energy),
			fmt.Sprintf("%d", p.Hydration),
			fmt.Sprintf("%d", p.Morale),
			fmt.Sprintf("%d", len(p.Ailments)),
		})
	}

	t := table.New().
		Border(dosBox).
		BorderStyle(border).
		BorderHeader(true).
		BorderRow(false).
		Headers("Player", "Sex", "Body", "Energy", "Hydration", "Morale", "Ailments").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return brightGreen.Bold(true)
			}
			return green
		})

	return t.Render()
}

func (m menuModel) controlsLine(totalWidth int) string {
	text := fmt.Sprintf(" Shift+N Next Day  |  Shift+S Save Slot %d  |  Shift+L Load  |  Auto Day: %dh  |  Cmd: hunt land|fish|air [raw] [liver] [p#] [grams]  use <item> <action> [p#]  actions [p#]  |  Shift+Q Menu ", m.activeSaveSlot, m.opts.dayHours)
	maxWidth := totalWidth - 2
	if maxWidth < 20 {
		maxWidth = 20
	}
	return border.Width(maxWidth).Render(strings.Repeat("-", maxWidth)) +
		"\n" + border.Width(maxWidth).Render(text) +
		"\n" + border.Width(maxWidth).Render(strings.Repeat("-", maxWidth))
}

func (m menuModel) footerText() string {
	var b strings.Builder

	out := m.run.EvaluateRun()
	if out.Status != game.RunOutcomeOngoing {
		b.WriteString(brightGreen.Render(string(out.Status)))
		b.WriteString("  ")
	}

	if m.status != "" {
		b.WriteString(brightGreen.Render(m.status))
	}

	if out.Status != game.RunOutcomeOngoing || m.status != "" {
		b.WriteString("\n")
	}

	if strings.TrimSpace(m.runInput) == "" {
		b.WriteString(dimGreen.Render("> "))
	} else {
		b.WriteString(brightGreen.Render("> " + m.runInput))
	}
	return b.String()
}

func saveRunToFile(path string, run game.RunState) error {
	if err := validateDataFilePath(path); err != nil {
		return err
	}

	payload := savedRun{
		FormatVersion: 1,
		SavedAt:       time.Now().UTC(),
		Run:           run,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	if len(data) > maxSaveFileBytes {
		return fmt.Errorf("save data exceeds %d bytes", maxSaveFileBytes)
	}

	return os.WriteFile(path, data, 0o600)
}

func savePathForSlot(slot int) string {
	return fmt.Sprintf("survive-it-save-%d.json", slot)
}

func validateDataFilePath(path string) error {
	clean := filepath.Clean(path)
	if filepath.IsAbs(clean) || strings.Contains(clean, string(filepath.Separator)) {
		return fmt.Errorf("invalid data file path: %s", path)
	}

	if clean == defaultCustomScenariosFile {
		return nil
	}

	if saveFilePattern.MatchString(clean) {
		return nil
	}

	return fmt.Errorf("unsupported data file path: %s", path)
}

func readDataFile(path string, maxBytes int) ([]byte, error) {
	if err := validateDataFilePath(path); err != nil {
		return nil, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > int64(maxBytes) {
		return nil, fmt.Errorf("data file exceeds %d bytes: %s", maxBytes, path)
	}

	// #nosec G304 -- path is restricted to a fixed allow-list by validateDataFilePath.
	return os.ReadFile(path)
}

func loadRunFromFile(path string, availableScenarios []game.Scenario) (game.RunState, error) {
	data, err := readDataFile(path, maxSaveFileBytes)
	if err != nil {
		return game.RunState{}, err
	}

	var payload savedRun
	if err := json.Unmarshal(data, &payload); err != nil {
		return game.RunState{}, err
	}

	scenariosForValidation := append([]game.Scenario{}, availableScenarios...)
	if payload.Run.Scenario.ID != "" {
		scenariosForValidation = append(scenariosForValidation, payload.Run.Scenario)
	}

	if err := validateRunConfigWithScenarios(payload.Run.Config, scenariosForValidation); err != nil {
		return game.RunState{}, fmt.Errorf("invalid saved run config: %w", err)
	}

	if payload.Run.SeasonSetID == "" {
		payload.Run.SeasonSetID = payload.Run.Scenario.DefaultSeasonSetID
	}
	payload.Run.EnsureWeather()

	return payload.Run, nil
}

func loadCustomScenarios(path string) ([]customScenarioRecord, error) {
	data, err := readDataFile(path, maxScenarioFileBytes)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var library customScenarioLibrary
	if err := json.Unmarshal(data, &library); err != nil {
		return nil, err
	}

	if len(library.Custom) > 0 {
		for i := range library.Custom {
			if library.Custom[i].PreferredMode == "" {
				library.Custom[i].PreferredMode = game.ModeAlone
			}
			if len(library.Custom[i].Scenario.SupportedModes) == 0 {
				library.Custom[i].Scenario.SupportedModes = []game.GameMode{library.Custom[i].PreferredMode}
			}
		}
		return library.Custom, nil
	}

	// Legacy format support.
	if len(library.Scenarios) > 0 {
		records := make([]customScenarioRecord, 0, len(library.Scenarios))
		for _, scenario := range library.Scenarios {
			if len(scenario.SupportedModes) == 0 {
				scenario.SupportedModes = []game.GameMode{game.ModeAlone}
			}
			records = append(records, customScenarioRecord{
				Scenario:      scenario,
				PreferredMode: game.ModeAlone,
			})
		}
		return records, nil
	}

	return nil, nil
}

func saveCustomScenarios(path string, scenarios []customScenarioRecord) error {
	if err := validateDataFilePath(path); err != nil {
		return err
	}

	library := customScenarioLibrary{
		FormatVersion: 1,
		Custom:        scenarios,
	}

	data, err := json.MarshalIndent(library, "", "  ")
	if err != nil {
		return err
	}
	if len(data) > maxScenarioFileBytes {
		return fmt.Errorf("scenario data exceeds %d bytes", maxScenarioFileBytes)
	}

	return os.WriteFile(path, data, 0o600)
}

func makeCustomScenarioID(name string, existing []game.Scenario) game.ScenarioID {
	base := strings.ToLower(strings.TrimSpace(name))
	replacer := strings.NewReplacer(
		" ", "_",
		"-", "_",
		".", "_",
		"/", "_",
		"\\", "_",
	)
	base = replacer.Replace(base)
	base = strings.Trim(base, "_")
	if base == "" {
		base = "scenario"
	}

	candidate := game.ScenarioID("custom_" + base)
	taken := make(map[game.ScenarioID]bool, len(existing))
	for _, s := range existing {
		taken[s.ID] = true
	}
	if !taken[candidate] {
		return candidate
	}

	for i := 2; ; i++ {
		next := game.ScenarioID(fmt.Sprintf("%s_%d", candidate, i))
		if !taken[next] {
			return next
		}
	}
}

func parseNonNegativeInt(raw string) (int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, fmt.Errorf("empty")
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if parsed < 0 {
		return 0, fmt.Errorf("negative")
	}
	return parsed, nil
}

func normalizedCSVList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, item)
	}
	return out
}

func selectedSeasonIndex(season game.SeasonID) int {
	options := builderSeasonOptions()
	for i, option := range options {
		if option == season {
			return i
		}
	}
	return 0
}

func indexOfInt(values []int, target int) int {
	for i, value := range values {
		if value == target {
			return i
		}
	}
	return -1
}

func validateRunConfigWithScenarios(c game.RunConfig, scenarios []game.Scenario) error {
	switch c.Mode {
	case game.ModeNakedAndAfraid:
	case game.ModeNakedAndAfraidXL:
	case game.ModeAlone:
	default:
		return fmt.Errorf("invalid mode: %s", c.Mode)
	}

	if c.PlayerCount < 1 || c.PlayerCount > 8 {
		return fmt.Errorf("player count must be between 1 and 8, got %d", c.PlayerCount)
	}

	found := c.ScenarioID == game.ScenarioRandomID
	if !found {
		for _, scenario := range scenarios {
			if scenario.ID == c.ScenarioID {
				found = true
				break
			}
		}
	}
	if !found {
		return fmt.Errorf("scenario not found: %s", c.ScenarioID)
	}

	if c.ScenarioID != game.ScenarioRandomID {
		scenario, ok := game.GetScenario(scenarios, c.ScenarioID)
		if !ok {
			return fmt.Errorf("scenario not found: %s", c.ScenarioID)
		}
		if !scenarioSupportsMode(scenario, c.Mode) {
			return fmt.Errorf("scenario %s does not support mode %s", c.ScenarioID, c.Mode)
		}
	}

	if !c.RunLength.IsValid() {
		return fmt.Errorf("invalid run length")
	}

	return nil
}

func newRunStateWithScenarios(config game.RunConfig, scenarios []game.Scenario) (game.RunState, error) {
	resolvedConfig := config

	if err := validateRunConfigWithScenarios(resolvedConfig, scenarios); err != nil {
		return game.RunState{}, err
	}

	if resolvedConfig.Seed == 0 {
		resolvedConfig.Seed = time.Now().UnixNano()
	}

	if resolvedConfig.ScenarioID == game.ScenarioRandomID {
		rng := seededRNG(resolvedConfig.Seed)
		resolvedConfig.ScenarioID = scenarios[rng.IntN(len(scenarios))].ID
	}

	scenario, found := game.GetScenario(scenarios, resolvedConfig.ScenarioID)
	if !found {
		return game.RunState{}, fmt.Errorf("scenario not found: %s", resolvedConfig.ScenarioID)
	}

	state := game.RunState{
		Config:      resolvedConfig,
		Scenario:    scenario,
		SeasonSetID: scenario.DefaultSeasonSetID,
		Day:         1,
		Players:     game.CreatePlayers(resolvedConfig),
	}
	state.EnsureWeather()
	return state, nil
}

func seededRNG(seed int64) *rand.Rand {
	// Non-cryptographic PRNG is intentional for deterministic simulation behavior.
	// #nosec G404
	return rand.New(rand.NewPCG(seedWord(seed, "a"), seedWord(seed, "b")))
}

func seedWord(seed int64, salt string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(fmt.Sprintf("%d:%s", seed, salt)))
	return h.Sum64()
}

func resizeTerminalBestEffort(cols, rows int) tea.Cmd {
	return func() tea.Msg {
		fmt.Printf("\x1b[8;%d;%dt", rows, cols) // CSI 8; rows; cols t
		return nil
	}
}
