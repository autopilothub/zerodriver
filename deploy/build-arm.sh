#!/usr/bin/env bash
set -euo pipefail

OUTPUT_DIR="bin"
TARGETS=(zerodriver imutest camtest motortest lidartest hwtest)

mkdir -p "$OUTPUT_DIR"

for target in "${TARGETS[@]}"; do
  echo "Building $target for Pi Zero W (linux/arm GOARM=6)..."
  GOOS=linux GOARCH=arm GOARM=6 go build -o "$OUTPUT_DIR/${target}-armv6" "./cmd/$target"
done

echo ""
echo "Done:"
for target in "${TARGETS[@]}"; do
  echo "  $OUTPUT_DIR/${target}-armv6"
done
echo ""
echo "Deploy to Pi:"
echo "  scp $OUTPUT_DIR/*-armv6 configs/zerodriver-hardware.yaml pi@raspberrypi.local:/home/pi/"
echo ""
echo "Pi test sequence:"
echo "  ./hwtest -config zerodriver-hardware.yaml"
echo "  ./hwtest -config zerodriver-hardware.yaml -motor -confirm"
echo "  ./zerodriver-armv6 -config zerodriver-hardware.yaml -mode hardware -v"
