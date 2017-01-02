package raft

import (
	"time"
	"math"
	"math/rand"
	"strconv"
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
				if req.Term > currTerm {
					r.setCurrentTerm(req.Term)
				}
			}

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
	r.Out("Transitioning to CANDIDATE_STATE")

	r.setCurrentTerm(r.GetCurrentTerm() + 1)
	r.State = CANDIDATE_STATE
	r.LeaderAddress = nil
	r.setVotedFor(r.Id)

	electionResults := make(chan bool)
	electionTimeout := makeElectionTimeout()
	r.requestVotes(electionResults)

	for {
		select {
		case shutdown := <-r.gracefulExit:
			if shutdown {
				return nil
			}

		case election := <-electionResults:
			if election {
				return r.doLeader
			} else {
				return r.doFollower
			}

		case vote := <-r.requestVote:
			req := vote.request
			candidate := req.CandidateId
			currTerm := r.GetCurrentTerm()
			if r.handleCompetingRequestVote(vote) {
				r.setCurrentTerm(req.Term)
				r.setVotedFor(candidate.Id)
				r.LeaderAddress = nil
				electionTimeout = makeElectionTimeout()
				return r.doFollower
			} else {
				// Other candidate has a less up to date log.
				if req.Term > currTerm {
					r.setCurrentTerm(req.Term)
				}
			}

		case appendEnt := <-r.appendEntries:
			req := appendEnt.request
			rep := appendEnt.reply
			currTerm := r.GetCurrentTerm()
			leader := req.LeaderId
			r.Out("Recieved heartbeat from %v", leader.Id)

			if req.Term >= currTerm {
				r.LeaderAddress = &leader
				r.setCurrentTerm(req.Term)
				r.setVotedFor("")
				rep <- AppendEntriesReply{currTerm, true}
				electionTimeout = makeElectionTimeout()
				return r.doFollower
			} else {
				rep <- AppendEntriesReply{currTerm, false}
			}

		case regClient := <-r.registerClient:
			rep := regClient.reply
			rep <- RegisterClientReply{ELECTION_IN_PROGRESS, 0, NodeAddr{"", ""}}

		case clientReq := <-r.clientRequest:
			rep := clientReq.reply
			rep <- ClientReply{ELECTION_IN_PROGRESS, "Not leader", NodeAddr{"", ""}}

		case <-electionTimeout:
		//It becomes a candidate again after the timeout.
			return r.doCandidate
		}
	}
	return nil
}

/**
 * This method contains the logic of a Raft node in the leader state.
 */
func (r *RaftNode) doLeader() state {
	r.Out("Transitioning to LEADER_STATE")
	r.State = LEADER_STATE
	r.setVotedFor("")
	r.LeaderAddress = r.GetLocalAddr()
	beats := r.makeHeartBeats()
	fallback := make(chan bool)
	finish := make(chan bool, 1)

	for _, n := range r.GetOtherNodes() {
		r.nextIndex[n.Id] = r.getLastLogIndex() + 1
	}

	r.sendNoop()
	r.Out("before noop The channel has %v\n", finish)
	finish <- true
	r.Out("The channel has %v\n", finish)
	for {
		select {

		case shutdown := <-r.gracefulExit:
			if shutdown {
				return nil
			}

		case <-beats:
				select {
				case <-finish:
					r.sendHeartBeats(fallback, finish)
					beats = r.makeHeartBeats()
				default:
					beats = time.After(time.Millisecond * 1)
				}
		case <-fallback:
			r.Out("heartbeat fallback")
			r.sendRequestFail()
			return r.doFollower

		case appendEnt := <-r.appendEntries:
			req := appendEnt.request
			rep := appendEnt.reply
			currTerm := r.GetCurrentTerm()
			leader := req.LeaderId
			r.Out("Recieved heartbeat from %v", leader.Id)

			if req.Term < currTerm {
				rep <- AppendEntriesReply{currTerm, false}
			} else {
				r.LeaderAddress = &leader
				r.setCurrentTerm(req.Term)
				r.setVotedFor("")
				rep <- AppendEntriesReply{currTerm, true}
				r.sendRequestFail()
				r.Out("appendEnt fallback")
				return r.doFollower
			}

		case vote := <-r.requestVote:
			r.Debug("request vote leader: entered\n")
			req := vote.request
			candidate := req.CandidateId
			currTerm := r.GetCurrentTerm()
			if r.handleCompetingRequestVote(vote) {
				r.setCurrentTerm(req.Term)
				r.setVotedFor(candidate.Id)
				r.LeaderAddress = nil
				r.sendRequestFail()
				r.Out("vote fallback by %v", candidate.Id)
				return r.doFollower
			} else {
				// Other candidate has a less up to date log.
				if req.Term > currTerm {
					r.setCurrentTerm(req.Term)
				}
			}

		case regClient := <-r.registerClient:
			req := regClient.request
			rep := regClient.reply

			entries := make([]LogEntry, 1)
			entries[0] = LogEntry{r.getLastLogIndex() + 1, r.GetCurrentTerm(), []byte(req.FromNode.Id), CLIENT_REGISTRATION, ""}
			r.appendLogEntry(entries[0])

			fallback, maj := r.sendAppendEntries(entries)

			if fallback {
				if maj {
					rep <- RegisterClientReply{OK, r.getLastLogIndex(), *r.LeaderAddress}
				} else {
					rep <- RegisterClientReply{REQUEST_FAILED, 0, *r.LeaderAddress}
				}
				r.sendRequestFail()
				r.LeaderAddress = nil
				r.Out("reg client fallback by %v")
				return r.doFollower
			}

			if !maj {
				rep <- RegisterClientReply{REQUEST_FAILED, 0, *r.LeaderAddress}
				r.truncateLog(r.getLastLogIndex())
			} else {
				rep <- RegisterClientReply{OK, r.getLastLogIndex(), *r.LeaderAddress}
			}

		case clientReq := <-r.clientRequest:
			req := clientReq.request
			rep := clientReq.reply
		//checking that it's registered
			entry := r.getLogEntry(req.ClientId)
			if entry.Command == CLIENT_REGISTRATION { 

				r.Out("The client is registered.")
				if false {
				//	need to think about oceanstore case here
				} else {
					entries := make([]LogEntry, 1)
					//Fill in the LogEntry based on the request data
					entries[0] = LogEntry{r.getLastLogIndex() + 1, r.GetCurrentTerm(), req.Data, req.Command, strconv.FormatUint(req.SequenceNumber, 10)}
					r.appendLogEntry(entries[0])

					r.requestMutex.Lock()
					r.requestMap[r.getLastLogIndex()] = clientReq
					r.requestMutex.Unlock()
				}
			} else {
				rep <- ClientReply{REQUEST_FAILED, "client is not registered.", *r.LeaderAddress}
			}

		}
	}

	return nil
}

