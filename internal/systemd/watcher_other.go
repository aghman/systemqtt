//go:build !linux

package systemd

import (
	"context"
	"errors"

	"github.com/systemqtt/systemqtt/internal/config"
)

// ErrUnsupported indicates systemd integration is unavailable on this platform.
var ErrUnsupported = errors.New("systemd watcher is only available on Linux")

// UnitEvent carries a unit state snapshot.
type UnitEvent struct {
	Unit        string
	ActiveState string
	SubState    string
}

// Watcher is a no-op stub on non-Linux platforms.
type Watcher struct{}

// NewWatcher returns ErrUnsupported on non-Linux.
func NewWatcher(_ *config.Config) (*Watcher, error) {
	return nil, ErrUnsupported
}

// Run is a stub.
func (w *Watcher) Run(ctx context.Context, emit func(UnitEvent)) error {
	return ErrUnsupported
}

// Close is a stub.
func (w *Watcher) Close() error { return nil }

// ListMatchingUnits is a stub.
func ListMatchingUnits(_ *config.Config) ([]UnitEvent, error) {
	return nil, ErrUnsupported
}
