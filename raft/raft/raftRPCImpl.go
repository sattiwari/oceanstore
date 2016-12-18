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