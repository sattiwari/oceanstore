package raft

type RequestVote struct {
	Term uint64
	CandidateId NodeAddr
}

type RequestVoteReply struct {
	Term uint64
	VoteGranted bool
}

func (r *RaftNode) requestVoteRPC(remoteNode *NodeAddr, request RequestVote) (*RequestVoteReply, error) {
	var reply RequestVoteReply
	err := makeRemoteCall(remoteNode, "requestVoteImpl", request, reply)
	if err != nil {
		return nil, err
	}
	return &reply, err
}