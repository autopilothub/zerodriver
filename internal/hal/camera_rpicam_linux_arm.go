//go:build linux && arm

package hal

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

const rpicamRunTimeoutMS = "86400000" // 24h; avoid -t 0 (instant exit on some builds)

// RpicamCamera captures frames via rpicam-vid/libcamera-vid (libcamera stack).
// Use this on Pi OS Bookworm+ where /dev/video0 is the raw unicam sensor node.
type RpicamCamera struct {
	mu       sync.Mutex
	cmd      *exec.Cmd
	reader   *bufio.Reader
	width    int // output width for line detector
	height   int // output height for line detector
	captureW int
	captureH int
	codec    string // "yuv420" or "mjpeg"
}

func findRpicamBinary() (string, error) {
	for _, name := range []string{"rpicam-vid", "libcamera-vid"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("rpicam-vid not found (install: sudo apt install -y rpicam-apps)")
}

// pickRpicamCaptureSize snaps to a sensor-friendly size (OV5647 min mode is 640x480).
func pickRpicamCaptureSize(wantW, wantH int) (int, int) {
	if wantW < 640 || wantH < 480 {
		return 640, 480
	}
	return wantW, wantH
}

// NewRpicamCamera starts rpicam-vid and reads frames from stdout.
func NewRpicamCamera(width, height int) (*RpicamCamera, error) {
	bin, err := findRpicamBinary()
	if err != nil {
		return nil, err
	}

	captureW, captureH := pickRpicamCaptureSize(width, height)
	var lastErr error
	for _, codec := range []string{"yuv420", "mjpeg"} {
		cam, err := startRpicamCamera(bin, width, height, captureW, captureH, codec)
		if err == nil {
			return cam, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func startRpicamCamera(bin string, outW, outH, captureW, captureH int, codec string) (*RpicamCamera, error) {
	args := []string{
		"-n",
		"-t", rpicamRunTimeoutMS,
		"--width", fmt.Sprintf("%d", captureW),
		"--height", fmt.Sprintf("%d", captureH),
		"--framerate", "15",
		"--codec", codec,
		"--inline",
		"--flush",
		"-o", "-",
	}
	if captureW == 640 && captureH == 480 {
		args = append(args, "--mode", "640:480:10")
	}

	cmd := exec.Command(bin, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start %s: %w", bin, err)
	}

	c := &RpicamCamera{
		cmd:      cmd,
		reader:   bufio.NewReader(stdout),
		width:    outW,
		height:   outH,
		captureW: captureW,
		captureH: captureH,
		codec:    codec,
	}

	if err := c.readFirstFrame(); err != nil {
		_ = c.Close()
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			if len(msg) > 300 {
				msg = msg[len(msg)-300:]
			}
			return nil, fmt.Errorf("rpicam first frame (%s %dx%d): %w; stderr: %s", codec, captureW, captureH, err, msg)
		}
		return nil, fmt.Errorf("rpicam first frame (%s %dx%d): %w", codec, captureW, captureH, err)
	}

	return c, nil
}

func (c *RpicamCamera) readFirstFrame() error {
	if c.codec == "yuv420" {
		_, err := c.readYUV420Frame()
		return err
	}
	_, err := c.readMJPEGFrame()
	return err
}

func (c *RpicamCamera) Capture() ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var rgb []byte
	var err error
	switch c.codec {
	case "yuv420":
		frame, err := c.readYUV420Frame()
		if err != nil {
			return nil, fmt.Errorf("read yuv420: %w", err)
		}
		rgb, err = yuv420ToRGB(frame, c.captureW, c.captureH)
	default:
		jpeg, err := c.readMJPEGFrame()
		if err != nil {
			return nil, fmt.Errorf("read mjpeg: %w", err)
		}
		rgb, err = decodeMJPEGToRGB(jpeg, c.captureW, c.captureH)
	}
	if err != nil {
		return nil, err
	}
	if c.captureW != c.width || c.captureH != c.height {
		rgb = scaleRGB(rgb, c.captureW, c.captureH, c.width, c.height)
	}
	return rgb, nil
}

func (c *RpicamCamera) Width() int  { return c.width }
func (c *RpicamCamera) Height() int { return c.height }

func (c *RpicamCamera) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cmd == nil || c.cmd.Process == nil {
		return nil
	}
	_ = c.cmd.Process.Kill()
	_ = c.cmd.Wait()
	c.cmd = nil
	return nil
}

func (c *RpicamCamera) readYUV420Frame() ([]byte, error) {
	size := c.captureW * c.captureH * 3 / 2
	buf := make([]byte, size)
	if _, err := io.ReadFull(c.reader, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func (c *RpicamCamera) readMJPEGFrame() ([]byte, error) {
	var buf bytes.Buffer

	for {
		b, err := c.reader.ReadByte()
		if err != nil {
			return nil, err
		}
		if b != 0xFF {
			continue
		}
		b2, err := c.reader.ReadByte()
		if err != nil {
			return nil, err
		}
		if b2 == 0xD8 {
			buf.WriteByte(0xFF)
			buf.WriteByte(0xD8)
			break
		}
	}

	var prev byte
	for {
		b, err := c.reader.ReadByte()
		if err != nil {
			return nil, err
		}
		buf.WriteByte(b)
		if prev == 0xFF && b == 0xD9 {
			break
		}
		prev = b
	}

	return buf.Bytes(), nil
}

func yuv420ToRGB(data []byte, width, height int) ([]byte, error) {
	want := width * height * 3 / 2
	if len(data) < want {
		return nil, fmt.Errorf("short yuv420 frame: got %d want %d", len(data), want)
	}

	rgb := make([]byte, width*height*3)
	ySize := width * height
	uvStride := width / 2
	uPlane := data[ySize : ySize+ySize/4]
	vPlane := data[ySize+ySize/4 : ySize+ySize/2]

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			yi := y*width + x
			uvIdx := (y/2)*uvStride + x/2
			writeYUV(rgb, yi*3, float64(data[yi]), float64(uPlane[uvIdx]), float64(vPlane[uvIdx]))
		}
	}
	return rgb, nil
}
