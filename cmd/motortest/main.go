// RC car motor test: steering servo + throttle PWM.
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
	speed := flag.Float64("speed", 0.3, "throttle 0.0-1.0")
	confirm := flag.Bool("confirm", false, "confirm motor will spin (required for hardware)")
	duration := flag.Duration("duration", 2*time.Second, "step duration")
	flag.Parse()

	if *mode == "hardware" && !*confirm {
		log.Fatal("hardware mode requires -confirm flag")
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

	log.Printf("RC motor test: throttle=%.2f mode=%s", *speed, cfg.Mode)
	log.Println("sequence: forward → stop → steer left → stop")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	runSequence(motor, *speed, *duration, sigCh)
}

func runSequence(motor hal.Motor, speed float64, dur time.Duration, sigCh chan os.Signal) {
	steps := []struct {
		name     string
		steering float64
		throttle float64
	}{
		{"forward", 0, speed},
		{"stop", 0, 0},
		{"steer left", -0.5, 0},
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

		log.Printf("→ %s (steer=%.2f throttle=%.2f)", step.name, step.steering, step.throttle)
		if err := motor.Drive(step.steering, step.throttle); err != nil {
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
