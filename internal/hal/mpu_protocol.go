package hal

// MPU WHO_AM_I register values (register 0x75).
const (
	mpuWHOAmIMP6500 = 0x70
	mpuWHOAmIMP9250 = 0x71
	mpuWHOAmIMP9255 = 0x73
)

// MPUChipName returns a human-readable name for a WHO_AM_I value.
func MPUChipName(who byte) string {
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

// IsSupportedMPU reports whether the WHO_AM_I value is a supported 6-axis IMU.
func IsSupportedMPU(who byte) bool {
	switch who {
	case mpuWHOAmIMP6500, mpuWHOAmIMP9250, mpuWHOAmIMP9255:
		return true
	default:
		return false
	}
}
