package game

func clampSkill(value int) int {
	return clamp(value, 0, 100)
}

func sumTraitModifier(traits []TraitModifier) int {
	total := 0
	for _, trait := range traits {
		total += trait.Modifier
	}
	return total
}

func positiveTraitModifier(traits []TraitModifier) int {
	total := 0
	for _, trait := range traits {
		if trait.Positive || trait.Modifier > 0 {
			total += max(0, trait.Modifier)
		}
	}
	return total
}

func negativeTraitModifier(traits []TraitModifier) int {
	total := 0
	for _, trait := range traits {
		if !trait.Positive || trait.Modifier < 0 {
			total += min(0, trait.Modifier)
		}
	}
	return total
}

func applySkillEffort(skill *int, effort int, success bool) {
	if skill == nil {
		return
	}
	if effort < 1 {
		effort = 1
	}
	gain := 1 + effort/25
	if success {
		gain++
	}
	*skill = clampSkill(*skill + gain)
}
