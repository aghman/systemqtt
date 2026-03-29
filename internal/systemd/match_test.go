package systemd

import "testing"

func TestMatchUnit(t *testing.T) {
	if !MatchUnit("nginx.service", []string{"*.service"}) {
		t.Fatal("expected glob match")
	}
	if MatchUnit("nginx.service", []string{"*.timer"}) {
		t.Fatal("expected no match")
	}
	if !MatchUnit("anything.service", nil) {
		t.Fatal("empty patterns should match all")
	}
}

func TestSplitPatterns(t *testing.T) {
	p := SplitPatterns("a.service, *.timer ")
	if len(p) != 2 {
		t.Fatalf("got %d", len(p))
	}
}
