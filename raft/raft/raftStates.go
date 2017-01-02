package raft

import (
	"math"
	"math/rand"
	"strconv"
	"time"
)

type state func() state

/**
 * This method contains the logic of a Raft node in the follower state.
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
		//rep := vote.reply
			currTerm := r.GetCurrentTerm()
			candidate := req.CandidateId
		//votedFor := r.GetVotedFor()

		//if votedFor == "" || votedFor == candidate.Id {
		// TODO: Check log to see si el vato la arma.
			if r.handleCompetingRequestVote(vote) {
				// return r.doFollower
				r.setCurrentTerm(req.Term)
				// Set voted for field (already voted)
				r.setVotedFor(candidate.Id)

				// Election in progess. There is no leader
				// Set to nil
				r.LeaderAddress = nil

				// Respond true, reset eleciton timeout.
				// rep <- RequestVoteReply{currTerm, true}
				electionTimeout = makeElectionTimeout()
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
		// r.Out("Recieved heartbeat from %v", req.LeaderId.Id)

			if req.Term < currTerm {
				r.Out("Rejected heartbeat from %v because of terms", req.LeaderId.Id)
				rep <- AppendEntriesReply{currTerm, false}
			} else {
				r.LeaderAddress = &req.LeaderId
				r.setCurrentTerm(req.Term)
				// Set voted for field
				r.setVotedFor("")

				entry := r.getLogEntry(req.PrevLogIndex)
				if entry == nil || entry.TermId != req.PrevLogTerm { //|| req.PrevLogIndex != r.getLastLogIndex() {
					r.Out("Rejected heartbeat from %v because of miss", req.LeaderId.Id)
					rep <- AppendEntriesReply{currTerm, false}
				} else {
					if req.PrevLogIndex+1 > r.commitIndex {
						r.truncateLog(req.PrevLogIndex + 1)
					} else {
						///// r.truncateLog(r.commitIndex + 1)
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
							//if reply.Status == REQ_FAILED {
							//	panic("This should not happen ever! PLZ FIX")
							//}
						}
						r.lastApplied = r.commitIndex
					}

					rep <- AppendEntriesReply{currTerm, true}
				}
			}
			electionTimeout = makeElectionTimeout()
		case regClient := <-r.registerClient:
		// req := regClient.request
			rep := regClient.reply
			if r.LeaderAddress != nil {
				rep <- RegisterClientReply{NOT_LEADER, 0, *r.LeaderAddress}
			} else {
				// Return election in progress.
				rep <- RegisterClientReply{ELECTION_IN_PROGRESS, 0, NodeAddr{"", ""}}
			}
		case <-electionTimeout:
			return r.doCandidate
		case clientReq := <-r.clientRequest:
		//req := clientReq.request
			rep := clientReq.reply
		// req := regClient.request
			if r.LeaderAddress != nil {
				rep <- ClientReply{NOT_LEADER, "Not leader", *r.LeaderAddress}
			} else {
				// Return election in progress.
				rep <- ClientReply{NOT_LEADER, "Not leader", NodeAddr{"", ""}}
			}
		}
	}
}

/**
 * This method contains the logic of a Raft node in the candidate state.
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

		/*if req.Term < currTerm {
			rep <- AppendEntriesReply{currTerm, false}
		} else {*/
			if req.Term >= currTerm {
				r.LeaderAddress = &leader
				r.setCurrentTerm(req.Term)
				r.setVotedFor("")
				rep <- AppendEntriesReply{currTerm, true}
				electionTimeout = makeElectionTimeout()
				return r.doFollower
				//}
			} else {
				rep <- AppendEntriesReply{currTerm, false}
			}

		case regClient := <-r.registerClient:
		// req := regClient.request
			rep := regClient.reply
			rep <- RegisterClientReply{ELECTION_IN_PROGRESS, 0, NodeAddr{"", ""}}
		case clientReq := <-r.clientRequest:
		//req := clientReq.request
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
	beats := r.makeBeats()
	fallback := make(chan bool)
	finish := make(chan bool, 1)

	// Set up all next index values
	for _, n := range r.GetOtherNodes() {
		r.nextIndex[n.Id] = r.getLastLogIndex() + 1
	}

	// r.sendHeartBeats()
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
		//TODO: Is it bc this one has to move down?
		case <-beats:
				select {
				case <-finish:
				// r.Out("sendHeartBeats: entered\n")
					r.sendHeartBeats(fallback, finish)
					beats = r.makeBeats()
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

				// Set voted for field (already voted)
				r.setVotedFor(candidate.Id)

				// Election in progess. There is no leader
				// Set to nil
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
			entries[0] = LogEntry{r.getLastLogIndex() + 1, r.GetCurrentTerm(), CLIENT_REGISTRATION, []byte(req.FromNode.Id), ""}
		//we add it to the log
			r.appendLogEntry(entries[0])

			fallback, maj := r.sendAppendEntries(entries)

			if fallback {
				if maj {
					rep <- RegisterClientReply{OK, r.getLastLogIndex(), *r.LeaderAddress}
				} else {
					rep <- RegisterClientReply{REQ_FAILED, 0, *r.LeaderAddress}
				}
				//not sure we need to do this
				r.sendRequestFail()
				r.LeaderAddress = nil
				r.Out("reg client fallback by %v")
				return r.doFollower
			}

			if !maj {
				rep <- RegisterClientReply{REQ_FAILED, 0, *r.LeaderAddress}
				//Truncate log, didnt happen
				r.truncateLog(r.getLastLogIndex())
			} else {
				//r.processLog(entries[0])
				//r.commitIndex++
				//r.sendAppendEntries(make([]LogEntry, 0))
				rep <- RegisterClientReply{OK, r.getLastLogIndex(), *r.LeaderAddress}
			}

		case clientReq := <-r.clientRequest:
			req := clientReq.request
			rep := clientReq.reply
		//checking that it's registered
			entry := r.getLogEntry(req.ClientId)
			if entry.Command == CLIENT_REGISTRATION { //&& string(entry.Data) == req. {

				r.Out("The client is registered.")
				//I think here we need to special case the GET
				if req.Command == GET {
					key := string(req.Data)
					var message string
					if req.Data == nil {
						message = "FAIL: The key cannot be nil"
					} else if val, ok := r.fileMap[key]; ok {
						message = "SUCCESS:" + val
					} else {
						message = "FAIL: The key is not in the dictionary"
					}
					if r.LeaderAddress != nil {
						rep <- ClientReply{OK, message, *r.LeaderAddress}
					} else {
						// Return election in progress.
						rep <- ClientReply{OK, message, NodeAddr{"", ""}}
					}
				} else {
					entries := make([]LogEntry, 1)
					//Fill in the LogEntry based on the request data
					entries[0] = LogEntry{r.getLastLogIndex() + 1, r.GetCurrentTerm(), req.Command, req.Data, strconv.FormatUint(req.SequenceNum, 10)}
					r.appendLogEntry(entries[0])

					//we now send it to everyone
					// fallback, maj := r.sendAppendEntries(entries)
					r.requestMutex.Lock()
					r.requestMap[r.getLastLogIndex()] = clientReq
					r.requestMutex.Unlock()
					/*
						if fallback {
							//Ahora no estoy tan seguro de hacer esto
							r.sendRequestFail()
							r.Out("client req fallback by %v")
							return r.doFollower
						}

						if !maj {
							//rep <- ClientReply{REQ_FAILED, 0, *r.LeaderAddr}
							//Truncate log, didn't happen
							//r.truncateLog(r.getLastLogIndex())
						} else {
							//Nos esperamos a que el heartbeat haga commit.
							//We now cache the response in the response map to reply to it later
							//r.requestMap[r.SequenceNum] = clientReq
						}
						// electionTimeout = makeElectionTimeout()
					*/
				}
			} else {
				rep <- ClientReply{REQ_FAILED, "client is not registered.", *r.LeaderAddress}
			}

		}
	}

	return nil
}

