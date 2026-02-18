package game

import "time"

func (s *RunState) ApplyRealtimeMetabolism(elapsed time.Duration, dayDuration time.Duration) {
	if s == nil || dayDuration <= 0 {
		return
	}
	s.EnsurePlayerRuntimeStats()

	target := clampFloat(float64(elapsed)/float64(dayDuration), 0, 1)
	delta := target - s.MetabolismProgress
	if delta <= 0 {
		return
	}

	for i := range s.Players {
		applyMetabolismFraction(&s.Players[i], delta)
	}
	s.MetabolismProgress = target
}

func (s *RunState) consumePendingDayMetabolism() {
	if s == nil {
		return
	}
	remaining := 1.0 - s.MetabolismProgress
	if remaining > 0 {
		for i := range s.Players {
			applyMetabolismFraction(&s.Players[i], remaining)
		}
	}
	s.MetabolismProgress = 0
}
