package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/appengine-ltd/survive-it/internal/game"
)

type docFile struct {
	Name    string
	Title   string
	Content string
}

func main() {
	root := filepath.Join("docs", "reference", "catalogs")
	if err := os.MkdirAll(root, 0o755); err != nil {
		fatal(err)
	}

	files := []docFile{
		generateAnimalsDoc(),
		generatePlantsDoc(),
		generateResourcesDoc(),
		generateTreesDoc(),
		generateCraftablesDoc(),
		generateTrapsDoc(),
		generateSheltersDoc(),
		generateScenariosDoc(),
		generateKitDoc(),
	}
	for _, f := range files {
		path := filepath.Join(root, f.Name)
		if err := os.WriteFile(path, []byte(f.Content), 0o644); err != nil {
			fatal(err)
		}
		fmt.Printf("wrote %s\n", path)
	}

	index := generateCatalogIndex(files)
	indexPath := filepath.Join(root, "README.md")
	if err := os.WriteFile(indexPath, []byte(index), 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("wrote %s\n", indexPath)
}

func generateCatalogIndex(files []docFile) string {
	var b strings.Builder
	b.WriteString("# Data Catalogs\n\n")
	b.WriteString("Generated from the current Go source using `go run ./cmd/docsgen`.\n\n")
	for _, f := range files {
		b.WriteString(fmt.Sprintf("- [%s](./%s)\n", f.Title, f.Name))
	}
	return b.String()
}

func generateAnimalsDoc() docFile {
	items := game.AnimalCatalog()
	domainRank := map[game.AnimalDomain]int{
		game.AnimalDomainLand:  0,
		game.AnimalDomainWater: 1,
		game.AnimalDomainAir:   2,
	}
	sort.Slice(items, func(i, j int) bool {
		ri := domainRank[items[i].Domain]
		rj := domainRank[items[j].Domain]
		if ri != rj {
			return ri < rj
		}
		if items[i].Name != items[j].Name {
			return items[i].Name < items[j].Name
		}
		return items[i].ID < items[j].ID
	})

	var b strings.Builder
	b.WriteString("# Animals\n\n")
	b.WriteString("Source: `internal/game/animals.go` (`AnimalCatalog`).\n\n")
	b.WriteString(fmt.Sprintf("Total animals: **%d**.\n\n", len(items)))
	b.WriteString("| ID | Name | Domain | Biome Tags | Weight (kg) | Edible Yield | Nutrition /100g | Disease Risks |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- | --- |\n")
	for _, a := range items {
		b.WriteString("| ")
		b.WriteString(escape(a.ID))
		b.WriteString(" | ")
		b.WriteString(escape(a.Name))
		b.WriteString(" | ")
		b.WriteString(escape(string(a.Domain)))
		b.WriteString(" | ")
		b.WriteString(escape(strings.Join(a.BiomeTags, ", ")))
		b.WriteString(" | ")
		b.WriteString(fmt.Sprintf("%.3g-%.3g", a.WeightMinKg, a.WeightMaxKg))
		b.WriteString(" | ")
		b.WriteString(fmt.Sprintf("%.0f%%", a.EdibleYieldRatio*100))
		b.WriteString(" | ")
		b.WriteString(escape(fmt.Sprintf("%dkcal %dgP %dgF %dgS", a.NutritionPer100g.CaloriesKcal, a.NutritionPer100g.ProteinG, a.NutritionPer100g.FatG, a.NutritionPer100g.SugarG)))
		b.WriteString(" | ")
		b.WriteString(escape(formatDiseaseRisks(a.DiseaseRisks)))
		b.WriteString(" |\n")
	}

	return docFile{Name: "animals.md", Title: "Animals", Content: b.String()}
}

func generatePlantsDoc() docFile {
	items := game.PlantCatalog()
	sort.Slice(items, func(i, j int) bool {
		if items[i].Category != items[j].Category {
			return items[i].Category < items[j].Category
		}
		if items[i].Name != items[j].Name {
			return items[i].Name < items[j].Name
		}
		return items[i].ID < items[j].ID
	})

	var b strings.Builder
	b.WriteString("# Plants\n\n")
	b.WriteString("Source: `internal/game/environment_resources.go` (`PlantCatalog`).\n\n")
	b.WriteString(fmt.Sprintf("Total plant foods: **%d**.\n\n", len(items)))
	b.WriteString("| ID | Name | Category | Biome Tags | Yield (g) | Nutrition /100g |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- |\n")
	for _, p := range items {
		b.WriteString("| ")
		b.WriteString(escape(p.ID))
		b.WriteString(" | ")
		b.WriteString(escape(p.Name))
		b.WriteString(" | ")
		b.WriteString(escape(string(p.Category)))
		b.WriteString(" | ")
		b.WriteString(escape(strings.Join(p.BiomeTags, ", ")))
		b.WriteString(" | ")
		b.WriteString(fmt.Sprintf("%d-%d", p.YieldMinG, p.YieldMaxG))
		b.WriteString(" | ")
		b.WriteString(escape(fmt.Sprintf("%dkcal %dgP %dgF %dgS", p.NutritionPer100g.CaloriesKcal, p.NutritionPer100g.ProteinG, p.NutritionPer100g.FatG, p.NutritionPer100g.SugarG)))
		b.WriteString(" |\n")
	}

	return docFile{Name: "plants.md", Title: "Plants", Content: b.String()}
}

func generateResourcesDoc() docFile {
	items := game.ResourceCatalog()
	sort.Slice(items, func(i, j int) bool {
		if items[i].Name != items[j].Name {
			return items[i].Name < items[j].Name
		}
		return items[i].ID < items[j].ID
	})

	var b strings.Builder
	b.WriteString("# Resources\n\n")
	b.WriteString("Source: `internal/game/environment_resources.go` (`ResourceCatalog`).\n\n")
	b.WriteString(fmt.Sprintf("Total utility resources: **%d**.\n\n", len(items)))
	b.WriteString("| ID | Name | Unit | Biome Tags | Gather Range | Dryness | Flammable | Uses |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- | --- |\n")
	for _, r := range items {
		b.WriteString("| ")
		b.WriteString(escape(r.ID))
		b.WriteString(" | ")
		b.WriteString(escape(r.Name))
		b.WriteString(" | ")
		b.WriteString(escape(r.Unit))
		b.WriteString(" | ")
		b.WriteString(escape(strings.Join(r.BiomeTags, ", ")))
		b.WriteString(" | ")
		b.WriteString(fmt.Sprintf("%.2g-%.2g", r.GatherMin, r.GatherMax))
		b.WriteString(" | ")
		b.WriteString(fmt.Sprintf("%.2f", r.Dryness))
		b.WriteString(" | ")
		b.WriteString(yesNo(r.Flammable))
		b.WriteString(" | ")
		b.WriteString(escape(strings.Join(r.Uses, ", ")))
		b.WriteString(" |\n")
	}

	return docFile{Name: "resources.md", Title: "Resources", Content: b.String()}
}

func generateTreesDoc() docFile {
	items := game.TreeCatalog()
	sort.Slice(items, func(i, j int) bool {
		if items[i].Name != items[j].Name {
			return items[i].Name < items[j].Name
		}
		return items[i].ID < items[j].ID
	})

	var b strings.Builder
	b.WriteString("# Trees\n\n")
	b.WriteString("Source: `internal/game/environment_resources.go` (`TreeCatalog`).\n\n")
	b.WriteString(fmt.Sprintf("Total trees: **%d**.\n\n", len(items)))
	b.WriteString("| ID | Name | Wood Type | Biome Tags | Gather (kg) | Heat Factor | Burn Factor | Spark Ease |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- | --- |\n")
	for _, t := range items {
		b.WriteString("| ")
		b.WriteString(escape(t.ID))
		b.WriteString(" | ")
		b.WriteString(escape(t.Name))
		b.WriteString(" | ")
		b.WriteString(escape(string(t.WoodType)))
		b.WriteString(" | ")
		b.WriteString(escape(strings.Join(t.BiomeTags, ", ")))
		b.WriteString(" | ")
		b.WriteString(fmt.Sprintf("%.1f-%.1f", t.GatherMinKg, t.GatherMaxKg))
		b.WriteString(" | ")
		b.WriteString(fmt.Sprintf("%.2f", t.HeatFactor))
		b.WriteString(" | ")
		b.WriteString(fmt.Sprintf("%.2f", t.BurnFactor))
		b.WriteString(" | ")
		b.WriteString(strconv.Itoa(t.SparkEase))
		b.WriteString(" |\n")
	}

	return docFile{Name: "trees.md", Title: "Trees", Content: b.String()}
}

func generateCraftablesDoc() docFile {
	items := game.CraftableCatalog()
	sort.Slice(items, func(i, j int) bool {
		ci := strings.TrimSpace(items[i].Category)
		cj := strings.TrimSpace(items[j].Category)
		if ci == "" {
			ci = "general"
		}
		if cj == "" {
			cj = "general"
		}
		if ci != cj {
			return ci < cj
		}
		if items[i].Name != items[j].Name {
			return items[i].Name < items[j].Name
		}
		return items[i].ID < items[j].ID
	})

	var b strings.Builder
	b.WriteString("# Craftables\n\n")
	b.WriteString("Source: `internal/game/environment_resources.go` (`CraftableCatalog`).\n\n")
	b.WriteString(fmt.Sprintf("Total craftables: **%d**.\n\n", len(items)))
	b.WriteString("| ID | Name | Category | Min Bushcraft | Time (h) | Portable | Req Fire | Req Shelter | Wood (kg) | Weight (kg) | Requires Items | Requires Resources | Biomes |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |\n")
	for _, c := range items {
		category := strings.TrimSpace(c.Category)
		if category == "" {
			category = "general"
		}
		b.WriteString("| ")
		b.WriteString(escape(c.ID))
		b.WriteString(" | ")
		b.WriteString(escape(c.Name))
		b.WriteString(" | ")
		b.WriteString(escape(category))
		b.WriteString(" | ")
		b.WriteString(strconv.Itoa(c.MinBushcraft))
		b.WriteString(" | ")
		b.WriteString(formatFloat(c.BaseHours))
		b.WriteString(" | ")
		b.WriteString(yesNo(c.Portable))
		b.WriteString(" | ")
		b.WriteString(yesNo(c.RequiresFire))
		b.WriteString(" | ")
		b.WriteString(yesNo(c.RequiresShelter))
		b.WriteString(" | ")
		b.WriteString(formatFloat(c.WoodKg))
		b.WriteString(" | ")
		b.WriteString(formatFloat(c.WeightKg))
		b.WriteString(" | ")
		b.WriteString(escape(strings.Join(c.RequiresItems, ", ")))
		b.WriteString(" | ")
		b.WriteString(escape(formatResourceReqs(c.RequiresResources)))
		b.WriteString(" | ")
		b.WriteString(escape(strings.Join(c.BiomeTags, ", ")))
		b.WriteString(" |\n")
	}

	return docFile{Name: "craftables.md", Title: "Craftables", Content: b.String()}
}

func generateTrapsDoc() docFile {
	items := game.TrapCatalog()
	sort.Slice(items, func(i, j int) bool {
		if items[i].Name != items[j].Name {
			return items[i].Name < items[j].Name
		}
		return items[i].ID < items[j].ID
	})

	var b strings.Builder
	b.WriteString("# Traps\n\n")
	b.WriteString("Source: `internal/game/trapping.go` (`TrapCatalog`).\n\n")
	b.WriteString(fmt.Sprintf("Total traps: **%d**.\n\n", len(items)))
	b.WriteString("| ID | Name | Targets | Min Bushcraft | Base Catch | Setup (h) | Cond Loss | Yield (kg) | Needs Water | Requires Crafted | Requires Resources | Requires Kit | Biomes |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |\n")
	for _, t := range items {
		b.WriteString("| ")
		b.WriteString(escape(t.ID))
		b.WriteString(" | ")
		b.WriteString(escape(t.Name))
		b.WriteString(" | ")
		b.WriteString(escape(strings.Join(t.Targets, ", ")))
		b.WriteString(" | ")
		b.WriteString(strconv.Itoa(t.MinBushcraft))
		b.WriteString(" | ")
		b.WriteString(fmt.Sprintf("%.0f%%", t.BaseChance*100))
		b.WriteString(" | ")
		b.WriteString(formatFloat(t.BaseHours))
		b.WriteString(" | ")
		b.WriteString(strconv.Itoa(t.ConditionLoss))
		b.WriteString(" | ")
		b.WriteString(fmt.Sprintf("%.2g-%.2g", t.YieldMinKg, t.YieldMaxKg))
		b.WriteString(" | ")
		b.WriteString(yesNo(t.NeedsWater))
		b.WriteString(" | ")
		b.WriteString(escape(strings.Join(t.RequiresCrafted, ", ")))
		b.WriteString(" | ")
		b.WriteString(escape(formatResourceReqs(t.RequiresResources)))
		b.WriteString(" | ")
		b.WriteString(escape(formatKitReqs(t.RequiresKit)))
		b.WriteString(" | ")
		b.WriteString(escape(strings.Join(t.BiomeTags, ", ")))
		b.WriteString(" |\n")
	}

	return docFile{Name: "traps.md", Title: "Traps", Content: b.String()}
}

func generateSheltersDoc() docFile {
	items := game.ShelterCatalog()
	sort.Slice(items, func(i, j int) bool {
		if items[i].Name != items[j].Name {
			return items[i].Name < items[j].Name
		}
		return items[i].ID < items[j].ID
	})

	var b strings.Builder
	b.WriteString("# Shelters\n\n")
	b.WriteString("Source: `internal/game/environment_resources.go` (`ShelterCatalog`).\n\n")
	b.WriteString(fmt.Sprintf("Total shelters: **%d**.\n\n", len(items)))
	b.WriteString("| ID | Name | Storage (kg) | Insulation | Rain | Wind | Insect | Durability Loss/Day | Build Cost (E/H2O) | Morale Bonus | Biomes |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |\n")
	for _, s := range items {
		b.WriteString("| ")
		b.WriteString(escape(string(s.ID)))
		b.WriteString(" | ")
		b.WriteString(escape(s.Name))
		b.WriteString(" | ")
		b.WriteString(formatFloat(s.StorageCapacityKg))
		b.WriteString(" | ")
		b.WriteString(strconv.Itoa(s.Insulation))
		b.WriteString(" | ")
		b.WriteString(strconv.Itoa(s.RainProtection))
		b.WriteString(" | ")
		b.WriteString(strconv.Itoa(s.WindProtection))
		b.WriteString(" | ")
		b.WriteString(strconv.Itoa(s.InsectProtection))
		b.WriteString(" | ")
		b.WriteString(strconv.Itoa(s.DurabilityPerDay))
		b.WriteString(" | ")
		b.WriteString(fmt.Sprintf("%d/%d", s.BuildEnergyCost, s.BuildHydrationCost))
		b.WriteString(" | ")
		b.WriteString(strconv.Itoa(s.BuildMoraleBonus))
		b.WriteString(" | ")
		b.WriteString(escape(strings.Join(s.BiomeTags, ", ")))
		b.WriteString(" |\n")
	}

	return docFile{Name: "shelters.md", Title: "Shelters", Content: b.String()}
}

func generateScenariosDoc() docFile {
	items := game.BuiltInScenarios()
	sort.Slice(items, func(i, j int) bool {
		if items[i].Name != items[j].Name {
			return items[i].Name < items[j].Name
		}
		return items[i].ID < items[j].ID
	})

	var b strings.Builder
	b.WriteString("# Scenarios\n\n")
	b.WriteString("Source: `internal/game/scenarios_builtin.go` (`BuiltInScenarios`).\n\n")
	b.WriteString(fmt.Sprintf("Total built-in scenarios: **%d**.\n\n", len(items)))
	b.WriteString("| ID | Name | Modes | Days | Location | Biome | Map (cells) | Default Season Set | Season Phases |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- | --- | --- |\n")
	for _, s := range items {
		modes := make([]string, 0, len(s.SupportedModes))
		for _, m := range s.SupportedModes {
			modes = append(modes, string(m))
		}
		b.WriteString("| ")
		b.WriteString(escape(string(s.ID)))
		b.WriteString(" | ")
		b.WriteString(escape(s.Name))
		b.WriteString(" | ")
		b.WriteString(escape(strings.Join(modes, ", ")))
		b.WriteString(" | ")
		b.WriteString(strconv.Itoa(s.DefaultDays))
		b.WriteString(" | ")
		b.WriteString(escape(s.Location))
		b.WriteString(" | ")
		b.WriteString(escape(s.Biome))
		b.WriteString(" | ")
		b.WriteString(fmt.Sprintf("%dx%d", s.MapWidthCells, s.MapHeightCells))
		b.WriteString(" | ")
		b.WriteString(escape(string(s.DefaultSeasonSetID)))
		b.WriteString(" | ")
		b.WriteString(escape(formatSeasonSets(s.SeasonSets)))
		b.WriteString(" |\n")
	}

	return docFile{Name: "scenarios.md", Title: "Scenarios", Content: b.String()}
}

func generateKitDoc() docFile {
	items := game.AllKitItems()
	sort.Slice(items, func(i, j int) bool {
		return string(items[i]) < string(items[j])
	})

	var b strings.Builder
	b.WriteString("# Kit Items\n\n")
	b.WriteString("Source: `internal/game/kit.go` (`AllKitItems`).\n\n")
	b.WriteString(fmt.Sprintf("Total kit items: **%d**.\n\n", len(items)))
	b.WriteString("| Item |\n")
	b.WriteString("| --- |\n")
	for _, k := range items {
		b.WriteString("| ")
		b.WriteString(escape(string(k)))
		b.WriteString(" |\n")
	}

	return docFile{Name: "kit-items.md", Title: "Kit Items", Content: b.String()}
}

func formatDiseaseRisks(items []game.DiseaseRisk) string {
	if len(items) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(items))
	for _, d := range items {
		parts = append(parts, fmt.Sprintf("%s (%.0f%%, %s)", d.Name, d.BaseChance*100, d.CarrierPart))
	}
	return strings.Join(parts, "; ")
}

