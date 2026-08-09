package hal

import (
	"testing"

	"github.com/autopilothub/zerodriver/internal/config"
)

func TestNewLidar_Disabled(t *testing.T) {
	disabled := false
	cfg := &config.Config{
		Mode: "hardware",
		Hardware: config.HardwareConfig{
			LidarEnabled: &disabled,
		},
	}

	lidar, err := NewLidar(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer lidar.Close()

	scan, err := lidar.Scan()
	if err != nil {
		t.Fatal(err)
	}
	if scan.FrontDistance != 999 {
		t.Fatalf("expected clear path, got %.0f", scan.FrontDistance)
	}
}
