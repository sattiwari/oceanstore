package raft

type ClientStatus int

const (
	OK ClientStatus = iota
	NOT_LEADER
	ELECTION_IN_PROGRESS
	REQUEST_FAILED
)

type RegisterClientRequest struct {
	// The client address invoking request
	FromNode NodeAddr
}

type RegisterClientReply struct {
	Status ClientStatus
}

