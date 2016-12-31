package raft

import (
	"fmt"
	"errors"
)

func (r *RaftNode) StartNode(request *StartNodeRequest) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.setOtherNodes(request.OtherNodes)
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
		errors.New("all nodes have joined the cluster")
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
	r.appendEntries <- AppendEntriesMsg{request:request, reply: reply}
	return <-reply, nil
}

func (r *RaftNode) RegisterClient(request *RegisterClientRequest) (RegisterClientReply, error) {
	Debug.Printf("Register client request received\n")
	reply := make(chan RegisterClientReply)
	r.registerClient <- RegisterClientMsg{request: request, reply: reply}
	return <-reply, nil
}

func (r *RaftNode) ClientRequest(request *ClientRequest) (ClientReply, error) {
	Debug.Printf("Client request received\n")
	reply := make(chan ClientReply)
	r.clientRequest <- ClientRequestMsg{request: request, reply: reply}
	return <-reply, nil
}

func (r *RaftNode) printOtherNodes(ctx string)  {
	otherStr := fmt.Sprintf("%v (%v) r.OtherNodes = [", ctx, r.Id)
	for _, otherNode := range r.GetOtherNodes() {
		otherStr += fmt.Sprintf("%v,", otherNode.Id)
	}
	Out.Printf(otherStr[:len(otherStr)-1] + "]\n")
}
