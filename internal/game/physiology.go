package game

type PhysiologyProfile struct {
	EnergyDrainPerDay    int
	HydrationDrainPerDay int
	MoralDrainPerDay     int
	CarryModifier        float64
}

func PhysiologyFor(body BodyType) PhysiologyProfile {
	switch body {
	case BodyTypeMale:
		return PhysiologyProfile{EnergyDrainPerDay: 12, HydrationDrainPerDay: 10, MoralDrainPerDay: 1, CarryModifier: 1.05}
	case BodyTypeFemale:
		return PhysiologyProfile{EnergyDrainPerDay: 11, HydrationDrainPerDay: 9, MoralDrainPerDay: 1, CarryModifier: 1.00}
	default:
		return PhysiologyProfile{EnergyDrainPerDay: 11, HydrationDrainPerDay: 10, MoralDrainPerDay: 1, CarryModifier: 1.00}
	}
}
