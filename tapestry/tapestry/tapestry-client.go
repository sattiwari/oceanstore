package tapestry

import "fmt"

// Invoke tapestry.Lookup on a remote tapestry node
func TapestryLookup(remote Node, key string) (nodes []Node, err error) {
	fmt.Printf("Making remote TapestryLookup call\n")
	var rsp LookupResponse
	err = makeRemoteNodeCall(remote, "TapestryLookup", LookupRequest{remote, key}, &rsp)
	nodes = rsp.Nodes
	return
}

// Get data from a tapestry node.  Looks up key then fetches directly
func TapestryGet(remote Node, key string) ([]byte, error) {
	fmt.Printf("Making remote TapestryGet call\n")
	// Lookup the key
	replicas, err := TapestryLookup(remote, key)
	if err != nil {
		return nil, err
	}
	if len(replicas) == 0 {
		return nil, fmt.Errorf("No replicas returned for key %v", key)
	}

	// Contact replicas
	var errs []error
	for _, replica := range replicas {
		blob, err := FetchRemoteBlob(replica, key)
		if err != nil {
			errs = append(errs, err)
		}
		if blob != nil {
			return *blob, nil
		}
	}

	return nil, fmt.Errorf("Error contacting replicas, %v: %v", replicas, errs)
}

func TapestryRemove(remote Node, key string) (success bool, err error) {
	fmt.Printf("Making remote TapestryRemove call\n")
	var rsp RemoveResponse
	err = makeRemoteNodeCall(remote, "TapestryRemove", RemoveRequest{remote, key}, &rsp)
	success = rsp.Removed
	return
}

type StoreRequest struct {
	To    Node
	Key   string
	Value []byte
}

type StoreResponse struct {
}

type LookupRequest struct {
	To  Node
	Key string
}

type LookupResponse struct {
	Nodes []Node
}

type RemoveRequest struct {
	To  Node
	Key string
}

type RemoveResponse struct {
	Removed bool
}

// Server: extension method to open up Store via RPC
func (server *TapestryRPCServer) TapestryStore(req StoreRequest, rsp *StoreResponse) (err error) {
	fmt.Printf("Received remote invocation of Tapestry.Store\n")
	return server.tapestry.Store(req.Key, req.Value)
}

// Server: extension method to open up Lookup via RPC
func (server *TapestryRPCServer) TapestryLookup(req LookupRequest, rsp *LookupResponse) (err error) {
	fmt.Printf("Received remote invocation of Tapestry.Lookup\n")
	rsp.Nodes, err = server.tapestry.Lookup(req.Key)
	return
}

// Server: extension method to open up Remove via RPC
func (server *TapestryRPCServer) TapestryRemove(req RemoveRequest, rsp *RemoveResponse) (err error) {
	fmt.Printf("Received remote invocation of Tapestry.Remove\n")
	rsp.Removed = server.tapestry.Remove(req.Key)
	return
}