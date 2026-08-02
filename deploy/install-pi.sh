#!/usr/bin/env bash
# Pi Zero W setup script for ZeroDriver.
# Usage: ./deploy/install-pi.sh
set -euo pipefail

INSTALL_DIR="${HOME}/zerodriver"
USER_NAME="${USER}"

echo "=== ZeroDriver Pi Install ==="

# Groups for hardware access
for grp in i2c dialout video; do
  if getent group "$grp" >/dev/null; then
    sudo usermod -aG "$grp" "$USER_NAME" 2>/dev/null || true
    echo "Added $USER_NAME to group $grp"
  fi
done

# I2C enable hint
if ! [ -e /dev/i2c-1 ]; then
  echo "WARN: /dev/i2c-1 not found. Enable I2C: sudo raspi-config → Interface Options → I2C"
fi

# Install binaries if present
mkdir -p "$INSTALL_DIR"
for bin in zerodriver hwtest imutest camtest motortest lidartest pcatest; do
  if [ -f "${bin}-armv6" ]; then
    cp "${bin}-armv6" "$INSTALL_DIR/$bin"
    chmod +x "$INSTALL_DIR/$bin"
    echo "Installed $INSTALL_DIR/$bin"
  fi
done

# Config
if [ -f zerodriver-hardware.yaml ]; then
  cp zerodriver-hardware.yaml "$INSTALL_DIR/zerodriver.yaml"
  echo "Installed config → $INSTALL_DIR/zerodriver.yaml"
fi

# systemd
if [ -f deploy/zerodriver.service ]; then
  sed "s|/home/pi|$HOME|g" deploy/zerodriver.service | sudo tee /etc/systemd/system/zerodriver.service >/dev/null
  sudo systemctl daemon-reload
  echo "Installed systemd service (not enabled by default)"
  echo "  Enable: sudo systemctl enable zerodriver"
  echo "  Start:  sudo systemctl start zerodriver"
fi

echo ""
echo "=== Verify ==="
echo "  i2cdetect -y 1        # 0x40=PCA9685, 0x68=IMU, 0x70=PCA9685 All Call"
echo "  ls /dev/ttyUSB0       # RPLidar A1"
echo "  $INSTALL_DIR/hwtest -config $INSTALL_DIR/zerodriver.yaml"
echo "  $INSTALL_DIR/pcatest -config $INSTALL_DIR/zerodriver.yaml -confirm"
echo ""
echo "Re-login required for group changes to take effect."
