package hal

import "testing"

func TestIsSupportedMPU(t *testing.T) {
	if !IsSupportedMPU(0x70) {
		t.Fatal("MPU-6500 core should be supported")
	}
	if !IsSupportedMPU(0x71) {
		t.Fatal("MPU-9250 core should be supported")
	}
	if IsSupportedMPU(0x68) {
		t.Fatal("0x68 is bus address, not WHO_AM_I")
	}
}

func TestResolveMPUChipName(t *testing.T) {
	if got := ResolveMPUChipName(0x70, true); got != "MPU-9250" {
		t.Fatalf("0x70+mag: got %s", got)
	}
	if got := ResolveMPUChipName(0x70, false); got != "MPU-6500" {
		t.Fatalf("0x70 no mag: got %s", got)
	}
	if got := ResolveMPUChipName(0x71, false); got != "MPU-9250" {
		t.Fatalf("0x71: got %s", got)
	}
}

func TestNormalizeIMUModel(t *testing.T) {
	if NormalizeIMUModel("9250") != IMUModelMPU9250 {
		t.Fatal("9250 alias")
	}
	if NormalizeIMUModel("bno085") != IMUModelBNO085 {
		t.Fatal("bno085 alias")
	}
	if ShouldProbeMagnetometer(IMUModelMPU6500) {
		t.Fatal("6500 should not probe mag")
	}
	if !RequireMagnetometer(IMUModelMPU9250) {
		t.Fatal("9250 should require mag")
	}
}
