package raft

type RequestVoteMsg struct {
	request *RequestVote
	reply chan RequestVoteReply
}
