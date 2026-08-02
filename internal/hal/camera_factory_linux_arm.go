//go:build linux && arm

package hal

import (
	"fmt"
	"log"

	"github.com/autopilothub/zerodriver/internal/config"
)

func newHardwareCamera(cfg *config.Config) (Camera, error) {
	backend := cfg.Hardware.CameraBackend
	if backend == "" {
		backend = "rpicam"
	}

	switch backend {
	case "rpicam":
		cam, err := NewRpicamCamera(cfg.Camera.Width, cfg.Camera.Height)
		if err != nil {
			return nil, err
		}
		log.Printf("camera: rpicam backend %dx%d", cam.Width(), cam.Height())
		return cam, nil
	case "v4l2":
		cam, err := NewV4L2Camera(cfg.Hardware.CameraDevice, cfg.Camera.Width, cfg.Camera.Height)
		if err != nil {
			return nil, err
		}
		log.Printf("camera: v4l2 backend %s %dx%d", cfg.Hardware.CameraDevice, cam.Width(), cam.Height())
		return cam, nil
	case "auto":
		cam, err := NewV4L2Camera(cfg.Hardware.CameraDevice, cfg.Camera.Width, cfg.Camera.Height)
		if err == nil {
			log.Printf("camera: v4l2 backend %s %dx%d", cfg.Hardware.CameraDevice, cam.Width(), cam.Height())
			return cam, nil
		}
		rpicam, err2 := NewRpicamCamera(cfg.Camera.Width, cfg.Camera.Height)
		if err2 == nil {
			log.Printf("camera: rpicam backend %dx%d (v4l2 failed: %v)", rpicam.Width(), rpicam.Height(), err)
			return rpicam, nil
		}
		return nil, fmt.Errorf("camera init failed (v4l2: %v; rpicam: %v)", err, err2)
	default:
		return nil, fmt.Errorf("unknown camera_backend: %q (use auto, v4l2, or rpicam)", backend)
	}
}
