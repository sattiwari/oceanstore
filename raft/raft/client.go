package raft

import (
	"time"
)

const MAX_RETRIES = 5

type Client struct {
	LocalAddr *NodeAddr
	Id        uint64
	Leader    NodeAddr
	SeqNum    uint64
}

func CreateClient(remoteAddr NodeAddr) (cp *Client, err error) {
	cp = new(Client)

	request := RegisterClientRequest{}

	var reply *RegisterClientReply

	retries := 0

	LOOP:
	for retries < MAX_RETRIES {
		reply, err = RegisterClientRPC(&remoteAddr, request)
		if err != nil {
			return
		}
		switch reply.Status {
		case OK:
			Out.Printf("%v is the leader. Client successfully created.\n", remoteAddr)
			break LOOP
		case REQ_FAILED:
			Error.Printf("Request failed...\n")
			retries++
		case NOT_LEADER:
			// The person we've contacted isn't the leader. Use
			// their hint to find the leader
			Out.Printf("%v is not the leader, but thinks that %v is\n", remoteAddr, reply.LeaderHint)
			remoteAddr = reply.LeaderHint
		case ELECTION_IN_PROGRESS:
			// An election is in progress. Accept the hint
			// and wait an appropriate amount of time, so the
			// election can finish.
			Out.Printf("%v is not the leader, but thinks that %v is\n", remoteAddr, reply.LeaderHint)
			remoteAddr = reply.LeaderHint
			time.Sleep(time.Millisecond * 200)
		default:
		}
	}

	// We've registered with the leader.
	cp.Id = reply.ClientId
	cp.Leader = remoteAddr

	return
}

func (c *Client) SendRequest(command FsmCommand, data []byte) (err error) {

	request := ClientRequest{
		c.Id,
		c.SeqNum,
		command,
		data,
	}
	c.SeqNum += 1

	var reply *ClientReply

	retries := 0

	LOOP:
	for retries < MAX_RETRIES {
		reply, err = ClientRequestRPC(&c.Leader, request)
		if err != nil {
			return
		}
		switch reply.Status {
		case OK:
			Debug.Printf("%v is the leader\n", c.Leader)
			Out.Printf("Request returned \"%v\".\n", reply.Response)
			break LOOP
		case REQ_FAILED:
			Error.Printf("Request failed: %v\n", reply.Response)
			retries++
			break LOOP
		case NOT_LEADER:
			// The person we've contacted isn't the leader. Use
			// their hint to find the leader
			c.Leader = reply.LeaderHint
		case ELECTION_IN_PROGRESS:
			// An election is in progress. Accept the hint
			// and wait an appropriate amount of time, so the
			// election can finish.
			c.Leader = reply.LeaderHint
			time.Sleep(time.Millisecond * 200)
		}
	}
	return
}
//Similar to the function above but it returns a response
func (c *Client) SendRequestWithResponse(command FsmCommand, data []byte) (reply *ClientReply, err error) {
	request := ClientRequest{
		c.Id,
		c.SeqNum,
		command,
		data,
	}
	c.SeqNum += 1

	//var reply *ClientReply

	retries := 0
	for retries < MAX_RETRIES {
		reply, err = ClientRequestRPC(&c.Leader, request)
		if err != nil {
			return nil, err
		}
		switch reply.Status {
		case OK:
			Debug.Printf("%v is the leader\n", c.Leader)
			Out.Printf("Request returned \"%v\".\n", reply.Response)
			return reply, nil
		case REQ_FAILED:
			Error.Printf("Request failed: %v\n", reply.Response)
			retries++
			return reply, nil
		case NOT_LEADER:
			// The person we've contacted isn't the leader. Use
			// their hint to find the leader
			c.Leader = reply.LeaderHint
		case ELECTION_IN_PROGRESS:
			// An election is in progress. Accept the hint
			// and wait an appropriate amount of time, so the
			// election can finish.
			c.Leader = reply.LeaderHint
			time.Sleep(time.Millisecond * 200)
		}
	}
	return nil, nil
}