package tapestry

import (
	"fmt"
	"testing"
	"time"
)

/*
PART 1) This test inserts several nodes and objects and
PART 2) then inserts others to make sure that the root changes to the new nodes.

PART 3) At the end it also makes sure that there is no replica in the previous node after the timeout.
*/
func TestChangeRoot(t *testing.T) {
	if DIGITS != 4 {
		t.Errorf("Test wont work unless DIGITS is set to 4.")
		return
	}
	if TIMEOUT > 3*time.Second && REPUBLISH > 2*time.Second {
		t.Errorf("Test will take too long unless TIMEOUT is set to 3 and REPUBLISH is set to 2.")
		return
	}
	//PART 1)
	port = 58000
	id := ID{5, 8, 3, 15}
	node0 := makeTapestry(id, "", t)
	id = ID{7, 0, 0xd, 1}
	node1 := makeTapestry(id, node0.local.node.Address, t)
	id = ID{9, 0, 0xf, 5}
	node2 := makeTapestry(id, node0.local.node.Address, t)
	id = ID{0xb, 0, 0xf, 0xa}
	node3 := makeTapestry(id, node0.local.node.Address, t)

	node0.Store("spoon", []byte("cuchara"))
	node1.Store("table", []byte("mesa"))
	node2.Store("chair", []byte("silla"))
	node3.Store("fork", []byte("tenedor"))

	//The root for the node is
	root, _ := node3.local.findRoot(node3.local.node, Hash("fork"))
	if !equal_ids(root.Id, node0.local.node.Id) {
		t.Errorf("The root for the fork is not node0, its %v\n", root.Id)
	}
	//PART 2) Now we insert a new node
	id = ID{0x5, 2, 0xa, 0xa}
	node4 := makeTapestry(id, node2.local.node.Address, t)
	node4.Store("napkin", []byte("servilleta"))

	//We wait the timeout
	time.Sleep(TIMEOUT + 1)

	//The root for fork should have changed to node4
	fmt.Printf("hash for fork: %v\n", Hash("fork"))
	fmt.Printf("hash for spoon: %v\n", Hash("spoon"))
	fmt.Printf("hash for table: %v\n", Hash("table"))
	fmt.Printf("hash for chair: %v\n", Hash("chair"))
	root2, _ := node2.local.findRoot(node2.local.node, Hash("fork"))
	if !equal_ids(root2.Id, node4.local.node.Id) {
		t.Errorf("The root for the fork is not node4, its %v\n", root2.Id)
	}
	//PART 3) We now make sure that the replica is no longer in the previous node
	replica := node0.local.store.Get("fork")
	if len(replica) != 0 {
		t.Errorf("This node still has a replica for another node %v", replica)
	}

	node1.Leave()
	node2.Leave()
	node3.Leave()
	node4.Leave()
	node0.Leave()
}

/*
This test is the same as the previous but it does not have a timeout. It tests the transfer of keys during the
AddNodeMulticast where keys are transfered to new joining node.*/
func TestTransferKeys(t *testing.T) {
	if DIGITS != 4 {
		t.Errorf("Test wont work unless DIGITS is set to 4.")
		return
	}
	if TIMEOUT > 3*time.Second && REPUBLISH > 2*time.Second {
		t.Errorf("Test will take too long unless TIMEOUT is set to 3 and REPUBLISH is set to 2.")
		return
	}
	port = 58000
	id := ID{5, 8, 3, 15}
	node0 := makeTapestry(id, "", t)
	id = ID{7, 0, 0xd, 1}
	node1 := makeTapestry(id, node0.local.node.Address, t)
	id = ID{9, 0, 0xf, 5}
	node2 := makeTapestry(id, node0.local.node.Address, t)
	id = ID{0xb, 0, 0xf, 0xa}
	node3 := makeTapestry(id, node0.local.node.Address, t)

	node0.Store("spoon", []byte("cuchara"))
	node1.Store("table", []byte("mesa"))
	node2.Store("chair", []byte("silla"))
	node3.Store("fork", []byte("tenedor"))

	//The root for the node is
	root, _ := node3.local.findRoot(node3.local.node, Hash("fork"))
	if !equal_ids(root.Id, node0.local.node.Id) {
		t.Errorf("The root for the fork is not node0, its %v\n", root.Id)
	}
	//Now we insert a new node
	id = ID{0x5, 2, 0xa, 0xa}
	node4 := makeTapestry(id, node2.local.node.Address, t)
	node4.Store("napkin", []byte("servilleta"))

	// //The root for spoon should have changed to node4
	fmt.Printf("hash for fork: %v\n", Hash("fork"))
	fmt.Printf("hash for spoon: %v\n", Hash("spoon"))
	fmt.Printf("hash for table: %v\n", Hash("table"))
	fmt.Printf("hash for chair: %v\n", Hash("chair"))
	root2, _ := node2.local.findRoot(node2.local.node, Hash("fork"))
	if !equal_ids(root2.Id, node4.local.node.Id) {
		t.Errorf("The root for the fork is not node4, its %v\n", root2.Id)
	}
	//We now make sure that the replica is no longer in the previous node
	replica := node0.local.store.Get("fork")
	if len(replica) != 0 {
		t.Errorf("This node still has a replica for another node %v", replica)
	}

	node1.Leave()
	node2.Leave()
	node3.Leave()
	node4.Leave()
	node0.Leave()
}
