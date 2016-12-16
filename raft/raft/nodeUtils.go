package raft

import "fmt"

func shutdown(off bool) {
	if (off) {
		return nil
	}
}

//TODO FIXME
func (r *RaftNode) getLogTerm(index uint64) uint64 {
	return 0
}

func (r *RaftNode) handleCompetingRequests(msg RequestVoteMsg) bool {
	request := msg.request
	reply := msg.reply
	currentTerm := r.currentTerm
	prevIndex := r.commitIndex
	prevTerm := r.getLogTerm(prevIndex)

	if prevTerm > request.CandidateLastLogTerm {
		reply <- RequestVoteReply{currentTerm, false}
		return false
	} else if prevTerm < request.CandidateLastLogTerm {
		reply <- RequestVoteReply{currentTerm, true}
		return true
	} else {
		if prevIndex > request.CandidateLastLogIndex {
			reply <- RequestVoteReply{currentTerm, false}
			return false
		} else if prevIndex < request.CandidateLastLogIndex {
			reply <- RequestVoteReply{currentTerm, true}
			return true
		} else {
			if currentTerm > request.Term {
				reply <- RequestVoteReply{currentTerm, false}
				return false
			} else if currentTerm < request.Term {
				reply <- RequestVoteReply{currentTerm, true}
				return true
			} else {
				if r.votedFor != "" && r.state == LEADERSTATE {
					reply <- RequestVoteReply{currentTerm, false}
					return false
				} else {
					reply <- RequestVoteReply{currentTerm, true}
					return true
				}
			}
		}
	}
}

func (r *RaftNode) sendNoop() {
	entries := make([]LogEntry, 1)
	entries[0] = LogEntry{r.getLastLogIndex() + 1, r.currentTerm, make([]byte, 0)}
	fmt.Println("NOOP logged with log index %d and term index %d", r.getLastLogIndex() + 1, r.currentTerm)
	r.appendEntries
}