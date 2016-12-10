package tapestry

import (
	"fmt"
	"time"
)

/*
   Implementation of the local tapestry node.  There are three kinds of methods defined in this file
       1.  Methods that can be invoked remotely via RPC by other Tapestry nodes (eg AddBackpointer, RemoveBackpointer)
       2.  Methods that are invoked by clients (eg Publish, Lookup)
       3.  Common utility methods that you can use in your implementation (eg findRoot)
   For RPC methods, to invoke the equivalent method on a remote node, see the methods defined in tapestry-remote.go
*/

/*
   The ID and location information for a node in the tapestry
*/
type Node struct {
	Id      ID
	Address string
}

/*
   The main struct for the local tapestry node.
   Methods can be invoked locally on this struct.
   Methods on remote TapestryNode instances can be called on the 'tapestry' member
*/
type TapestryNode struct {
	node         Node          // the ID and address of this node
	table        *RoutingTable // the routing table
	backpointers *Backpointers // backpointers to keep track of other nodes that point to us
	store        *ObjectStore  // stores keys for which this node is the root
	tapestry     *Tapestry     // api to the tapestry mesh
}

/*
   Called in tapestry initialization to create a tapestry node struct
*/
func newTapestryNode(node Node, tapestry *Tapestry) *TapestryNode {
	n := new(TapestryNode)
	n.node = node
	n.table = NewRoutingTable(node)
	n.backpointers = NewBackpointers(node)
	n.store = NewObjectStore()
	n.tapestry = tapestry
	return n
}

/*
   This method is invoked over RPC by other Tapestry nodes.
   *    Returns the best candidate from our routing table for routing to the provided ID
*/
func (local *TapestryNode) GetNextHop(id ID) (morehops bool, nexthop Node, err error) {

	// Call routingtable.go method
	nexthop = local.table.GetNextHop(id)

	// If all digits match (aka is equal), no more hops are needed.
	sharedDigits := SharedPrefixLength(local.node.Id, nexthop.Id)
	if DIGITS == sharedDigits {
		morehops = false
		return
	}

	morehops = true
	return
}

/*
   This method is invoked over RPC by other Tapestry nodes.
   The provided nodes are bad and we should discard them
   *    Remove each node from our routing table
   *    Remove each node from our set of backpointers
*/
func (local *TapestryNode) RemoveBadNodes(badnodes []Node) (err error) {
	for _, badnode := range badnodes {
		if local.table.Remove(badnode) {
			fmt.Printf("Removed bad node %v\n", badnode)
		}
		if local.backpointers.Remove(badnode) {
			fmt.Printf("Removed bad node backpointer %v\n", badnode)
		}
	}
	return
}

/*
   Utility function for iteratively contacting nodes to get the root node for the provided ID
   *    Starting from the specified node, iteratively contact nodes calling getNextHop until we reach the root node
   *    Also keep track of any bad nodes that errored during lookup
   *    At each step, notify the next-hop node of all of the bad nodes we have encountered along the way
*/
func (local *TapestryNode) findRoot(start Node, id ID) (Node, error) {
	fmt.Printf("Routing to %v\n", id)

	next := start
	var current, prev Node
	var badNodes []Node
	var err error
	var hasNext bool

	for {
		prev = current
		current = next
		hasNext, next, err = local.tapestry.getNextHop(current, id)

		if err != nil {
			badNodes = append(badNodes, current)
			current = prev
		}

		err = local.tapestry.removeBadNodes(current, badNodes)

		// Current node does not exist anymore. Cannot recover.
		if err != nil {
			return current, err
		}

		if !hasNext {
			break
		}
	}

	return current, nil
}

/*
   Client API.  Publishes the key to tapestry.
   *    Route to the root node for the key
   *    Register our local node on the root
   *    Start periodically republishing the key
   *    Return a channel for cancelling the publish
*/
func (local *TapestryNode) Publish(key string) (done chan bool, err error) {
	done = make(chan bool)

	root, err := local.findRoot(local.node, Hash(key))

	// If is not root
	for i := 0; i < RETRIES; i++ {
		isRoot, err := local.tapestry.register(root, local.node, key)

		if isRoot && err == nil {
			break
		}
	}

	//Periodically checking
	go func() {
		for {
			select {
			case <-done:
				return
			case <-time.After(REPUBLISH):
				root, _ := local.findRoot(local.node, Hash(key))
				local.tapestry.register(root, local.node, key)
				for i := 0; i < RETRIES; i++ {
					isRoot, err := local.tapestry.register(root, local.node, key)

					if isRoot && err == nil {
						break
					}
				}
			}
		}
	}()

	return done, err
}

/*
   This method is invoked over RPC by other Tapestry nodes.
   *    Check that we are the root node for the requested key
   *    Return all nodes that are registered in the local object store for this key
*/
func (local *TapestryNode) Fetch(key string) (isRoot bool, replicas []Node, err error) {
	root, _ := local.findRoot(local.node, Hash(key))
	if !equal_ids(root.Id, local.node.Id) {
		isRoot = false
		return
	}
	isRoot = true

	replicas = local.store.Get(key)
	return
}

/*
   Client API.  Look up the Tapestry nodes that are storing the blob for the specified key.
   *    Find the root node for the key
   *    Fetch the replicas from the root's object store
   *    Attempt up to RETRIES many times
*/
func (local *TapestryNode) Lookup(key string) (nodes []Node, err error) {
	for i := 0; i < RETRIES; i++ {
		root, err := local.findRoot(local.node, Hash(key))
		if err != nil {
			continue
		}

		wasRoot, replicas, err := local.tapestry.fetch(root, key)

		if wasRoot && err == nil {
			nodes = replicas
			break
		}
	}

	return
}

