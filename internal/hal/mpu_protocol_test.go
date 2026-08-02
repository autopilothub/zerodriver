package hal

import "testing"

func TestIsSupportedMPU(t *testing.T) {
	if !IsSupportedMPU(0x70) {
		t.Fatal("MPU-6500 should be supported")
	}
	if !IsSupportedMPU(0x71) {
		t.Fatal("MPU-9250 should be supported")
	}
	if IsSupportedMPU(0x68) {
		t.Fatal("0x68 is not WHO_AM_I")
	}
}

func TestMPUHasMagnetometer(t *testing.T) {
	if !MPUHasMagnetometer(0x71) {
		t.Fatal("MPU-9250 should have magnetometer")
	}
	if MPUHasMagnetometer(0x70) {
		t.Fatal("MPU-6500 should not have magnetometer by ID")
	}
}

func TestMPUChipName(t *testing.T) {
	if MPUChipName(0x70) != "MPU-6500" {
		t.Fatalf("got %s", MPUChipName(0x70))
	}
}