func (r *RaftNode) sendRequestFail()  {

}

func (r *RaftNode) sendNoop() bool {
	entries := make([]LogEntry, 1)
	entries[0] = LogEntry{r.GetLastLogIndex() + 1, r.GetCurrentTerm(), make([]byte, 0), NOOP, ""}
	r.Out("NOOP logged with log index %d and term index %d", r.GetLastLogIndex() + 1, r.GetCurrentTerm())
	r.appendLogEntry(entries[0])
	return false
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

	if prevTerm > request.LastLogTerm {
		reply <- RequestVoteReply{currentTerm, false}
		return false
	} else if prevTerm < request.LastLogTerm {
		reply <- RequestVoteReply{currentTerm, true}
		return true
	} else {
		if prevIndex > request.LastLogIndex {
			reply <- RequestVoteReply{currentTerm, false}
			return false
		} else if prevIndex < request.LastLogIndex {
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
			request := RequestVoteRequest{r.GetCurrentTerm(), *r.GetLocalAddr(), r.commitIndex, r.getLogTerm(r.commitIndex), r.GetLastLogIndex() }
			reply, _ := r.RequestVoteRPC(&node, request)
			if reply == nil {
				continue
			}
			if r.GetCurrentTerm() < reply.Term {
				Out.Printf("[Term outdated] Current = %d, Remote = %d\n", r.GetCurrentTerm(), reply.Term)
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
	nodes := r.GetOtherNodes()
	for _, n := range nodes {
		if n.Id == r.Id {
			continue
		}
		r.AppendEntriesRPC(&n,
			AppendEntriesRequest{r.GetCurrentTerm(), *r.GetLocalAddr(),
				r.getLastLogIndex(), r.getLastLogTerm(), make([]LogEntry, 0),
				r.commitIndex})
	}
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
	go func() {
		nodes := r.GetOtherNodes()
		succ_nodes := 1
		for _, n := range nodes {
			if n.Id == r.Id {
				continue
			}
			if r.State != LEADER_STATE {
				return
			}
			prevLogIndex := r.nextIndex[n.Id] - 1
			if prevLogIndex > r.getLastLogIndex() {
				r.nextIndex[n.Id] = r.getLastLogIndex() + 1
				prevLogIndex = r.getLastLogIndex()
			}
			prevLogTerm := r.getLogEntry(prevLogIndex).Term
			reply, _ := r.AppendEntriesRPC(&n,
				AppendEntriesRequest{r.GetCurrentTerm(), *r.GetLocalAddr(),
					prevLogIndex, prevLogTerm, make([]LogEntry, 0),
					r.commitIndex})

			if reply == nil {
				continue
			}

			if r.GetCurrentTerm() < reply.Term {
				r.setCurrentTerm(reply.Term)
				fallback <- true
				return
			}

			if reply.Success {
				succ_nodes++
				nextIndex := r.nextIndex[n.Id]
				r.matchIndex[n.Id] = r.nextIndex[n.Id] - 1
				if (nextIndex - 1) != r.getLastLogIndex() {
					entries := r.getLogEntries(nextIndex, r.getLastLogIndex())
					reply, _ = r.AppendEntriesRPC(&n,
						AppendEntriesRequest{r.GetCurrentTerm(), *r.GetLocalAddr(),
							prevLogIndex, prevLogTerm, entries,
							r.commitIndex})

					if reply != nil && reply.Success {
						r.nextIndex[n.Id] = r.getLastLogIndex() + 1
						r.matchIndex[n.Id] = r.nextIndex[n.Id] - 1
					}
				}

			} else {
				r.nextIndex[n.Id]--
			}
		}

		for N := r.getLastLogIndex(); N > r.commitIndex; N-- {
			if r.hasMajority(N) && r.getLogTerm(N) == r.GetCurrentTerm() {
				r.commitIndex = N
			}
		}

		if r.lastApplied != r.commitIndex {
			i := r.lastApplied
			if r.lastApplied != 0 {
				i++
			}
			for ; i <= r.commitIndex; i++ {
				_ = r.processLog(*r.getLogEntry(i))
			}
			r.lastApplied = r.commitIndex
		}

		finish <- true
	}()
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