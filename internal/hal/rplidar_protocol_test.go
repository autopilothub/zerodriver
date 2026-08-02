package hal

import (
	"testing"
)

func TestParseScanNode_Valid(t *testing.T) {
	// angle=90° (5760 = 0x1680), distance=1000mm (4000 = 0x0FA0)
	// quality=30 → byte0 = (30<<2)|0x02 = 0x7A (check bit set)
	data := [5]byte{0x7A, 0x00, 0x2D, 0xA0, 0x0F}
	node, ok := ParseScanNode(data)
	if !ok {
		t.Fatal("expected valid node")
	}
	if node.AngleDeg < 89 || node.AngleDeg > 91 {
		t.Fatalf("expected ~90°, got %f", node.AngleDeg)
	}
	if node.DistCM < 99 || node.DistCM > 101 {
		t.Fatalf("expected ~100cm, got %f", node.DistCM)
	}
}

func TestParseScanNode_InvalidCheckBit(t *testing.T) {
	data := [5]byte{0x01, 0x00, 0x2D, 0xA0, 0x0F} // check bit = 0
	_, ok := ParseScanNode(data)
	if ok {
		t.Fatal("expected invalid node")
	}
}

func TestFrontMinDistance(t *testing.T) {
	nodes := []ScanNode{
		{AngleDeg: 0, DistCM: 50},
		{AngleDeg: 10, DistCM: 30},
		{AngleDeg: 90, DistCM: 10},
		{AngleDeg: -5, DistCM: 40},
	}
	min, front := FrontMinDistance(nodes, 30)
	if min != 30 {
		t.Fatalf("expected 30cm, got %f", min)
	}
	if len(front) != 3 {
		t.Fatalf("expected 3 front nodes, got %d", len(front))
	}
}

func TestStreamParser_Resync(t *testing.T) {
	p := newStreamParser()
	garbage := []byte{0xFF, 0xFF}
	valid := [5]byte{0x7A, 0x00, 0x2D, 0xA0, 0x0F}

	data := append(garbage, valid[:]...)
	nodes := p.feed(data)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node after resync, got %d", len(nodes))
	}
}
