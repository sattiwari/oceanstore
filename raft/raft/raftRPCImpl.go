package raft

import (
	"fmt"
	"net/rpc"
)

type RaftRPCServer struct {
	node *RaftNode
}

func (server *RaftRPCServer) startRpcServer() {
	for {
		if server.node.IsShutDown {
			fmt.Printf("(%v) Shutting down RPC server\n", server.node.Id)
			return
		}
		conn, err := server.node.Listener.Accept()
		if err != nil {
			if !server.node.IsShutDown {
				fmt.Printf("(%v) Raft RPC server accept error: %v\n", server.node.Id, err)
			}
			continue
		}
		if !server.node.IsShutDown {
			go rpc.ServeConn(conn)
		} else {
			conn.Close()
		}
	}
}

func (server *RaftRPCServer) JoinImpl(req *JoinRequest, reply *JoinReply) error {
	err := server.node.Join(req)
	reply.Success = err == nil
	return err
}

func (server *RaftRPCServer) StartNodeImpl(req *StartNodeRequest, reply *StartNodeReply) error {
	err := server.node.StartNode(req)
	reply.Success = err == nil
	return err
}

func (server *RaftRPCServer) RequestVoteImpl(req *RequestVoteRequest, reply *RequestVoteReply) error {
	if server.node.Testing.IsDenied(req.CandidateId, *server.node.GetLocalAddr()) {
		return ErrorTestingPolicyDenied
	}
	rvreply, err := server.node.RequestVote(req)
	*reply = rvreply
	return err
}

func (server *RaftRPCServer) ClientRequestImpl(req *ClientRequest, reply *ClientReply) error {
	rvreply, err := server.node.ClientRequest(req)
	*reply = rvreply
	return err
}

func (server *RaftRPCServer) RegisterClientImpl(req *RegisterClientRequest, reply *RegisterClientReply) error {
	rvreply, err := server.node.RegisterClient(req)
	*reply = rvreply
	return err
}

func (server *RaftRPCServer) AppendEntriesImpl(req *AppendEntriesRequest, reply *AppendEntriesReply) error {
	if server.node.Testing.IsDenied(req.LeaderId, *server.node.GetLocalAddr()) {
		return ErrorTestingPolicyDenied
	}
	aereply, err := server.node.AppendEntries(req)
	*reply = aereply
	return err
}