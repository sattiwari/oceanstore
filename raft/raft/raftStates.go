package raft

import (
	"fmt"
	"time"
	"math"
	"math/rand"
)

type state func() state

/**
 * This method contains the logic of a Raft node in the follower state.
 Followers are passive: they issue no requests on their own but simply respond to requests from leaders and candidates.
 */
func (r *RaftNode) doFollower() state {
	r.Out("Transitioning to FOLLOWER_STATE")
	r.State = FOLLOWER_STATE
	electionTimeout := makeElectionTimeout()
	for {
		select {
		case shutdown := <-r.gracefulExit:
			if shutdown {
				return nil
			}

		//	response to candidate request for votes
		case vote := <-r.requestVote:
			r.Out("Request vote received")
			req := vote.request
			currTerm := r.GetCurrentTerm()
			candidate := req.CandidateId

			if r.handleCompetingRequestVote(vote) {
				r.setCurrentTerm(req.Term)
				r.setVotedFor(candidate.Id)
				r.LeaderAddress = nil
				electionTimeout = makeElectionTimeout()
			} else {
				// Other candidate has a less up to date log.
				if req.Term > currTerm {
					r.setCurrentTerm(req.Term)
				}
			}

		//	response for leader's heartbeat message
		case appendEnt := <-r.appendEntries:
			req := appendEnt.request
			rep := appendEnt.reply
			currTerm := r.GetCurrentTerm()

			if req.Term < currTerm {
				r.Out("Rejected heartbeat from %v because of terms", req.LeaderId.Id)
				rep <- AppendEntriesReply{currTerm, false}
			} else {
				r.LeaderAddress = &req.LeaderId
				r.setCurrentTerm(req.Term)
				r.setVotedFor("")

				entry := r.getLogEntry(req.PrevLogIndex)
				if entry == nil || entry.Term != req.PrevLogTerm {
					r.Out("Rejected heartbeat from %v because of miss", req.LeaderId.Id)
					rep <- AppendEntriesReply{currTerm, false}
				} else {
					if req.PrevLogIndex+1 > r.commitIndex {
						r.truncateLog(req.PrevLogIndex + 1)
					} else {
						r.truncateLog(req.PrevLogIndex + 1)
						r.commitIndex = req.PrevLogIndex
					}

					if len(req.Entries) > 0 {
						for i := range req.Entries {
							r.appendLogEntry(req.Entries[i])
						}
					}

					if req.LeaderCommit > r.commitIndex {
						r.commitIndex = uint64(math.Min(float64(req.LeaderCommit), float64(r.getLastLogIndex())))
					}

					// Calls process log from lastApplied until commit index.
					if r.lastApplied < r.commitIndex {
						i := r.lastApplied
						if r.lastApplied != 0 {
							i++
						}
						for ; i <= r.commitIndex; i++ {
							_ = r.processLog(*r.getLogEntry(i))
						}
						r.lastApplied = r.commitIndex
					}

					rep <- AppendEntriesReply{currTerm, true}
				}
			}
			electionTimeout = makeElectionTimeout()

		//	if a client contacts a follower, the follower redirects it to the leader
		case regClient := <-r.registerClient:
			rep := regClient.reply
			if r.LeaderAddress != nil {
				rep <- RegisterClientReply{NOT_LEADER, 0, *r.LeaderAddress}
			} else {
				rep <- RegisterClientReply{ELECTION_IN_PROGRESS, 0, NodeAddr{"", ""}}
			}

		case <-electionTimeout:
			return r.doCandidate

		case clientReq := <-r.clientRequest:
			rep := clientReq.reply
			if r.LeaderAddress != nil {
				rep <- ClientReply{NOT_LEADER, "Not leader", *r.LeaderAddress}
			} else {
				rep <- ClientReply{NOT_LEADER, "Not leader", NodeAddr{"", ""}}
			}
		}
	}
}

/**
 * This method contains the logic of a Raft node in the candidate state.
 This state is used to elect a new leader. If a candidate wins the election, then it serves as leader for the rest of the term.
 In some situations an election will result in a split vote. In this case the term will end with no leader; a new term (with a new election)
 */
