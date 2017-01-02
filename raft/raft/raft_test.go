package raft

import (
	"testing"
	//"fmt"
	"time"
	"fmt"
)

func TestLeaderElection(t *testing.T) {
	config := DefaultConfig()
	config.ClusterSize = 5
	config.LogPath = randSeq(10)

	nodes, err := CreateCluster(config)
	if err != nil {
		t.Errorf("Could not create nodes")
		return
	}
	time.Sleep(time.Millisecond * 500)
	if !checkNodes(nodes, config.ClusterSize) {
		t.Errorf("CreateLocalCluster FAILED")
		return
	}

	fmt.Printf("Before loop\n")
	getLeader(nodes)
	leader := getLeader(nodes)
	fmt.Printf("after loop\n")
	if leader == nil {
		t.Errorf("Not found the leader")
		fmt.Printf("# nodes: %v\n", len(nodes))
		printNodes(nodes)
		return
	}

	time.Sleep(time.Millisecond * 500)
	if !checkMajorityTerms(nodes) {
		t.Errorf("Nodes are not on the same term (%v)", leader.GetCurrentTerm())
	}
	if !checkMajorityCommitIndex(nodes) {
		t.Errorf("Nodes dont have the same commit index (%v)", leader.commitIndex)
	}
	if !checkLogOrder(nodes) {
		t.Errorf("Nodes logs are not in an ok order")
		printNodes(nodes)
	}

	fmt.Printf("The disabled node is: %v\n", leader.Id)
	leader.Testing.PauseWorld(true)
	disableLeader := leader
	time.Sleep(time.Millisecond * 100)
	leader = getLeader(nodes)
	if leader == nil {
		t.Errorf("Leader is not the same %v is not located in node", leader.Id)
		return
	}

	fmt.Printf("We now enable %v\n", disableLeader.Id)
	disableLeader.Testing.PauseWorld(false)
	time.Sleep(time.Millisecond * 100)
	leader = getLeader(nodes)
	if leader == nil {
		t.Errorf("Leader is not the same %v is not located in node", leader.Id)
		return
	}
	time.Sleep(time.Millisecond * 500)
	if !checkMajorityTerms(nodes) {
		t.Errorf("Nodes are not on the same term (%v)", leader.GetCurrentTerm())
	}
	if !checkMajorityCommitIndex(nodes) {
		t.Errorf("Nodes dont have the same commit index (%v)", leader.commitIndex)
	}
	if !checkLogOrder(nodes) {
		t.Errorf("Nodes logs are not in an ok order")
		printNodes(nodes)
	}

	fmt.Println("TestLeaderElection pass")
	shutdownNodes(nodes)
}