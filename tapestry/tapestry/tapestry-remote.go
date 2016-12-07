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
