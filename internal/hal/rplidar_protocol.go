package hal

import "math"

const (
	rplidarSyncByte     = 0xA5
	rplidarCmdScan      = 0x20
	rplidarCmdStop      = 0x25
	rplidarCmdReset     = 0x40
	rplidarBaudA1       = 115200
	rplidarMaxDistCM    = 1200 // 12m
	rplidarMinQuality   = 0.1
	rplidarHeaderLen    = 7
	rplidarNodeLen      = 5
)

// ScanNode is a parsed RPLidar A1 measurement.
type ScanNode struct {
	Start    bool
	Quality  float64
	AngleDeg float64
	DistCM   float64
}

// ParseScanNode decodes a 5-byte RPLidar standard scan node.
// See Slamtec RPLidar protocol: check bit (byte0 bit1) must be 1.
func ParseScanNode(data [5]byte) (node ScanNode, ok bool) {
	if (data[0]>>1)&0x01 != 1 {
		return ScanNode{}, false
	}

	node.Start = data[0]&0x01 != 0
	node.Quality = float64(data[0]>>2) / 63.0

	angleRaw := uint16(data[1])>>1 | uint16(data[2])<<7
	node.AngleDeg = float64(angleRaw) / 64.0

	distRaw := uint16(data[3]) | uint16(data[4])<<8
	distMM := float64(distRaw) / 4.0
	node.DistCM = distMM / 10.0

	if node.DistCM < 1 || node.DistCM > rplidarMaxDistCM {
		return ScanNode{}, false
	}
	if node.Quality < rplidarMinQuality {
		return ScanNode{}, false
	}

	return node, true
}

// FrontMinDistance returns the nearest obstacle within ±halfAngleDeg of 0°.
func FrontMinDistance(nodes []ScanNode, halfAngleDeg float64) (minCM float64, front []ScanNode) {
	minCM = 999
	for _, n := range nodes {
		diff := math.Abs(normalizeAngle(n.AngleDeg))
		if diff <= halfAngleDeg && n.DistCM < minCM {
			minCM = n.DistCM
		}
		if diff <= halfAngleDeg {
			front = append(front, n)
		}
	}
	return minCM, front
}

func normalizeAngle(a float64) float64 {
	for a > 180 {
		a -= 360
	}
	for a < -180 {
		a += 360
	}
	return a
}

// streamParser buffers serial bytes and extracts valid 5-byte nodes.
type streamParser struct {
	buf []byte
}

func newStreamParser() *streamParser {
	return &streamParser{buf: make([]byte, 0, 512)}
}

func (p *streamParser) feed(data []byte) []ScanNode {
	p.buf = append(p.buf, data...)
	var nodes []ScanNode

	for len(p.buf) >= rplidarNodeLen {
		var raw [5]byte
		copy(raw[:], p.buf[:rplidarNodeLen])
		node, ok := ParseScanNode(raw)
		if !ok {
			p.buf = p.buf[1:]
			continue
		}
		p.buf = p.buf[rplidarNodeLen:]
		nodes = append(nodes, node)
	}

	if len(p.buf) > 1024 {
		p.buf = p.buf[len(p.buf)-256:]
	}

	return nodes
}