func formatResourceReqs(items []game.ResourceRequirement) string {
	if len(items) == 0 {
		return ""
	}
	parts := make([]string, 0, len(items))
	for _, r := range items {
		parts = append(parts, fmt.Sprintf("%s %.2g", r.ID, r.Qty))
	}
	return strings.Join(parts, ", ")
}

func formatKitReqs(items []game.KitItem) string {
	if len(items) == 0 {
		return ""
	}
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, string(item))
	}
	return strings.Join(parts, ", ")
}

func formatSeasonSets(sets []game.SeasonSet) string {
	if len(sets) == 0 {
		return ""
	}
	parts := make([]string, 0, len(sets))
	for _, set := range sets {
		phases := make([]string, 0, len(set.Phases))
		for _, phase := range set.Phases {
			days := "end"
			if phase.Days > 0 {
				days = strconv.Itoa(phase.Days)
			}
			phases = append(phases, fmt.Sprintf("%s:%s", phase.Season, days))
		}
		parts = append(parts, fmt.Sprintf("%s[%s]", set.ID, strings.Join(phases, ",")))
	}
	return strings.Join(parts, "; ")
}

func formatFloat(v float64) string {
	if v == 0 {
		return "0"
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func escape(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	v = strings.ReplaceAll(v, "|", "\\|")
	v = strings.ReplaceAll(v, "\n", "<br>")
	return v
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
