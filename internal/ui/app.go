package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/appengine-ltd/survive-it/internal/game"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/appengine-ltd/survive-it/internal/update"
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
	itemCheckUpdate
	itemInstallUpdate
	itemQuit
)

type menuModel struct {
	w, h   int
	cfg    AppConfig
	idx    int
	screen screen

	run    *game.RunState
	status string
	busy   bool
	err    string
}

func newMenuModel(cfg AppConfig) menuModel {
	return menuModel{cfg: cfg, idx: 0}
}

func (m menuModel) Init() tea.Cmd {
	return resizeTerminalBestEffort(120, 35)
}

type updateResultMsg struct {
	status string
	err    error
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.screen == screenRun {
			return m.updateRun(msg)
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
	}
	return m, nil
}

func (m menuModel) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.busy {
		// Ignore input while update check runs.
		return m, nil
	}
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		m.idx = (m.idx + 2) % 4
		return m, nil
	case "down", "j":
		m.idx = (m.idx + 1) % 4
		return m, nil
	case "enter":
		switch menuItem(m.idx) {
		case itemStart:
			cfg := game.RunConfig{
				Mode:        game.ModeAlone,          // or ModeNakedAndAfraid later via setup
				ScenarioID:  game.ScenarioRandomID,   // quick-start
				PlayerCount: 2,                       // quick-start
				RunLength:   game.RunLength{Days: 7}, // or OpenEnded: true
				Seed:        0,                       // auto-gen
				Players:     nil,                     // auto-gen names
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

func (m menuModel) viewRun() string {
	w := m.w
	h := m.h
	if w <= 0 {
		w = 120
	}
	if h <= 0 {
		h = 35
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1).
		Width(w - 1) // leave 1 column margin to avoid right-edge scroll

	// Render header first (auto height)
	header := box.Render(m.headerText())

	// Render footer first (auto height)
	footer := box.Render(m.footerText())

	usedHeight := lipgloss.Height(header) + lipgloss.Height(footer)

	bodyHeight := h - usedHeight
	if bodyHeight < 3 {
		bodyHeight = 3
	}

	body := box.Height(bodyHeight).Render(m.bodyText())

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m menuModel) viewMenu() string {
	title := brightGreen.Render("SURVIVE IT") + dimGreen.Render("  alpha")
	ver := dimGreen.Render(fmt.Sprintf("v%s  (%s)  %s", m.cfg.Version, m.cfg.Commit, m.cfg.BuildDate))

	items := []string{
		"Start (quick run)",
		"Check for updates",
		"Install Update",
		"Quit",
	}

	var b strings.Builder
	b.WriteString(title + "\n" + ver + "\n")
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
	b.WriteString(brightGreen.Render("SURVIVE IT") + dimGreen.Render("  run\n"))
	b.WriteString(dimGreen.Render(fmt.Sprintf("Day %d  |  %s  |  Season: %s",
		m.run.Day, m.run.Scenario.Name, seasonStr,
	)))
	return b.String()
}

func (m menuModel) bodyText() string {
	var b strings.Builder
	b.WriteString(green.Render("Players:\n"))
	for _, p := range m.run.Players {
		b.WriteString(fmt.Sprintf(" - %s [%s/%s] E:%d H:%d M:%d\n",
			p.Name, p.Sex, p.BodyType, p.Energy, p.Hydration, p.Morale,
		))
	}
	return b.String()
}

func (m menuModel) footerText() string {
	var b strings.Builder

	out := m.run.EvaluateRun()
	if out.Status != game.RunOutcomeOngoing {
		b.WriteString(brightGreen.Render(string(out.Status)) + " ")
	}

	if m.status != "" {
		b.WriteString(green.Render(m.status))
	}

	if out.Status != game.RunOutcomeOngoing || m.status != "" {
		b.WriteString("\n")
	}

	b.WriteString(dimGreen.Render("Enter/n: next day   q/esc: back"))
	return b.String()
}

func resizeTerminalBestEffort(cols, rows int) tea.Cmd {
	return func() tea.Msg {
		fmt.Printf("\x1b[8;%d;%dt", rows, cols) // CSI 8; rows; cols t
		return nil
	}
}
