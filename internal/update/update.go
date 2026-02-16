package update

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
)

// Alpha1: updater is "wired" but intentionally conservative.
//
// How it works (when enabled):
// - Reads repo owner/name from env (SURVIVEIT_REPO="owner/name").
// - Compares current version with latest GitHub Release tag.
// - (Future) Downloads the right binary asset, verifies checksum, swaps binary.
//
// For alpha1 we only do a very lightweight check and return a message.

type CheckParams struct {
	CurrentVersion string
}

func Check(p CheckParams) (string, error) {
	repo := strings.TrimSpace(os.Getenv("SURVIVEIT_REPO"))
	if repo == "" {
		return "Self-update not configured. Set SURVIVEIT_REPO=owner/name after you publish releases.", nil
	}

	// In alpha1 we don't call GitHub yet (keeps early builds dependency-free and avoids rate-limit auth).
	// We still return something useful.
	return fmt.Sprintf("Self-update configured for %s (%s/%s). Implement release checks in alpha2.", repo, runtime.GOOS, runtime.GOARCH), nil
}

var ErrNotImplemented = errors.New("self-update not implemented")
