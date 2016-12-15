package raft

import "fmt"

func (r *RaftNode) requestVotes(electionResults chan bool) {
	go func() {
		nodes := r.otherNodes
		numNodes := len(nodes)
		votes := 1
		for _, node := range nodes {
			if node.id == r.id {
				continue
			}
			request := RequestVoteMsg{RequestVote{r.currentTerm, r.localAddr}}
			reply, _ := r.requestVoteRPC(&node, request)
			if reply == nil {
				continue
			}
			if r.currentTerm < reply.Term {
				fmt.Println("[Term outdated] Current = %d, Remote = %d", r.currentTerm, reply.Term)
				electionResults <-  false
				return
			}
			if reply.VoteGranted {
				votes++
			}
		}
		if votes > numNodes / 2 {
			electionResults <- true
		}
		electionResults <- false
	}()
}
