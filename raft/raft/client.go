package raft

import "time"

const MAX_RETRIES = 5

type Client struct {
	LocalAddr *NodeAddr
	Id uint64
	Leader NodeAddr
	SeqNum uint64
}

func CreateClient(remoteAddr NodeAddr) (cp *Client, err error) {
	cp = new(Client)
	request := RegisterClientRequest{}
	var reply *RegisterClientReply
	retries := 0

	LOOP:
	for retries < MAX_RETRIES  {
		reply, err = RegisterClientRPC(&remoteAddr, request)
		if err != nil {
			return
		}
		switch reply.Status {
		case OK:
			Out.Printf("%v is leader. Client successfully created.\n", remoteAddr)
			break LOOP
		case REQUEST_FAILED:
			Error.Printf("request failed\n")
			retries++
		case NOT_LEADER:
			Out.Printf("%v is not leader, but thinks that %v is.\n", remoteAddr, reply.LeaderHint)
			remoteAddr = reply.LeaderHint
		case ELECTION_IN_PROGRESS:
		//	when election is in progress, accept the hint and wait for appropriate ammount of time so that election can finish
			Out.Printf("Election in progress. %v is not leader but thinks %v is.", remoteAddr, reply.LeaderHint)
			remoteAddr = reply.LeaderHint
			time.Sleep(time.Millisecond * 200) //TODO move this in config
		default:
		}
	}

	cp.Id = reply.ClientId
	cp.Leader = remoteAddr
	return 
}

func (c *Client) SendRequest(command FsmCommand, data []byte) (err error) {
	request := ClientRequest{c.Id, c.SeqNum, command, data}
	c.SeqNum += 1
	var reply *ClientReply
	retries := 0

	LOOP:
		for retries < MAX_RETRIES {
			reply, err = ClientRequestRPC(c.Leader, request)
			if err != nil {
				return
			}
			switch reply.Status {
			case OK:
				Debug.Printf("%v is leader\n", c.Leader)
				Out.Printf("Request returned %v\n", reply.Response)
				break LOOP
			case REQUEST_FAILED:
				Error.Printf("Request failed: %v", reply.Response)
				retries++
				break LOOP
			case NOT_LEADER:
				c.Leader = reply.LeaderHint
			case ELECTION_IN_PROGRESS:
				c.Leader = reply.LeaderHint
				time.Sleep(time.Millisecond * 200)
			}
		}
	return
}

func (c *Client) SendRequestWithReply(command FsmCommand, data []byte) (reply ClientReply, err error) {
	return nil, nil
}