package raft

import (
	"testing"
	"fmt"
)

func TestNodeClusterCreation(t *testing.T)  {
	conf := DefaultConfig()
	nodes, err := createCluster(conf)
	leader := getLeader(nodes)
	fmt.Print(leader, err)
}
