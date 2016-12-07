package tapestry

import "sync"

/*
	A routing table has a number of levels equal to the number of digits in an ID (default 40)
	Each level has a number of slots equal to the digit base (default 16)
	A node that exists on level n thereby shares a prefix of length n with the local node.
	Access to the routing table is managed by a lock
*/
type RoutingTable struct {
	local Node                  // the local tapestry node
	mutex sync.Mutex            // to manage concurrent access to the routing table. could have a per-level mutex though
	rows  [DIGITS][BASE]*[]Node // the rows of the routing table
}

/*
	Creates and returns a new routing table, placing the local node at the appropriate slot in each level of the table
*/
func NewRoutingTable(me Node) *RoutingTable {
	t := new(RoutingTable)
	t.local = me

	// Create the node lists with capacity of SLOTSIZE
	for i := 0; i < DIGITS; i++ {
		for j := 0; j < BASE; j++ {
			slot := make([]Node, 0, SLOTSIZE)
			t.rows[i][j] = &slot
		}
	}

	// Make sure each row has at least our node in it
	for i := 0; i < DIGITS; i++ {
		slot := t.rows[i][t.local.Id[i]]
		*slot = append(*slot, t.local)
	}

	return t
}

func GetFurthest(id ID, nodes []Node) int {
	furthest := 0
	for i := 1; i < SLOTSIZE; i++ {
		if id.Closer(nodes[furthest].Id, nodes[i].Id) {
			furthest = i
		}
	}
	return furthest
}

/*
	Adds the given node to the routing table
	Returns true if the node did not previously exist in the table and was subsequently added
	Returns the previous node in the table, if one was overwritten
*/
func (t *RoutingTable) Add(node Node) (added bool, previous *Node) {
	t.mutex.Lock()

	// Find table slot.
	level := SharedPrefixLength(node.Id, t.local.Id)

	if level == DIGITS {
		added = false
		t.mutex.Unlock()
		return
	}

	// fmt.Printf("%v, %v\n", i, node.Id[i])
	slot := t.rows[level][node.Id[level]]

	// Check if it exists; if it does return false
	for i := 0; i < len(*slot); i++ {
		if SharedPrefixLength((*slot)[i].Id, node.Id) == DIGITS {
			added = false
			t.mutex.Unlock()
			return
		}
	}

	// Append new slot and make sure theres a 3 node maximum.

	for i := 0; i <= level; i++ {
		slot = t.rows[i][node.Id[i]]
		*slot = append(*slot, node)
		if len(*slot) > SLOTSIZE {
			furthest := GetFurthest(t.local.Id, *slot)
			previous = &(*slot)[furthest]
			*slot = append((*slot)[:furthest], (*slot)[furthest+1:]...)
		}
	}

	added = true
	t.mutex.Unlock()
	return
}

/*
	Removes the specified node from the routing table, if it exists
	Returns true if the node was in the table and was successfully removed
*/
func (t *RoutingTable) Remove(node Node) (wasRemoved bool) {
	t.mutex.Lock()

	// Get the table slot
	level := SharedPrefixLength(node.Id, t.local.Id)
	if level == DIGITS {
		// Never delete youself on your own routing table.
		wasRemoved = false
		t.mutex.Unlock()
		return
	}

	wasRemoved = false

	for j := 0; j <= level; j++ {
		slot := t.rows[j][node.Id[j]]

		// Find and remove node
		for i := 0; i < len(*slot); i++ {
			if SharedPrefixLength((*slot)[i].Id, node.Id) == DIGITS {
				*slot = append((*slot)[:i], (*slot)[i+1:]...) // This is remove in Go
				wasRemoved = true
			}
		}
	}

	// Return false if node was not found.
	t.mutex.Unlock()
	return
}

