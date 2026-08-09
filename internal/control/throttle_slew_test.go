package control

import "testing"

func TestThrottleSlew_RampsGradually(t *testing.T) {
	s := NewThrottleSlew(0.5) // 0.5/s
	const dt = 0.02

	out := s.Step(1.0, dt)
	if out >= 0.02 {
		t.Fatalf("first step should be small, got %f", out)
	}

	for i := 0; i < 100; i++ {
		out = s.Step(1.0, dt)
	}
	if out < 0.9 {
		t.Fatalf("expected near 1.0 after 2s, got %f", out)
	}
}

func TestThrottleSlew_Reset(t *testing.T) {
	s := NewThrottleSlew(1.0)
	s.Step(0.8, 0.1)
	s.Reset()
	if s.Out() != 0 {
		t.Fatalf("reset: got %f", s.Out())
	}
}
