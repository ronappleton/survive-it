package gui

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/appengine-ltd/survive-it/internal/game"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const defaultPlayerProfilesFile = "survive-it-player-profiles.json"

type playerProfile struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	CreatedAt         time.Time         `json:"created_at"`
	LastPlayedAt      time.Time         `json:"last_played_at,omitempty"`
	RunsPlayed        int               `json:"runs_played"`
	TotalDaysSurvived int               `json:"total_days_survived"`
	Config            game.PlayerConfig `json:"config"`
}

type playerProfilesPayload struct {
	FormatVersion int             `json:"format_version"`
	ActiveID      string          `json:"active_id,omitempty"`
	Profiles      []playerProfile `json:"profiles"`
}

func loadPlayerProfiles(path string) ([]playerProfile, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, "", nil
		}
		return nil, "", err
	}
	var payload playerProfilesPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, "", err
	}
	profiles := append([]playerProfile(nil), payload.Profiles...)
	return profiles, strings.TrimSpace(payload.ActiveID), nil
}

func savePlayerProfiles(path string, profiles []playerProfile, activeID string) error {
	payload := playerProfilesPayload{
		FormatVersion: 1,
		ActiveID:      strings.TrimSpace(activeID),
		Profiles:      append([]playerProfile(nil), profiles...),
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func (ui *gameUI) initPlayerProfiles() {
	profiles, activeID, err := loadPlayerProfiles(defaultPlayerProfilesFile)
	if err != nil {
		ui.status = "Profile load failed: " + err.Error()
		profiles = nil
	}
	profiles, _ = normalizeProfiles(profiles)
	if len(profiles) == 0 {
		profiles = []playerProfile{newPlayerProfile("You", ui.selectedMode())}
		activeID = profiles[0].ID
	}
	if profileIndexByID(profiles, activeID) < 0 {
		activeID = profiles[0].ID
	}
	ui.profiles = profiles
	ui.selectedProfileID = activeID
}

func normalizeProfiles(in []playerProfile) ([]playerProfile, bool) {
	profiles := append([]playerProfile(nil), in...)
	changed := false
	if len(profiles) == 0 {
		return profiles, false
	}
	taken := make(map[string]struct{}, len(profiles))
	for i := range profiles {
		p := &profiles[i]
		if strings.TrimSpace(p.Name) == "" {
			p.Name = fmt.Sprintf("Survivor %d", i+1)
			changed = true
		}
		if strings.TrimSpace(p.ID) == "" {
			p.ID = uniqueProfileID(slugifyName(p.Name), taken)
			changed = true
		}
		if _, exists := taken[p.ID]; exists {
			p.ID = uniqueProfileID(p.ID, taken)
			changed = true
		}
		taken[p.ID] = struct{}{}
		if p.CreatedAt.IsZero() {
			p.CreatedAt = time.Now().UTC()
			changed = true
		}
		cfg := sanitizeProfileConfig(p.Config, game.ModeAlone)
		if strings.TrimSpace(cfg.Name) == "" {
			cfg.Name = p.Name
			changed = true
		}
		p.Config = cfg
	}
	sort.SliceStable(profiles, func(i, j int) bool {
		return strings.ToLower(profiles[i].Name) < strings.ToLower(profiles[j].Name)
	})
	return profiles, changed
}

func sanitizeProfileConfig(cfg game.PlayerConfig, mode game.GameMode) game.PlayerConfig {
	cfg.Name = strings.TrimSpace(cfg.Name)
	cfg.Strength = clampInt(cfg.Strength, -3, 3)
	cfg.MentalStrength = clampInt(cfg.MentalStrength, -3, 3)
	cfg.Agility = clampInt(cfg.Agility, -3, 3)
	cfg.Endurance = cfg.Strength
	cfg.Mental = cfg.MentalStrength
	cfg.Bushcraft = cfg.Agility
	cfg.Hunting = clampInt(cfg.Hunting, 0, 100)
	cfg.Fishing = clampInt(cfg.Fishing, 0, 100)
	cfg.Foraging = clampInt(cfg.Foraging, 0, 100)
	cfg.Crafting = clampInt(cfg.Crafting, 0, 100)
	cfg.Gathering = clampInt(cfg.Gathering, 0, 100)
	cfg.Trapping = clampInt(cfg.Trapping, 0, 100)
	cfg.Firecraft = clampInt(cfg.Firecraft, 0, 100)
	cfg.Sheltercraft = clampInt(cfg.Sheltercraft, 0, 100)
	cfg.Cooking = clampInt(cfg.Cooking, 0, 100)
	cfg.Navigation = clampInt(cfg.Navigation, 0, 100)
	if strings.TrimSpace(cfg.CurrentTask) == "" {
		cfg.CurrentTask = "Idle"
	}
	if cfg.KitLimit <= 0 {
		cfg.KitLimit = defaultKitLimitForMode(mode)
	}
	maxLimit := maxKitLimitForMode(mode)
	if cfg.KitLimit > maxLimit {
		cfg.KitLimit = maxLimit
	}
	if len(cfg.Kit) > cfg.KitLimit {
		cfg.Kit = append([]game.KitItem(nil), cfg.Kit[:cfg.KitLimit]...)
	}
	if cfg.Sex == "" {
		cfg.Sex = game.SexOther
	}
	if cfg.BodyType == "" {
		cfg.BodyType = game.BodyTypeNeutral
	}
	if cfg.WeightKg <= 0 {
		cfg.WeightKg = 75
	}
	if cfg.HeightFt <= 0 {
		cfg.HeightFt = 5
	}
	if cfg.HeightIn < 0 || cfg.HeightIn > 11 {
		cfg.HeightIn = 10
	}
	return cfg
}

func newPlayerProfile(name string, mode game.GameMode) playerProfile {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Survivor"
	}
	cfg := sanitizeProfileConfig(defaultPlayerConfig(0, mode), mode)
	cfg.Name = name
	now := time.Now().UTC()
	return playerProfile{
		ID:        slugifyName(name),
		Name:      name,
		CreatedAt: now,
		Config:    cfg,
	}
}

func slugifyName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return "survivor"
	}
	b := strings.Builder{}
	lastDash := false
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			lastDash = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "survivor"
	}
	return out
}

