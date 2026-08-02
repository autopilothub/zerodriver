package perception

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWritePPM(t *testing.T) {
	rgb := make([]byte, 4*4*3)
	for i := 0; i < len(rgb); i += 3 {
		rgb[i], rgb[i+1], rgb[i+2] = 255, 0, 0
	}
	path := filepath.Join(t.TempDir(), "test.ppm")
	if err := WritePPM(path, rgb, 4, 4); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Size() < 50 {
		t.Fatalf("ppm file too small: %v", err)
	}
}
