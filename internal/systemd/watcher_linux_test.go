//go:build linux

package systemd

import "testing"

func TestListUnitsRows(t *testing.T) {
	t.Parallel()
	row := []interface{}{"foo.service", "", "", "active", "running"}
	t.Run("matrix form from godbus", func(t *testing.T) {
		t.Parallel()
		in := [][]interface{}{row}
		got, err := listUnitsRows(in)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0][0] != "foo.service" {
			t.Fatalf("got %#v", got)
		}
	})
	t.Run("slice of interface rows", func(t *testing.T) {
		t.Parallel()
		in := []interface{}{row}
		got, err := listUnitsRows(in)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0][0] != "foo.service" {
			t.Fatalf("got %#v", got)
		}
	})
	t.Run("reject non-row element", func(t *testing.T) {
		t.Parallel()
		in := []interface{}{"not-a-tuple"}
		_, err := listUnitsRows(in)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