func uniqueProfileID(base string, taken map[string]struct{}) string {
	id := slugifyName(base)
	if taken == nil {
		return id
	}
	if _, exists := taken[id]; !exists {
		return id
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", id, i)
		if _, exists := taken[candidate]; !exists {
			return candidate
		}
	}
}

func profileIndexByID(profiles []playerProfile, id string) int {
	id = strings.TrimSpace(id)
	if id == "" {
		return -1
	}
	for i := range profiles {
		if profiles[i].ID == id {
			return i
		}
	}
	return -1
}

func (ui *gameUI) selectedProfileIndex() int {
	if len(ui.profiles) == 0 {
		return -1
	}
	idx := profileIndexByID(ui.profiles, ui.selectedProfileID)
	if idx >= 0 {
		return idx
	}
	ui.selectedProfileID = ui.profiles[0].ID
	return 0
}

func (ui *gameUI) selectedProfile() (playerProfile, bool) {
	idx := ui.selectedProfileIndex()
	if idx < 0 || idx >= len(ui.profiles) {
		return playerProfile{}, false
	}
	return ui.profiles[idx], true
}

func (ui *gameUI) selectedProfileSummary() string {
	profile, ok := ui.selectedProfile()
	if !ok {
		return "None"
	}
	if profile.RunsPlayed <= 0 {
		return profile.Name + " (new)"
	}
	return fmt.Sprintf("%s (%d runs)", profile.Name, profile.RunsPlayed)
}

func (ui *gameUI) selectProfileByIndex(idx int) {
	if idx < 0 || idx >= len(ui.profiles) {
		return
	}
	ui.selectedProfileID = ui.profiles[idx].ID
	ui.saveProfilesToDisk()
}

func (ui *gameUI) applySelectedProfileToSetupPrimary() {
	if len(ui.pcfg.Players) == 0 {
		return
	}
	profile, ok := ui.selectedProfile()
	if !ok {
		return
	}
	cfg := sanitizeProfileConfig(profile.Config, ui.selectedMode())
	if strings.TrimSpace(cfg.Name) == "" {
		cfg.Name = profile.Name
	}
	ui.pcfg.Players[0] = cfg
}

