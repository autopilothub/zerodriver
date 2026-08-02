# ZeroDriver — AI Agent Steering Rules

## Project

Raspberry Pi Zero W line tracer RC car. Go 1.22+, AWS AI-DLC methodology.

## Architecture

- `internal/domain` — pure types, no external deps
- `internal/hal` — hardware interfaces + mock/real implementations
- `internal/perception` — sensor data processing
- `internal/fusion` — multi-sensor fusion
- `internal/control` — PID, state machine, motor control
- `cmd/zerodriver` — main entry point

## Coding Standards

- Go 1.22+, standard library preferred
- `internal/` packages must not be imported externally
- HAL: always define interface + mock implementation first
- Every exported function needs a unit test
- Config via YAML (`configs/zerodriver.yaml`)
- No global mutable state; pass dependencies explicitly

## Development Modes

- `mock` — simulated sensors/motors (default, works on dev PC)
- `hardware` — real Pi GPIO/I2C/UART (linux/arm only)

## AI-DLC Process

Follow bolt-based development documented in `docs/05-aidlc-process.md`.
Each bolt requires human validation before proceeding.

## Hardware Notes

- Pi Zero W: ARM1176, 512MB RAM — keep algorithms lightweight
- Camera: 320x240, ROI bottom third only
- LiDAR: 5Hz scan rate sufficient
- MPU-9250: I2C 0x68

## Do Not

- Add heavy dependencies without justification
- Skip tests for control/fusion logic
- Hardcode pin numbers (use config)
- Commit secrets or AWS credentials