//El pedo con esta es que si si se aprobo pero el lider tuvo fallback
//le va a mandar un fail antes de que process log le de en caso de que si
// se haya logrado.
func (r *RaftNode) sendRequestFail() {
	//r.requestMutex.Lock()
	////We need to make sure that a leader does not hang in case
	////the client is no longer available
	//for k, v := range r.requestMap {
	//	r.Out("The leader address is: %v\n", r.LeaderAddr)
	//	r.Out("The reply is: %v\n", v.reply)
	//	var leader NodeAddr
	//	if r.LeaderAddr != nil {
	//		leader = *r.LeaderAddr
	//	} else {
	//		leader = NodeAddr{"",""}
	//	}
	//	select{
	//		case v.reply <- ClientReply{REQ_FAILED, "", leader}:
	//			//If we handle it we no longer need it
	//			delete(r.requestMap, k)
	//			fmt.Println("Message successfully sent!")
	//		default:
	//			fmt.Println("Client no longer available")
	//		}
	//}
	//r.requestMutex.Unlock()
}

func (r *RaftNode) sendNoop() bool {
	// Send NOOP as first entry to all nodes.
	entries := make([]LogEntry, 1)
	entries[0] = LogEntry{r.getLastLogIndex() + 1, r.GetCurrentTerm(), NOOP, make([]byte, 0), ""}
	r.Out("Noop logged with inde %v and term %v\n", r.getLastLogIndex()+1, r.GetCurrentTerm())
	r.appendLogEntry(entries[0])
	//f, _ := r.sendAppendEntries(entries)

	//Leader should fall back, but just got elected. How?
	//if f {
	//	return true
	//}
	//No processLog until we increase the commit index. Which happens in
	//heartbeat.
	// if maj {
	// 	r.processLog(entries[0])
	// 	r.commitIndex++
	// }
	// r.sendAppendEntries(make([]LogEntry, 0))

	return false
}

