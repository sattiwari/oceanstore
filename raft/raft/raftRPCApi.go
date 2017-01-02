package raft

import (
	"fmt"
	"net/rpc"
)

/* RPC connection map cache */
var connMap = make(map[string]*rpc.Client)

/*                                                                     */
/* Join RPC, used when a node in the cluster is first starting up so   */
/* it can notify a leader what their listening address is.             */
/*                                                                     */
type JoinRequest struct {
	RemoteNode NodeAddr
	FromAddr   NodeAddr
}

type JoinReply struct {
	Success bool
}

func JoinRPC(remoteNode *NodeAddr, fromAddr *NodeAddr) error {
	request := JoinRequest{RemoteNode: *remoteNode, FromAddr: *fromAddr}
	var reply JoinReply
	err := makeRemoteCall(remoteNode, "JoinImpl", request, &reply)
	if err != nil {
		return err
	}
	if !reply.Success {
		return fmt.Errorf("Unable to join Raft cluster\n")
	}
	return err
}

/*                                                                     */
/* StartNode RPC, once the first node in the cluster has all of the    */
/* addresses for all other nodes in the cluster it can then tell them  */
/* to transition into Follower state and start the Raft protocol.      */
/*                                                                     */
type StartNodeRequest struct {
	RemoteNode NodeAddr
	OtherNodes []NodeAddr
}

type StartNodeReply struct {
	Success bool
}

func StartNodeRPC(remoteNode NodeAddr, otherNodes []NodeAddr) error {
	request := StartNodeRequest{}
	request.RemoteNode = remoteNode

	request.OtherNodes = make([]NodeAddr, len(otherNodes))
	for i, n := range otherNodes {
		request.OtherNodes[i].Addr = n.Addr
		request.OtherNodes[i].Id = n.Id
	}

	var reply StartNodeReply
	err := makeRemoteCall(&remoteNode, "StartNodeImpl", request, &reply)
	if err != nil {
		return err
	}
	return err
}

/*                                                                     */
/* Raft RequestVote RPC, invoked by candidates to gather votes         */
/*                                                                     */
type RequestVoteRequest struct {
	/* The candidate's current term Id */
	Term uint64

	/* The cadidate Id currently requesting a node to vote for it. */
	CandidateId NodeAddr

	/* The index of the candidate's last log entry */
	LastLogIndex uint64

	/* The term of the candidate's last log entry */
	LastLogTerm uint64

	CurrentIndex uint64
}

type RequestVoteReply struct {
	/* The current term, for candidate to update itself */
	Term uint64

	/* True means candidate received vote */
	VoteGranted bool
}

func (r *RaftNode) RequestVoteRPC(remoteNode *NodeAddr, request RequestVoteRequest) (*RequestVoteReply, error) {
	if r.Testing.IsDenied(*r.GetLocalAddr(), *remoteNode) {
		return nil, ErrorTestingPolicyDenied
	}
	var reply RequestVoteReply
	err := makeRemoteCall(remoteNode, "RequestVoteImpl", request, &reply)
	if err != nil {
		return nil, err
	}
	return &reply, err
}

/*                                                                     */
/* Raft AppendEntries RPC, invoked by leader to replicate log entries; */
/* also used as a heartbeat between leaders and followers.             */
/*                                                                     */
type AppendEntriesRequest struct {
	/* The leader's term */
	Term uint64

	/* The ID of the leader, so that followers can redirect clients */
	LeaderId NodeAddr

	/* The index of the log entry immediately preceding new ones */
	PrevLogIndex uint64

	/* The term of the prevLogIndex entry */
	PrevLogTerm uint64

	/* The log entries the follower needs to store (empty for */
	/* heartbeat; may send more than one for efficiency)      */
	Entries []LogEntry

	/* The leader's commitIndex */
	LeaderCommit uint64
}

type AppendEntriesReply struct {
	/* The current term, for leader to update itself */
	Term uint64

	/* True if follower contained entry matching prevLogIndex and prevLogTerm*/
	Success bool
}

func (r *RaftNode) AppendEntriesRPC(remoteNode *NodeAddr, request AppendEntriesRequest) (*AppendEntriesReply, error) {
	if r.Testing.IsDenied(*r.GetLocalAddr(), *remoteNode) {
		return nil, ErrorTestingPolicyDenied
	}
	var reply AppendEntriesReply
	err := makeRemoteCall(remoteNode, "AppendEntriesImpl", request, &reply)
	if err != nil {
		return nil, err
	}
	return &reply, err
}

/* Node's can be in three possible states */
type ClientStatus int

const (
	OK ClientStatus = iota
	NOT_LEADER
	ELECTION_IN_PROGRESS
	REQ_FAILED
)

type FsmCommand int

const (
	HASH_CHAIN_ADD FsmCommand = iota
	HASH_CHAIN_INIT
	CLIENT_REGISTRATION
	INIT
	NOOP
	//Adding commands to interact with filesystem
	REMOVE //For deleting
	//Commands that modify the map that raft is in charge of
	GET //for querying
	SET //for modifying files
	LOCK
	UNLOCK
)

type ClientRequest struct {
	/* The unique client ID associated with this client session (received    */
	/* via a previous RegisterClient call).                                  */
	ClientId uint64

	/* A sequence number is associated to request to avoid duplicates */
	SequenceNum uint64

	/* Command to be executed on the state machine; it may affect state */
	Command FsmCommand

	/* Data to accompany the command to the state machine; it may affect state */
	Data []byte
}

type ClientReply struct {
	/* OK if state machine successfully applied command */
	Status ClientStatus

	/* State machine output, if successful */
	Response string

	/* In cases where the client contacted a non-leader, the node should     */
	/* reply with the correct current leader.                                */
	LeaderHint NodeAddr
}

func ClientRequestRPC(remoteNode *NodeAddr, request ClientRequest) (*ClientReply, error) {
	var reply ClientReply
	err := makeRemoteCall(remoteNode, "ClientRequestImpl", request, &reply)
	if err != nil {
		return nil, err
	}
	return &reply, err
}

type RegisterClientRequest struct {
	/* The client address invoking request */
	FromNode NodeAddr
}

type RegisterClientReply struct {
	/* OK if state machine registered client */
	Status ClientStatus

	/* Unique ID for client session */
	ClientId uint64

	/* In cases where the client contacted a non-leader, the node should */
	/* reply with the correct current leader.                            */
	LeaderHint NodeAddr
}

func RegisterClientRPC(remoteNode *NodeAddr, request RegisterClientRequest) (*RegisterClientReply, error) {
	var reply RegisterClientReply
	err := makeRemoteCall(remoteNode, "RegisterClientImpl", request, &reply)
	if err != nil {
		return nil, err
	}
	return &reply, err
}

/* Helper function to make a call to a remote node */
func makeRemoteCall(remoteNode *NodeAddr, method string, req interface{}, rsp interface{}) error {
	// Dial the server if we don't already have a connection to it
	remoteNodeAddrStr := remoteNode.Addr
	var err error
	client, ok := connMap[remoteNodeAddrStr]
	if !ok {
		client, err = rpc.Dial("tcp", remoteNode.Addr)
		if err != nil {
			return err
		}
		connMap[remoteNodeAddrStr] = client
	}

	// Make the request
	uniqueMethodName := fmt.Sprintf("%v.%v", remoteNodeAddrStr, method)
	// fmt.Println(uniqueMethodName)
	err = client.Call(uniqueMethodName, req, rsp)
	if err != nil {
		client.Close()
		delete(connMap, remoteNodeAddrStr)
		return err
	}

	return nil
}
