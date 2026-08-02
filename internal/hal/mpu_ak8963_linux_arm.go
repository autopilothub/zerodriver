//go:build linux && arm

package hal

import (
	"fmt"
	"time"

	"periph.io/x/conn/v3/i2c"
)

const (
	mpuRegIntPinCfg = 0x37

	ak8963I2CAddr   = 0x0C
	ak8963RegWIA    = 0x00
	ak8963RegHXL    = 0x03
	ak8963RegCNTL1  = 0x0A
	ak8963WHOAmI    = 0x48
	ak8963Mode16Hz  = 0x16 // 16-bit, continuous 100Hz
	ak8963ScaleUT   = 4912.0 / 32768.0
)

func initAK8963(mpu *i2c.Dev, bus i2c.Bus) (i2c.Dev, bool, error) {
	if err := i2cWriteReg(mpu, mpuRegIntPinCfg, 0x02); err != nil {
		return i2c.Dev{}, false, fmt.Errorf("enable i2c bypass: %w", err)
	}
	time.Sleep(10 * time.Millisecond)

	mag := i2c.Dev{Addr: ak8963I2CAddr, Bus: bus}
	wia, err := i2cReadReg(&mag, ak8963RegWIA)
	if err != nil || wia != ak8963WHOAmI {
		return i2c.Dev{}, false, nil
	}

	_ = i2cWriteReg(&mag, ak8963RegCNTL1, 0x00)
	time.Sleep(10 * time.Millisecond)
	if err := i2cWriteReg(&mag, ak8963RegCNTL1, ak8963Mode16Hz); err != nil {
		return i2c.Dev{}, false, fmt.Errorf("start magnetometer: %w", err)
	}
	time.Sleep(10 * time.Millisecond)
	return mag, true, nil
}

func readAK8963(mag *i2c.Dev) (mx, my, mz float64, ok bool, err error) {
	buf := make([]byte, 7)
	if err := mag.Tx([]byte{ak8963RegHXL}, buf); err != nil {
		return 0, 0, 0, false, err
	}
	if buf[6]&0x08 != 0 {
		return 0, 0, 0, false, nil
	}

	hx := int16(buf[1])<<8 | int16(buf[0])
	hy := int16(buf[3])<<8 | int16(buf[2])
	hz := int16(buf[5])<<8 | int16(buf[4])
	return float64(hx) * ak8963ScaleUT, float64(hy) * ak8963ScaleUT, float64(hz) * ak8963ScaleUT, true, nil
}
