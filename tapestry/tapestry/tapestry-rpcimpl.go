package tapestry

import (
	"net"
	"net/rpc"
	"fmt"
)

/*
	Receives remote invocations of methods for the local tapestry node
*/
type TapestryRPCServer struct {
	tapestry *Tapestry
	listener net.Listener
	rpc      *rpc.Server
}

/*
	Creates the tapestry RPC server of a tapestry node.  The RPC server receives function invocations,
	and proxies them to the tapestrynode implementations
*/
func newTapestryRPCServer(port int, tapestry *Tapestry) (server *TapestryRPCServer, err error) {
	// Create the RPC server
	server = new(TapestryRPCServer)
	server.tapestry = tapestry
	server.rpc = rpc.NewServer()
	server.rpc.Register(server)
	server.rpc.Register(NewBlobStoreRPC(tapestry.blobstore))
	server.listener, err = net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return nil, fmt.Errorf("Tapestry RPC server unable to listen on tcp port %v, reason: %v", port, err)
	}

	// Start the RPC server
	go func() {
		for {
			cxn, err := server.listener.Accept()
			if err != nil {
				fmt.Printf("Server %v closing: %s\n", port, err)
				return
			}
			go server.rpc.ServeConn(cxn)
		}
	}()

	return
}