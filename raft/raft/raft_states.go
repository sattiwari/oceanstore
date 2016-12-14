package raft

import (
	"time"
)

type state func() state

/**
 * This method contains the logic of a Raft node in the follower state.
 */
func (r *RaftNode) doFollower() state {
	electionTimeOut := makeElectionTimeout()
	for {
		select {
		case shutdown := <- r.gracefulExcit:
			if shutdown {
				return nil
			}
		case _ = <- r.requestVote:
		case _ = <- r.appendEntries:
		case _ = <- r.registerClient:
		case _ = <- r.clientRequest:
		case _ = <- electionTimeOut:
		}
	}
	return nil
}

/**
 * This method contains the logic of a Raft node in the candidate state.
 */
func (r *RaftNode) doCandidate() state {
	electionResults := make(chan bool)
	for {
		select {
		case shutdown := <- r.gracefulExcit:
			if shutdown {
				return nil
			}
		case result := <- electionResults:
			if result {
				r.doLeader()
			} else {
				r.doFollower()
			}
		case _ = r.requestVote:
		case _ = r.appendEntries:
		case _ = r.registerClient:
		case _ = r.clientRequest:
		case _ = r.electionTimeOut():
			return r.doFollower()
		}
	}
	return nil
}

/**
 * This method contains the logic of a Raft node in the leader state.
 */
func (r *RaftNode) doLeader() state {
}

func (r *RaftNode) handleCompetingRequestVote(msg RequestVoteMsg) bool {

}

/**
 * This function is called to request votes from all other nodes. It takes
 * a channel which the result of the vote should be sent over: true for
 * successful election, false otherwise.
 */
func (r *RaftNode) requestVotes(electionResults chan bool) {

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

/**
 * This function will use time.After to create a random timeout.
 */
func makeElectionTimeout() <-chan time.Time {

}


