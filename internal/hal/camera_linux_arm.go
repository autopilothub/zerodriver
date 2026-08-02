//go:build linux && arm

package hal

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"strings"
	"sync"

	"github.com/blackjack/webcam"
)

// V4L2 pixel format fourCC codes.
const (
	v4l2PixFmtMJPEG = webcam.PixelFormat(0x47504A50) // 'MJPG'
	v4l2PixFmtYUYV  = webcam.PixelFormat(0x56595559) // 'YUYV'
)

// V4L2Camera captures frames from a V4L2 device (Pi Camera via libcamera-compat).
type V4L2Camera struct {
	mu       sync.Mutex
	cam      *webcam.Webcam
	width    int
	height   int
	format   webcam.PixelFormat
	isMJPEG  bool
}

// NewV4L2Camera opens the V4L2 device and configures capture format.
func NewV4L2Camera(device string, width, height int) (*V4L2Camera, error) {
	cam, err := webcam.Open(device)
	if err != nil {
		return nil, fmt.Errorf("open camera %q: %w", device, err)
	}

	format, isMJPEG := selectPixelFormat(cam.GetSupportedFormats())
	if format == 0 {
		cam.Close()
		return nil, fmt.Errorf("no supported pixel format on %q", device)
	}

	actualFmt, actualW, actualH, err := cam.SetImageFormat(format, uint32(width), uint32(height))
	if err != nil {
		cam.Close()
		return nil, fmt.Errorf("set image format: %w", err)
	}

	_ = cam.SetBufferCount(2)

	c := &V4L2Camera{
		cam:     cam,
		width:   int(actualW),
		height:  int(actualH),
		format:  actualFmt,
		isMJPEG: isMJPEG,
	}

	if c.width == 0 {
		c.width = width
	}
	if c.height == 0 {
		c.height = height
	}

	if err := cam.StartStreaming(); err != nil {
		cam.Close()
		return nil, fmt.Errorf("start streaming: %w", err)
	}

	return c, nil
}

func selectPixelFormat(formats map[webcam.PixelFormat]string) (webcam.PixelFormat, bool) {
	// Prefer MJPEG for lower CPU on Pi Zero W.
	if _, ok := formats[v4l2PixFmtMJPEG]; ok {
		return v4l2PixFmtMJPEG, true
	}
	for f, desc := range formats {
		if strings.Contains(strings.ToUpper(desc), "MJPG") {
			return f, true
		}
	}
	if _, ok := formats[v4l2PixFmtYUYV]; ok {
		return v4l2PixFmtYUYV, false
	}
	for f, desc := range formats {
		if strings.Contains(strings.ToUpper(desc), "YUYV") {
			return f, false
		}
	}
	for f := range formats {
		return f, false
	}
	return 0, false
}

func (c *V4L2Camera) Capture() ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.cam.WaitForFrame(500); err != nil {
		return nil, fmt.Errorf("wait for frame: %w", err)
	}

	frame, err := c.cam.ReadFrame()
	if err != nil {
		return nil, fmt.Errorf("read frame: %w", err)
	}
	if len(frame) == 0 {
		return nil, fmt.Errorf("empty frame")
	}

	if c.isMJPEG {
		return decodeMJPEGToRGB(frame, c.width, c.height)
	}
	return yuyvToRGB(frame, c.width, c.height)
}

func (c *V4L2Camera) Width() int  { return c.width }
func (c *V4L2Camera) Height() int { return c.height }

func (c *V4L2Camera) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cam == nil {
		return nil
	}
	_ = c.cam.StopStreaming()
	return c.cam.Close()
}

func decodeMJPEGToRGB(data []byte, width, height int) ([]byte, error) {
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode jpeg: %w", err)
	}

	rgb := make([]byte, width*height*3)
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y && y < height; y++ {
		for x := bounds.Min.X; x < bounds.Max.X && x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			idx := (y*width + x) * 3
			rgb[idx] = byte(r >> 8)
			rgb[idx+1] = byte(g >> 8)
			rgb[idx+2] = byte(b >> 8)
		}
	}
	return rgb, nil
}

func yuyvToRGB(data []byte, width, height int) ([]byte, error) {
	rgb := make([]byte, width*height*3)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x += 2 {
			idx := (y*width + x) * 2
			if idx+3 >= len(data) {
				break
			}
			y0, u, y1, v := float64(data[idx]), float64(data[idx+1]), float64(data[idx+2]), float64(data[idx+3])
			writeYUV(rgb, (y*width+x)*3, y0, u, v)
			if x+1 < width {
				writeYUV(rgb, (y*width+x+1)*3, y1, u, v)
			}
		}
	}
	return rgb, nil
}

func writeYUV(rgb []byte, idx int, y, u, v float64) {
	if idx+2 >= len(rgb) {
		return
	}
	r := y + 1.402*(v-128)
	g := y - 0.344136*(u-128) - 0.714136*(v-128)
	b := y + 1.772*(u-128)
	rgb[idx] = clampByte(r)
	rgb[idx+1] = clampByte(g)
	rgb[idx+2] = clampByte(b)
}

func clampByte(v float64) byte {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return byte(v)
}
