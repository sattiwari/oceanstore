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

/*
PART 1) This test first adds several nodes, and adds several objects to the nodes.
It then checks for the existance of objects from several nodes.

PART 2)Then it deletes one node and makes sure that the object it had
"spoon" is no longer available

PART 3)Then a new node with the "spoon" object joins and makes sure that it is available through another node.
*/
func TestPublishAndRegister(t *testing.T) {
	if DIGITS != 4 {
		t.Errorf("Test wont work unless DIGITS is set to 4.")
		return
	}
	if TIMEOUT > 3*time.Second && REPUBLISH > 2*time.Second {
		t.Errorf("Test will take too long unless TIMEOUT is set to 3 and REPUBLISH is set to 2.")
		return
	}
	//PART 1
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

	time.Sleep(time.Second * 5)

	// Objects should persist after TIMEOUT seconds because
	// publish is called every two seconds.
	result, err := node1.Get("spoon")
	CheckGet(err, result, "cuchara", t)
	result, err = node2.Get("table")
	CheckGet(err, result, "mesa", t)
	result, err = node3.Get("chair")
	CheckGet(err, result, "silla", t)
	result, err = node0.Get("fork")
	CheckGet(err, result, "tenedor", t)

	// PART 2) Root node of Hash(spoon) should no longer have a record
	// of this object after node0 leaves after TIMEOUT seconds.
	root := FindRootOfHash([]*Tapestry{node1, node2, node3}, Hash("chair"))
	fmt.Printf("The root is: %v and the node0 id is: %v", root, node0.local.node.Id)
	node0.Leave()
	//fmt.Printf("The root is: %v and the node0 id is: %v", root.local.node.Id, node0.local.node.Id)
	if root == nil {
		t.Errorf("Could not find Root of Hash")
	} else {
		replicas := root.local.store.Get("spoon")
		if len(replicas) == 0 && len(replicas) > 1 {
			t.Errorf("Replica of 'spoon' not in root node. What?")
		} else {
			time.Sleep(time.Second * 5)
			replicas = root.local.store.Get("spoon")
			if len(replicas) != 0 {
				t.Errorf("Replica of 'spoon' is in root node after node containing it left.")
			}
		}
	}
	//PART 3) We add a new node that contains spoon and we should find it.
	id = ID{0x5, 2, 0xa, 0xa}
	node4 := makeTapestry(id, node2.local.node.Address, t)
	node4.Store("spoon", []byte("cuchara"))
	time.Sleep(time.Second * 5)
	replicas, _ := node1.local.tapestry.Get("spoon")
	fmt.Printf("id of root is: %v\n", root.local.node.Id)
	if len(replicas) == 0 {
		t.Errorf("'spoon' is not there even after a new node containing it joined")
	}

	node1.Leave()
	node2.Leave()
	node3.Leave()
	node4.Leave()
}

/*Helper function to compare a result with an expected string.*/
func CheckGet(err error, result []byte, expected string, t *testing.T) {
	if err != nil {
		t.Errorf("Get errored out. returned: %v", err)
		return
	}

	if string(result) != expected {
		t.Errorf("Get(\"%v\") did not return expected result '%v'",
			string(result), expected)
	}
}
/*Helper function that returns the root of an ID from a slice of nodes*/
func FindRootOfHash(nodes []*Tapestry, hash ID) *Tapestry {
	if len(nodes) == 0 {
		return nil
	}
	root, _ := nodes[0].local.findRoot(nodes[0].local.node, hash)

	for _, node := range nodes {
		if equal_ids(node.local.node.Id, root.Id) {
			return node
		}
	}

	return nil
}