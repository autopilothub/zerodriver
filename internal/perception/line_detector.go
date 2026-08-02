package perception

import (
	"time"

	"github.com/autopilothub/zerodriver/internal/domain"
)

// LineDetector extracts line position from camera frames.
type LineDetector struct {
	width  int
	height int
	roiY   int
}

func NewLineDetector(width, height, roiY int) *LineDetector {
	return &LineDetector{width: width, height: height, roiY: roiY}
}

// Detect finds the line centroid in a raw RGB frame.
// Returns offset normalized to [-1.0, +1.0] where 0 = centered.
func (d *LineDetector) Detect(frame []byte) domain.LinePosition {
	now := time.Now()
	if len(frame) == 0 {
		return domain.LinePosition{Detected: false, Timestamp: now}
	}

	var sumX, count float64
	for y := d.roiY; y < d.height; y++ {
		for x := 0; x < d.width; x++ {
			idx := (y*d.width + x) * 3
			if idx+2 >= len(frame) {
				break
			}
			r, g, b := frame[idx], frame[idx+1], frame[idx+2]
			// Dark pixel threshold (black line on white surface)
			if int(r)+int(g)+int(b) < 100 {
				sumX += float64(x)
				count++
			}
		}
	}

	if count == 0 {
		return domain.LinePosition{Detected: false, Timestamp: now}
	}

	centroid := sumX / count
	center := float64(d.width) / 2
	offset := (centroid - center) / center

	if offset < -1 {
		offset = -1
	}
	if offset > 1 {
		offset = 1
	}

	return domain.LinePosition{
		Offset:    offset,
		Detected:  true,
		Timestamp: now,
	}
}
