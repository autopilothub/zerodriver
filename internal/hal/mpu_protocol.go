package hal

// MPU WHO_AM_I register values (register 0x75 on the 6-axis core at 0x68).
const (
	mpuWHOAmIMP6500 = 0x70 // MPU-6500 core; also the MPU-9250 6-axis core ID
	mpuWHOAmIMP9250 = 0x71 // Some MPU-9250 boards report this value
	mpuWHOAmIMP9255 = 0x73
)

// MPUCoreName returns the 6-axis core name from WHO_AM_I alone.
// Note: MPU-9250 packages use the MPU-6500 core and also return 0x70.
func MPUCoreName(who byte) string {
	switch who {
	case mpuWHOAmIMP6500:
		return "MPU-6500 core"
	case mpuWHOAmIMP9250:
		return "MPU-9250 core"
	case mpuWHOAmIMP9255:
		return "MPU-9255 core"
	default:
		return "unknown core"
	}
}

// MPUChipName is an alias for MPUCoreName (backward compatible).
func MPUChipName(who byte) string {
	return MPUCoreName(who)
}

// ResolveMPUChipName returns the user-facing module name.
// WHO_AM_I 0x70 is shared by standalone MPU-6500 and MPU-9250; AK8963 probe disambiguates.
func ResolveMPUChipName(who byte, hasMag bool) string {
	if hasMag {
		switch who {
		case mpuWHOAmIMP6500, mpuWHOAmIMP9250:
			return "MPU-9250"
		case mpuWHOAmIMP9255:
			return "MPU-9255"
		}
	}
	switch who {
	case mpuWHOAmIMP6500:
		return "MPU-6500"
	case mpuWHOAmIMP9250:
		return "MPU-9250"
	case mpuWHOAmIMP9255:
		return "MPU-9255"
	default:
		return "unknown"
	}
}

// IsSupportedMPU reports whether the WHO_AM_I value is a supported IMU core.
func IsSupportedMPU(who byte) bool {
	switch who {
	case mpuWHOAmIMP6500, mpuWHOAmIMP9250, mpuWHOAmIMP9255:
		return true
	default:
		return false
	}
}

// IMU model config values for hardware.imu_model.
const (
	IMUModelAuto    = "auto"
	IMUModelMPU6500 = "mpu6500"
	IMUModelMPU9250 = "mpu9250"
)

// NormalizeIMUModel normalizes config imu_model to auto/mpu6500/mpu9250.
func NormalizeIMUModel(model string) string {
	switch model {
	case "", "auto":
		return IMUModelAuto
	case IMUModelMPU6500, "6500":
		return IMUModelMPU6500
	case IMUModelMPU9250, "9250":
		return IMUModelMPU9250
	default:
		return model
	}
}

// ShouldProbeMagnetometer reports whether AK8963 should be initialized.
func ShouldProbeMagnetometer(model string) bool {
	switch NormalizeIMUModel(model) {
	case IMUModelMPU6500:
		return false
	default:
		return true
	}
}

// RequireMagnetometer reports whether init must fail if AK8963 is missing.
func RequireMagnetometer(model string) bool {
	return NormalizeIMUModel(model) == IMUModelMPU9250
}
