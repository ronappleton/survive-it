package game

import "strings"

func hasPersonalItem(player PlayerState, id string) bool {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		return false
	}
	for _, item := range player.PersonalItems {
		if strings.ToLower(strings.TrimSpace(item.ID)) == id && item.Qty > 0 {
			return true
		}
	}
	return false
}

func hasAnyKitItem(player PlayerState, issued []KitItem, target KitItem) bool {
	if slicesContainsKit(player.Kit, target) {
		return true
	}
	return slicesContainsKit(issued, target)
}

func (s *RunState) applyCraftedWeatherModifiersForPlayer(impact statDelta, player PlayerState, weather WeatherType, tempC int) statDelta {
	out := impact

	hasHideJacket := hasPersonalItem(player, "hide_jacket")
	hasGrassCape := hasPersonalItem(player, "grass_cape")
	hasWovenTunic := hasPersonalItem(player, "woven_tunic")
	hasMoccasins := hasPersonalItem(player, "hide_moccasins") || hasPersonalItem(player, "bast_sandals")

	hasThermal := hasAnyKitItem(player, s.Config.IssuedKit, KitThermalLayer)
	hasRainJacket := hasAnyKitItem(player, s.Config.IssuedKit, KitRainJacket)
	hasWool := hasAnyKitItem(player, s.Config.IssuedKit, KitWoolBlanket)

	if tempC <= 2 {
		if hasHideJacket {
			out.Energy++
			out.Morale++
		}
		if hasThermal || hasWool {
			out.Energy++
		}
		if hasMoccasins {
			out.Energy++
		}
	}

	if tempC >= 32 {
		if hasGrassCape || hasWovenTunic {
			out.Hydration++
			out.Morale++
		}
		if hasMoccasins {
			out.Energy++
		}
	}

	if isRainyWeather(weather) {
		if hasRainJacket || hasGrassCape || hasHideJacket {
			out.Energy++
			out.Morale++
		}
		if hasWovenTunic {
			out.Morale++
		}
	}
	if isSevereWeather(weather) {
		if hasHideJacket || hasRainJacket {
			out.Energy++
		}
	}

	return out
}
