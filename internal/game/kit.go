package game

type KitItem string

const (
	KitHatchet             KitItem = "Hatchet"
	KitSixInchKnife        KitItem = "6-Inch Knife"
	KitMachete             KitItem = "Machete"
	KitFoldingSaw          KitItem = "Folding Saw"
	KitParacord50ft        KitItem = "Paracord (50 ft)"
	KitFerroRod            KitItem = "Ferro Rod"
	KitFirePlunger         KitItem = "Fire Plunger"
	KitMagnifyingLens      KitItem = "Magnifying Lens"
	KitCookingPot          KitItem = "Cooking Pot"
	KitMetalCup            KitItem = "Metal Cup"
	KitCanteen             KitItem = "Canteen"
	KitWaterFilter         KitItem = "Water Filter"
	KitPurificationTablets KitItem = "Purification Tablets"
	KitFishingLineHooks    KitItem = "Fishing Line + Hooks"
	KitGillNet             KitItem = "Gill Net"
	KitSpear               KitItem = "Fishing Spear"
	KitSnareWire           KitItem = "Snare Wire"
	KitBowArrows           KitItem = "Bow + Arrows"
	KitTarp                KitItem = "Tarp"
	KitSleepingBag         KitItem = "Sleeping Bag"
	KitWoolBlanket         KitItem = "Wool Blanket"
	KitThermalLayer        KitItem = "Thermal Layer"
	KitRainJacket          KitItem = "Rain Jacket"
	KitMosquitoNet         KitItem = "Mosquito Net"
	KitInsectRepellent     KitItem = "Insect Repellent"
	KitCompass             KitItem = "Compass"
	KitMap                 KitItem = "Map"
	KitHeadlamp            KitItem = "Headlamp"
	KitSignalMirror        KitItem = "Signal Mirror"
	KitWhistle             KitItem = "Whistle"
	KitMultiTool           KitItem = "Multi-tool"
	KitDuctTape            KitItem = "Duct Tape"
	KitSewingKit           KitItem = "Sewing Kit"
	KitShovel              KitItem = "Shovel"
	KitClimbingRope        KitItem = "Climbing Rope"
	KitCarabiners          KitItem = "Carabiners"
	KitFirstAidKit         KitItem = "First Aid Kit"
	KitSalt                KitItem = "Salt"
	KitEmergencyRations    KitItem = "Emergency Rations"
	KitDryBag              KitItem = "Dry Bag"
)

func AllKitItems() []KitItem {
	return []KitItem{
		KitHatchet,
		KitSixInchKnife,
		KitMachete,
		KitFoldingSaw,
		KitParacord50ft,
		KitFerroRod,
		KitFirePlunger,
		KitMagnifyingLens,
		KitCookingPot,
		KitMetalCup,
		KitCanteen,
		KitWaterFilter,
		KitPurificationTablets,
		KitFishingLineHooks,
		KitGillNet,
		KitSpear,
		KitSnareWire,
		KitBowArrows,
		KitTarp,
		KitSleepingBag,
		KitWoolBlanket,
		KitThermalLayer,
		KitRainJacket,
		KitMosquitoNet,
		KitInsectRepellent,
		KitCompass,
		KitMap,
		KitHeadlamp,
		KitSignalMirror,
		KitWhistle,
		KitMultiTool,
		KitDuctTape,
		KitSewingKit,
		KitShovel,
		KitClimbingRope,
		KitCarabiners,
		KitFirstAidKit,
		KitSalt,
		KitEmergencyRations,
		KitDryBag,
	}
}
