package util

import (
	"fmt"
	"testing"
)

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

func TestMod(t *testing.T) {
	one := Identifier([]byte{1})
	res := one.CalculateStart(2, 160)
	fmt.Println(res)
}