func (ui *gameUI) openProfilesScreen(returnTo screen) {
	if len(ui.profiles) == 0 {
		ui.initPlayerProfiles()
	}
	ui.profilesUI.EditingNew = false
	ui.profilesUI.EditingID = ""
	ui.profilesUI.NameBuffer = ""
	ui.profilesUI.ReturnTo = returnTo
	idx := ui.selectedProfileIndex()
	if idx < 0 {
		idx = 0
	}
	ui.profilesUI.Cursor = idx
	ui.screen = screenProfiles
}

func (ui *gameUI) updateProfiles() {
	rows := len(ui.profiles) + 2
	if rows < 2 {
		rows = 2
	}
	if ui.profilesUI.Cursor < 0 || ui.profilesUI.Cursor >= rows {
		ui.profilesUI.Cursor = 0
	}

	if ui.profilesUI.EditingNew {
		captureTextInput(&ui.profilesUI.NameBuffer, 32)
		if rl.IsKeyPressed(rl.KeyEscape) {
			ui.profilesUI.EditingNew = false
			ui.profilesUI.EditingID = ""
			ui.profilesUI.NameBuffer = ""
			return
		}
		if rl.IsKeyPressed(rl.KeyEnter) {
			ui.commitProfileEdit()
		}
		return
	}

	if rl.IsKeyPressed(rl.KeyEscape) {
		ui.screen = ui.profileReturnScreen()
		return
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		ui.profilesUI.Cursor = wrapIndex(ui.profilesUI.Cursor+1, rows)
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		ui.profilesUI.Cursor = wrapIndex(ui.profilesUI.Cursor-1, rows)
	}
	if ShiftPressedKey(rl.KeyN) {
		ui.startProfileCreate()
		return
	}
	if ShiftPressedKey(rl.KeyR) && ui.profilesUI.Cursor < len(ui.profiles) {
		ui.startProfileRename(ui.profilesUI.Cursor)
		return
	}

	if rl.IsKeyPressed(rl.KeyEnter) {
		switch ui.profilesUI.Cursor {
		case len(ui.profiles):
			ui.startProfileCreate()
		case len(ui.profiles) + 1:
			ui.screen = ui.profileReturnScreen()
		default:
			ui.selectProfileByIndex(ui.profilesUI.Cursor)
			ui.applySelectedProfileToSetupPrimary()
			ui.ensureSetupPlayers()
			ui.profilesUI.Status = "Active profile set for New Run Setup."
		}
	}
}

func (ui *gameUI) profileReturnScreen() screen {
	if ui.profilesUI.ReturnTo == 0 {
		return screenMenu
	}
	return ui.profilesUI.ReturnTo
}

func (ui *gameUI) startProfileCreate() {
	ui.profilesUI.EditingNew = true
	ui.profilesUI.EditingID = ""
	ui.profilesUI.NameBuffer = ""
	ui.profilesUI.Status = "Enter a name for the new profile."
}

func (ui *gameUI) startProfileRename(idx int) {
	if idx < 0 || idx >= len(ui.profiles) {
		return
	}
	ui.profilesUI.EditingNew = true
	ui.profilesUI.EditingID = ui.profiles[idx].ID
	ui.profilesUI.NameBuffer = ui.profiles[idx].Name
	ui.profilesUI.Status = "Rename profile and press Enter to save."
}

