package hal

import "testing"

func TestThrottleToPulseUs_Linear(t *testing.T) {
	cfg := ThrottlePulseConfig{MinUs: 1000, MaxUs: 2000, Map: ThrottleMapLinear}
	if got := ThrottleToPulseUs(0, cfg); got != 1000 {
		t.Fatalf("stop: %d", got)
	}
	if got := ThrottleToPulseUs(0.5, cfg); got != 1500 {
		t.Fatalf("half: %d", got)
	}
	if got := ThrottleToPulseUs(1, cfg); got != 2000 {
		t.Fatalf("max: %d", got)
	}
}

func TestThrottleToPulseUs_Bidirectional(t *testing.T) {
	cfg := ThrottlePulseConfig{
		NeutralUs: 1500, MaxUs: 2000, ReverseUs: 1000, Map: ThrottleMapBidirectional,
	}
	if got := ThrottleToPulseUs(-1, cfg); got != 1000 {
		t.Fatalf("reverse: %d", got)
	}
	if got := ThrottleToPulseUs(0, cfg); got != 1500 {
		t.Fatalf("neutral: %d", got)
	}
	if got := ThrottleToPulseUs(1, cfg); got != 2000 {
		t.Fatalf("forward: %d", got)
	}
}

func TestThrottleToPulseUs_BidirectionalReverseStart(t *testing.T) {
	cfg := ThrottlePulseConfig{
		NeutralUs: 1500, MaxUs: 2000, ReverseUs: 1000,
		ReverseStartUs: 1400, Map: ThrottleMapBidirectional,
	}
	if got := ThrottleToPulseUs(-0.1, cfg); got != 1400 {
		t.Fatalf("reverse deadband: %d", got)
	}
}

func TestThrottleToPulseUs_ForwardStart(t *testing.T) {
	cfg := ThrottlePulseConfig{
		NeutralUs: 1500, MaxUs: 2000, Map: ThrottleMapNeutralForward, ForwardStartUs: 1600,
	}
	if got := ThrottleToPulseUs(0.1, cfg); got != 1600 {
		t.Fatalf("deadband: %d", got)
	}
}