func (r *RaftNode) doCandidate() state {
	//fmt.Println("Transitioning to candidate state")
	//electionResults := make(chan bool)
	//electionTimeOut := r.makeElectionTimeout()
	//
	//r.State = CANDIDATE_STATE
	//r.LeaderAddress = nil
	//r.VotedFor = r.Id
	//
	//r.requestVotes(electionResults)
	//
	//
	//for {
	//	select {
	//	case off := <- r.gracefulExit:
	//		shutdown(off)
	//	case result := <- electionResults:
	//		if result {
	//			r.doLeader()
	//		} else {
	//			r.doFollower()
	//		}
	//	case vote := <- r.requestVote:
	//		fmt.Println("Request vote received by candidate")
	//		request := vote.request
	//		if r.handleCompetingRequests(vote) {
	//			r.currentTerm = request.Term
	//			r.LeaderAddress = nil
	//			r.VotedFor = request.CandidateId
	//			r.doFollower()
	//		} else {
	//			if r.currentTerm < request.Term {
	//				r.currentTerm = request.Term
	//			}
	//		}
	//	case _ = r.appendEntries:
	//	case _ = r.registerClient:
	//	case _ = r.clientRequest:
	//	case _ = electionTimeOut:
	//		return r.doCandidate()
	//	}
	//}
	return nil
}

/**
 * This method contains the logic of a Raft node in the leader state.
 */
func (r *RaftNode) doLeader() state {
	fmt.Println("Transitioning to leader state")
	//r.LeaderAddress = r.LocalAddr
	//r.VotedFor = ""
	//r.State = LEADER_STATE
	//
	//beats := r.makeHeartBeats()
	//fallback := make(chan bool)
	//finish := make(chan bool, 1)
	//
	//for {
	//	select {
	//	case off := <- r.gracefulExit:
	//		shutdown(off)
	//	case _ = <- r.appendEntries:
	//	case _ = <- beats:
	//		select {
	//			case <- finish:
	//			default:
	//				time.After(time.Millisecond * 1)
	//			}
	//
	//	case _ = <- fallback:
	//	case _ = <- r.appendEntries:
	//	case _ = <- r.registerClient:
	//	case _ = <- r.clientRequest:
	//
	//	}
	//}
	return nil
}

func (r *RaftNode) sendRequestFail()  {

}

func (r *RaftNode) sendNoop() {
	entries := make([]LogEntry, 1)
	entries[0] = LogEntry{r.GetLastLogIndex() + 1, r.GetCurrentTerm(), make([]byte, 0)}
	fmt.Println("NOOP logged with log index %d and term index %d", r.GetLastLogIndex() + 1, r.GetCurrentTerm())
	r.appendEntries
}

/*
The RequestVote RPC includes information about the candidate’s log, and the voter denies its vote if its own log is more up-to-date than that of the candidate.
Raft determines which of two logs is more up-to-date by comparing the index and term of the last entries in the logs.
If the logs have last entries with different terms, then the log with the later term is more up-to-date. If the logs end with the same term,
then whichever log is longer is more up-to-date.
 */
func (r *RaftNode) handleCompetingRequestVote(msg RequestVoteMsg) bool {
	request := msg.request
	reply := msg.reply
	currentTerm := r.GetCurrentTerm()
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
				if r.GetVotedFor() != "" && r.State == LEADER_STATE {
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
		nodes := r.GetOtherNodes()
		numNodes := len(nodes)
		votes := 1
		for _, node := range nodes {
			if node.Id == r.Id {
				continue
			}
			request := RequestVoteMsg{RequestVoteRequest{r.GetCurrentTerm(), r.GetLocalAddr()}}
			reply, _ := r.RequestVoteRPC(&node, request)
			if reply == nil {
				continue
			}
			if r.GetCurrentTerm() < reply.Term {
				fmt.Println("[Term outdated] Current = %d, Remote = %d", r.GetCurrentTerm(), reply.Term)
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

func makeElectionTimeout() <- chan time.Time {
	millis := rand.Int()%400 + 150
	return time.After(time.Millisecond * time.Duration(millis))
}

func (r *RaftNode) makeHeartBeats() <- chan time.Time {
	return time.After(r.conf.HeartbeatFrequency)
}

func shutdown(off bool) {
	if (off) {
		return nil
	}
}