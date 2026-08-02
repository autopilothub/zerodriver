package perception

import (
	"fmt"
	"os"
)

// WritePPM saves an RGB frame as a PPM image (viewable with most image viewers).
func WritePPM(path string, rgb []byte, width, height int) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	header := fmt.Sprintf("P6\n%d %d\n255\n", width, height)
	if _, err := f.WriteString(header); err != nil {
		return err
	}
	if len(rgb) < width*height*3 {
		return fmt.Errorf("rgb buffer too small: got %d, need %d", len(rgb), width*height*3)
	}
	_, err = f.Write(rgb[:width*height*3])
	return err
}
