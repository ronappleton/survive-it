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

	Downloaded     bool
	DownloadedByID map[string]bool

	ActiveModelID string
	ModelID       string
	ModelIndex    int
	ModelPacks    []ai.ModelPack

	Downloading     bool
	DownloadedBytes int64
	DownloadTotal   int64

	Status string

	downloadProgressCh chan ai.Progress
	downloadDoneCh     chan error
}

type aiRowAction int

const (
	aiRowActiveModel aiRowAction = iota
	aiRowModel
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
	packs := ai.AvailableModelPacks()
	cfg, cfgErr := ai.LoadConfig()

	modelID := ai.NormalizeModelID(cfg.ModelID)
	modelIndex := modelPackIndexByID(packs, modelID)
	if modelIndex < 0 {
		modelIndex = 0
		if len(packs) > 0 {
			modelID = packs[0].ID
		}
	}

	downloaded, modelErr := ai.ModelExists(modelID)

	ui.ai = aiSettingsState{
		Cursor:         0,
		Downloaded:     downloaded,
		DownloadedByID: map[string]bool{},
		ActiveModelID:  ai.NormalizeModelID(cfg.ModelID),
		ModelID:        modelID,
		ModelIndex:     modelIndex,
		ModelPacks:     packs,
	}
	ui.refreshAIModelStatus()

	if len(packs) == 0 {
		ui.ai.Status = "No AI models are set up yet."
	}

	switch {
	case cfgErr != nil:
		ui.ai.Status = "Could not load AI settings: " + cfgErr.Error()
	case modelErr != nil:
		ui.ai.Status = "Could not check model status: " + modelErr.Error()
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
		delta := -1
		if rl.IsKeyPressed(rl.KeyRight) {
			delta = 1
		}
		switch rows[ui.ai.Cursor].Action {
		case aiRowActiveModel:
			ui.adjustActiveAIModel(delta)
		case aiRowModel:
			ui.adjustAIModel(delta)
		}
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		switch rows[ui.ai.Cursor].Action {
		case aiRowActiveModel:
			ui.adjustActiveAIModel(1)
		case aiRowModel:
			ui.adjustAIModel(1)
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
	activeLabel := "None (download a model)"
	if pack, ok := ai.ModelPackByID(ui.ai.ActiveModelID); ok && ui.ai.DownloadedByID[pack.ID] {
		activeLabel = pack.Name
	}

	modelLabel := "(none)"
	if pack, ok := ui.currentAIModelPack(); ok {
		modelLabel = pack.Name
	}

	rows := []aiSettingsRow{
		{Label: "AI Model In Use", Value: truncateForUI(activeLabel, 24), Action: aiRowActiveModel},
		{Label: "Model Browser", Value: truncateForUI(modelLabel, 24), Action: aiRowModel},
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

func (ui *gameUI) currentAIModelPack() (ai.ModelPack, bool) {
	if len(ui.ai.ModelPacks) == 0 {
		return ai.ModelPack{}, false
	}
	if ui.ai.ModelIndex >= 0 && ui.ai.ModelIndex < len(ui.ai.ModelPacks) {
		return ui.ai.ModelPacks[ui.ai.ModelIndex], true
	}
	pack, ok := ai.ModelPackByID(ui.ai.ModelID)
	if ok {
		return pack, true
	}
	return ui.ai.ModelPacks[0], true
}

func (ui *gameUI) downloadedAIModelPacks() []ai.ModelPack {
	if len(ui.ai.ModelPacks) == 0 || len(ui.ai.DownloadedByID) == 0 {
		return nil
	}
	out := make([]ai.ModelPack, 0, len(ui.ai.ModelPacks))
	for _, pack := range ui.ai.ModelPacks {
		if ui.ai.DownloadedByID[pack.ID] {
			out = append(out, pack)
		}
	}
	return out
}

func (ui *gameUI) adjustActiveAIModel(delta int) {
	choices := ui.downloadedAIModelPacks()
	if len(choices) == 0 {
		ui.ai.Status = "No downloaded model available yet."
		return
	}
	index := 0
	for i, pack := range choices {
		if pack.ID == ui.ai.ActiveModelID {
			index = i
			break
		}
	}
	index = wrapIndex(index+delta, len(choices))
	ui.ai.ActiveModelID = choices[index].ID
	if err := ai.SaveConfig(ai.Config{AIEnabled: true, ModelID: ui.ai.ActiveModelID}); err != nil {
		ui.ai.Status = "Could not save active AI model: " + err.Error()
		return
	}
	ui.ai.Status = "Active AI model updated."
}

func (ui *gameUI) adjustAIModel(delta int) {
	if ui.ai.Downloading || len(ui.ai.ModelPacks) == 0 {
		return
	}
	ui.ai.ModelIndex = wrapIndex(ui.ai.ModelIndex+delta, len(ui.ai.ModelPacks))
	ui.ai.ModelID = ui.ai.ModelPacks[ui.ai.ModelIndex].ID
	ui.refreshAIModelStatus()
	ui.ai.Status = "Browsing model options."
}

func (ui *gameUI) deleteAIModel() {
	if err := ai.DeleteModel(ui.ai.ModelID); err != nil {
		ui.ai.Status = "Could not delete model: " + err.Error()
		return
	}
	ui.refreshAIModelStatus()
	ui.ai.Status = "Model deleted."
}

func (ui *gameUI) startAIDownload() {
	if ui.ai.Downloading {
		return
	}
	if _, ok := ui.currentAIModelPack(); !ok {
		ui.ai.Status = "Select a model first."
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
	modelID := ui.ai.ModelID

	go func(progressCh chan<- ai.Progress, doneCh chan<- error, modelID string) {
		err := ai.DownloadModel(context.Background(), modelID, func(p ai.Progress) {
			select {
			case progressCh <- p:
			default:
			}
		})
		doneCh <- err
	}(progressCh, doneCh, modelID)
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
		ui.ai.Status = "Model download complete."
	default:
	}
}

func (ui *gameUI) refreshAIModelStatus() {
	ui.ai.DownloadedByID = map[string]bool{}
	for _, pack := range ui.ai.ModelPacks {
		exists, err := ai.ModelExists(pack.ID)
		if err != nil {
			ui.ai.Status = "Could not check model status: " + err.Error()
			ui.ai.Downloaded = false
			return
		}
		ui.ai.DownloadedByID[pack.ID] = exists
	}
	ui.ai.Downloaded = ui.ai.DownloadedByID[ui.ai.ModelID]

	prevActive := ui.ai.ActiveModelID
	if ui.ai.ActiveModelID == "" || !ui.ai.DownloadedByID[ui.ai.ActiveModelID] {
		downloadedChoices := ui.downloadedAIModelPacks()
		if len(downloadedChoices) > 0 {
			ui.ai.ActiveModelID = downloadedChoices[0].ID
			if ui.ai.ActiveModelID != prevActive {
				_ = ai.SaveConfig(ai.Config{AIEnabled: true, ModelID: ui.ai.ActiveModelID})
			}
		} else {
			ui.ai.ActiveModelID = ""
			if prevActive != "" {
				_ = ai.SaveConfig(ai.Config{AIEnabled: false, ModelID: ""})
			}
		}
	}
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
	drawText("Left/Right to change, Enter to select, Esc to go back", int32(left.X)+22, int32(left.Y+left.Height)-38, 18, colorDim)

	if ui.ai.Downloading {
		drawText(formatAIDownloadProgress(ui.ai.DownloadedBytes, ui.ai.DownloadTotal), int32(right.X)+24, int32(right.Y)+82, 18, colorAccent)
	}

	if pack, ok := ui.currentAIModelPack(); ok {
		drawWrappedText("Why pick this one: "+pack.Summary, right, 110, 16, colorText)
	}
	nextY := ui.drawAIModelComparisonTable(right, 260)

	if strings.TrimSpace(ui.ai.Status) != "" {
		statusColor := colorAccent
		statusLower := strings.ToLower(ui.ai.Status)
		if strings.Contains(statusLower, "failed") || strings.Contains(statusLower, "error") {
			statusColor = colorWarn
		}
		drawWrappedText(ui.ai.Status, right, nextY+8, 18, statusColor)
	}
}

func (ui *gameUI) drawAIModelComparisonTable(panel rl.Rectangle, y int32) int32 {
	packs := ui.modelComparisonPacks()
	if len(packs) == 0 {
		drawWrappedText("No model choices are available.", panel, y, 16, colorWarn)
		return y + 28
	}

	tableTop := int32(panel.Y) + y
	leftX := int32(panel.X) + 16
	rowH := int32(30)
	usableW := int32(panel.Width) - 32
	colModel := leftX + 8
	colSize := leftX + int32(float32(usableW)*0.42)
	colSpeed := leftX + int32(float32(usableW)*0.55)
	colQuality := leftX + int32(float32(usableW)*0.69)
	colBestFor := leftX + int32(float32(usableW)*0.82)

	drawText("Model comparison (selected model is listed first)", leftX, tableTop, 17, colorAccent)
	tableTop += 26

	headerRect := rl.NewRectangle(float32(leftX), float32(tableTop), panel.Width-32, float32(rowH))
	rl.DrawRectangleRounded(headerRect, 0.2, 8, rl.Fade(colorBorder, 0.25))
	drawText("Model", colModel, tableTop+7, 16, colorText)
	drawText("Size", colSize, tableTop+7, 16, colorText)
	drawText("Speed", colSpeed, tableTop+7, 16, colorText)
	drawText("Quality", colQuality, tableTop+7, 16, colorText)
	drawText("Best for", colBestFor, tableTop+7, 16, colorText)
	tableTop += rowH + 4

	maxRows := 6
	for i, pack := range packs {
		if i >= maxRows {
			break
		}
		rowRect := rl.NewRectangle(float32(leftX), float32(tableTop), panel.Width-32, float32(rowH))
		if i == 0 {
			rl.DrawRectangleRounded(rowRect, 0.2, 8, rl.Fade(colorAccent, 0.2))
			rl.DrawRectangleRoundedLinesEx(rowRect, 0.2, 8, 1, colorAccent)
		} else {
			rl.DrawRectangleRounded(rowRect, 0.2, 8, rl.Fade(colorPanel, 0.55))
			rl.DrawRectangleRoundedLinesEx(rowRect, 0.2, 8, 1, rl.Fade(colorBorder, 0.7))
		}

		name := truncateForUI(pack.Name, 30)
		if i == 0 {
			name = "Current: " + truncateForUI(pack.Name, 21)
		}
		drawText(name, colModel, tableTop+7, 15, colorText)
		drawText(fmt.Sprintf("%.2f GB", pack.SizeGB), colSize, tableTop+7, 15, colorText)
		drawText(pack.Speed, colSpeed, tableTop+7, 15, colorText)
		drawText(pack.Quality, colQuality, tableTop+7, 15, colorText)
		drawText(truncateForUI(pack.BestFor, 24), colBestFor, tableTop+7, 15, colorText)
		tableTop += rowH + 4
	}
	return int32(float32(tableTop - int32(panel.Y)))
}

func (ui *gameUI) modelComparisonPacks() []ai.ModelPack {
	if len(ui.ai.ModelPacks) == 0 {
		return nil
	}
	selectedID := strings.ToLower(strings.TrimSpace(ui.ai.ModelID))
	out := make([]ai.ModelPack, 0, len(ui.ai.ModelPacks))
	for _, pack := range ui.ai.ModelPacks {
		if strings.ToLower(strings.TrimSpace(pack.ID)) == selectedID {
			out = append(out, pack)
			break
		}
	}
	for _, pack := range ui.ai.ModelPacks {
		if strings.ToLower(strings.TrimSpace(pack.ID)) == selectedID {
			continue
		}
		out = append(out, pack)
	}
	return out
}

func modelPackIndexByID(packs []ai.ModelPack, id string) int {
	id = strings.ToLower(strings.TrimSpace(id))
	for i, pack := range packs {
		if strings.ToLower(strings.TrimSpace(pack.ID)) == id {
			return i
		}
	}
	return -1
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
