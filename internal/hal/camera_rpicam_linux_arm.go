//go:build linux && arm

package hal

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// RpicamCamera captures frames via rpicam-vid/libcamera-vid (libcamera stack).
// Use this on Pi OS Bookworm+ where /dev/video0 is the raw unicam sensor node.
type RpicamCamera struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	reader *bufio.Reader
	width  int
	height int
}

func findRpicamBinary() (string, error) {
	for _, name := range []string{"rpicam-vid", "libcamera-vid"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("rpicam-vid not found (install: sudo apt install -y rpicam-apps)")
}

// NewRpicamCamera starts rpicam-vid and reads MJPEG frames from stdout.
func NewRpicamCamera(width, height int) (*RpicamCamera, error) {
	bin, err := findRpicamBinary()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(bin,
		"-t", "0",
		"-n",
		"--width", fmt.Sprintf("%d", width),
		"--height", fmt.Sprintf("%d", height),
		"--codec", "mjpeg",
		"--flush",
		"-o", "-",
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start %s: %w", bin, err)
	}

	c := &RpicamCamera{
		cmd:    cmd,
		reader: bufio.NewReader(stdout),
		width:  width,
		height: height,
	}

	if _, err := c.readMJPEGFrame(); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("rpicam first frame: %w", err)
	}

	return c, nil
}

func (c *RpicamCamera) Capture() ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	jpeg, err := c.readMJPEGFrame()
	if err != nil {
		return nil, fmt.Errorf("read mjpeg: %w", err)
	}
	return decodeMJPEGToRGB(jpeg, c.width, c.height)
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
