package web

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
)

// RGBToJPEG encodes an interleaved RGB buffer as JPEG.
func RGBToJPEG(rgb []byte, width, height, quality int) ([]byte, error) {
	if quality <= 0 {
		quality = 70
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		row := y * width * 3
		for x := 0; x < width; x++ {
			i := row + x*3
			img.SetRGBA(x, y, color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
