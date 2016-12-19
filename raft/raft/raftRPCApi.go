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
	request := StartNodeRequest{RemoteNode: otherNodes}
	var reply StartNodeReply
	err := makeRemoteCall(remoteNode, "StartNodeImpl", request, &reply)
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
	Term uint64
	CandidateId NodeAddr
	CandidateLastLogTerm uint64
	CandidateLastLogIndex uint64
}

type RequestVoteReply struct {
	Term uint64
	VoteGranted bool
}

func (r *RaftNode) RequestVoteRPC(remoteNode *NodeAddr, request RequestVoteRequest) (*RequestVoteReply, error) {
	var reply RequestVoteReply
	err := makeRemoteCall(remoteNode, "requestVoteImpl", request, reply)
	if err != nil {
		return nil, err
	}
	return &reply, err
}

//Invoked by leader to replicate log entries. Also, used as heartbeat between leaders and followers.
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