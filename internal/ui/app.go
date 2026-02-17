package ui

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
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
	screenLoadRun
	screenScenarioBuilder
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
)

// --- Menu model ---

type menuItem int

const (
	itemStart menuItem = iota
	itemLoadRun
	itemScenarioBuilder
	itemCheckUpdate
	itemInstallUpdate
	itemQuit
)

const (
	defaultCustomScenariosFile = "survive-it-scenarios.json"
	saveSlotCount              = 3
)

type menuModel struct {
	w, h   int
	cfg    AppConfig
	idx    int
	screen screen
	setup  setupState
	load   loadRunState
	build  scenarioBuilderState

	run             *game.RunState
	runInput        string
	activeSaveSlot  int
	customScenarios []game.Scenario
	loadReturn      screen
	status          string
	busy            bool
	err             string
}

func newMenuModel(cfg AppConfig) menuModel {
	customScenarios, _ := loadCustomScenarios(defaultCustomScenariosFile)

	return menuModel{
		cfg:             cfg,
		idx:             0,
		setup:           newSetupState(),
		load:            newLoadRunState(),
		build:           newScenarioBuilderState(),
		activeSaveSlot:  1,
		customScenarios: customScenarios,
	}
}

func (m menuModel) Init() tea.Cmd {
	// Approximate 1024x768 using typical terminal cells at ~8x16 px.
	return resizeTerminalBestEffort(128, 48)
}

type updateResultMsg struct {
	status string
	err    error
}

type savedRun struct {
	FormatVersion int           `json:"format_version"`
	SavedAt       time.Time     `json:"saved_at"`
	Run           game.RunState `json:"run"`
}

type customScenarioLibrary struct {
	FormatVersion int             `json:"format_version"`
	Scenarios     []game.Scenario `json:"scenarios"`
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
		if m.screen == screenLoadRun {
			return m.updateLoadRun(msg)
		}
		if m.screen == screenScenarioBuilder {
			return m.updateScenarioBuilder(msg)
		}

		return m.updateMenu(msg)
	case updateResultMsg:
		m.busy = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Update check failed: %v", msg.err)

			return m, nil
		}
		m.status = msg.status

		return m, nil
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
	if m.screen == screenLoadRun {
		return m.viewLoadRun()
	}
	if m.screen == screenScenarioBuilder {
		return m.viewScenarioBuilder()
	}
	if m.screen == screenRun {
		return m.viewRun()
	}
	return m.viewMenu()
}

func (m menuModel) updateRun(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "Q", "esc":
		m.screen = screenMenu
		return m, nil

	case "N":
		m.advanceRunDay()
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
		m.load = newLoadRunState()
		m.loadReturn = screenRun
		m.screen = screenLoadRun
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
		m.load = newLoadRunState()
		m.loadReturn = screenRun
		m.screen = screenLoadRun
		return m, nil
	case "menu", "back":
		m.screen = screenMenu
		return m, nil
	default:
		m.status = "Unknown command. Try: next, save, load, menu."
		return m, nil
	}
}

func menuItems() []string {
	return []string{
		"Start",
		"Load Run",
		"Scenario Builder",
		"Check for updates",
		"Install Update",
		"Quit",
	}
}

func (m menuModel) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.busy {
		// Ignore input while update check runs.
		return m, nil
	}
	itemCount := len(menuItems())
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
		switch menuItem(m.idx) {
		case itemStart:
			m.setup = newSetupState()
			m.screen = screenSetup
			m.status = ""
			return m, nil
		case itemLoadRun:
			m.load = newLoadRunState()
			m.loadReturn = screenMenu
			m.screen = screenLoadRun
			return m, nil
		case itemScenarioBuilder:
			m.build = newScenarioBuilderState()
			m.screen = screenScenarioBuilder
			return m, nil
		case itemCheckUpdate:
			if m.cfg.NoUpdate {
				m.status = "Update checks disabled (--no-update)."
				return m, nil
			}
			m.busy = true
			m.status = "Checking for updates…"
			return m, checkUpdateCmd(m.cfg.Version)
		case itemInstallUpdate:
			if m.cfg.NoUpdate {
				m.status = "Update checks disabled (--no-update)."
				return m, nil
			}
			m.busy = true
			m.status = "Installing update…"
			return m, applyUpdateCmd(m.cfg.Version)
		case itemQuit:
			return m, tea.Quit
		}
	}

	return m, nil
}