func (ui *gameUI) commitProfileEdit() {
	name := strings.TrimSpace(ui.profilesUI.NameBuffer)
	if name == "" {
		ui.profilesUI.Status = "Profile name cannot be empty."
		return
	}
	if strings.TrimSpace(ui.profilesUI.EditingID) == "" {
		profile := newPlayerProfile(name, ui.selectedMode())
		taken := make(map[string]struct{}, len(ui.profiles))
		for _, existing := range ui.profiles {
			taken[existing.ID] = struct{}{}
		}
		profile.ID = uniqueProfileID(profile.ID, taken)
		ui.profiles = append(ui.profiles, profile)
		sort.SliceStable(ui.profiles, func(i, j int) bool {
			return strings.ToLower(ui.profiles[i].Name) < strings.ToLower(ui.profiles[j].Name)
		})
		idx := profileIndexByID(ui.profiles, profile.ID)
		if idx >= 0 {
			ui.profilesUI.Cursor = idx
			ui.selectProfileByIndex(idx)
		}
		ui.applySelectedProfileToSetupPrimary()
		ui.ensureSetupPlayers()
		ui.profilesUI.Status = "Profile created."
	} else {
		idx := profileIndexByID(ui.profiles, ui.profilesUI.EditingID)
		if idx >= 0 {
			ui.profiles[idx].Name = name
			ui.profiles[idx].Config.Name = name
			if ui.profiles[idx].ID == ui.selectedProfileID {
				ui.applySelectedProfileToSetupPrimary()
				ui.ensureSetupPlayers()
			}
			ui.profilesUI.Status = "Profile renamed."
			ui.saveProfilesToDisk()
		}
	}
	ui.profilesUI.EditingNew = false
	ui.profilesUI.EditingID = ""
	ui.profilesUI.NameBuffer = ""
}

func (ui *gameUI) drawProfiles() {
	DrawFrame(ui.width, ui.height)
	left := rl.NewRectangle(20, 20, float32(ui.width)*0.4, float32(ui.height-40))
	right := rl.NewRectangle(left.X+left.Width+20, 20, float32(ui.width)-left.Width-60, float32(ui.height-40))
	drawPanel(left, "Player Profiles")
	drawPanel(right, "Profile Details")

	y := int32(left.Y) + 60
	for i, profile := range ui.profiles {
		if y > int32(left.Y+left.Height)-120 {
			break
		}
		if i == ui.profilesUI.Cursor {
			drawListRowFrame(rl.NewRectangle(left.X+10, float32(y-8), left.Width-20, 38), true)
		}
		prefix := "  "
		if profile.ID == ui.selectedProfileID {
			prefix = "* "
		}
		drawText(prefix+profile.Name, int32(left.X)+16, y, typeScale.Body, colorText)
		y += 42
	}

	addRow := len(ui.profiles)
	backRow := addRow + 1
	addY := int32(left.Y) + int32(left.Height) - 102
	if ui.profilesUI.Cursor == addRow {
		drawListRowFrame(rl.NewRectangle(left.X+10, float32(addY-8), left.Width-20, 36), true)
	}
	drawText("Add New Profile", int32(left.X)+16, addY, typeScale.Body, colorAccent)
	if ui.profilesUI.Cursor == backRow {
		drawListRowFrame(rl.NewRectangle(left.X+10, float32(addY+30), left.Width-20, 36), true)
	}
	drawText("Back", int32(left.X)+16, addY+38, typeScale.Body, colorText)

	DrawHintText("Enter select | Shift+N new | Shift+R rename | Esc back", int32(left.X)+16, int32(left.Y+left.Height)-30)

	profile, ok := ui.selectedProfile()
	if !ok {
		drawWrappedText("No profiles available.", right, 48, typeScale.Body, colorWarn)
		return
	}
	lines := []string{
		fmt.Sprintf("Name: %s", profile.Name),
		fmt.Sprintf("Profile ID: %s", profile.ID),
		fmt.Sprintf("Runs Played: %d", profile.RunsPlayed),
		fmt.Sprintf("Total Days Survived: %d", profile.TotalDaysSurvived),
	}
	if !profile.LastPlayedAt.IsZero() {
		lines = append(lines, "Last Played: "+profile.LastPlayedAt.Local().Format("2006-01-02 15:04"))
	} else {
		lines = append(lines, "Last Played: never")
	}
	lines = append(lines,
		"",
		"Progressed Stats (persist across runs):",
		fmt.Sprintf("Strength %+d | Mental %+d | Agility %+d", profile.Config.Strength, profile.Config.MentalStrength, profile.Config.Agility),
		fmt.Sprintf("Hunt %d | Fish %d | Forage %d", profile.Config.Hunting, profile.Config.Fishing, profile.Config.Foraging),
		fmt.Sprintf("Craft %d | Gather %d | Trap %d", profile.Config.Crafting, profile.Config.Gathering, profile.Config.Trapping),
		fmt.Sprintf("Fire %d | Shelter %d | Cook %d | Nav %d", profile.Config.Firecraft, profile.Config.Sheltercraft, profile.Config.Cooking, profile.Config.Navigation),
		"",
		"Use Enter on a profile to make it active",
		"for New Run Setup (Player 1 / YOU).",
	)
	drawLines(right, 50, typeScale.Body, lines, colorText)

	if strings.TrimSpace(ui.profilesUI.Status) != "" {
		drawWrappedText(ui.profilesUI.Status, right, int32(right.Height)-76, typeScale.Small, colorAccent)
	}
	if ui.profilesUI.EditingNew {
		r := rl.NewRectangle(left.X+20, left.Y+left.Height-176, left.Width-40, 84)
		drawDialogPanel(r)
		title := "New Profile Name"
		if strings.TrimSpace(ui.profilesUI.EditingID) != "" {
			title = "Rename Profile"
		}
		drawText(title, int32(r.X)+12, int32(r.Y)+10, 18, colorAccent)
		drawText(truncateForUI(ui.profilesUI.NameBuffer, 42)+"_", int32(r.X)+12, int32(r.Y)+34, 24, colorText)
	}
}

