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

