package raft

import "net/rpc"

type RaftRPCServer struct {
	node *RaftNode
}

func (server *RaftRPCServer) startRPCServer() {
	for {
		if server.node.IsShutDown {
			Out.Printf("%v shutting down RPC server \n", server.node.Id)
			return
		}
		conn, err := server.node.Listener.Accept()
		if err != nil {
			if !server.node.IsShutDown {
				Out.Printf("%v Raft RPC accept error %v\n", server.node.Id, err)
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

func (server *RaftRPCServer) StartNodeImpl(request *StartNodeRequest, reply *StartNodeReply) error {
	err := server.node.StartNode(request)
	reply.Success = err == nil
	return err
}

func (server *RaftRPCServer) JoinImpl(request *JoinRequest, reply *JoinReply) error {
	err := server.node.Join(request)
	reply.Success = err == nil
	return err
}

//func (server *RaftRPCServer) RequestVoteImpl(request *RequestVoteRequest, reply *RequestVoteReply) error {
//	rvReply, err := server.node.RequestVote(request)
//	reply.VoteGranted = err == nil
//	return err
//}
//
//func (server *RaftRPCServer) AppendEntriesImpl(request *AppendEntriesRequest, reply *AppendEntriesReply) error {
//	err := server.node.AppendEntries(request)
//	reply.Success = err == nil
//	return err
//}
//
//func (server *RaftRPCServer) RegisterClientImpl(request *RegisterClientRequest, reply *RegisterClientReply) error {
//	err := server.node.RegisterClient(request)
//	reply.Success = err == nil
//	return err
//}
//
//func (server *RaftRPCServer) ClientRequestImpl(request *ClientRequest, reply *ClientReply) error {
//	err := server.node.StartNode(request)
//	reply.Success = err == nil
//	return err
//}