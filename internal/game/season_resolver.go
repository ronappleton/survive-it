package game

// CurrentSeason returns the season for the current run day.
// ok=false means the SeasonSetID wasn't found or the schedule is invalid/empty.
func (s *RunState) CurrentSeason() (season SeasonID, ok bool) {
	return s.SeasonForDay(s.Day)
}

// SeasonForDay returns the season for a given day number in the run (1-based).
func (s *RunState) SeasonForDay(day int) (SeasonID, bool) {
	if day < 1 {
		return "", false
	}

	set, ok := s.getSeasonSet(s.SeasonSetID)
	if !ok || len(set.Phases) == 0 {
		return "", false
	}

	remaining := day

	for _, phase := range set.Phases {
		// Days == 0 => until end.
		if phase.Days == 0 {
			return phase.Season, true
		}

		// Defensive: negative durations are invalid.
		if phase.Days < 0 {
			return "", false
		}

		if remaining <= phase.Days {
			return phase.Season, true
		}

		remaining -= phase.Days
	}

	// If we got here, schedule ended without an "until end" phase.
	// Fallback to the last phase's season if present (or treat it as invalid).
	last := set.Phases[len(set.Phases)-1]
	return last.Season, true
}

func (s *RunState) getSeasonSet(id SeasonSetID) (SeasonSet, bool) {
	for _, ss := range s.Scenario.SeasonSets {
		if ss.ID == id {
			return ss, true
		}
	}
	return SeasonSet{}, false
}