type setupScenarioOption struct {
	id    game.ScenarioID
	label string
}

type saveSlotMeta struct {
	Slot      int
	Path      string
	Exists    bool
	Summary   string
	ErrDetail string
}

type loadRunState struct {
	cursor int
}

type scenarioBuilderState struct {
	cursor           int
	name             string
	biomeIdx         int
	defaultDaysIdx   int
	seasonProfileIdx int
}

type runLengthOption struct {
	label     string
	openEnded bool
	days      int
}

type setupState struct {
	cursor         int
	modeIdx        int
	scenarioIdx    int
	playerCountIdx int
	runLengthIdx   int
}

func newSetupState() setupState {
	return setupState{
		cursor:         0,
		modeIdx:        0,
		scenarioIdx:    0,
		playerCountIdx: 1, // 2 players by default
		runLengthIdx:   0,
	}
}

func newLoadRunState() loadRunState {
	return loadRunState{cursor: 0}
}

func newScenarioBuilderState() scenarioBuilderState {
	return scenarioBuilderState{
		cursor:           0,
		name:             "",
		biomeIdx:         0,
		defaultDaysIdx:   1,
		seasonProfileIdx: 0,
	}
}

func setupModes() []game.GameMode {
	return []game.GameMode{
		game.ModeAlone,
		game.ModeNakedAndAfraid,
	}
}

func (m menuModel) setupScenarioOptions() []setupScenarioOption {
	options := []setupScenarioOption{
		{id: game.ScenarioRandomID, label: "Random"},
	}

	for _, scenario := range m.availableScenarios() {
		label := scenario.Name
		if isCustomScenarioID(scenario.ID) {
			label = scenario.Name + " (Custom)"
		}
		options = append(options, setupScenarioOption{
			id:    scenario.ID,
			label: label,
		})
	}

	return options
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

func builderBiomes() []string {
	return []string{"Forest", "Jungle", "Arctic", "Coast", "Mountain", "Desert"}
}

func builderDefaultDays() []int {
	return []int{7, 14, 30, 60, 90}
}

func builderSeasonProfiles() []struct {
	label string
	set   game.SeasonSet
} {
	return []struct {
		label string
		set   game.SeasonSet
	}{
		{
			label: "Autumn -> Winter",
			set: game.SeasonSet{
				ID: game.SeasonSetAloneDefaultID,
				Phases: []game.SeasonPhase{
					{Season: game.SeasonAutumn, Days: 14},
					{Season: game.SeasonWinter, Days: 0},
				},
			},
		},
		{
			label: "Wet",
			set: game.SeasonSet{
				ID: game.SeasonSetWetDefaultID,
				Phases: []game.SeasonPhase{
					{Season: game.SeasonWet, Days: 0},
				},
			},
		},
		{
			label: "Winter",
			set: game.SeasonSet{
				ID: game.SeasonSetWinterDefaultID,
				Phases: []game.SeasonPhase{
					{Season: game.SeasonWinter, Days: 0},
				},
			},
		},
	}
}

func wrapIndex(current, delta, size int) int {
	next := current + delta
	for next < 0 {
		next += size
	}
	return next % size
}

func (m menuModel) updateSetup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	const rowCount = 6 // mode, scenario, players, run length, start, cancel

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "Q", "q", "esc":
		m.screen = screenMenu
		return m, nil
	case "up", "k":
		m.setup.cursor = wrapIndex(m.setup.cursor, -1, rowCount)
		return m, nil
	case "down", "j":
		m.setup.cursor = wrapIndex(m.setup.cursor, 1, rowCount)
		return m, nil
	case "left":
		m = m.adjustSetupChoice(-1)
		return m, nil
	case "right":
		m = m.adjustSetupChoice(1)
		return m, nil
	case "enter":
		switch m.setup.cursor {
		case 4:
			return m.startRunFromSetup()
		case 5:
			m.screen = screenMenu
			return m, nil
		default:
			m = m.adjustSetupChoice(1)
			return m, nil
		}
	}

	return m, nil
}

