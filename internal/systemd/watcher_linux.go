//go:build linux

package systemd

import (
	"context"
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/systemqtt/systemqtt/internal/config"
)

// UnitEvent carries a unit state snapshot.
type UnitEvent struct {
	Unit        string
	ActiveState string
	SubState    string
}

// Watcher subscribes to systemd unit property changes.
type Watcher struct {
	cfg  *config.Config
	conn *dbus.Conn
	pats []string
}

// NewWatcher connects to the system bus and subscribes to systemd signals.
func NewWatcher(cfg *config.Config) (*Watcher, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("system dbus: %w", err)
	}
	w := &Watcher{
		cfg:  cfg,
		conn: conn,
		pats: SplitPatterns(cfg.Systemd.Unit.Filter),
	}
	return w, nil
}

// ListMatchingUnits returns current state for all units matching the filter.
func ListMatchingUnits(cfg *config.Config) ([]UnitEvent, error) {
	w, err := NewWatcher(cfg)
	if err != nil {
		return nil, err
	}
	defer w.conn.Close()
	return w.listUnits()
}

func (w *Watcher) listUnits() ([]UnitEvent, error) {
	obj := w.conn.Object("org.freedesktop.systemd1", dbus.ObjectPath("/org/freedesktop/systemd1"))
	call := obj.Call("org.freedesktop.systemd1.Manager.ListUnits", 0)
	if call.Err != nil {
		return nil, call.Err
	}
	raw, ok := call.Body[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("ListUnits: unexpected body")
	}
	var out []UnitEvent
	for _, row := range raw {
		tuple, ok := row.([]interface{})
		if !ok || len(tuple) < 5 {
			continue
		}
		name, _ := tuple[0].(string)
		active, _ := tuple[3].(string)
		sub, _ := tuple[4].(string)
		if !MatchUnit(name, w.pats) {
			continue
		}
		out = append(out, UnitEvent{Unit: name, ActiveState: active, SubState: sub})
	}
	return out, nil
}

// Run blocks until ctx is cancelled, emitting unit events on changes.
func (w *Watcher) Run(ctx context.Context, emit func(UnitEvent)) error {
	obj := w.conn.Object("org.freedesktop.systemd1", dbus.ObjectPath("/org/freedesktop/systemd1"))
	if call := obj.Call("org.freedesktop.systemd1.Manager.Subscribe", 0); call.Err != nil {
		return fmt.Errorf("Subscribe: %w", call.Err)
	}

	match := "type='signal',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged',path_namespace='/org/freedesktop/systemd1/unit'"
	if call := w.conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, match); call.Err != nil {
		return fmt.Errorf("AddMatch: %w", call.Err)
	}

	ch := make(chan *dbus.Signal, 128)
	w.conn.Signal(ch)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case sig := <-ch:
			if sig == nil {
				continue
			}
			if sig.Name != "org.freedesktop.DBus.Properties.PropertiesChanged" {
				continue
			}
			if len(sig.Body) < 2 {
				continue
			}
			iface, ok := sig.Body[0].(string)
			if !ok || iface != "org.freedesktop.systemd1.Unit" {
				continue
			}
			ev, ok := w.eventFromPath(sig.Path)
			if !ok {
				continue
			}
			emit(ev)
		}
	}
}

func (w *Watcher) eventFromPath(path dbus.ObjectPath) (UnitEvent, bool) {
	obj := w.conn.Object("org.freedesktop.systemd1", path)
	call := obj.Call("org.freedesktop.DBus.Properties.Get", 0, "org.freedesktop.systemd1.Unit", "Id")
	if call.Err != nil {
		return UnitEvent{}, false
	}
	var v dbus.Variant
	if err := call.Store(&v); err != nil {
		return UnitEvent{}, false
	}
	name, ok := v.Value().(string)
	if !ok || !MatchUnit(name, w.pats) {
		return UnitEvent{}, false
	}
	active, sub := w.readActiveSub(path)
	return UnitEvent{Unit: name, ActiveState: active, SubState: sub}, true
}

func (w *Watcher) readActiveSub(path dbus.ObjectPath) (string, string) {
	obj := w.conn.Object("org.freedesktop.systemd1", path)
	var av, sv dbus.Variant
	if c := obj.Call("org.freedesktop.DBus.Properties.Get", 0, "org.freedesktop.systemd1.Unit", "ActiveState"); c.Err == nil {
		_ = c.Store(&av)
	}
	if c := obj.Call("org.freedesktop.DBus.Properties.Get", 0, "org.freedesktop.systemd1.Unit", "SubState"); c.Err == nil {
		_ = c.Store(&sv)
	}
	a, _ := av.Value().(string)
	s, _ := sv.Value().(string)
	return a, s
}

// Close releases the bus connection.
func (w *Watcher) Close() error {
	if w.conn != nil {
		return w.conn.Close()
	}
	return nil
}

// DecodeUnitPath is exported for tests: maps a systemd object path fragment to a unit name.
func DecodeUnitPath(path dbus.ObjectPath) string {
	s := string(path)
	idx := strings.LastIndex(s, "/")
	if idx < 0 {
		return ""
	}
	frag := s[idx+1:]
	return dbusPathUnescape(frag)
}

func dbusPathUnescape(s string) string {
	s = strings.ReplaceAll(s, "_2d", "-")
	s = strings.ReplaceAll(s, "_2e", ".")
	s = strings.ReplaceAll(s, "_40", "@")
	return s
}
