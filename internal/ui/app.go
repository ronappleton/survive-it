package ui

import (
	"fmt"
	"time"

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
	itemQuit
)

type menuModel struct {
	cfg AppConfig
	idx int

	status string
	busy   bool
}

func newMenuModel(cfg AppConfig) menuModel {
	return menuModel{cfg: cfg, idx: 0}
}

func (m menuModel) Init() tea.Cmd {
	// Alpha1: keep update check opt-in via menu.
	return nil
}

type updateResultMsg struct {
	status string
	err    error
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.busy {
			// Ignore input while update check runs.
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			m.idx = (m.idx + 2) % 3
			return m, nil
		case "down", "j":
			m.idx = (m.idx + 1) % 3
			return m, nil
		case "enter":
			switch menuItem(m.idx) {
			case itemStart:
				m.status = "Game loop not implemented yet (alpha1)."
				return m, nil
			case itemCheckUpdate:
				if m.cfg.NoUpdate {
					m.status = "Update checks disabled (--no-update)."
					return m, nil
				}
				m.busy = true
				m.status = "Checking for updates…"
				return m, checkUpdateCmd(m.cfg.Version)
			case itemQuit:
				return m, tea.Quit
			}
		}
	case updateResultMsg:
		m.busy = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Update check failed: %v", msg.err)
			return m, nil
		}
		m.status = msg.status
		return m, nil
	}

	return m, nil
}

func (m menuModel) View() string {
	title := brightGreen.Render("SURVIVE IT") + dimGreen.Render("  alpha1")
	ver := dimGreen.Render(fmt.Sprintf("v%s  (%s)  %s", m.cfg.Version, m.cfg.Commit, m.cfg.BuildDate))

	items := []string{
		"Start (coming soon)",
		"Check for updates",
		"Quit",
	}

	out := ""
	out += title + "\n" + ver + "\n"
	out += border.Render("----------------------------------------") + "\n\n"

	for i, it := range items {
		cursor := "  "
		line := it
		if i == m.idx {
			cursor = "> "
			line = brightGreen.Render(it)
		} else {
			line = green.Render(it)
		}
		out += cursor + line + "\n"
	}

	out += "\n" + border.Render("----------------------------------------") + "\n"
	out += dimGreen.Render("↑/↓ to move, Enter to select, q to quit") + "\n"
	if m.status != "" {
		out += "\n" + green.Render(m.status) + "\n"
	}
	return out
}

func checkUpdateCmd(currentVersion string) tea.Cmd {
	return func() tea.Msg {
		// Tiny delay so the UI visibly switches to busy state.
		time.Sleep(150 * time.Millisecond)

		res, err := update.Apply(currentVersion)
		if err != nil {
			return updateResultMsg{err: err}
		}
		return updateResultMsg{status: res}
	}
}
