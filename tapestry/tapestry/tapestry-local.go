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