func (ui *gameUI) saveProfilesToDisk() {
	if len(ui.profiles) == 0 {
		return
	}
	if err := savePlayerProfiles(defaultPlayerProfilesFile, ui.profiles, ui.selectedProfileID); err != nil {
		ui.profilesUI.Status = "Profile save failed: " + err.Error()
	}
}

func playerConfigFromState(player game.PlayerState) game.PlayerConfig {
	return game.PlayerConfig{
		Name:           player.Name,
		Sex:            player.Sex,
		BodyType:       player.BodyType,
		WeightKg:       player.WeightKg,
		HeightFt:       player.HeightFt,
		HeightIn:       player.HeightIn,
		Endurance:      player.Endurance,
		Bushcraft:      player.Bushcraft,
		Mental:         player.Mental,
		Strength:       player.Strength,
		MentalStrength: player.MentalStrength,
		Agility:        player.Agility,
		Hunting:        player.Hunting,
		Fishing:        player.Fishing,
		Foraging:       player.Foraging,
		Crafting:       player.Crafting,
		Gathering:      player.Gathering,
		Trapping:       player.Trapping,
		Firecraft:      player.Firecraft,
		Sheltercraft:   player.Sheltercraft,
		Cooking:        player.Cooking,
		Navigation:     player.Navigation,
		CurrentTask:    player.CurrentTask,
		Traits:         append([]game.TraitModifier(nil), player.Traits...),
		KitLimit:       player.KitLimit,
		Kit:            append([]game.KitItem(nil), player.Kit...),
	}
}

func (ui *gameUI) persistActiveRunProfileProgress() {
	if ui == nil || ui.run == nil || len(ui.run.Players) == 0 || len(ui.profiles) == 0 {
		return
	}
	profileID := strings.TrimSpace(ui.runProfileID)
	if profileID == "" {
		profileID = strings.TrimSpace(ui.selectedProfileID)
	}
	idx := profileIndexByID(ui.profiles, profileID)
	if idx < 0 {
		return
	}
	profile := &ui.profiles[idx]
	player := ui.run.Players[0]
	cfg := sanitizeProfileConfig(playerConfigFromState(player), ui.run.Config.Mode)
	name := strings.TrimSpace(player.Name)
	if name == "" {
		name = profile.Name
	}
	if name == "" {
		name = "You"
	}
	profile.Name = name
	cfg.Name = name
	profile.Config = cfg
	profile.RunsPlayed++
	profile.TotalDaysSurvived += maxInt(0, ui.run.Day)
	now := time.Now().UTC()
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = now
	}
	profile.LastPlayedAt = now
	ui.selectedProfileID = profile.ID
	ui.saveProfilesToDisk()
}

func (ui *gameUI) leaveRunToMenu() {
	ui.persistActiveRunProfileProgress()
	ui.runProfileID = ""
	ui.enterMenu()
}
