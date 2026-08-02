//go:build linux && arm

package hal

import (
	"fmt"
	"sync"
	"time"

	"go.bug.st/serial"

	"github.com/autopilothub/zerodriver/internal/domain"
)

// RPLidar reads scan data from an RPLidar A1 over USB serial (/dev/ttyUSB0).
type RPLidar struct {
	mu         sync.Mutex
	port       serial.Port
	frontAngle float64
	nodes      []ScanNode
	parser     *streamParser
	scanning   bool
	done       chan struct{}
}

// NewRPLidar opens the serial port and starts SCAN mode.
func NewRPLidar(portName string, baud int, frontAngleDeg float64) (*RPLidar, error) {
	if baud == 0 {
		baud = rplidarBaudA1
	}

	mode := &serial.Mode{
		BaudRate: baud,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, fmt.Errorf("open serial %q: %w", portName, err)
	}

	_ = port.SetReadTimeout(100 * time.Millisecond)
	_ = port.ResetInputBuffer()

	r := &RPLidar{
		port:       port,
		frontAngle: frontAngleDeg,
		nodes:      make([]ScanNode, 0, 360),
		parser:     newStreamParser(),
		done:       make(chan struct{}),
	}

	if err := r.startScan(); err != nil {
		port.Close()
		return nil, err
	}

	go r.readLoop()

	return r, nil
}

func (r *RPLidar) startScan() error {
	// Reset device before scan (A1 recommended on startup)
	reset := []byte{rplidarSyncByte, rplidarCmdReset}
	if _, err := r.port.Write(reset); err != nil {
		return fmt.Errorf("send RESET: %w", err)
	}
	time.Sleep(500 * time.Millisecond)
	_ = r.port.ResetInputBuffer()

	cmd := []byte{rplidarSyncByte, rplidarCmdScan}
	if _, err := r.port.Write(cmd); err != nil {
		return fmt.Errorf("send SCAN: %w", err)
	}

	header := make([]byte, rplidarHeaderLen)
	n, err := r.port.Read(header)
	if err != nil || n < rplidarHeaderLen {
		return fmt.Errorf("read scan header: %w (got %d bytes)", err, n)
	}
	if header[0] != rplidarSyncByte || header[1] != 0x5A {
		return fmt.Errorf("invalid scan header: %02X %02X", header[0], header[1])
	}

	r.scanning = true
	return nil
}

func (r *RPLidar) readLoop() {
	buf := make([]byte, 256)
	for r.scanning {
		select {
		case <-r.done:
			return
		default:
		}

		n, err := r.port.Read(buf)
		if err != nil || n == 0 {
			continue
		}

		parsed := r.parser.feed(buf[:n])
		if len(parsed) == 0 {
			continue
		}

		r.mu.Lock()
		r.nodes = append(r.nodes, parsed...)
		if len(r.nodes) > 720 {
			r.nodes = r.nodes[len(r.nodes)-360:]
		}
		r.mu.Unlock()
	}
}

func (r *RPLidar) Scan() (domain.ObstacleScan, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if len(r.nodes) == 0 {
		return domain.ObstacleScan{FrontDistance: 999, Timestamp: now}, nil
	}

	halfAngle := r.frontAngle / 2
	frontMin, frontNodes := FrontMinDistance(r.nodes, halfAngle)

	distances := make([]float64, len(frontNodes))
	angles := make([]float64, len(frontNodes))
	for i, n := range frontNodes {
		distances[i] = n.DistCM
		angles[i] = n.AngleDeg
	}

	r.nodes = r.nodes[:0]

	return domain.ObstacleScan{
		FrontDistance: frontMin,
		Distances:     distances,
		Angles:        angles,
		Timestamp:     now,
	}, nil
}

func (r *RPLidar) Close() error {
	r.scanning = false
	close(r.done)

	stop := []byte{rplidarSyncByte, rplidarCmdStop}
	_, _ = r.port.Write(stop)
	time.Sleep(50 * time.Millisecond)

	return r.port.Close()
}

// NodeCount returns buffered scan nodes (for diagnostics).
func (r *RPLidar) NodeCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.nodes)
}
