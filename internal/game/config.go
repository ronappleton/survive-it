package game

import (
	"errors"
)

type GameMode string

const (
	ModeNakedAndAfraid GameMode = "naked_and_afraid"
	ModelAlive         GameMode = "alone"
)

type RunConfig struct {
	Mode GameMode
}

func (c RunConfig) Validate() error {
	switch c.Mode {
	case ModeNakedAndAfraid:
	case ModelAlive:
	default:
		return
	}
}
