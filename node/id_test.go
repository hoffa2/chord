package node

import "testing"

func TestNodeID(t *testing.T) {
	nID1 := StringToID("2")
	nID2 := StringToID("1")
	if !nID1.IsLarger(nID2) {
		t.Errorf("nID1 should be larger than a")
	}

	nID1 = StringToID("2")
	nID2 = StringToID("2")
	if !nID1.IsEqual(nID2) {
		t.Errorf("nID1 and NID2 should be equal")
	}
}

func TestNodeIDKeySpace(t *testing.T) {

}
