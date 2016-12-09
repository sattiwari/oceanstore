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

type RegisterRequest struct {
	To   Node
	From Node
	Key  string
}

type RegisterResponse struct {
	IsRoot bool
}

type NextHopRequest struct {
	To Node
	Id ID
}
type NextHopResponse struct {
	HasNext bool
	Next    Node
}

type RemoveBadNodesRequest struct {
	To       Node
	BadNodes []Node
}

type FetchRequest struct {
	To  Node
	Key string
}
type FetchResponse struct {
	To     Node
	IsRoot bool
	Values []Node
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

func (server *TapestryRPCServer) validate(expect Node) error {
	if server.tapestry.local.node != expect {
		return fmt.Errorf("Remote node expected us to be %v, but we are %v", expect, server.tapestry.local.node)
	}
	return nil
}

func (server *TapestryRPCServer) GetNextHop(req NextHopRequest, rsp *NextHopResponse) (err error) {
	err = server.validate(req.To)
	if err == nil {
		rsp.HasNext, rsp.Next, err = server.tapestry.local.GetNextHop(req.Id)
	}
	return
}

// Server: proxies a remote method invocation to the local node
func (server *TapestryRPCServer) RemoveBadNodes(req RemoveBadNodesRequest, rsp *Node) error {
	err := server.validate(req.To)
	if err != nil {
		return err
	}
	return server.tapestry.local.RemoveBadNodes(req.BadNodes)
}

func (server *TapestryRPCServer) Fetch(req FetchRequest, rsp *FetchResponse) (err error) {
	err = server.validate(req.To)
	if err == nil {
		rsp.IsRoot, rsp.Values, err = server.tapestry.local.Fetch(req.Key)
	}
	return
}