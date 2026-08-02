// Motor diagnostic tool. Spins wheels briefly for wiring verification.
// Usage: ./motortest -confirm -speed 0.3 -duration 2s
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
	configPath := flag.String("config", "configs/zerodriver.yaml", "config file")
	mode := flag.String("mode", "hardware", "mode: mock or hardware")
	speed := flag.Float64("speed", 0.3, "motor speed 0.0-1.0")
	confirm := flag.Bool("confirm", false, "confirm motor will spin (required for hardware)")
	duration := flag.Duration("duration", 2*time.Second, "spin duration")
	flag.Parse()

	if *mode == "hardware" && !*confirm {
		log.Fatal("hardware mode requires -confirm flag (wheels will spin)")
	}
	if *speed < 0 || *speed > 1 {
		log.Fatal("speed must be 0.0-1.0")
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

	log.Printf("motor test: speed=%.2f duration=%s mode=%s", *speed, *duration, cfg.Mode)
	log.Println("sequence: forward → stop → turn left → stop")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	runSequence(motor, *speed, *duration, sigCh)
}

func runSequence(motor hal.Motor, speed float64, dur time.Duration, sigCh chan os.Signal) {
	steps := []struct {
		name  string
		left  float64
		right float64
	}{
		{"forward", speed, speed},
		{"stop", 0, 0},
		{"turn left", -speed, speed},
		{"stop", 0, 0},
	}

	for _, step := range steps {
		select {
		case <-sigCh:
			motor.Stop()
			log.Println("interrupted")
			return
		default:
		}

		log.Printf("→ %s (L=%.2f R=%.2f)", step.name, step.left, step.right)
		if err := motor.Set(step.left, step.right); err != nil {
			log.Printf("motor error: %v", err)
		}

		wait := dur / 4
		if step.name == "stop" {
			wait = 500 * time.Millisecond
		}
		time.Sleep(wait)
	}

	motor.Stop()
	log.Println("done")
}
