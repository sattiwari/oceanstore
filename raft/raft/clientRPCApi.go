package raft

type ClientStatus int

const (
	OK ClientStatus = iota
	NOT_LEADER
	ELECTION_IN_PROGRESS
	REQUEST_FAILED
)

type FsmCommand int

const (
	HASH_CHAIN_ADD FsmCommand = iota
	HASH_CHAIN_INIT
	CLIENT_REGISTRATION
	INIT
	NOOP
)

type ClientRequest struct {
	//unique id associated with client session. Recevied via previous register client call
	ClientId uint64

	//avoids duplicates
	SequenceNumber uint64

	//Command to be executed by state machine. It may affect state
	Command FsmCommand

	//Data to accompany the command to state machine
	Data []byte
}

type ClientReply struct {
	//OK if the state machine successfully applied command
	Status ClientStatus

	//state machine response
	Response string

	//a non leader node should reply the correct leader
	LeaderHint NodeAddr
}


type RegisterClientRequest struct {
	// The client address invoking request
	FromNode NodeAddr
}

type RegisterClientReply struct {
	//ok if the state machine registered client
	Status ClientStatus

//	unique id for the client session
	ClientId uint64

	// if the node contacted is not leader, it tells the correct leader
	LeaderHint NodeAddr
}

func RegisterClientRPC(remoteNode *NodeAddr, request RegisterClientRequest) (*RegisterClientReply, error) {
	var reply RegisterClientReply
	err := makeRemoteCall(remoteNode, "RegisterClientImpl", request, &reply)
	if err != nil {
		return nil, err
	}
	return &reply, nil
}

func ClientRequestRPC(remoteNode *NodeAddr, request ClientRequest) (*ClientReply, error) {
	var reply ClientReply
	err := makeRemoteCall(remoteNode, "ClientRequestImpl", request, &reply)
	if err != nil {
		return nil, err
	}
	return &reply, nil
}

