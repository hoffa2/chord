package launch

import (
	"fmt"
	"os"
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

func TestLaunchSSH(t *testing.T) {
	cwd, err := os.Getwd()

	err = runNodes(cwd, "1", "fake")
	if err != nil {
		t.Error(err)
	}
	
}
