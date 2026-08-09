package perception

import (
	"time"

	"github.com/autopilothub/zerodriver/internal/domain"
)

// LineDetector extracts line position from camera frames.
type LineDetector struct {
	width         int
	height        int
	roiY          int
	lookaheadRows int
	threshold     int
}

func NewLineDetector(width, height, roiY, lookaheadRows, threshold int) *LineDetector {
	if threshold <= 0 {
		threshold = 100
	}
	if lookaheadRows <= 0 {
		lookaheadRows = 40
	}
	return &LineDetector{
		width: width, height: height, roiY: roiY,
		lookaheadRows: lookaheadRows, threshold: threshold,
	}
}

// Detect finds the line centroid in a raw RGB frame.
// Returns offset at bottom ROI and lookahead row for Pure Pursuit.
func (d *LineDetector) Detect(frame []byte) domain.LinePosition {
	now := time.Now()
	if len(frame) == 0 {
		return domain.LinePosition{Detected: false, Timestamp: now}
	}

	bottomOff, bottomN := d.offsetInBand(frame, d.roiY, d.height)
	lookY := d.roiY - d.lookaheadRows
	if lookY < 0 {
		lookY = 0
	}
	lookEnd := lookY + 8
	if lookEnd > d.roiY {
		lookEnd = d.roiY
	}
	lookOff, lookN := d.offsetInBand(frame, lookY, lookEnd)

	if bottomN == 0 && lookN == 0 {
		return domain.LinePosition{Detected: false, Timestamp: now}
	}

	offset := bottomOff
	if bottomN == 0 {
		offset = lookOff
	}
	lookahead := lookOff
	if lookN == 0 {
		lookahead = offset
	}

	return domain.LinePosition{
		Offset:          offset,
		LookaheadOffset: lookahead,
		Detected:        true,
		Timestamp:       now,
	}
}

func (d *LineDetector) offsetInBand(frame []byte, yStart, yEnd int) (offset float64, count int) {
	if yStart < 0 {
		yStart = 0
	}
	if yEnd > d.height {
		yEnd = d.height
	}

	var sumX float64
	for y := yStart; y < yEnd; y++ {
		for x := 0; x < d.width; x++ {
			idx := (y*d.width + x) * 3
			if idx+2 >= len(frame) {
				break
			}
			r, g, b := frame[idx], frame[idx+1], frame[idx+2]
			if int(r)+int(g)+int(b) < d.threshold {
				sumX += float64(x)
				count++
			}
		}
	}

	if count == 0 {
		return 0, 0
	}

	centroid := sumX / float64(count)
	center := float64(d.width) / 2
	off := (centroid - center) / center
	if off < -1 {
		off = -1
	}
	if off > 1 {
		off = 1
	}
	return off, count
}
