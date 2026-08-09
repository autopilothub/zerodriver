package perception

import (
	"testing"
)

func TestLineDetector_CenteredLine(t *testing.T) {
	d := NewLineDetector(320, 240, 160, 40, 100)
	frame := make([]byte, 320*240*3)
	for i := range frame {
		frame[i] = 255
	}
	// Draw black line at center (x=160) in ROI
	for y := 160; y < 240; y++ {
		for x := 158; x <= 162; x++ {
			idx := (y*320 + x) * 3
			frame[idx], frame[idx+1], frame[idx+2] = 0, 0, 0
		}
	}
	pos := d.Detect(frame)
	if !pos.Detected {
		t.Fatal("line should be detected")
	}
	if pos.Offset < -0.05 || pos.Offset > 0.05 {
		t.Fatalf("expected ~0 offset, got %f", pos.Offset)
	}
}

func TestLineDetector_OffsetRight(t *testing.T) {
	d := NewLineDetector(320, 240, 160, 40, 100)
	frame := make([]byte, 320*240*3)
	for i := range frame {
		frame[i] = 255
	}
	for y := 160; y < 240; y++ {
		for x := 238; x <= 242; x++ {
			idx := (y*320 + x) * 3
			frame[idx], frame[idx+1], frame[idx+2] = 0, 0, 0
		}
	}
	pos := d.Detect(frame)
	if !pos.Detected {
		t.Fatal("line should be detected")
	}
	if pos.Offset <= 0 {
		t.Fatalf("expected positive offset, got %f", pos.Offset)
	}
}

func TestLineDetector_NoLine(t *testing.T) {
	d := NewLineDetector(320, 240, 160, 40, 100)
	frame := make([]byte, 320*240*3)
	for i := range frame {
		frame[i] = 255
	}
	pos := d.Detect(frame)
	if pos.Detected {
		t.Fatal("should not detect line on blank frame")
	}
}
