// Camera + line detection diagnostic tool.
// Usage: go run ./cmd/camtest -config configs/zerodriver-hardware.yaml -save frame.ppm
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/autopilothub/zerodriver/internal/config"
	"github.com/autopilothub/zerodriver/internal/hal"
	"github.com/autopilothub/zerodriver/internal/perception"
)

func main() {
	configPath := flag.String("config", "configs/zerodriver.yaml", "config file")
	mode := flag.String("mode", "hardware", "mode: mock or hardware")
	save := flag.String("save", "", "save last frame as PPM")
	duration := flag.Duration("duration", 0, "run duration (0 = until Ctrl+C)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	cfg.Mode = *mode

	camera, err := hal.NewCamera(cfg)
	if err != nil {
		log.Fatalf("init camera: %v", err)
	}
	defer camera.Close()

	detector := perception.NewLineDetector(cfg.Camera.Width, cfg.Camera.Height, cfg.Camera.ROIY, cfg.Camera.LineThreshold)

	log.Printf("camera %s %dx%d ROI_y=%d (mode=%s)",
		cfg.Hardware.CameraDevice, cfg.Camera.Width, cfg.Camera.Height, cfg.Camera.ROIY, cfg.Mode)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	deadline := time.Now().Add(*duration)
	if *duration == 0 {
		deadline = time.Time{}
	}

	var lastFrame []byte

	for {
		select {
		case <-sigCh:
			if *save != "" && len(lastFrame) > 0 {
				saveFrame(*save, lastFrame, cfg.Camera.Width, cfg.Camera.Height)
			}
			log.Println("stopped")
			return
		case <-ticker.C:
			if !deadline.IsZero() && time.Now().After(deadline) {
				if *save != "" && len(lastFrame) > 0 {
					saveFrame(*save, lastFrame, cfg.Camera.Width, cfg.Camera.Height)
				}
				log.Println("done")
				return
			}
			frame, err := camera.Capture()
			if err != nil {
				log.Printf("capture error: %v", err)
				continue
			}
			lastFrame = frame
			pos := detector.Detect(frame)
			if pos.Detected {
				log.Printf("line offset=%.3f (detected)", pos.Offset)
			} else {
				log.Printf("line not detected")
			}
		}
	}
}

func saveFrame(path string, rgb []byte, w, h int) {
	if err := perception.WritePPM(path, rgb, w, h); err != nil {
		log.Printf("save error: %v", err)
		return
	}
	log.Printf("saved %s", path)
}
