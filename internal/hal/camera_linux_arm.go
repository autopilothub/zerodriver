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
	v4l2PixFmtRGB3  = webcam.PixelFormat(0x33424752) // 'RGB3'
	v4l2PixFmtBGR3  = webcam.PixelFormat(0x33524742) // 'BGR3'
)

type v4l2PixelMode int

const (
	v4l2ModeMJPEG v4l2PixelMode = iota
	v4l2ModeYUYV
	v4l2ModeRGB3
	v4l2ModeBGR3
)

// V4L2Camera captures frames from a V4L2 device.
// On Pi OS Bookworm+, /dev/video0 is often the raw unicam node and may not stream
// without libcamerify; use camera_backend: rpicam or auto instead.
type V4L2Camera struct {
	mu       sync.Mutex
	cam      *webcam.Webcam
	width    int // output width for line detector
	height   int // output height for line detector
	captureW int
	captureH int
	format   webcam.PixelFormat
	mode     v4l2PixelMode
}

// NewV4L2Camera opens the V4L2 device and configures capture format.
func NewV4L2Camera(device string, width, height int) (*V4L2Camera, error) {
	cam, err := webcam.Open(device)
	if err != nil {
		return nil, fmt.Errorf("open camera %q: %w", device, err)
	}

	format, mode := selectPixelFormat(cam.GetSupportedFormats())
	if format == 0 {
		cam.Close()
		return nil, fmt.Errorf("no supported pixel format on %q", device)
	}

	captureW, captureH := pickCaptureSize(cam, format, width, height)

	var actualFmt webcam.PixelFormat
	var actualW, actualH uint32
	if err := cam.SetBufferCount(2); err != nil {
		cam.Close()
		return nil, fmt.Errorf("set buffer count: %w", err)
	}
	actualFmt, actualW, actualH, err = cam.SetImageFormat(format, uint32(captureW), uint32(captureH))
	if err != nil {
		cam.Close()
		return nil, fmt.Errorf("set image format: %w", err)
	}

	c := &V4L2Camera{
		cam:      cam,
		width:    width,
		height:   height,
		captureW: int(actualW),
		captureH: int(actualH),
		format:   actualFmt,
		mode:     mode,
	}
	if c.captureW == 0 {
		c.captureW = captureW
	}
	if c.captureH == 0 {
		c.captureH = captureH
	}

	if err := cam.StartStreaming(); err != nil {
		cam.Close()
		return nil, fmt.Errorf("start streaming: %w", err)
	}

	return c, nil
}

func selectPixelFormat(formats map[webcam.PixelFormat]string) (webcam.PixelFormat, v4l2PixelMode) {
	if _, ok := formats[v4l2PixFmtMJPEG]; ok {
		return v4l2PixFmtMJPEG, v4l2ModeMJPEG
	}
	for f, desc := range formats {
		if strings.Contains(strings.ToUpper(desc), "MJPG") {
			return f, v4l2ModeMJPEG
		}
	}
	if _, ok := formats[v4l2PixFmtYUYV]; ok {
		return v4l2PixFmtYUYV, v4l2ModeYUYV
	}
	for f, desc := range formats {
		if strings.Contains(strings.ToUpper(desc), "YUYV") {
			return f, v4l2ModeYUYV
		}
	}
	if _, ok := formats[v4l2PixFmtRGB3]; ok {
		return v4l2PixFmtRGB3, v4l2ModeRGB3
	}
	if _, ok := formats[v4l2PixFmtBGR3]; ok {
		return v4l2PixFmtBGR3, v4l2ModeBGR3
	}
	for f := range formats {
		return f, v4l2ModeYUYV
	}
	return 0, v4l2ModeYUYV
}