/*
   This method is invoked over RPC by other Tapestry nodes.
   *    Add the from node to our backpointers
   *    Possibly add the node to our routing table, if appropriate
*/
func (local *TapestryNode) AddBackpointer(from Node) (err error) {
	if local.backpointers.Add(from) {
		fmt.Printf("Added backpointer %v\n", from)
	}
	local.addRoute(from)
	return
}

/*
   This method is invoked over RPC by other Tapestry nodes.
   *    Remove the from node from our backpointers
*/
func (local *TapestryNode) RemoveBackpointer(from Node) (err error) {
	if local.backpointers.Remove(from) {
		fmt.Printf("Removed backpointer %v\n", from)
	}
	return
}

/*
   This method is invoked over RPC by other Tapestry nodes.
   *    Get all backpointers at the level specified
   *    Possibly add the node to our routing table, if appropriate
*/
func (local *TapestryNode) GetBackpointers(from Node, level int) (backpointers []Node, err error) {
	fmt.Printf("Sending level %v backpointers to %v\n", level, from)
	backpointers = local.backpointers.Get(level)
	local.addRoute(from)
	return
}

/*
   Utility function that adds a node to our routing table
   *    Adds the provided node to the routing table, if appropriate.
   *    If the node was added to the routing table, notify the node of a backpointer
   *    If an old node was removed from the routing table, notify the old node of a removed backpointer
*/

func (local *TapestryNode) addRoute(node Node) (err error) {
	added, prev := local.table.Add(node)

	// In any case, error was loss of communication with node.
	if added {
		err = local.tapestry.addBackpointer(node, local.node)
	}

	if prev != nil {
		err = local.tapestry.removeBackpointer(node, local.node)
	}

	return
}

/*
   This method is invoked over RPC by other Tapestry nodes.
   We are the root node for some new node joining the tapestry.
   *    Begin the acknowledged multicast
   *    Return the neighbourset from the multicast
*/
func (local *TapestryNode) AddNode(node Node) (neighbourset []Node, err error) {
	return local.AddNodeMulticast(node, SharedPrefixLength(node.Id, local.node.Id))
}

/*
   This method is invoked over RPC by other Tapestry nodes
   *    Register all of the provided objects in the local object store
   *    If appropriate, add the from node to our local routing table
*/
func (local *TapestryNode) Transfer(from Node, replicamap map[string][]Node) error {
	local.store.RegisterAll(replicamap, TIMEOUT)

	err := local.addRoute(from)
	return err
}

/*
   This method is invoked over RPC by other Tapestry nodes.
   A new node is joining the tapestry, and we are a need-to-know node participating in the multicast.
   *    Propagate the multicast to the specified row in our routing table
   *    Await multicast response and return the neighbourset
   *    Begin transfer of appropriate replica info to the new node
*/
func (local *TapestryNode) AddNodeMulticast(newnode Node, level int) (neighbours []Node, err error) {
	fmt.Printf("Add node multicast %v at level %v\n", newnode, level)
	neighbours = local.table.GetLevel(level)
	results := make([]Node, 0)
	for _, target := range neighbours {
		result, err := local.tapestry.addNodeMulticast(
			target, newnode, level+1)

		// bad node, continue
		if err != nil {
			continue
		}

		results = append(results, result...)
	}

	transferReg := local.store.GetTransferRegistrations(local.node, newnode)
	err = local.tapestry.transfer(newnode, local.node, transferReg)

	// Could not add stuff to new node. We need to reregister what we removed.
	if err != nil {
		local.store.RegisterAll(transferReg, TIMEOUT)
	}

	neighbours = append(neighbours, local.node)
	neighbours = append(neighbours, results...)
	return
}

/*
   Invoked when starting the local node, if we are connecting to an existing Tapestry.
   *    Find the root for our node's ID
   *    Call AddNode on our root to initiate the multicast and receive our initial neighbour set
   *    Iteratively get backpointers from the neighbour set and populate routing table
*/
func (local *TapestryNode) Join(otherNode Node) error {
	fmt.Printf("Joining\n", otherNode)

	// Route to our root
	root, err := local.findRoot(otherNode, local.node.Id)
	if err != nil {
		return fmt.Errorf("Error joining existing tapestry node %v, reason: %v", otherNode, err)
	}

	// Add ourselves to our root by invoking AddNode on the remote node
	neighbours, err := local.tapestry.addNode(root, local.node)
	if err != nil {
		return fmt.Errorf("Error adding ourselves to root node %v, reason: %v", root, err)
	}

	/*
		"The nodes returned by AddNodeMulticast() will go into the
		Joining node's routing table, but all these nodes are of length n
		and greater. This means that rows 0 through n-1 of the node's
		routing table still need to be filled via backpointer traversal."
	*/

	level := SharedPrefixLength(local.node.Id, root.Id)
	for {

		for _, n := range neighbours {
			local.addRoute(n)
		}

		if level > 0 {
			nextNeighbours := make([]Node, 0)
			for _, neighbour := range neighbours {
				result, err := local.tapestry.getBackpointers(neighbour, local.node, level-1)

				if err != nil {
					continue
				}

				nextNeighbours = append(nextNeighbours, result...)

			}
			if len(nextNeighbours) == 0 {
				nextNeighbours = append(nextNeighbours, neighbours...)
			}

			neighbours = nextNeighbours
		} else {
			break
		}

		level--
	}

	return err
}
