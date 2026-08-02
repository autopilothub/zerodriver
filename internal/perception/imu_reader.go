package perception

import (
	"github.com/autopilothub/zerodriver/internal/domain"
	"github.com/autopilothub/zerodriver/internal/hal"
)

// IMUReader wraps the HAL IMU interface.
type IMUReader struct {
	imu hal.IMU
}

func NewIMUReader(imu hal.IMU) *IMUReader {
	return &IMUReader{imu: imu}
}

func (r *IMUReader) Read() (domain.Attitude, error) {
	return r.imu.Read()
}