func (m menuModel) adjustSetupChoice(delta int) menuModel {
	scenarios := m.setupScenarioOptions()
	switch m.setup.cursor {
	case 0:
		m.setup.modeIdx = wrapIndex(m.setup.modeIdx, delta, len(setupModes()))
	case 1:
		m.setup.scenarioIdx = wrapIndex(m.setup.scenarioIdx, delta, len(scenarios))
	case 2:
		m.setup.playerCountIdx = wrapIndex(m.setup.playerCountIdx, delta, len(setupPlayerCounts()))
	case 3:
		m.setup.runLengthIdx = wrapIndex(m.setup.runLengthIdx, delta, len(setupRunLengths()))
	}

	return m
}

func (m menuModel) startRunFromSetup() (tea.Model, tea.Cmd) {
	modes := setupModes()
	scenarioOptions := m.setupScenarioOptions()
	playerCounts := setupPlayerCounts()
	runLengths := setupRunLengths()

	runLength := runLengths[m.setup.runLengthIdx]
	cfg := game.RunConfig{
		Mode:        modes[m.setup.modeIdx],
		ScenarioID:  scenarioOptions[m.setup.scenarioIdx].id,
		PlayerCount: playerCounts[m.setup.playerCountIdx],
		RunLength: game.RunLength{
			OpenEnded: runLength.openEnded,
			Days:      runLength.days,
		},
		Seed:    0,
		Players: nil,
	}

	state, err := newRunStateWithScenarios(cfg, m.availableScenarios())
	if err != nil {
		m.status = fmt.Sprintf("Failed to start: %v", err)
		return m, nil
	}

	m.run = &state
	m.screen = screenRun
	m.status = ""
	return m, nil
}

func (m menuModel) viewSetup() string {
	modes := setupModes()
	scenarios := m.setupScenarioOptions()
	playerCounts := setupPlayerCounts()
	runLengths := setupRunLengths()

	rows := []struct {
		label string
		value string
	}{
		{label: "Mode", value: modeLabel(modes[m.setup.modeIdx])},
		{label: "Scenario", value: scenarios[m.setup.scenarioIdx].label},
		{label: "Players", value: fmt.Sprintf("%d", playerCounts[m.setup.playerCountIdx])},
		{label: "Run Length", value: runLengths[m.setup.runLengthIdx].label},
		{label: "Start Run", value: ""},
		{label: "Cancel", value: ""},
	}

	var b strings.Builder
	b.WriteString(lipgloss.JoinVertical(
		lipgloss.Left,
		brightGreen.Render("NEW RUN WIZARD"),
		dimGreen.Render("Choose options, then select Start Run"),
	) + "\n")
	b.WriteString(border.Render("----------------------------------------") + "\n\n")

	for i, row := range rows {
		cursor := "  "
		lineStyle := green
		if i == m.setup.cursor {
			cursor = "> "
			lineStyle = brightGreen
		}

		if row.value == "" {
			b.WriteString(cursor + lineStyle.Render(row.label) + "\n")
			continue
		}

		b.WriteString(cursor + lineStyle.Render(fmt.Sprintf("%-10s %s", row.label+":", row.value)) + "\n")
	}

	b.WriteString("\n" + border.Render("----------------------------------------") + "\n")
	b.WriteString(dimGreen.Render("↑/↓ move  ←/→ change  Enter select  Shift+Q back") + "\n")
	if m.status != "" {
		b.WriteString("\n" + green.Render(m.status) + "\n")
	}

	return b.String()
}

func (m menuModel) availableScenarios() []game.Scenario {
	scenarios := append([]game.Scenario{}, game.BuiltInScenarios()...)
	scenarios = append(scenarios, m.customScenarios...)
	return scenarios
}

func isCustomScenarioID(id game.ScenarioID) bool {
	return strings.HasPrefix(string(id), "custom_")
}

