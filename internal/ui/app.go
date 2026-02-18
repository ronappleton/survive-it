package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"os"
	"path/filepath"
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
	screenPlayerConfig
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
	pcfg   playerConfigState
	load   loadRunState
	build  scenarioBuilderState

	run             *game.RunState
	runInput        string
	activeSaveSlot  int
	customScenarios []customScenarioRecord
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
		pick:            newScenarioPickerState(),
		pcfg:            newPlayerConfigState(),
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
		if m.screen == screenPlayerConfig {
			return m.updatePlayerConfig(msg)
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
	if m.screen == screenScenarioPicker {
		return m.viewScenarioPicker()
	}
	if m.screen == screenPlayerConfig {
		return m.viewPlayerConfig()
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
			m.pcfg = newPlayerConfigState()
			m = m.ensureSetupScenarioSelection()
			m = m.ensureSetupPlayers()
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
	scenario game.Scenario
	label    string
}

type scenarioPickerState struct {
	cursor int
}

type playerConfigState struct {
	cursor       int
	playerIdx    int
	nameIdx      int
	addKitIdx    int
	addIssuedIdx int
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
	cursor         int
	selectedIdx    int // 0 = new, 1..N = existing custom scenario index+1
	name           string
	modeIdx        int
	biomeIdx       int
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

type scenarioBuilderRowKind int

const (
	builderRowScenario scenarioBuilderRowKind = iota
	builderRowName
	builderRowMode
	builderRowBiome
	builderRowDaysMode
	builderRowDaysPreset
	builderRowDaysCustom
	builderRowSeasonProfileID
	builderRowPhaseSeason
	builderRowPhaseDays
	builderRowAddPhase
	builderRowRemovePhase
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

type playerConfigRowKind int

const (
	playerRowPlayer playerConfigRowKind = iota
	playerRowName
	playerRowQuickName
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
	playerRowAddPersonalKit
	playerRowRemovePersonalKit
	playerRowIssuedKit
	playerRowAddIssuedKit
	playerRowRemoveIssuedKit
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
		playerCountIdx: 1, // 2 players by default
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

func newPlayerConfigState() playerConfigState {
	return playerConfigState{
		cursor:       0,
		playerIdx:    0,
		nameIdx:      0,
		addKitIdx:    0,
		addIssuedIdx: 0,
	}
}

func newLoadRunState() loadRunState {
	return loadRunState{cursor: 0}
}

func newScenarioBuilderState() scenarioBuilderState {
	return scenarioBuilderState{
		cursor:         0,
		selectedIdx:    0,
		name:           "",
		modeIdx:        0,
		biomeIdx:       0,
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

func playerConfigSuggestedNames() []string {
	return []string{
		"Sophia", "Daniel", "Maya", "Ethan", "Harper",
		"Riley", "Quinn", "Rowan", "Kai", "Avery",
		"Emma", "Olivia", "Jack", "Tom", "Freya",
		"Jordan", "Morgan", "Taylor", "Alex", "Noah",
	}
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

	if mode == game.ModeNakedAndAfraid {
		return firstNKitItems(kit, 2)
	}
	if mode == game.ModeNakedAndAfraidXL {
		return firstNKitItems(kit, 3)
	}
	return firstNKitItems(kit, 4)
}

func firstNKitItems(items []game.KitItem, n int) []game.KitItem {
	if len(items) <= n {
		return append([]game.KitItem(nil), items...)
	}
	return append([]game.KitItem(nil), items[:n]...)
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

func builderScenarioChoices(custom []customScenarioRecord) []string {
	choices := []string{"New Scenario"}
	for _, record := range custom {
		choices = append(choices, record.Scenario.Name)
	}
	return choices
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

func (m menuModel) updateSetup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	const rowCount = 7 // mode, scenario, players, run length, configure players, start, cancel
	m = m.ensureSetupScenarioSelection()
	m = m.ensureSetupPlayers()

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
		if m.setup.cursor == 1 {
			m = m.openScenarioPicker()
			return m, nil
		}
		if m.setup.cursor == 4 {
			m.screen = screenPlayerConfig
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
			m.screen = screenPlayerConfig
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
			m.screen = screenPlayerConfig
			return m, nil
		case 5:
			return m.startRunFromSetup()
		case 6:
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

	rows := []struct {
		label string
		value string
	}{
		{label: "Mode", value: modeLabel(modes[m.setup.modeIdx])},
		{label: "Scenario", value: scenarioLabel},
		{label: "Players", value: fmt.Sprintf("%d", playerCounts[m.setup.playerCountIdx])},
		{label: "Run Length", value: runLengths[m.setup.runLengthIdx].label},
		{label: "Configure Players", value: ""},
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
	b.WriteString(dimGreen.Render("↑/↓ move  ←/→ change  Enter select  (Scenario opens picker)  Shift+Q back") + "\n")
	if m.status != "" {
		b.WriteString("\n" + green.Render(m.status) + "\n")
	}

	return b.String()
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
		m.screen = screenSetup
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
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(listWidth).
		Height(contentHeight).
		Render(list.String())
	detailPane := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(detailWidth).
		Height(contentHeight).
		Render(detailText)

	var b strings.Builder
	b.WriteString(brightGreen.Render("SCENARIO SELECT") + "\n")
	b.WriteString(dimGreen.Render("Mode: "+modeLabel(m.setupMode())+"  |  ↑/↓ browse, Enter select, Shift+Q back") + "\n\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane))
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
	b.WriteString(brightGreen.Render(s.Name) + "\n")
	b.WriteString(green.Render(fmt.Sprintf("Biome: %s", s.Biome)) + "\n")
	b.WriteString(green.Render(fmt.Sprintf("Default Days: %d", s.DefaultDays)) + "\n\n")
	b.WriteString(dimGreen.Render(desc) + "\n\n")
	b.WriteString(brightGreen.Render("Daunting") + "\n")
	b.WriteString(green.Render(daunting) + "\n\n")
	b.WriteString(brightGreen.Render("Motivation") + "\n")
	b.WriteString(green.Render(motivation))
	return b.String()
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
		case playerRowQuickName:
			names := playerConfigSuggestedNames()
			if len(names) == 0 || len(m.setup.players) == 0 {
				return m, nil
			}
			m.setup.players[m.pcfg.playerIdx].Name = names[m.pcfg.nameIdx]
		case playerRowAddPersonalKit:
			m = m.addPersonalKitItem()
		case playerRowRemovePersonalKit:
			m = m.removePersonalKitItem()
		case playerRowAddIssuedKit:
			m = m.addIssuedKitItem()
		case playerRowRemoveIssuedKit:
			m = m.removeIssuedKitItem()
		case playerRowResetIssuedKit:
			m = m.resetIssuedKitRecommendations()
		case playerRowBack:
			m.screen = screenSetup
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

	var b strings.Builder
	b.WriteString(brightGreen.Render("PLAYER CONFIGURATION") + "\n")
	b.WriteString(dimGreen.Render(fmt.Sprintf("Mode: %s  |  Scenario: %s  |  Player %d/%d", modeLabel(m.setupMode()), scenarioLabel, activePlayer, playerCount)) + "\n")
	b.WriteString(dimGreen.Render(fmt.Sprintf("Series limit: up to %d personal kit item(s) per player", maxKitLimitForMode(m.setupMode()))) + "\n")
	b.WriteString(border.Render("------------------------------------------------------------") + "\n\n")

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
			b.WriteString(cursor + lineStyle.Render(row.label) + "\n")
			continue
		}

		value := row.value
		if row.kind == playerRowName && strings.TrimSpace(value) == "" {
			value = "<auto/random or type custom name>"
		}
		b.WriteString(cursor + lineStyle.Render(fmt.Sprintf("%-24s %s", row.label+":", value)) + "\n")
	}

	b.WriteString("\n" + border.Render("------------------------------------------------------------") + "\n")
	b.WriteString(dimGreen.Render("↑/↓ move  ←/→ change  Enter select  type to edit Name  Shift+Q back") + "\n")
	b.WriteString(dimGreen.Render("Use Add rows for kit picks and Remove rows to undo selections.") + "\n")
	if m.status != "" {
		b.WriteString("\n" + green.Render(m.status) + "\n")
	}

	return b.String()
}

func (m menuModel) playerConfigRows() []playerConfigRow {
	if len(m.setup.players) == 0 {
		return nil
	}

	allKit := game.AllKitItems()
	names := playerConfigSuggestedNames()
	if len(names) == 0 {
		names = []string{"Player"}
	}
	playerIdx := m.pcfg.playerIdx
	if playerIdx < 0 || playerIdx >= len(m.setup.players) {
		playerIdx = 0
	}
	nameIdx := m.pcfg.nameIdx
	if nameIdx < 0 || nameIdx >= len(names) {
		nameIdx = 0
	}
	addKitIdx := m.pcfg.addKitIdx
	addIssuedIdx := m.pcfg.addIssuedIdx
	if len(allKit) > 0 {
		if addKitIdx < 0 || addKitIdx >= len(allKit) {
			addKitIdx = 0
		}
		if addIssuedIdx < 0 || addIssuedIdx >= len(allKit) {
			addIssuedIdx = 0
		}
	}

	p := m.setup.players[playerIdx]
	addKitLabel := ""
	addIssuedLabel := ""
	if len(allKit) > 0 {
		addKitLabel = string(allKit[addKitIdx])
		addIssuedLabel = string(allKit[addIssuedIdx])
	}

	return []playerConfigRow{
		{label: "Player", value: fmt.Sprintf("%d / %d", playerIdx+1, len(m.setup.players)), kind: playerRowPlayer, active: true},
		{label: "Name", value: p.Name, kind: playerRowName, active: true},
		{label: "Quick Name", value: names[nameIdx], kind: playerRowQuickName, active: true},
		{label: "Sex", value: string(p.Sex), kind: playerRowSex, active: true},
		{label: "Body Type", value: string(p.BodyType), kind: playerRowBodyType, active: true},
		{label: "Weight (kg)", value: fmt.Sprintf("%d", p.WeightKg), kind: playerRowWeightKg, active: true},
		{label: "Height (ft)", value: fmt.Sprintf("%d", p.HeightFt), kind: playerRowHeightFt, active: true},
		{label: "Height (in)", value: fmt.Sprintf("%d", p.HeightIn), kind: playerRowHeightIn, active: true},
		{label: "Endurance Modifier", value: fmt.Sprintf("%+d", p.Endurance), kind: playerRowEndurance, active: true},
		{label: "Bushcraft Modifier", value: fmt.Sprintf("%+d", p.Bushcraft), kind: playerRowBushcraft, active: true},
		{label: "Mental Modifier", value: fmt.Sprintf("%+d", p.Mental), kind: playerRowMental, active: true},
		{label: "Kit Limit", value: fmt.Sprintf("%d", p.KitLimit), kind: playerRowKitLimit, active: true},
		{label: "Personal Kit", value: kitSummary(p.Kit, p.KitLimit), kind: playerRowPersonalKit, active: true},
		{label: "Add Personal Item", value: addKitLabel, kind: playerRowAddPersonalKit, active: len(allKit) > 0 && len(p.Kit) < p.KitLimit},
		{label: "Remove Personal Item", value: "", kind: playerRowRemovePersonalKit, active: len(p.Kit) > 0},
		{label: "Issued Kit", value: kitSummary(m.setup.issuedKit, 0), kind: playerRowIssuedKit, active: true},
		{label: "Add Issued Item", value: addIssuedLabel, kind: playerRowAddIssuedKit, active: len(allKit) > 0},
		{label: "Remove Issued Item", value: "", kind: playerRowRemoveIssuedKit, active: len(m.setup.issuedKit) > 0},
		{label: "Reset Issued (Recommended)", value: "", kind: playerRowResetIssuedKit, active: true},
		{label: "Back To Wizard", value: "", kind: playerRowBack, active: true},
	}
}

func (m menuModel) playerConfigRowSupportsCycle(kind playerConfigRowKind) bool {
	switch kind {
	case playerRowPlayer, playerRowQuickName, playerRowSex, playerRowBodyType,
		playerRowWeightKg, playerRowHeightFt, playerRowHeightIn, playerRowEndurance,
		playerRowBushcraft, playerRowMental, playerRowKitLimit, playerRowAddPersonalKit,
		playerRowAddIssuedKit:
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
	case playerRowQuickName:
		m.pcfg.nameIdx = wrapIndex(m.pcfg.nameIdx, delta, len(playerConfigSuggestedNames()))
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
	case playerRowAddPersonalKit:
		m.pcfg.addKitIdx = wrapIndex(m.pcfg.addKitIdx, delta, len(game.AllKitItems()))
	case playerRowAddIssuedKit:
		m.pcfg.addIssuedIdx = wrapIndex(m.pcfg.addIssuedIdx, delta, len(game.AllKitItems()))
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

func (m menuModel) addPersonalKitItem() menuModel {
	if len(m.setup.players) == 0 {
		return m
	}
	m = m.normalizePlayerConfigState()
	items := game.AllKitItems()
	if len(items) == 0 {
		return m
	}
	p := &m.setup.players[m.pcfg.playerIdx]
	if len(p.Kit) >= p.KitLimit {
		m.status = fmt.Sprintf("Player kit limit reached (%d).", p.KitLimit)
		return m
	}
	item := items[m.pcfg.addKitIdx]
	if hasKitItem(p.Kit, item) {
		m.status = fmt.Sprintf("%s already selected for this player.", item)
		return m
	}
	p.Kit = append(p.Kit, item)
	m.status = fmt.Sprintf("Added %s to player kit.", item)
	return m
}

func (m menuModel) removePersonalKitItem() menuModel {
	if len(m.setup.players) == 0 {
		return m
	}
	m = m.normalizePlayerConfigState()
	p := &m.setup.players[m.pcfg.playerIdx]
	if len(p.Kit) == 0 {
		return m
	}

	items := game.AllKitItems()
	if len(items) > 0 {
		selected := items[m.pcfg.addKitIdx]
		if idx := indexOfKitItem(p.Kit, selected); idx >= 0 {
			p.Kit = removeKitItemAt(p.Kit, idx)
			m.status = fmt.Sprintf("Removed %s from player kit.", selected)
			return m
		}
	}

	removed := p.Kit[len(p.Kit)-1]
	p.Kit = p.Kit[:len(p.Kit)-1]
	m.status = fmt.Sprintf("Removed %s from player kit.", removed)
	return m
}

func (m menuModel) addIssuedKitItem() menuModel {
	m = m.normalizePlayerConfigState()
	items := game.AllKitItems()
	if len(items) == 0 {
		return m
	}
	item := items[m.pcfg.addIssuedIdx]
	if hasKitItem(m.setup.issuedKit, item) {
		m.status = fmt.Sprintf("%s is already in issued kit.", item)
		return m
	}
	m.setup.issuedKit = append(m.setup.issuedKit, item)
	m.setup.issuedCustom = true
	m.status = fmt.Sprintf("Added %s to issued kit.", item)
	return m
}

func (m menuModel) removeIssuedKitItem() menuModel {
	m = m.normalizePlayerConfigState()
	if len(m.setup.issuedKit) == 0 {
		return m
	}

	items := game.AllKitItems()
	if len(items) > 0 {
		selected := items[m.pcfg.addIssuedIdx]
		if idx := indexOfKitItem(m.setup.issuedKit, selected); idx >= 0 {
			m.setup.issuedKit = removeKitItemAt(m.setup.issuedKit, idx)
			m.setup.issuedCustom = true
			m.status = fmt.Sprintf("Removed %s from issued kit.", selected)
			return m
		}
	}

	removed := m.setup.issuedKit[len(m.setup.issuedKit)-1]
	m.setup.issuedKit = m.setup.issuedKit[:len(m.setup.issuedKit)-1]
	m.setup.issuedCustom = true
	m.status = fmt.Sprintf("Removed %s from issued kit.", removed)
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

	names := playerConfigSuggestedNames()
	if len(names) > 0 {
		if m.pcfg.nameIdx < 0 || m.pcfg.nameIdx >= len(names) {
			m.pcfg.nameIdx = 0
		}
	} else {
		m.pcfg.nameIdx = 0
	}

	items := game.AllKitItems()
	if len(items) > 0 {
		if m.pcfg.addKitIdx < 0 || m.pcfg.addKitIdx >= len(items) {
			m.pcfg.addKitIdx = 0
		}
		if m.pcfg.addIssuedIdx < 0 || m.pcfg.addIssuedIdx >= len(items) {
			m.pcfg.addIssuedIdx = 0
		}
	} else {
		m.pcfg.addKitIdx = 0
		m.pcfg.addIssuedIdx = 0
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

	data, err := readDataFile(path, maxSaveFileBytes)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			meta.Exists = false
			meta.Summary = "Empty"
			return meta
		}
		meta.Exists = true
		meta.Summary = "Unreadable save"
		meta.ErrDetail = err.Error()
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
		m.screen = screenMenu
		return m, nil
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
		case builderRowSave:
			return m.saveScenarioFromBuilder()
		case builderRowDelete:
			return m.deleteScenarioFromBuilder()
		case builderRowCancel:
			m.screen = screenMenu
			return m, nil
		case builderRowAddPhase:
			if len(m.build.phases) >= maxBuilderPhases {
				m.status = fmt.Sprintf("Maximum phases reached (%d).", maxBuilderPhases)
				return m, nil
			}
			m.build.phases = append(m.build.phases, phaseBuilderPhase{
				seasonIdx: m.build.phases[len(m.build.phases)-1].seasonIdx,
				days:      "0",
			})
			return m, nil
		case builderRowRemovePhase:
			if len(m.build.phases) <= 1 {
				m.status = "At least one phase is required."
				return m, nil
			}
			m.build.phases = m.build.phases[:len(m.build.phases)-1]
			return m, nil
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

	seasonOptions := builderSeasonOptions()
	scenarioChoices := builderScenarioChoices(m.customScenarios)
	switch row.kind {
	case builderRowScenario:
		next := wrapIndex(m.build.selectedIdx, delta, len(scenarioChoices))
		m = m.loadScenarioBuilderSelection(next)
	case builderRowMode:
		m.build.modeIdx = wrapIndex(m.build.modeIdx, delta, len(setupModes()))
	case builderRowBiome:
		m.build.biomeIdx = wrapIndex(m.build.biomeIdx, delta, len(builderBiomes()))
	case builderRowDaysMode:
		m.build.useCustomDays = !m.build.useCustomDays
	case builderRowDaysPreset:
		m.build.defaultDaysIdx = wrapIndex(m.build.defaultDaysIdx, delta, len(builderDefaultDays()))
	case builderRowPhaseSeason:
		idx := row.phaseIdx
		if idx >= 0 && idx < len(m.build.phases) {
			m.build.phases[idx].seasonIdx = wrapIndex(m.build.phases[idx].seasonIdx, delta, len(seasonOptions))
		}
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

	scenarioID := game.ScenarioID("")
	if m.build.selectedIdx == 0 {
		scenarioID = makeCustomScenarioID(name, m.availableScenarios())
	} else {
		scenarioID = m.customScenarios[m.build.selectedIdx-1].Scenario.ID
	}

	scenario := game.Scenario{
		ID:          scenarioID,
		Name:        name,
		Biome:       builderBiomes()[m.build.biomeIdx],
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
	if m.build.selectedIdx == 0 {
		updated = append(updated, record)
		m.build.selectedIdx = len(updated)
	} else {
		updated[m.build.selectedIdx-1] = record
	}

	if err := saveCustomScenarios(defaultCustomScenariosFile, updated); err != nil {
		m.status = fmt.Sprintf("Failed to save scenario: %v", err)
		return m, nil
	}

	m.customScenarios = updated
	m.status = fmt.Sprintf("Saved scenario: %s", scenario.Name)
	m = m.loadScenarioBuilderSelection(m.build.selectedIdx)
	return m, nil
}

func (m menuModel) deleteScenarioFromBuilder() (tea.Model, tea.Cmd) {
	if m.build.selectedIdx == 0 {
		m.status = "Select an existing scenario to delete."
		return m, nil
	}

	idx := m.build.selectedIdx - 1
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
	case builderRowScenario, builderRowMode, builderRowBiome, builderRowDaysMode, builderRowDaysPreset, builderRowPhaseSeason:
		return true
	default:
		return false
	}
}

func (m menuModel) scenarioBuilderRows() []scenarioBuilderRow {
	scenarioChoices := builderScenarioChoices(m.customScenarios)
	selectedScenarioIdx := m.build.selectedIdx
	if selectedScenarioIdx < 0 || selectedScenarioIdx >= len(scenarioChoices) {
		selectedScenarioIdx = 0
	}

	rows := []scenarioBuilderRow{
		{label: "Scenario", value: scenarioChoices[selectedScenarioIdx], kind: builderRowScenario, active: true},
		{label: "Name", value: m.build.name, kind: builderRowName, active: true},
		{label: "Game Mode", value: modeLabel(setupModes()[m.build.modeIdx]), kind: builderRowMode, active: true},
		{label: "Biome", value: builderBiomes()[m.build.biomeIdx], kind: builderRowBiome, active: true},
		{label: "Days Mode", value: map[bool]string{true: "Custom", false: "Preset"}[m.build.useCustomDays], kind: builderRowDaysMode, active: true},
		{label: "Days Preset", value: fmt.Sprintf("%d", builderDefaultDays()[m.build.defaultDaysIdx]), kind: builderRowDaysPreset, active: !m.build.useCustomDays},
		{label: "Days Custom", value: m.build.customDays, kind: builderRowDaysCustom, active: m.build.useCustomDays},
		{label: "Season Profile ID", value: m.build.seasonSetID, kind: builderRowSeasonProfileID, active: true},
	}

	seasonOptions := builderSeasonOptions()
	for i := range m.build.phases {
		rows = append(rows,
			scenarioBuilderRow{
				label:    fmt.Sprintf("Phase %d Season", i+1),
				value:    builderSeasonLabel(seasonOptions[m.build.phases[i].seasonIdx]),
				kind:     builderRowPhaseSeason,
				phaseIdx: i,
				active:   true,
			},
			scenarioBuilderRow{
				label:    fmt.Sprintf("Phase %d Days", i+1),
				value:    m.build.phases[i].days,
				kind:     builderRowPhaseDays,
				phaseIdx: i,
				active:   true,
			},
		)
	}

	rows = append(rows,
		scenarioBuilderRow{label: "Add Phase", value: fmt.Sprintf("%d/%d", len(m.build.phases), maxBuilderPhases), kind: builderRowAddPhase, active: len(m.build.phases) < maxBuilderPhases},
		scenarioBuilderRow{label: "Remove Last Phase", value: "", kind: builderRowRemovePhase, active: len(m.build.phases) > 1},
		scenarioBuilderRow{label: "Save Scenario", value: "", kind: builderRowSave, active: true},
		scenarioBuilderRow{label: "Delete Scenario", value: "", kind: builderRowDelete, active: m.build.selectedIdx > 0},
		scenarioBuilderRow{label: "Cancel", value: "", kind: builderRowCancel, active: true},
	)

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
	case builderRowPhaseDays:
		if row.phaseIdx >= 0 && row.phaseIdx < len(m.build.phases) {
			for _, r := range runes {
				if r >= '0' && r <= '9' {
					m.build.phases[row.phaseIdx].days += string(r)
				}
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
	case builderRowDaysCustom:
		m.build.customDays = backspace(m.build.customDays)
	case builderRowSeasonProfileID:
		m.build.seasonSetID = backspace(m.build.seasonSetID)
	case builderRowPhaseDays:
		if row.phaseIdx >= 0 && row.phaseIdx < len(m.build.phases) {
			m.build.phases[row.phaseIdx].days = backspace(m.build.phases[row.phaseIdx].days)
		}
	}

	return m
}

func (m menuModel) loadScenarioBuilderSelection(selected int) menuModel {
	cursor := m.build.cursor
	loaded := newScenarioBuilderState()
	loaded.cursor = cursor
	loaded.selectedIdx = selected

	if selected == 0 {
		m.build = loaded
		return m
	}

	if selected-1 < 0 || selected-1 >= len(m.customScenarios) {
		m.build = loaded
		return m
	}

	record := m.customScenarios[selected-1]
	loaded.name = record.Scenario.Name
	loaded.modeIdx = selectedModeIndex(record.PreferredMode)
	loaded.biomeIdx = selectedBiomeIndex(record.Scenario.Biome)

	if idx := indexOfInt(builderDefaultDays(), record.Scenario.DefaultDays); idx >= 0 {
		loaded.useCustomDays = false
		loaded.defaultDaysIdx = idx
		loaded.customDays = fmt.Sprintf("%d", record.Scenario.DefaultDays)
	} else {
		loaded.useCustomDays = true
		loaded.customDays = fmt.Sprintf("%d", record.Scenario.DefaultDays)
	}

	loaded.seasonSetID = sanitizeSeasonSetID(string(record.Scenario.DefaultSeasonSetID))
	if loaded.seasonSetID == "" {
		loaded.seasonSetID = "custom_profile"
	}

	set := game.SeasonSet{}
	found := false
	for _, seasonSet := range record.Scenario.SeasonSets {
		if seasonSet.ID == record.Scenario.DefaultSeasonSetID {
			set = seasonSet
			found = true
			break
		}
	}
	if !found && len(record.Scenario.SeasonSets) > 0 {
		set = record.Scenario.SeasonSets[0]
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

func (m menuModel) viewScenarioBuilder() string {
	rows := m.scenarioBuilderRows()

	var b strings.Builder
	b.WriteString(brightGreen.Render("SCENARIO BUILDER / EDITOR") + "\n")
	b.WriteString(dimGreen.Render("Create or edit scenarios with dynamic season phases") + "\n")
	b.WriteString(border.Render("----------------------------------------") + "\n\n")

	if m.build.cursor < 0 || m.build.cursor >= len(rows) {
		m.build.cursor = 0
	}

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
			b.WriteString(cursor + lineStyle.Render(row.label) + "\n")
			continue
		}

		value := row.value
		if row.kind == builderRowName && strings.TrimSpace(value) == "" {
			value = "<type name>"
		}
		if row.kind == builderRowDaysCustom && strings.TrimSpace(value) == "" {
			value = "<type days>"
		}
		if row.kind == builderRowSeasonProfileID && strings.TrimSpace(value) == "" {
			value = "<profile_id>"
		}
		if row.kind == builderRowPhaseDays && strings.TrimSpace(value) == "" {
			value = "0"
		}
		b.WriteString(cursor + lineStyle.Render(fmt.Sprintf("%-18s %s", row.label+":", value)) + "\n")
	}

	b.WriteString("\n" + border.Render("----------------------------------------") + "\n")
	b.WriteString(dimGreen.Render("↑/↓ move  ←/→ change  Enter select  type in text fields  Shift+Q back") + "\n")
	b.WriteString(dimGreen.Render("Phase days: 0 means until end (only valid on the last phase)") + "\n")
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

	for slot := 1; slot <= saveSlotCount; slot++ {
		if clean == savePathForSlot(slot) {
			return nil
		}
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

	return game.RunState{
		Config:      resolvedConfig,
		Scenario:    scenario,
		SeasonSetID: scenario.DefaultSeasonSetID,
		Day:         1,
		Players:     game.CreatePlayers(resolvedConfig),
	}, nil
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
