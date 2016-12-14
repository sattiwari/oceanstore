package raft

type AppendEntriesMsg struct {

}

type RequestVoteMsg struct {
	request *RequestVote
	reply chan RequestVoteReply
}

type ClientRequestMsg struct {

}

type ClientRegistration struct {

}