func (m menuModel) updateLoadRun(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	rowCount := saveSlotCount + 1 // slots + cancel

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "Q", "q", "esc":
		if m.loadReturn == 0 {
			m.screen = screenMenu
		} else {
			m.screen = m.loadReturn
		}
		return m, nil
	case "up", "k":
		m.load.cursor = wrapIndex(m.load.cursor, -1, rowCount)
		return m, nil
	case "down", "j":
		m.load.cursor = wrapIndex(m.load.cursor, 1, rowCount)
		return m, nil
	case "enter":
		if m.load.cursor == saveSlotCount {
			if m.loadReturn == 0 {
				m.screen = screenMenu
			} else {
				m.screen = m.loadReturn
			}
			return m, nil
		}

		slot := m.load.cursor + 1
		state, err := loadRunFromFile(savePathForSlot(slot), m.availableScenarios())
		if err != nil {
			m.status = fmt.Sprintf("Load failed: %v", err)
			return m, nil
		}

		m.run = &state
		m.activeSaveSlot = slot
		m.screen = screenRun
		m.status = fmt.Sprintf("Loaded slot %d", slot)
		return m, nil
	}

	return m, nil
}

func loadSlotMetadata(slot int) saveSlotMeta {
	path := savePathForSlot(slot)
	meta := saveSlotMeta{
		Slot: slot,
		Path: path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		meta.Exists = false
		meta.Summary = "Empty"
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
	meta.Summary = fmt.Sprintf("%s | Day %d | %s",
		payload.Run.Scenario.Name,
		payload.Run.Day,
		payload.SavedAt.Local().Format("2006-01-02 15:04"),
	)
	return meta
}

func (m menuModel) viewLoadRun() string {
	var b strings.Builder

	b.WriteString(brightGreen.Render("LOAD RUN") + "\n")
	b.WriteString(dimGreen.Render("Pick a save slot to load") + "\n")
	b.WriteString(border.Render("----------------------------------------") + "\n\n")

	for i := 0; i < saveSlotCount; i++ {
		cursor := "  "
		lineStyle := green
		if i == m.load.cursor {
			cursor = "> "
			lineStyle = brightGreen
		}

		meta := loadSlotMetadata(i + 1)
		line := fmt.Sprintf("Slot %d: %s", meta.Slot, meta.Summary)
		b.WriteString(cursor + lineStyle.Render(line) + "\n")
	}

	cursor := "  "
	lineStyle := green
	if m.load.cursor == saveSlotCount {
		cursor = "> "
		lineStyle = brightGreen
	}
	b.WriteString(cursor + lineStyle.Render("Cancel") + "\n")

	b.WriteString("\n" + border.Render("----------------------------------------") + "\n")
	b.WriteString(dimGreen.Render("↑/↓ move  Enter select  Shift+Q back") + "\n")
	if m.status != "" {
		b.WriteString("\n" + green.Render(m.status) + "\n")
	}

	return b.String()
}

func (m menuModel) updateScenarioBuilder(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	const rowCount = 6 // name, biome, days, season, save, cancel

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "Q", "q", "esc":
		m.screen = screenMenu
		return m, nil
	case "up", "k":
		m.build.cursor = wrapIndex(m.build.cursor, -1, rowCount)
		return m, nil
	case "down", "j":
		m.build.cursor = wrapIndex(m.build.cursor, 1, rowCount)
		return m, nil
	case "left":
		m = m.adjustScenarioBuilderChoice(-1)
		return m, nil
	case "right":
		m = m.adjustScenarioBuilderChoice(1)
		return m, nil
	case "backspace", "ctrl+h":
		if m.build.cursor == 0 && len(m.build.name) > 0 {
			m.build.name = m.build.name[:len(m.build.name)-1]
			return m, nil
		}
	case "enter":
		switch m.build.cursor {
		case 4:
			return m.saveScenarioFromBuilder()
		case 5:
			m.screen = screenMenu
			return m, nil
		case 0:
			m.build.name += " "
			return m, nil
		default:
			m = m.adjustScenarioBuilderChoice(1)
			return m, nil
		}
	}

	if m.build.cursor == 0 && len(msg.Runes) > 0 {
		m.build.name += string(msg.Runes)
		return m, nil
	}

	return m, nil
}

