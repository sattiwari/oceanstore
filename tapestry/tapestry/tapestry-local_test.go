package tapestry

import (
	"testing"
)
/*Helper function to make sure that a target node is the same
as the expected node*/
func CheckFindRoot(node *TapestryNode, target ID, expected ID,
t *testing.T) {
	result, _ := node.findRoot(node.node, target)
	if !equal_ids(result.Id, expected) {
		t.Errorf("%v: findRoot of %v is not %v (gives %v)", node.node.Id,
			target, expected, result.Id)
	}
}

/* This test checks that find root works from all nodes */
func TestFindRootAndLeave(t *testing.T) {
	if DIGITS != 4 {
		t.Errorf("Test wont work unless DIGITS is set to 4.")
	}

	port = 58000
	id := ID{5, 8, 3, 15}
	mainNode := makeTapestryNode(id, "", t)
	id = ID{7, 0, 0xd, 1}
	node1 := makeTapestryNode(id, mainNode.node.Address, t)
	id = ID{7, 0, 0xf, 5}
	node2 := makeTapestryNode(id, mainNode.node.Address, t)
	id = ID{7, 0, 0xf, 0xa}
	node3 := makeTapestryNode(id, mainNode.node.Address, t)

	// Checks all possible combinations between all nodes to find
	// a given route.
	id = ID{3, 0xf, 8, 0xa}
	CheckFindRoot(mainNode, id, mainNode.node.Id, t)
	CheckFindRoot(node1, id, mainNode.node.Id, t)
	CheckFindRoot(node2, id, mainNode.node.Id, t)
	CheckFindRoot(node3, id, mainNode.node.Id, t)
	id = ID{5, 2, 0, 0xc}
	CheckFindRoot(mainNode, id, mainNode.node.Id, t)
	CheckFindRoot(node1, id, mainNode.node.Id, t)
	CheckFindRoot(node2, id, mainNode.node.Id, t)
	CheckFindRoot(node3, id, mainNode.node.Id, t)
	id = ID{5, 8, 0xf, 0xf}
	CheckFindRoot(mainNode, id, mainNode.node.Id, t)
	CheckFindRoot(node1, id, mainNode.node.Id, t)
	CheckFindRoot(node2, id, mainNode.node.Id, t)
	CheckFindRoot(node3, id, mainNode.node.Id, t)
	id = ID{7, 0, 0xc, 3}
	CheckFindRoot(mainNode, id, node1.node.Id, t)
	CheckFindRoot(node1, id, node1.node.Id, t)
	CheckFindRoot(node2, id, node1.node.Id, t)
	CheckFindRoot(node3, id, node1.node.Id, t)
	id = ID{6, 0, 0xf, 4}
	CheckFindRoot(mainNode, id, node2.node.Id, t)
	CheckFindRoot(node1, id, node2.node.Id, t)
	CheckFindRoot(node2, id, node2.node.Id, t)
	CheckFindRoot(node3, id, node2.node.Id, t)
	id = ID{7, 0, 0xa, 2}
	CheckFindRoot(mainNode, id, node1.node.Id, t)
	CheckFindRoot(node1, id, node1.node.Id, t)
	CheckFindRoot(node2, id, node1.node.Id, t)
	CheckFindRoot(node3, id, node1.node.Id, t)
	id = ID{6, 3, 9, 5}
	CheckFindRoot(mainNode, id, node1.node.Id, t)
	CheckFindRoot(node1, id, node1.node.Id, t)
	CheckFindRoot(node2, id, node1.node.Id, t)
	CheckFindRoot(node3, id, node1.node.Id, t)
	id = ID{6, 8, 3, 0xf}
	CheckFindRoot(mainNode, id, node1.node.Id, t)
	CheckFindRoot(node1, id, node1.node.Id, t)
	CheckFindRoot(node2, id, node1.node.Id, t)
	CheckFindRoot(node3, id, node1.node.Id, t)
	id = ID{6, 3, 0xe, 5}
	CheckFindRoot(mainNode, id, node2.node.Id, t)
	CheckFindRoot(node1, id, node2.node.Id, t)
	CheckFindRoot(node2, id, node2.node.Id, t)
	CheckFindRoot(node3, id, node2.node.Id, t)
	id = ID{6, 3, 0xe, 9}
	CheckFindRoot(mainNode, id, node3.node.Id, t)
	CheckFindRoot(node1, id, node3.node.Id, t)
	CheckFindRoot(node2, id, node3.node.Id, t)
	CheckFindRoot(node3, id, node3.node.Id, t)
	id = ID{0xb, 0xe, 0xe, 0xf}
	CheckFindRoot(mainNode, id, mainNode.node.Id, t)
	CheckFindRoot(node1, id, mainNode.node.Id, t)
	CheckFindRoot(node2, id, mainNode.node.Id, t)
	CheckFindRoot(node3, id, mainNode.node.Id, t)

	// Check if after node leaves, tables get updated.
	mainNode.tapestry.Leave()

	id = ID{3, 0xf, 8, 0xa}
	CheckFindRoot(node1, id, node1.node.Id, t)
	CheckFindRoot(node2, id, node1.node.Id, t)
	CheckFindRoot(node3, id, node1.node.Id, t)
	id = ID{5, 2, 0, 0xc}
	CheckFindRoot(node1, id, node1.node.Id, t)
	CheckFindRoot(node2, id, node1.node.Id, t)
	CheckFindRoot(node3, id, node1.node.Id, t)
	id = ID{5, 8, 0xf, 0xf}
	CheckFindRoot(node1, id, node2.node.Id, t)
	CheckFindRoot(node2, id, node2.node.Id, t)
	CheckFindRoot(node3, id, node2.node.Id, t)

	node1.tapestry.Leave()
	node2.tapestry.Leave()
	node3.tapestry.Leave()
}