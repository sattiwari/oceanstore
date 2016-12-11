package raft

import (
	"testing"
	"fmt"
)

func TestNodeClusterCreation(t *testing.T)  {
	conf := DefaultConfig()
	nodes, err := createCluster(conf)
	fmt.Print(nodes, err)
}