func (m menuModel) adjustScenarioBuilderChoice(delta int) menuModel {
	switch m.build.cursor {
	case 1:
		m.build.biomeIdx = wrapIndex(m.build.biomeIdx, delta, len(builderBiomes()))
	case 2:
		m.build.defaultDaysIdx = wrapIndex(m.build.defaultDaysIdx, delta, len(builderDefaultDays()))
	case 3:
		m.build.seasonProfileIdx = wrapIndex(m.build.seasonProfileIdx, delta, len(builderSeasonProfiles()))
	}

	return m
}

func (m menuModel) saveScenarioFromBuilder() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.build.name)
	if name == "" {
		m.status = "Scenario name is required."
		return m, nil
	}

	seasonProfiles := builderSeasonProfiles()
	profile := seasonProfiles[m.build.seasonProfileIdx]
	scenarioID := makeCustomScenarioID(name, m.availableScenarios())

	newScenario := game.Scenario{
		ID:          scenarioID,
		Name:        name,
		Biome:       builderBiomes()[m.build.biomeIdx],
		DefaultDays: builderDefaultDays()[m.build.defaultDaysIdx],
		IssuedKit:   game.IssuedKit{},
		SeasonSets: []game.SeasonSet{
			profile.set,
		},
		DefaultSeasonSetID: profile.set.ID,
	}

	updated := append(append([]game.Scenario{}, m.customScenarios...), newScenario)
	if err := saveCustomScenarios(defaultCustomScenariosFile, updated); err != nil {
		m.status = fmt.Sprintf("Failed to save scenario: %v", err)
		return m, nil
	}

	m.customScenarios = updated
	m.status = fmt.Sprintf("Saved custom scenario: %s", newScenario.Name)
	m.screen = screenMenu
	return m, nil
}

func (m menuModel) viewScenarioBuilder() string {
	biomes := builderBiomes()
	days := builderDefaultDays()
	seasonProfiles := builderSeasonProfiles()

	rows := []struct {
		label string
		value string
	}{
		{label: "Name", value: m.build.name},
		{label: "Biome", value: biomes[m.build.biomeIdx]},
		{label: "Default Days", value: fmt.Sprintf("%d", days[m.build.defaultDaysIdx])},
		{label: "Season Profile", value: seasonProfiles[m.build.seasonProfileIdx].label},
		{label: "Save Scenario", value: ""},
		{label: "Cancel", value: ""},
	}

	var b strings.Builder
	b.WriteString(brightGreen.Render("SCENARIO BUILDER") + "\n")
	b.WriteString(dimGreen.Render("Create a custom scenario for the run wizard") + "\n")
	b.WriteString(border.Render("----------------------------------------") + "\n\n")

	for i, row := range rows {
		cursor := "  "
		lineStyle := green
		if i == m.build.cursor {
			cursor = "> "
			lineStyle = brightGreen
		}

		if row.value == "" {
			b.WriteString(cursor + lineStyle.Render(row.label) + "\n")
			continue
		}

		value := row.value
		if i == 0 && strings.TrimSpace(value) == "" {
			value = "<type name>"
		}
		b.WriteString(cursor + lineStyle.Render(fmt.Sprintf("%-14s %s", row.label+":", value)) + "\n")
	}

	b.WriteString("\n" + border.Render("----------------------------------------") + "\n")
	b.WriteString(dimGreen.Render("↑/↓ move  ←/→ change  type name  Enter select  Shift+Q back") + "\n")
	if m.status != "" {
		b.WriteString("\n" + green.Render(m.status) + "\n")
	}

	return b.String()
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

	headerRows := 4
	controlsRows := 1
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
		Border(lipgloss.NormalBorder()).
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
	title := lipgloss.JoinVertical(
		lipgloss.Left,
		brightGreen.Render("SURVIVE IT"),
		dimGreen.Render("alpha"),
	)
	ver := dimGreen.Render(fmt.Sprintf("v%s  (%s)  %s", m.cfg.Version, m.cfg.Commit, m.cfg.BuildDate))

	items := menuItems()

	var b strings.Builder
	b.WriteString(title + "\n")
	b.WriteString(ver + "\n")
	b.WriteString(border.Render("----------------------------------------") + "\n\n")

	for i, it := range items {
		cursor := "  "
		line := it
		if i == m.idx {
			cursor = "> "
			line = brightGreen.Render(it)
		} else {
			line = green.Render(it)
		}
		b.WriteString(cursor + line + "\n")
	}

	b.WriteString("\n" + border.Render("----------------------------------------") + "\n")
	b.WriteString(dimGreen.Render("↑/↓ to move, Enter to select, Shift+Q to quit") + "\n")
	if m.status != "" {
		b.WriteString("\n" + green.Render(m.status) + "\n")
	}

	return b.String()
}

