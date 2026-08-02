#!/usr/bin/env bash
# Run field-test diagnostics on Pi in order.
# Usage: ./deploy/verify-pi.sh [config]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

CONFIG="${1:-configs/zerodriver-hardware.yaml}"
BIN_DIR="${BIN_DIR:-bin}"
PASS=0
FAIL=0

run_step() {
  local name="$1"
  shift
  echo ""
  echo "=== $name ==="
  if "$@"; then
    echo "OK: $name"
    PASS=$((PASS + 1))
  else
    echo "FAIL: $name"
    FAIL=$((FAIL + 1))
  fi
}

pick_bin() {
  local name="$1"
  if [ -x "$BIN_DIR/${name}-armv6" ]; then
    echo "$BIN_DIR/${name}-armv6"
  elif [ -x "$ROOT/${name}" ]; then
    echo "$ROOT/${name}"
  else
    echo "$BIN_DIR/${name}-armv6"
  fi
}

HWTEST="$(pick_bin hwtest)"
IMUTEST="$(pick_bin imutest)"
CAMTEST="$(pick_bin camtest)"
LIDARTEST="$(pick_bin lidartest)"
PCATEST="$(pick_bin pcatest)"

echo "ZeroDriver Pi verify"
echo "  config: $CONFIG"
echo "  hwtest: $HWTEST"

if ! command -v rpicam-vid >/dev/null 2>&1 && ! command -v libcamera-vid >/dev/null 2>&1; then
  echo "WARN: rpicam-vid not found. Install: sudo apt install -y rpicam-apps"
fi

run_step "I2C bus" test -e /dev/i2c-1
run_step "IMU" "$IMUTEST" -config "$CONFIG" -duration 3s
run_step "Camera" "$CAMTEST" -config "$CONFIG" -duration 2s
run_step "LiDAR" "$LIDARTEST" -port /dev/ttyUSB0 -duration 3s
run_step "Hardware summary" "$HWTEST" -config "$CONFIG"

echo ""
echo "=== Summary: $PASS passed, $FAIL failed ==="
echo "Next (vehicle lifted): $PCATEST -config $CONFIG -confirm"
echo "Next (line on floor):  $(pick_bin zerodriver) -config $CONFIG -mode hardware -v"

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
