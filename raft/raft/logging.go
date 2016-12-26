package raft

import (
	"log"
	"io/ioutil"
	"os"
	"fmt"
)

var Debug *log.Logger
var Out *log.Logger
var Error *log.Logger

//initialize the loggers
func init()  {
	Debug = log.New(ioutil.Discard, "", log.Ltime | log.Lshortfile)
	Out = log.New(os.Stdout, "", log.Ltime | log.Lshortfile)
	Error = log.New(os.Stdout, "ERROR: ", log.Ltime | log.Lshortfile)
}

func (r *RaftNode) Out(formatString string, args ...interface{}) {
	Out.Output(2, fmt.Sprintf("(%v/%v) %v", r.Id, r.State, fmt.Sprintf(formatString, args ...)))
}

func (r *RaftNode) Debug(formatString string, args ...interface{}) {
	Debug.Output(2, fmt.Sprintf("(%v/%v) %v", r.Id, r.State, fmt.Sprintf(formatString, args ...)))
}

func (r *RaftNode) Error(formatString string, args ...interface{}) {
	Debug.Output(2, fmt.Sprintf("(%v/%v) %v", r.Id, r.State, fmt.Sprintf(formatString, args ...)))
}

func (s NodeState) String() string {
	switch s {
	case FOLLOWER_STATE:
		return "follower"
	case CANDIDATE_STATE:
		return "candidate"
	case LEADER_STATE:
		return "leader"
	case JOIN_STATE:
		return "joining"
	default:
		return "unknown"
	}
}

func FsmCommandString(cmd FsmCommand) string {
	return ""
}

func (r *RaftNode) ShowState()  {

}

func (r *RaftNode) PrintLogCache() {

}