func checkUpdateCmd(currentVersion string) tea.Cmd {
	return func() tea.Msg {
		// Tiny delay so the UI visibly switches to a busy state.
		time.Sleep(150 * time.Millisecond)

		res, err := update.Check(update.CheckParams{
			CurrentVersion: currentVersion,
		})
		if err != nil {
			return updateResultMsg{err: err}
		}
		return updateResultMsg{status: res}
	}
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

func (m menuModel) headerText() string {
	season, ok := m.run.CurrentSeason()
	seasonStr := "unknown"
	if ok {
		seasonStr = string(season)
	}

	var b strings.Builder
	b.WriteString(lipgloss.JoinVertical(
		lipgloss.Left,
		brightGreen.Render("SURVIVE IT!"),
		dimGreen.Render(fmt.Sprintf("Mode: %s  |  Scenario: %s  |  Season: %s  |  Day: %d",
			modeLabel(m.run.Config.Mode), m.run.Scenario.Name, seasonStr, m.run.Day)),
	))
	return b.String()
}

func modeLabel(mode game.GameMode) string {
	switch mode {
	case game.ModeAlone:
		return "Alone"
	case game.ModeNakedAndAfraid:
		return "Naked and Afraid"
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
		})
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(border).
		BorderHeader(true).
		BorderRow(false).
		Headers("Player", "Sex", "Body", "Energy", "Hydration", "Morale").
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
	text := fmt.Sprintf(" Shift+N Next Day  |  Shift+S Save Slot %d  |  Shift+L Load  |  Shift+Q Menu ", m.activeSaveSlot)
	maxWidth := totalWidth - 2
	if maxWidth < 20 {
		maxWidth = 20
	}
	return border.Width(maxWidth).Render(text)
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
	payload := savedRun{
		FormatVersion: 1,
		SavedAt:       time.Now().UTC(),
		Run:           run,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func savePathForSlot(slot int) string {
	return fmt.Sprintf("survive-it-save-%d.json", slot)
}

func loadRunFromFile(path string, availableScenarios []game.Scenario) (game.RunState, error) {
	data, err := os.ReadFile(path)
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

	return payload.Run, nil
}

func loadCustomScenarios(path string) ([]game.Scenario, error) {
	data, err := os.ReadFile(path)
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

	return library.Scenarios, nil
}

func saveCustomScenarios(path string, scenarios []game.Scenario) error {
	library := customScenarioLibrary{
		FormatVersion: 1,
		Scenarios:     scenarios,
	}

	data, err := json.MarshalIndent(library, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
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

func validateRunConfigWithScenarios(c game.RunConfig, scenarios []game.Scenario) error {
	switch c.Mode {
	case game.ModeNakedAndAfraid:
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
		seed1 := uint64(resolvedConfig.Seed)
		seed2 := seed1 ^ uint64(0x9e3779b97f4a7c15)
		rng := rand.New(rand.NewPCG(seed1, seed2))
		resolvedConfig.ScenarioID = scenarios[rng.IntN(len(scenarios))].ID
	}

	scenario, found := game.GetScenario(scenarios, resolvedConfig.ScenarioID)
	if !found {
		return game.RunState{}, fmt.Errorf("scenario not found: %s", resolvedConfig.ScenarioID)
	}

	return game.RunState{
		Config:      resolvedConfig,
		Scenario:    scenario,
		SeasonSetID: scenario.DefaultSeasonSetID,
		Day:         1,
		Players:     game.CreatePlayers(resolvedConfig),
	}, nil
}

func resizeTerminalBestEffort(cols, rows int) tea.Cmd {
	return func() tea.Msg {
		fmt.Printf("\x1b[8;%d;%dt", rows, cols) // CSI 8; rows; cols t
		return nil
	}
}
