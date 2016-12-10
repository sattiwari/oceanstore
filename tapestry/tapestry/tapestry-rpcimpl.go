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

type GetBackpointersRequest struct {
	To    Node
	From  Node
	Level int
}

type TransferRequest struct {
	To   Node
	From Node
	Data map[string][]Node
}

type NodeRequest struct {
	To   Node
	Node Node
}

type AddNodeMulticastRequest struct {
	To      Node
	NewNode Node
	Level   int
}

type NotifyLeaveRequest struct {
	To          Node
	From        Node
	Replacement *Node
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

func (server *TapestryRPCServer) Hello(req Node, rsp *Node) (err error) {
	*rsp = server.tapestry.local.node
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

func (server *TapestryRPCServer) AddBackpointer(req NodeRequest, rsp *Node) error {
	err := server.validate(req.To)
	if err != nil {
		return err
	}
	return server.tapestry.local.AddBackpointer(req.Node)
}

func (server *TapestryRPCServer) RemoveBackpointer(req NodeRequest, rsp *Node) error {
	err := server.validate(req.To)
	if err != nil {
		return err
	}
	return server.tapestry.local.RemoveBackpointer(req.Node)
}

func (server *TapestryRPCServer) GetBackpointers(req GetBackpointersRequest, rsp *[]Node) (err error) {
	err = server.validate(req.To)
	if err != nil {
		return err
	}
	backpointers, err := server.tapestry.local.GetBackpointers(req.From, req.Level)
	*rsp = append(*rsp, backpointers...)
	return
}

func (server *TapestryRPCServer) AddNode(req NodeRequest, rsp *[]Node) (err error) {
	err = server.validate(req.To)
	if err != nil {
		return
	}
	neighbours, err := server.tapestry.local.AddNode(req.Node)
	*rsp = append(*rsp, neighbours...)
	return
}

func (server *TapestryRPCServer) AddNodeMulticast(req AddNodeMulticastRequest, rsp *[]Node) (err error) {
	err = server.validate(req.To)
	if err != nil {
		return err
	}
	neighbours, err := server.tapestry.local.AddNodeMulticast(req.NewNode, req.Level)
	*rsp = append(*rsp, neighbours...)
	return err
}

func (server *TapestryRPCServer) Transfer(req TransferRequest, rsp *Node) error {
	err := server.validate(req.To)
	if err != nil {
		return err
	}
	return server.tapestry.local.Transfer(req.From, req.Data)
}

func (server *TapestryRPCServer) NotifyLeave(req NotifyLeaveRequest, rsp *Node) error {
	err := server.validate(req.To)
	if err != nil {
		return err
	}
	return server.tapestry.local.NotifyLeave(req.From, req.Replacement)
}

/*
   This method is invoked over RPC by other Tapestry nodes.
   Register the specified node as an advertiser of the specified key.

   *    Check that we are the root node for the key
   *    Add the node to the object store
   *    Kick off a timer to remove the node if it's not advertised again after a set amount of time
*/
func (server *TapestryRPCServer) Register(req RegisterRequest, rsp *RegisterResponse) (err error) {
	err = server.validate(req.To)
	if err == nil {
		rsp.IsRoot, err = server.tapestry.local.Register(req.Key, req.From)
	}
	return
}