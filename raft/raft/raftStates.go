package raft

import (
	"fmt"
	"time"
)

type state func() state

/**
 * This method contains the logic of a Raft node in the follower state.
 */
func (r *RaftNode) doFollower() state {
	fmt.Println("Transitioning to follower state")
	electionTimeOut := r.makeElectionTimeOut()
	for {
		select {
		case off := <- r.gracefulExit:
			shutdown(off)
		case vote := <- r.requestVote:
			fmt.Println("Request Vote Received by Follower")
			request := vote.request
			if r.handleCompetingRequests(vote) {
				r.currentTerm = request.Term
				r.VotedFor = request.CandidateId
				r.LeaderAddress = nil
				electionTimeOut = r.makeElectionTimeOut()
			} else {
				if request.Term > r.currentTerm {
					r.currentTerm = request.Term
				}
			}
		case _ = <- r.appendEntries:
		case _ = <- r.registerClient:
		case _ = <- r.clientRequest:
		case _ = <- electionTimeOut:
			return r.doCandidate()
		}
	}
	return nil
}

/**
 * This method contains the logic of a Raft node in the candidate state.
 */
func (r *RaftNode) doCandidate() state {
	fmt.Println("Transitioning to candidate state")
	electionResults := make(chan bool)
	electionTimeOut := r.makeElectionTimeOut()

	r.State = CANDIDATE_STATE
	r.LeaderAddress = nil
	r.VotedFor = r.Id

	r.requestVotes(electionResults)


	for {
		select {
		case off := <- r.gracefulExit:
			shutdown(off)
		case result := <- electionResults:
			if result {
				r.doLeader()
			} else {
				r.doFollower()
			}
		case vote := <- r.requestVote:
			fmt.Println("Request vote received by candidate")
			request := vote.request
			if r.handleCompetingRequests(vote) {
				r.currentTerm = request.Term
				r.LeaderAddress = nil
				r.VotedFor = request.CandidateId
				r.doFollower()
			} else {
				if r.currentTerm < request.Term {
					r.currentTerm = request.Term
				}
			}
		case _ = r.appendEntries:
		case _ = r.registerClient:
		case _ = r.clientRequest:
		case _ = electionTimeOut:
			return r.doCandidate()
		}
	}
	return nil
}

/**
 * This method contains the logic of a Raft node in the leader state.
 */
func (r *RaftNode) doLeader() state {
	fmt.Println("Transitioning to leader state")
	r.LeaderAddress = r.LocalAddr
	r.VotedFor = ""
	r.State = LEADER_STATE

	beats := r.makeHeartBeats()
	fallback := make(chan bool)
	finish := make(chan bool, 1)

	for {
		select {
		case off := <- r.gracefulExit:
			shutdown(off)
		case _ = <- r.appendEntries:
		case _ = <- beats:
			select {
				case <- finish:
				default:
					time.After(time.Millisecond * 1)
				}

		case _ = <- fallback:
		case _ = <- r.appendEntries:
		case _ = <- r.registerClient:
		case _ = <- r.clientRequest:

		}
	}
}

func (r *RaftNode) sendRequestFail()  {

}

func (r *RaftNode) sendNoop() {
	entries := make([]LogEntry, 1)
	entries[0] = LogEntry{r.GetLastLogIndex() + 1, r.currentTerm, make([]byte, 0)}
	fmt.Println("NOOP logged with log index %d and term index %d", r.GetLastLogIndex() + 1, r.currentTerm)
	r.appendEntries
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
				if r.VotedFor != "" && r.State == LEADER_STATE {
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
func (r *RaftNode) requestVotes(electionResults chan bool) {
	go func() {
		nodes := r.otherNodes
		numNodes := len(nodes)
		votes := 1
		for _, node := range nodes {
			if node.id == r.Id {
				continue
			}
			request := RequestVoteMsg{RequestVoteRequest{r.currentTerm, r.LocalAddr}}
			reply, _ := r.RequestVoteRPC(&node, request)
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

func (r *RaftNode) sendEmptyHeartBeats()  {

}

/**
 * This function is used by the leader to send out heartbeats to each of
 * the other nodes. It returns true if the leader should fall back to the
 * follower state. (This happens if we discover that we are in an old term.)
 *
 * If another node isn't up-to-date, then the leader should attempt to
 * update them, and, if an index has made it to a quorum of nodes, commit
 * up to that index. Once committed to that index, the replicated state
 * machine should be given the new log entries via processLog.
 */
func (r *RaftNode) sendHeartBeats(fallback, finish chan bool) {

}

func (r *RaftNode) sendAppendEntries(entries []LogEntry) (fallBack, sentToMajority bool) {
	fallBack = true
	sentToMajority = true
	return
}

func (r *RaftNode) makeElectionTimeOut() <- chan time.Time {
	return time.After(r.conf.ElectionTimeout)
}

func (r *RaftNode) makeHeartBeats() <- chan time.Time {
	return time.After(r.conf.HeartbeatFrequency)
}

func shutdown(off bool) {
	if (off) {
		return nil
	}
}