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

func TestGetHosts(t *testing.T) {

	hosts, err := getNodeList("3")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(hosts)
}
