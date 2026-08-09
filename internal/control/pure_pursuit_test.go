package control

import (
	"math"
	"testing"
)

func TestPurePursuit_Center(t *testing.T) {
	pp := NewPurePursuit(0.35, 1.0)
	if s := pp.Steering(0); math.Abs(s) > 0.001 {
		t.Fatalf("center steering=%f", s)
	}
}

func TestPurePursuit_TurnsTowardLine(t *testing.T) {
	pp := NewPurePursuit(0.35, 1.0)
	left := pp.Steering(-0.5)
	right := pp.Steering(0.5)
	if left >= 0 {
		t.Fatalf("line left should steer left (negative), got %f", left)
	}
	if right <= 0 {
		t.Fatalf("line right should steer right (positive), got %f", right)
	}
}