func pickCaptureSize(cam *webcam.Webcam, format webcam.PixelFormat, wantW, wantH int) (int, int) {
	sizes := cam.GetSupportedFrameSizes(format)
	bestW, bestH := wantW, wantH
	bestScore := -1

	for _, s := range sizes {
		for _, wh := range enumerateFrameSizes(s) {
			w, h := wh[0], wh[1]
			score := abs(w-wantW) + abs(h-wantH)
			if bestScore < 0 || score < bestScore {
				bestScore = score
				bestW, bestH = w, h
			}
		}
	}

	if bestScore < 0 {
		if wantW > 0 && wantH > 0 {
			return wantW, wantH
		}
		return 640, 480
	}
	return bestW, bestH
}

func enumerateFrameSizes(s webcam.FrameSize) [][2]int {
	if s.StepWidth == 0 && s.StepHeight == 0 {
		return [][2]int{{int(s.MinWidth), int(s.MinHeight)}}
	}

	var out [][2]int
	for w := int(s.MinWidth); w <= int(s.MaxWidth); w += max(1, int(s.StepWidth)) {
		for h := int(s.MinHeight); h <= int(s.MaxHeight); h += max(1, int(s.StepHeight)) {
			out = append(out, [2]int{w, h})
			if len(out) >= 32 {
				return out
			}
		}
	}
	return out
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

	var rgb []byte
	switch c.mode {
	case v4l2ModeMJPEG:
		rgb, err = decodeMJPEGToRGB(frame, c.width, c.height)
	case v4l2ModeYUYV:
		rgb, err = yuyvToRGB(frame, c.captureW, c.captureH)
	case v4l2ModeRGB3:
		rgb, err = rgb24ToRGB(frame, c.captureW, c.captureH, false)
	case v4l2ModeBGR3:
		rgb, err = rgb24ToRGB(frame, c.captureW, c.captureH, true)
	default:
		rgb, err = yuyvToRGB(frame, c.captureW, c.captureH)
	}
	if err != nil {
		return nil, err
	}
	if c.captureW != c.width || c.captureH != c.height {
		rgb = scaleRGB(rgb, c.captureW, c.captureH, c.width, c.height)
	}
	return rgb, nil
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
	srcW := bounds.Dx()
	srcH := bounds.Dy()
	if srcW == 0 || srcH == 0 {
		return nil, fmt.Errorf("empty jpeg image")
	}

	for y := 0; y < height; y++ {
		sy := bounds.Min.Y + y*srcH/height
		for x := 0; x < width; x++ {
			sx := bounds.Min.X + x*srcW/width
			r, g, b, _ := img.At(sx, sy).RGBA()
			idx := (y*width + x) * 3
			rgb[idx] = byte(r >> 8)
			rgb[idx+1] = byte(g >> 8)
			rgb[idx+2] = byte(b >> 8)
		}
	}
	return rgb, nil
}

func rgb24ToRGB(data []byte, width, height int, bgr bool) ([]byte, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid rgb size %dx%d", width, height)
	}
	rgb := make([]byte, width*height*3)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcIdx := (y*width + x) * 3
			if srcIdx+2 >= len(data) {
				break
			}
			dstIdx := srcIdx
			if bgr {
				rgb[dstIdx] = data[srcIdx+2]
				rgb[dstIdx+1] = data[srcIdx+1]
				rgb[dstIdx+2] = data[srcIdx]
			} else {
				copy(rgb[dstIdx:dstIdx+3], data[srcIdx:srcIdx+3])
			}
		}
	}
	return rgb, nil
}

func scaleRGB(src []byte, srcW, srcH, dstW, dstH int) []byte {
	if srcW == dstW && srcH == dstH {
		return src
	}
	dst := make([]byte, dstW*dstH*3)
	for y := 0; y < dstH; y++ {
		sy := y * srcH / dstH
		for x := 0; x < dstW; x++ {
			sx := x * srcW / dstW
			srcIdx := (sy*srcW + sx) * 3
			dstIdx := (y*dstW + x) * 3
			if srcIdx+2 < len(src) {
				copy(dst[dstIdx:dstIdx+3], src[srcIdx:srcIdx+3])
			}
		}
	}
	return dst
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

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
