package game

type CraftQuality string

const (
	CraftQualityPoor      CraftQuality = "poor"
	CraftQualityFair      CraftQuality = "fair"
	CraftQualityGood      CraftQuality = "good"
	CraftQualityExcellent CraftQuality = "excellent"
)

type CraftOutcome struct {
	Spec       CraftableSpec
	Quality    CraftQuality
	HoursSpent float64
	StoredAt   string
}

func qualityFromScore(score float64) CraftQuality {
	switch {
	case score >= 6.5:
		return CraftQualityExcellent
	case score >= 3.0:
		return CraftQualityGood
	case score >= 0.5:
		return CraftQualityFair
	default:
		return CraftQualityPoor
	}
}

func qualityTimeReduction(quality CraftQuality) float64 {
	switch quality {
	case CraftQualityExcellent:
		return 0.35
	case CraftQualityGood:
		return 0.2
	case CraftQualityPoor:
		return -0.2
	default:
		return 0
	}
}

func qualityCraftEffectBonus(quality CraftQuality) int {
	switch quality {
	case CraftQualityExcellent:
		return 2
	case CraftQualityGood:
		return 1
	case CraftQualityPoor:
		return -1
	default:
		return 0
	}
}

func qualityCatchBonus(quality CraftQuality) float64 {
	switch quality {
	case CraftQualityExcellent:
		return 0.12
	case CraftQualityGood:
		return 0.07
	case CraftQualityPoor:
		return -0.05
	default:
		return 0
	}
}
