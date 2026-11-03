package config

import (
	"testing"
	"time"
)

func TestGetOrAndIntDuration(t *testing.T) {
	t.Setenv("DEMO_PORT", "9090")
	t.Setenv("DEMO_TIMEOUT", "2s")

	if GetOr("MISSING", "x") != "x" {
		t.Fatal("fallback failed")
	}
	if Int("DEMO_PORT", 80) != 9090 {
		t.Fatal("Int failed")
	}
	if Duration("DEMO_TIMEOUT", time.Second) != 2*time.Second {
		t.Fatal("Duration failed")
	}
}
