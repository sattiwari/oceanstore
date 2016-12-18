package raft

type StartNodeRequest struct {
	RemoteNode NodeAddr
	OtherNodes []NodeAddr
}

type StartNodeReply struct {
	Success bool
}

type RequestVote struct {
	Term uint64
	CandidateId NodeAddr
	CandidateLastLogTerm uint64
	CandidateLastLogIndex uint64
}

type RequestVoteReply struct {
	Term uint64
	VoteGranted bool
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

func JoinRPC(remoteNode *NodeAddr, fromAddr *NodeAddr) error {

}

func (r *RaftNode) requestVoteRPC(remoteNode *NodeAddr, request RequestVote) (*RequestVoteReply, error) {
	var reply RequestVoteReply
	err := makeRemoteCall(remoteNode, "requestVoteImpl", request, reply)
	if err != nil {
		return nil, err
	}
	return &reply, err
}