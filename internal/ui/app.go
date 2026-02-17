package ui

import (
	"encoding/json"
	"fmt"
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
	itemSaveRun
	itemCheckUpdate
	itemInstallUpdate
	itemQuit
)

const defaultSaveFile = "survive-it-save.json"

type menuModel struct {
	w, h   int
	cfg    AppConfig
	idx    int
	screen screen
	setup  setupState

	run    *game.RunState
	status string
	busy   bool
	err    string
}

func newMenuModel(cfg AppConfig) menuModel {
	return menuModel{
		cfg:   cfg,
		idx:   0,
		setup: newSetupState(),
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

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.screen == screenRun {
			return m.updateRun(msg)
		}
		if m.screen == screenSetup {
			return m.updateSetup(msg)
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
	if m.screen == screenRun {
		return m.viewRun()
	}
	return m.viewMenu()
}

func (m menuModel) updateRun(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.screen = screenMenu
		return m, nil

	case "enter", "n":
		m.run.AdvanceDay()
		out := m.run.EvaluateRun()
		if out.Status == game.RunOutcomeCritical {
			m.status = fmt.Sprintf("CRITICAL: %v", out.CriticalPlayerIDs)
		} else if out.Status == game.RunOutcomeCompleted {
			m.status = "COMPLETED"
		} else {
			m.status = ""
		}
		return m, nil
	case "s":
		if m.run == nil {
			m.status = "No active run to save."
			return m, nil
		}
		if err := saveRunToFile(defaultSaveFile, *m.run); err != nil {
			m.status = fmt.Sprintf("Save failed: %v", err)
			return m, nil
		}
		m.status = fmt.Sprintf("Saved run to %s", defaultSaveFile)
		return m, nil
	case "l":
		state, err := loadRunFromFile(defaultSaveFile)
		if err != nil {
			m.status = fmt.Sprintf("Load failed: %v", err)
			return m, nil
		}
		m.run = &state
		m.status = fmt.Sprintf("Loaded run from %s", defaultSaveFile)
		return m, nil
	}
	return m, nil
}

func menuItems() []string {
	return []string{
		"Start",
		"Load Run",
		"Save Run",
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
	case "ctrl+c", "q":
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
			state, err := loadRunFromFile(defaultSaveFile)
			if err != nil {
				m.status = fmt.Sprintf("Load failed: %v", err)
				return m, nil
			}
			m.run = &state
			m.screen = screenRun
			m.status = fmt.Sprintf("Loaded run from %s", defaultSaveFile)
			return m, nil
		case itemSaveRun:
			if m.run == nil {
				m.status = "No active run to save."
				return m, nil
			}
			if err := saveRunToFile(defaultSaveFile, *m.run); err != nil {
				m.status = fmt.Sprintf("Save failed: %v", err)
				return m, nil
			}
			m.status = fmt.Sprintf("Saved run to %s", defaultSaveFile)
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

func setupModes() []game.GameMode {
	return []game.GameMode{
		game.ModeAlone,
		game.ModeNakedAndAfraid,
	}
}

func setupScenarioOptions() []setupScenarioOption {
	options := []setupScenarioOption{
		{id: game.ScenarioRandomID, label: "Random"},
	}

	for _, scenario := range game.BuiltInScenarios() {
		options = append(options, setupScenarioOption{
			id:    scenario.ID,
			label: scenario.Name,
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
	case "q", "esc":
		m.screen = screenMenu
		return m, nil
	case "up", "k":
		m.setup.cursor = wrapIndex(m.setup.cursor, -1, rowCount)
		return m, nil
	case "down", "j":
		m.setup.cursor = wrapIndex(m.setup.cursor, 1, rowCount)
		return m, nil
	case "left", "h":
		m = m.adjustSetupChoice(-1)
		return m, nil
	case "right", "l":
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
	switch m.setup.cursor {
	case 0:
		m.setup.modeIdx = wrapIndex(m.setup.modeIdx, delta, len(setupModes()))
	case 1:
		m.setup.scenarioIdx = wrapIndex(m.setup.scenarioIdx, delta, len(setupScenarioOptions()))
	case 2:
		m.setup.playerCountIdx = wrapIndex(m.setup.playerCountIdx, delta, len(setupPlayerCounts()))
	case 3:
		m.setup.runLengthIdx = wrapIndex(m.setup.runLengthIdx, delta, len(setupRunLengths()))
	}

	return m
}

func (m menuModel) startRunFromSetup() (tea.Model, tea.Cmd) {
	modes := setupModes()
	scenarios := setupScenarioOptions()
	playerCounts := setupPlayerCounts()
	runLengths := setupRunLengths()

	runLength := runLengths[m.setup.runLengthIdx]
	cfg := game.RunConfig{
		Mode:        modes[m.setup.modeIdx],
		ScenarioID:  scenarios[m.setup.scenarioIdx].id,
		PlayerCount: playerCounts[m.setup.playerCountIdx],
		RunLength: game.RunLength{
			OpenEnded: runLength.openEnded,
			Days:      runLength.days,
		},
		Seed:    0,
		Players: nil,
	}

	state, err := game.NewRunState(cfg)
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
	scenarios := setupScenarioOptions()
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
	b.WriteString(dimGreen.Render("↑/↓ move  ←/→ change  Enter select  esc back") + "\n")
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
	footerRows := 4
	bodyRows := totalHeight - headerRows - footerRows
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
	footer := paneStyle.Copy().
		Foreground(lipgloss.Color("#FAFAFA")).
		Height(contentHeight(footerRows)).
		Render(m.footerText())

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
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
	b.WriteString(dimGreen.Render("↑/↓ to move, Enter to select, q to quit") + "\n")
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

	b.WriteString(brightGreen.Render("[enter] next day  [s] save  [l] load  [q/esc] back"))
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

func loadRunFromFile(path string) (game.RunState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return game.RunState{}, err
	}

	var payload savedRun
	if err := json.Unmarshal(data, &payload); err != nil {
		return game.RunState{}, err
	}

	if err := payload.Run.Config.Validate(); err != nil {
		return game.RunState{}, fmt.Errorf("invalid saved run config: %w", err)
	}

	return payload.Run, nil
}

func resizeTerminalBestEffort(cols, rows int) tea.Cmd {
	return func() tea.Msg {
		fmt.Printf("\x1b[8;%d;%dt", rows, cols) // CSI 8; rows; cols t
		return nil
	}
}
