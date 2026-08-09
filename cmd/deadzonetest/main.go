// deadzonetest sweeps ESC pulse widths to find forward/reverse deadzones.
// Usage: ./deadzonetest -confirm -config configs/zerodriver-hardware.yaml
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/autopilothub/zerodriver/internal/calibration"
	"github.com/autopilothub/zerodriver/internal/config"
	"github.com/autopilothub/zerodriver/internal/hal"
)

func main() {
	configPath := flag.String("config", "configs/zerodriver-hardware.yaml", "config file")
	mode := flag.String("mode", "hardware", "mode: mock or hardware")
	confirm := flag.Bool("confirm", false, "confirm vehicle will move")
	flag.Parse()

	if *mode == "hardware" && !*confirm {
		log.Fatal("hardware mode requires -confirm (vehicle will move briefly)")
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	cfg.Mode = *mode

	imu, err := hal.NewIMU(cfg)
	if err != nil {
		log.Fatalf("init imu: %v", err)
	}
	defer imu.Close()

	motor, err := hal.NewMotor(cfg)
	if err != nil {
		log.Fatalf("init motor: %v", err)
	}
	defer motor.Close()

	log.Println("=== ESC deadzone calibration ===")
	log.Println("Car on ground, room to roll forward/back briefly.")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		motor.Stop()
		os.Exit(1)
	}()

	log.Println("ESC arm 2s...")
	motor.Drive(0, 0)
	time.Sleep(2 * time.Second)

	res, err := calibration.RunDeadzone(
		motor.SetThrottleUS,
		imu.Read,
		calibration.DeadzoneConfig{
			NeutralUs:      cfg.Hardware.ThrottleNeutral(),
			ForwardEndUs:   cfg.Hardware.ThrottleForwardEnd(),
			ReverseEndUs:   cfg.Hardware.ThrottleReverseEnd(),
			ForwardStartUs: cfg.Hardware.ThrottleForwardStartUs,
			ReverseStartUs: cfg.Hardware.ThrottleReverseStartUs,
			Inverted:       cfg.Hardware.ThrottleInverted(),
		},
		calibration.DeadzoneOptions{
			OnProgress: func(phase string, pulse int) {
				log.Printf("  %s %dµs", phase, pulse)
			},
		},
	)
	motor.Stop()
	if err != nil {
		log.Fatalf("deadzone: %v", err)
	}

	fmt.Println("\n--- Results ---")
	for _, note := range res.Notes {
		fmt.Println(" ", note)
	}
	fmt.Println("\nSuggested config:")
	fmt.Printf("  throttle_forward_start_us: %d\n", res.ThrottleForwardStartUs)
	fmt.Printf("  throttle_reverse_start_us: %d\n", res.ThrottleReverseStartUs)
	fmt.Println("\nCopy values into configs/zerodriver-hardware.yaml")
}