/**
 * This function is called when the node is a candidate or leader, and a
 * competing RequestVote is called. It will return true if the caller should
 * fall back to the follower state.
 */
// DICE QUE ES SOLO CANDIDATE O LEADER, PERO SEGUN YO FOLLOWER JALA TAMBIEN NO?
func (r *RaftNode) handleCompetingRequestVote(msg RequestVoteMsg) bool {
	req := msg.request
	rep := msg.reply
	prevIndex := r.commitIndex
	prevTerm := r.getLogTerm(prevIndex)
	currTerm := r.GetCurrentTerm()

	//if r.GetVotedFor() != "" {
	//	rep <- RequestVoteReply{currTerm, false}
	//	return false
	//}

	if prevTerm > req.LastLogTerm { // My commit term is greater, no vote.
		rep <- RequestVoteReply{currTerm, false}
		return false
	} else if prevTerm < req.LastLogTerm { // My commit term is smaller, vote.
		rep <- RequestVoteReply{currTerm, true}
		return true
	} else { // Same commit indexes
		if prevIndex > req.LastLogIndex { // My commit index is greater, no vote.
			rep <- RequestVoteReply{currTerm, false}
			return false
		} else if prevIndex < req.LastLogIndex { // My commit index is smaller, vote.
			rep <- RequestVoteReply{currTerm, true}
			return true
		} else { // Commit index and term are the same. Tiebreaker with currTerms
			// Previous election or election that I already voted for or
			// currently pursuing. No vote.
			if req.CurrentIndex < r.getLastLogIndex() {
				rep <- RequestVoteReply{currTerm, false}
				return false
			} else if req.CurrentIndex > r.getLastLogIndex() {
				rep <- RequestVoteReply{currTerm, true}
				return true
			} else {
				if req.Term < currTerm {
					rep <- RequestVoteReply{currTerm, false}
					return false
				} else if req.Term == currTerm {
					if r.GetVotedFor() != "" || r.State == LEADER_STATE {
						rep <- RequestVoteReply{currTerm, false}
						return false
					} else {
						rep <- RequestVoteReply{currTerm, true}
						return true
					}

				} else { // Greater election term, give vote.
					rep <- RequestVoteReply{currTerm, true}
					return true
				}
			}
		}
	}
}

/**
 * This function is called to request votes from all other nodes. It takes
 * a channel which the result of the vote should be sent over: true for
 * successful election, false otherwise.
 */
