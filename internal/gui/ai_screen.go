package gui

import (
	"context"
	"fmt"
	"strings"

	"github.com/appengine-ltd/survive-it/internal/ai"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type aiSettingsState struct {
	Cursor int

	Enabled    bool
	Downloaded bool

	Downloading     bool
	DownloadedBytes int64
	DownloadTotal   int64

	Status     string
	ModelPath  string
	ConfigPath string

	downloadProgressCh chan ai.Progress
	downloadDoneCh     chan error
}

type aiRowAction int

const (
	aiRowToggle aiRowAction = iota
	aiRowDownload
	aiRowDelete
	aiRowBack
)

type aiSettingsRow struct {
	Label  string
	Value  string
	Action aiRowAction
}

func (ui *gameUI) openAISettings() {
	cfg, cfgErr := ai.LoadConfig()
	configPath, configPathErr := ai.ConfigPath()
	modelPath, modelPathErr := ai.ModelPath()
	downloaded, modelErr := ai.ModelExists()

	ui.ai = aiSettingsState{
		Cursor:     0,
		Enabled:    cfg.AIEnabled,
		Downloaded: downloaded,
		ModelPath:  modelPath,
		ConfigPath: configPath,
	}

	if configPathErr != nil {
		ui.ai.ConfigPath = "(unavailable)"
	}
	if modelPathErr != nil {
		ui.ai.ModelPath = "(unavailable)"
	}

	switch {
	case cfgErr != nil:
		ui.ai.Status = "Config load failed: " + cfgErr.Error()
	case configPathErr != nil:
		ui.ai.Status = "Config path unavailable: " + configPathErr.Error()
	case modelPathErr != nil:
		ui.ai.Status = "Model path unavailable: " + modelPathErr.Error()
	case modelErr != nil:
		ui.ai.Status = "Model status check failed: " + modelErr.Error()
	}

	ui.screen = screenAISettings
}

func (ui *gameUI) updateAISettings() {
	ui.pollAIDownloadEvents()
	if !ui.ai.Downloading {
		ui.refreshAIModelStatus()
	}

	rows := ui.aiRows()
	if len(rows) == 0 {
		return
	}
	ui.ai.Cursor = clampInt(ui.ai.Cursor, 0, len(rows)-1)

	if rl.IsKeyPressed(rl.KeyEscape) {
		ui.enterMenu()
		return
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		ui.ai.Cursor = wrapIndex(ui.ai.Cursor+1, len(rows))
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		ui.ai.Cursor = wrapIndex(ui.ai.Cursor-1, len(rows))
	}
	if rl.IsKeyPressed(rl.KeyLeft) || rl.IsKeyPressed(rl.KeyRight) {
		if rows[ui.ai.Cursor].Action == aiRowToggle {
			ui.toggleAIEnabled()
		}
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		switch rows[ui.ai.Cursor].Action {
		case aiRowToggle:
			ui.toggleAIEnabled()
		case aiRowDownload:
			ui.startAIDownload()
		case aiRowDelete:
			ui.deleteAIModel()
		case aiRowBack:
			ui.enterMenu()
		}
	}
}

func (ui *gameUI) aiRows() []aiSettingsRow {
	rows := []aiSettingsRow{
		{
			Label:  "AI Enabled",
			Value:  map[bool]string{true: "On", false: "Off"}[ui.ai.Enabled],
			Action: aiRowToggle,
		},
	}
	if !ui.ai.Downloading {
		if ui.ai.Downloaded {
			rows = append(rows, aiSettingsRow{Label: "Delete", Value: "Enter", Action: aiRowDelete})
		} else {
			rows = append(rows, aiSettingsRow{Label: "Download", Value: "Enter", Action: aiRowDownload})
		}
	}
	rows = append(rows, aiSettingsRow{Label: "Back", Value: "Enter", Action: aiRowBack})
	return rows
}

func (ui *gameUI) toggleAIEnabled() {
	next := !ui.ai.Enabled
	if err := ai.SaveConfig(ai.Config{AIEnabled: next}); err != nil {
		ui.ai.Status = "Failed to save config: " + err.Error()
		return
	}
	ui.ai.Enabled = next
	ui.ai.Status = "AI enabled setting saved."
}

func (ui *gameUI) deleteAIModel() {
	if err := ai.DeleteModel(); err != nil {
		ui.ai.Status = "Delete failed: " + err.Error()
		return
	}
	ui.refreshAIModelStatus()
	ui.ai.Status = "Model file deleted."
}

func (ui *gameUI) startAIDownload() {
	if ui.ai.Downloading {
		return
	}
	ui.ai.Downloading = true
	ui.ai.DownloadedBytes = 0
	ui.ai.DownloadTotal = 0
	ui.ai.Status = "Downloading model..."

	progressCh := make(chan ai.Progress, 16)
	doneCh := make(chan error, 1)
	ui.ai.downloadProgressCh = progressCh
	ui.ai.downloadDoneCh = doneCh

	go func(progressCh chan<- ai.Progress, doneCh chan<- error) {
		err := ai.DownloadModel(context.Background(), func(p ai.Progress) {
			select {
			case progressCh <- p:
			default:
			}
		})
		doneCh <- err
	}(progressCh, doneCh)
}

func (ui *gameUI) pollAIDownloadEvents() {
	if ui.ai.downloadProgressCh != nil {
		for {
			select {
			case p := <-ui.ai.downloadProgressCh:
				ui.ai.DownloadedBytes = p.DownloadedBytes
				ui.ai.DownloadTotal = p.TotalBytes
			default:
				goto doneProgress
			}
		}
	}

doneProgress:
	if ui.ai.downloadDoneCh == nil {
		return
	}
	select {
	case err := <-ui.ai.downloadDoneCh:
		ui.ai.Downloading = false
		ui.ai.downloadProgressCh = nil
		ui.ai.downloadDoneCh = nil
		ui.refreshAIModelStatus()
		if err != nil {
			ui.ai.Status = "Download failed: " + err.Error()
			return
		}
		ui.ai.Status = "Model downloaded."
	default:
	}
}

func (ui *gameUI) refreshAIModelStatus() {
	downloaded, err := ai.ModelExists()
	if err != nil {
		ui.ai.Status = "Model status check failed: " + err.Error()
		ui.ai.Downloaded = false
		return
	}
	ui.ai.Downloaded = downloaded
}

func (ui *gameUI) drawAISettings() {
	left := rl.NewRectangle(20, 20, float32(ui.width)*0.47, float32(ui.height-40))
	right := rl.NewRectangle(left.X+left.Width+16, 20, float32(ui.width)-left.Width-56, float32(ui.height-40))
	drawPanel(left, "AI Settings")
	drawPanel(right, "Model")

	rows := ui.aiRows()
	ui.ai.Cursor = clampInt(ui.ai.Cursor, 0, maxInt(0, len(rows)-1))
	for i, row := range rows {
		y := int32(left.Y) + 62 + int32(i*56)
		if i == ui.ai.Cursor {
			rl.DrawRectangle(int32(left.X)+16, y-8, int32(left.Width)-32, 42, rl.Fade(colorAccent, 0.2))
		}
		drawText(row.Label, int32(left.X)+26, y, 24, colorText)
		drawText(row.Value, int32(left.X)+286, y, 24, colorAccent)
	}
	drawText("Left/Right or Enter to toggle, Esc back", int32(left.X)+22, int32(left.Y+left.Height)-38, 18, colorDim)

	modelIndicator := "☐"
	if ui.ai.Downloaded {
		modelIndicator = "✓"
	}
	drawText(fmt.Sprintf("Model: %s ai-pack-v1.gguf", modelIndicator), int32(right.X)+24, int32(right.Y)+56, 24, colorText)

	progressColor := colorDim
	progressText := "Idle"
	if ui.ai.Downloading {
		progressText = formatAIDownloadProgress(ui.ai.DownloadedBytes, ui.ai.DownloadTotal)
		progressColor = colorAccent
	}
	drawText(progressText, int32(right.X)+24, int32(right.Y)+96, 20, progressColor)

	drawWrappedText("Model Path: "+ui.ai.ModelPath, right, 136, 16, colorDim)
	drawWrappedText("Config Path: "+ui.ai.ConfigPath, right, 178, 16, colorDim)

	if strings.TrimSpace(ui.ai.Status) != "" {
		statusColor := colorAccent
		statusLower := strings.ToLower(ui.ai.Status)
		if strings.Contains(statusLower, "failed") || strings.Contains(statusLower, "error") {
			statusColor = colorWarn
		}
		drawWrappedText(ui.ai.Status, right, 234, 19, statusColor)
	}
}

func formatAIDownloadProgress(downloaded, total int64) string {
	if total > 0 {
		percent := float64(downloaded) / float64(total) * 100
		if percent < 0 {
			percent = 0
		}
		if percent > 100 {
			percent = 100
		}
		return fmt.Sprintf("Downloading: %.0f%% (%s / %s)", percent, formatByteCount(downloaded), formatByteCount(total))
	}
	return "Downloading: " + formatByteCount(downloaded)
}

func formatByteCount(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	units := []string{"KB", "MB", "GB", "TB"}
	value := float64(bytes)
	unit := units[0]
	for i := 0; i < len(units); i++ {
		value /= 1024
		unit = units[i]
		if value < 1024 || i == len(units)-1 {
			break
		}
	}
	if value >= 100 {
		return fmt.Sprintf("%.0f %s", value, unit)
	}
	if value >= 10 {
		return fmt.Sprintf("%.1f %s", value, unit)
	}
	return fmt.Sprintf("%.2f %s", value, unit)
}
