package tapestry

import (
	"net/rpc"
	"fmt"
)

/*
	The methods defined in this file parallel the methods defined in tapestry-local.
	These methods take an additional argument, the node on which the method should be invoked.
	Calling any of these methods will invoke the corresponding method on the specified remote node.
*/

// Remote API: ping an address to get tapestry node info
func (tapestry *Tapestry) hello(address string) (rsp Node, err error) {
	err = makeRemoteCall(address, "TapestryRPCServer", "Hello", tapestry.local.node, &rsp)
	return
}

// Helper function to makes a remote call
func makeRemoteNodeCall(remote Node, method string, req interface{}, rsp interface{}) error {
	fmt.Printf("%v(%v)\n", method, req)
	return makeRemoteCall(remote.Address, "TapestryRPCServer", method, req, rsp)
}

// Helper function to makes a remote call
func makeRemoteCall(address string, structtype string, method string, req interface{}, rsp interface{}) error {
	// Dial the server
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return err
	}

	// Make the request
	fqm := fmt.Sprintf("%v.%v", structtype, method)
	err = client.Call(fqm, req, rsp)

	client.Close()
	if err != nil {
		return err
	}

	return nil
}

// Remote API: makes a remote call to the Register function
func (tapestry *Tapestry) register(remote Node, replica Node, key string) (bool, error) {
	var rsp RegisterResponse
	err := makeRemoteNodeCall(remote, "Register", RegisterRequest{remote, replica, key}, &rsp)
	return rsp.IsRoot, err
}

// Remote API: makes a remote call to the GetNextHop function
func (tapestry *Tapestry) getNextHop(remote Node, id ID) (bool, Node, error) {
	var rsp NextHopResponse
	err := makeRemoteNodeCall(remote, "GetNextHop", NextHopRequest{remote, id}, &rsp)
	return rsp.HasNext, rsp.Next, err
}

// Remote API: makes a remote call to the RemoveBadNodes function
func (tapestry *Tapestry) removeBadNodes(remote Node, toremove []Node) error {
	return makeRemoteNodeCall(remote, "RemoveBadNodes", RemoveBadNodesRequest{remote, toremove}, &Node{})
}

// Remote API: makes a remote call to the Fetch function
func (tapestry *Tapestry) fetch(remote Node, key string) (bool, []Node, error) {
	var rsp FetchResponse
	err := makeRemoteNodeCall(remote, "Fetch", FetchRequest{remote, key}, &rsp)
	return rsp.IsRoot, rsp.Values, err
}

// Remote API: makes a remote call to the AddBackpointer function
func (tapestry *Tapestry) addBackpointer(remote Node, toAdd Node) error {
	return makeRemoteNodeCall(remote, "AddBackpointer", NodeRequest{remote, toAdd}, &Node{})
}

// Remote API: makes a remote call to the RemoveBackpointer function
func (tapestry *Tapestry) removeBackpointer(remote Node, toRemove Node) error {
	return makeRemoteNodeCall(remote, "RemoveBackpointer", NodeRequest{remote, toRemove}, &Node{})
}

// Remote API: makes a remote call to the GetBackpointers function
func (tapestry *Tapestry) getBackpointers(remote Node, from Node, level int) (neighbours []Node, err error) {
	err = makeRemoteNodeCall(remote, "GetBackpointers", GetBackpointersRequest{remote, from, level}, &neighbours)
	return
}

// Remote API: makes a remote call to the AddNode function
func (tapestry *Tapestry) addNode(remote Node, newnode Node) (neighbours []Node, err error) {
	err = makeRemoteNodeCall(remote, "AddNode", NodeRequest{remote, newnode}, &neighbours)
	return
}

// Remote API: makes a remote call to the AddNodeMulticast function
func (tapestry *Tapestry) addNodeMulticast(remote Node, newnode Node, level int) (neighbours []Node, err error) {
	err = makeRemoteNodeCall(remote, "AddNodeMulticast", AddNodeMulticastRequest{remote, newnode, level}, &neighbours)
	return
}

func (tapestry *Tapestry) transfer(remote Node, from Node, data map[string][]Node) error {
	return makeRemoteNodeCall(remote, "Transfer", TransferRequest{remote, from, data}, &Node{})
}

// Remote API: makes a remote call to the NotifyLeave function
func (tapestry *Tapestry) notifyLeave(remote Node, from Node, replacement *Node) (err error) {
	return makeRemoteNodeCall(remote, "NotifyLeave", NotifyLeaveRequest{remote, from, replacement}, &Node{})
}
