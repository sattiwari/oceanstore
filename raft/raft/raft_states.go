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
	electionTimeOut := r.electionTimeOut()
	for {
		select {
		case off := <- r.gracefulExit:
			shutdown(off)
		case vote := <- r.requestVote:
			fmt.Println("Request Vote Received by Follower")
			request := vote.request
			if r.handleCompetingRequests(vote) {
				r.currentTerm = request.Term
				r.votedFor = request.CandidateId
				r.leaderAddress = nil
				electionTimeOut = r.electionTimeOut()
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
	electionTimeOut := r.electionTimeOut()

	r.state = CANDIDATESTATE
	r.leaderAddress = nil
	r.votedFor = r.id

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
				r.leaderAddress = nil
				r.votedFor = request.CandidateId
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
	r.leaderAddress = r.localAddr
	r.votedFor = ""
	r.state = LEADERSTATE

	beats := r.heartBeats()
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

func (r *RaftNode) handleCompetingRequestVote(msg RequestVoteMsg) bool {

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

}