// PCA9685 diagnostic: sweep steering servo and pulse throttle.
// Usage: ./pcatest -confirm -config zerodriver-hardware.yaml
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
)

func main() {
	configPath := flag.String("config", "configs/zerodriver-hardware.yaml", "config file")
	mode := flag.String("mode", "hardware", "mode: mock or hardware")
	confirm := flag.Bool("confirm", false, "confirm servo/ESC will move")
	flag.Parse()

	if *mode == "hardware" && !*confirm {
		log.Fatal("hardware mode requires -confirm")
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	cfg.Mode = *mode

	motor, err := hal.NewMotor(cfg)
	if err != nil {
		log.Fatalf("init motor: %v", err)
	}
	defer motor.Close()

	log.Printf("PCA9685 test (mode=%s)", cfg.Mode)
	log.Printf("channels: steering=%d throttle=%d", cfg.Hardware.SteeringChannel, cfg.Hardware.ThrottleChannel)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// ESC arm
	log.Println("ESC arm: throttle=0 for 2s...")
	motor.Drive(0, 0)
	sleepOrAbort(2*time.Second, sigCh)

	steps := []struct {
		name     string
		steering float64
		throttle float64
		wait     time.Duration
	}{
		{"center", 0, 0, 1 * time.Second},
		{"steer left", -0.5, 0, 1 * time.Second},
		{"steer right", 0.5, 0, 1 * time.Second},
		{"center", 0, 0, 500 * time.Millisecond},
		{"throttle 30%", 0, 0.3, 1 * time.Second},
		{"stop", 0, 0, 500 * time.Millisecond},
	}

	for _, step := range steps {
		select {
		case <-sigCh:
			motor.Stop()
			log.Println("interrupted")
			return
		default:
		}
		log.Printf("→ %s (steer=%.1f throttle=%.1f)", step.name, step.steering, step.throttle)
		if err := motor.Drive(step.steering, step.throttle); err != nil {
			log.Fatalf("drive error: %v", err)
		}
		sleepOrAbort(step.wait, sigCh)
	}

	motor.Stop()
	log.Println("done")
}

func sleepOrAbort(d time.Duration, sigCh chan os.Signal) {
	select {
	case <-sigCh:
	case <-time.After(d):
	}
}
