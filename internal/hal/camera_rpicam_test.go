//go:build linux && arm

package hal

import (
	"testing"
)

func TestPickRpicamCaptureSize(t *testing.T) {
	w, h := pickRpicamCaptureSize(320, 240)
	if w != 640 || h != 480 {
		t.Fatalf("got %dx%d, want 640x480", w, h)
	}

	w, h = pickRpicamCaptureSize(800, 600)
	if w != 800 || h != 600 {
		t.Fatalf("got %dx%d, want 800x600", w, h)
	}
}

func TestYUV420ToRGB(t *testing.T) {
	// 2x2 gray: Y=128, U=128, V=128 -> neutral gray RGB
	data := []byte{
		128, 128,
		128, 128,
		128, 128, 128, 128,
	}
	rgb, err := yuv420ToRGB(data, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(rgb) != 2*2*3 {
		t.Fatalf("len %d", len(rgb))
	}
	for i := 0; i < len(rgb); i += 3 {
		if rgb[i] < 120 || rgb[i] > 136 {
			t.Fatalf("unexpected gray pixel at %d: %v", i, rgb[i:i+3])
		}
	}
}
