package oceanstore

import (
	"net"
	"net/rpc"
	"fmt"
)

type OceanRPCServer struct {
	node     *OceanNode
	listener net.Listener
	rpc      *rpc.Server
}

func newOceanstoreRPCServer(ocean *OceanNode) (server *OceanRPCServer) {
	server = new(OceanRPCServer)
	server.node = ocean
	server.rpc = rpc.NewServer()
	listener, _, err := OpenListener()
	server.rpc.RegisterName(listener.Addr().String(), server)
	server.listener = listener

	if err != nil {
		panic("AA")
	}

	go func() {
		for {
			conn, err := server.listener.Accept()
			if err != nil {
				fmt.Printf("(%v) Raft RPC server accept error: %v\n", err)
				continue
			}
			go server.rpc.ServeConn(conn)
		}
	}()

	return
}

func (server *OceanRPCServer) ConnectImpl(req *ConnectRequest, rep *ConnectReply) error {
	rvreply, err := server.node.connect(req)
	*rep = rvreply
	return err
}
