package raft

import (
	"testing"
	"fmt"
)

func TestNodeClusterCreation(t *testing.T)  {
	conf := DefaultConfig()
	nodes, err := CreateCluster(conf)
	leader := getLeader(nodes)
	fmt.Print(leader, err)
}
