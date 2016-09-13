package launch

import (
	"fmt"
	"testing"
)

func TestNodeSort(t *testing.T) {
	m, h := hashValuesSorted([]string{"ccccc", "aaaaa", "xxxxx"})
	for _, val := range h {
		fmt.Println(m[val])
	}
}
