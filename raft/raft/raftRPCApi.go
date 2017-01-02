package raft

type StartNodeRequest struct {
	RemoteNode NodeAddr
	OtherNodes []NodeAddr
}

type StartNodeReply struct {
	Success bool
}

// Once the first node in the cluster has addresses of all other nodes, it can tell them to transition to Follower state
//and start the raft protocol
func StartNodeRPC(remoteNode NodeAddr, otherNodes []NodeAddr) error {
	request := StartNodeRequest{RemoteNode: remoteNode}
	request.OtherNodes = make([]NodeAddr, len(otherNodes))
	for i, n := range otherNodes {
		request.OtherNodes[i].Address = n.Address
		request.OtherNodes[i].Id = n.Id
	}
	var reply StartNodeReply
	err := makeRemoteCall(&remoteNode, "StartNodeImpl", request, &reply)
	if err != nil {
		return err
	}
	return nil
}

type JoinRequest struct {
	RemoteNode NodeAddr
	FromNode NodeAddr
}

type JoinReply struct {
	Success bool
}

func JoinRPC(remoteNode *NodeAddr, fromAddr *NodeAddr) error {
	request := JoinRequest{RemoteNode: *remoteNode, FromNode: *fromAddr}
	var reply JoinReply
	err := makeRemoteCall(remoteNode, "JoinImpl", request, &reply)
	if err != nil {
		return err
	}
	if !reply.Success {
		Error.Printf("%v Unable to join cluster", fromAddr.Id)
	}
	return err
}


type RequestVoteRequest struct {
	Term         uint64
	CandidateId  NodeAddr
	LastLogIndex uint64
	LastLogTerm  uint64
	CurrentIndex uint64
}

type RequestVoteReply struct {
	Term uint64
	VoteGranted bool
}

func (r *RaftNode) RequestVoteRPC(remoteNode *NodeAddr, request RequestVoteRequest) (*RequestVoteReply, error) {
	if r.Testing.IsDenied(*r.GetLocalAddr(), *remoteNode) {
		return nil, ErrorTestingPolicyDenied
	}
	var reply RequestVoteReply
	err := makeRemoteCall(remoteNode, "RequestVoteImpl", request, &reply)
	if err != nil {
		return nil, err
	}
	return &reply, err
}

//Invoked by leader to replicate log entries. Also, used as heartbeat between leaders and followers.
/*
Log Matching Proprty - If two entries in different logs have the same index and term, then the logs are identical in all preceding entries.
This property is guaranteed by a simple consistency check performed by AppendEntries. When sending an AppendEntries RPC,
the leader includes the index and term of the entry in its log that immediately precedes the new entries.
If the follower does not find an entry in its log with the same index and term, then it refuses the new entries.
The consistency check acts as an induction step: the initial empty state of the logs satisfies the Log Matching Property,
and the consistency check preserves the Log Matching Property whenever logs are extended. As a result, whenever AppendEntries returns successfully,
the leader knows that the follower’s log is identical to its own log up through the new entries.
 */
type AppendEntriesRequest struct {
	//	leader's term
	Term uint64

	//	id of the leader so that followers can redirect client
	LeaderId NodeAddr

	// index of the log entry immediately precceding the new ones
	PrevLogIndex uint64

	// the term of the previous logIndex entry
	PrevLogTerm uint64

	//	the log entries follower may need to store - empty for heart beats
	Entries []LogEntry

	//	the leader's commit index
	LeaderCommit uint64
}

type AppendEntriesReply struct {
	//	current term for the leader to update itself
	Term uint64

	//	true if follower contained entry matching prevLogIndex & prevLogTerm
	Success bool
}

func (r *RaftNode) AppendEntriesRPC(remoteNode *NodeAddr, request AppendEntriesRequest) (*AppendEntriesReply, error)  {
	if r.Testing.IsDenied(*r.GetLocalAddr(), *remoteNode) {
		return nil, ErrorTestingPolicyDenied
	}
	var reply AppendEntriesReply
	err := makeRemoteCall(remoteNode, "AppendEnteriesImpl", request, &reply)
	if err != nil {
		return nil, err
	}
	return &reply, err
}