func (r *RaftNode) requestVotes(electionResults chan bool) {
	go func() {
		nodes := r.GetOtherNodes()
		num_nodes := len(nodes)
		votes := 1
		// r.setVotedFor(r.Id)
		for _, n := range nodes {
			if n.Id == r.Id {
				continue
			}
			//reply, _ := r.RequestVoteRPC(&n,
			//	RequestVoteRequest{r.GetCurrentTerm(), *r.GetLocalAddr(),
			//		r.getLastLogIndex(), r.getLastLogTerm()})
			reply, _ := r.RequestVoteRPC(&n,
				RequestVoteRequest{r.GetCurrentTerm(), *r.GetLocalAddr(),
					r.commitIndex, r.getLogTerm(r.commitIndex), r.getLastLogIndex()})

			if reply == nil {
				// Could not reach node for vote.
				continue
			}

			if r.GetCurrentTerm() < reply.Term {
				r.Out("YA VALIO %d < %d\n", r.GetCurrentTerm(), reply.Term)
				electionResults <- false // YA VALIO MADRE
				return
			}

			// r.Debug("RequestVotes: votes %v term %d\n", reply.VoteGranted, r.GetCurrentTerm())
			if reply.VoteGranted {
				votes++
			}

		}
		if votes > num_nodes/2 {
			electionResults <- true
			r.Debug("RequestVotes: won with %d votes\n", votes)
			return
		}
		electionResults <- false
		r.Out("RequestVotes: lost with %d votes\n", votes)
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
		num_nodes := len(nodes)
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
			prevLogTerm := r.getLogEntry(prevLogIndex).TermId
			//r.Out("HeartBeat to %v", n.Id)
			reply, _ := r.AppendEntriesRPC(&n,
				AppendEntriesRequest{r.GetCurrentTerm(), *r.GetLocalAddr(),
					prevLogIndex, prevLogTerm, make([]LogEntry, 0),
					r.commitIndex})

			if reply == nil {
				continue
			}

			if r.GetCurrentTerm() < reply.Term {
				r.setCurrentTerm(reply.Term)
				// return true, false
				fallback <- true
				return
			}

			if reply.Success {
				succ_nodes++
				nextIndex := r.nextIndex[n.Id]
				//Aqui segun you se vuelven iguales porque ya apendearon
				r.matchIndex[n.Id] = r.nextIndex[n.Id] - 1
				if (nextIndex - 1) != r.getLastLogIndex() {
					// TODO
					// r.Out("nextIndex vs lastLog %v, %v", nextIndex, r.getLastLogIndex())
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

		// Updates commit index according to what the followers have in their logs already.
		for N := r.getLastLogIndex(); N > r.commitIndex; N-- {
			if r.hasMajority(N) && r.getLogTerm(N) == r.GetCurrentTerm() {
				r.commitIndex = N
			}
		}

		// Calls process log from lastApplied until commit index.
		if r.lastApplied != r.commitIndex {
			i := r.lastApplied
			if r.lastApplied != 0 {
				i++
			}
			for ; i <= r.commitIndex; i++ {
				_ = r.processLog(*r.getLogEntry(i))
				//if reply.Status == REQ_FAILED {
				//	panic("This should not happen ever! PLZ FIX")
				//}
			}
			r.lastApplied = r.commitIndex
		}

		if succ_nodes > num_nodes/2 {
			// return false, true
		}

		// return false, false
		finish <- true
	}()
}

func (r *RaftNode) sendAppendEntries(entries []LogEntry) (fallBack, sentToMajority bool) {
	nodes := r.GetOtherNodes()
	num_nodes := len(nodes)
	succ_nodes := 1

	for _, n := range nodes {
		if n.Id == r.Id {
			continue
		}
		//we have the - 1 because we just added it (before this call) to our own log
		prevLogIndex := r.getLastLogIndex() - 1
		prevLogTerm := r.getLogTerm(prevLogIndex)
		reply, _ := r.AppendEntriesRPC(&n,
			AppendEntriesRequest{r.GetCurrentTerm(), *r.GetLocalAddr(),
				prevLogIndex, prevLogTerm, entries,
				r.commitIndex})

		if reply == nil {
			continue
		}

		if r.GetCurrentTerm() < reply.Term {
			r.setCurrentTerm(reply.Term)
			return true, false
		}

		if reply.Success {
			succ_nodes++
			//We increase the next index for this node
			//It's lastLogIndex + 1 porque en lastLogIndex es donde lo
			//acaba de agregar.
			r.nextIndex[n.Id] = r.getLastLogIndex() + 1
		} else {
			continue
		}

	}

	if succ_nodes > num_nodes/2 {
		return false, true
	}

	return false, false
}

/**
 * This function will use time.After to create a random timeout.
 */
func makeElectionTimeout() <-chan time.Time {
	millis := rand.Int()%400 + 150
	return time.After(time.Millisecond * time.Duration(millis))
}

func (r *RaftNode) makeBeats() <-chan time.Time {
	return time.After(r.conf.HeartbeatFrequency)
}