package raft

import (
	"fmt"
)

func (r *RaftNode) Join(req *JoinRequest) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(r.GetOtherNodes()) == r.conf.ClusterSize {
		for _, otherNode := range r.GetOtherNodes() {
			if otherNode.Id == req.FromAddr.Id {
				StartNodeRPC(otherNode, r.GetOtherNodes())
				return nil
			}
		}
		r.Error("Warning! Unrecognized node tried to join after all other nodes have joined.\n")
		return fmt.Errorf("All nodes have already joined this Raft cluster\n")
	} else {
		r.AppendOtherNodes(req.FromAddr)
	}
	return nil
}

func (r *RaftNode) StartNode(req *StartNodeRequest) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.SetOtherNodes(req.OtherNodes)
	r.printOtherNodes("StartNode")

	// Start the Raft finite-state-machine, initially in follower state
	go r.run()

	return nil
}

type RequestVoteMsg struct {
	request *RequestVoteRequest
	reply   chan RequestVoteReply
}

func (r *RaftNode) RequestVote(req *RequestVoteRequest) (RequestVoteReply, error) {
	r.Out("RequestVote request received\n")
	reply := make(chan RequestVoteReply)
	r.requestVote <- RequestVoteMsg{req, reply}
	return <-reply, nil
}

type AppendEntriesMsg struct {
	request *AppendEntriesRequest
	reply   chan AppendEntriesReply
}

func (r *RaftNode) AppendEntries(req *AppendEntriesRequest) (AppendEntriesReply, error) {
	r.Debug("AppendEntries request received\n")
	reply := make(chan AppendEntriesReply)
	r.appendEntries <- AppendEntriesMsg{req, reply}
	return <-reply, nil
}

type ClientRequestMsg struct {
	request *ClientRequest
	reply   chan ClientReply
}

func (r *RaftNode) ClientRequest(req *ClientRequest) (ClientReply, error) {
	r.Debug("ClientRequest request received\n")
	reply := make(chan ClientReply)
	cr, exists := r.CheckRequestCache(*req)
	if exists {
		return *cr, nil
	} else {
		r.clientRequest <- ClientRequestMsg{req, reply}
		return <-reply, nil
	}
}

type RegisterClientMsg struct {
	request *RegisterClientRequest
	reply   chan RegisterClientReply
}

func (r *RaftNode) RegisterClient(req *RegisterClientRequest) (RegisterClientReply, error) {
	r.Debug("ClientRequest request received\n")
	reply := make(chan RegisterClientReply)
	r.registerClient <- RegisterClientMsg{req, reply}
	return <-reply, nil
}

func (r *RaftNode) printOtherNodes(ctx string) {
	otherStr := fmt.Sprintf("%v (%v) r.OtherNodes = [", ctx, r.Id)
	for _, otherNode := range r.GetOtherNodes() {
		otherStr += fmt.Sprintf("%v,", otherNode.Id)
	}
	Out.Printf(otherStr[:len(otherStr)-1] + "]\n")
}
