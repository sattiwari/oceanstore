package raft

func (r *RaftNode) StartNode(request *StartNodeRequest) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.SetOtherNodes(request.OtherNodes)
	go r.run()
	return nil
}

func (r *RaftNode) Join(request *JoinRequest) error {
	if len(r.GetOtherNodes()) == r.conf.ClusterSize {
		for _, otherNode := range r.GetOtherNodes() {
			if otherNode.Id != r.Id {
				StartNodeRPC(otherNode, r.GetOtherNodes())
				return nil
			}
		}
		error("all nodes have joined the cluster")
	} else {
		r.AppendOtherNodes(request.FromNode)
	}
	return nil
}

func (r *RaftNode) RequestVote(request *RequestVoteRequest) (RequestVoteReply, error) {
	Debug.Printf("RequestVote request received\n")
	reply := make(chan RequestVoteReply)
	r.requestVote <- RequestVoteMsg{request:request, reply: reply}
	return <-reply, nil
}

func (r *RaftNode) AppendEntries(request *AppendEntriesRequest) (AppendEntriesReply, error) {
	Debug.Printf("AppendEntries request received\n")
	reply := make(chan AppendEntriesReply)
	r.requestVote <- AppendEntriesMsg{request:request, reply: reply}
	return <-reply, nil
}

//func (r *RaftNode) RegisterClient(request *RegisterClientRequest) error {
//
//}
//
//func (r *RaftNode) ClientRequest(request *ClientRequest) error {
//
//}
