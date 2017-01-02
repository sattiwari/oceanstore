package raft

import (
	"fmt"
	"time"
	"math/rand"
)

//TODO - move away the code to find majority element
func getLeader(nodes []*RaftNode) *RaftNode {
	it := 1
	var leader *RaftNode = nil
	for leader == nil && it < 50 {
		fmt.Printf("iteration %v\n", it)
		time.Sleep(time.Millisecond * 200)
		clusterSize := nodes[0].conf.ClusterSize
		idCountMap := make(map[string]int, clusterSize)
		for _, n := range nodes {
			if n.LeaderAddress != nil {
				idCountMap[n.LeaderAddress.Id]++
			}
		}
		fmt.Printf("node id to count map %v\n\n", idCountMap)
		var id string
		max := -1
		for k,v := range idCountMap {
			if max < v {
				max = v
				id = k
			}
		}
		if max > clusterSize / 2 {
			for _,node := range nodes {
				if node.LeaderAddress.Id == id {
					return node
				}
			}
		}
		it++
	}
	return leader
}

func randSeq(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func checkNodes(nodes []*RaftNode, clusterSize int) bool {
	for _, n := range nodes {
		if len(n.GetOtherNodes()) != clusterSize {
			Error.Println(len(n.GetOtherNodes()), clusterSize)
			return false
		}
	}
	return true
}

func printNodes(nodes []*RaftNode) {
	for _, n := range nodes {
		n.PrintLogCache()
		n.ShowState()
	}
}

func checkMajorityTerms(nodes []*RaftNode) bool {
	sums := make(map[uint64]int, nodes[0].conf.ClusterSize)
	for _, n := range nodes {
		sums[n.GetCurrentTerm()]++
	}
	max := -1
	for _, v := range sums {
		if v > max {
			max = v
		}
	}

	if max > len(nodes)/2 {
		return true
	}
	return false
}

func checkMajorityCommitIndex(nodes []*RaftNode) bool {
	sums := make(map[uint64]int, nodes[0].conf.ClusterSize)
	for _, n := range nodes {
		sums[n.commitIndex]++
	}
	max := -1
	for _, v := range sums {
		if v > max {
			max = v
		}
	}

	if max > len(nodes)/2 {
		return true
	}
	return false
}

func checkLogOrder(nodes []*RaftNode) bool {
	for _, n := range nodes {
		prevIndex := int64(-1)
		prevTerm := int64(-1)
		seen := make(map[uint64]bool)
		for _, entry := range n.logCache {
			if seen[entry.Index] || int64(entry.Index)-1 != prevIndex || int64(entry.Term) < prevTerm {
				return false
			}

			seen[entry.Index] = true
			prevIndex = int64(entry.Index)
			prevTerm = int64(entry.Term)
		}
	}
	return true
}

func shutdownNodes(nodes []*RaftNode) {
	for _, n := range nodes {
		n.IsShutDown = true
		n.gracefulExit <- true
	}
	time.Sleep(time.Millisecond * 200)
}