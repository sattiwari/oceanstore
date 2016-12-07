package tapestry

import (
	"testing"
)

/*This test adds 100,000 nodes to the table and removes them, checking
 that all were deleted.*/

func TestAddAndRemove(t *testing.T) {
	NUM_NODES := 100000
	me := Node{RandomID(), ""}
	table := NewRoutingTable(me)
	nodes := make([]Node, NUM_NODES)
	for i := 0; i < NUM_NODES; i++ {
		nodes[i] = Node{RandomID(), ""}
		table.Add(nodes[i])
	}
	for i := 0; i < NUM_NODES; i++ {
		table.Remove(nodes[i])
	}

	for i := 0; i < DIGITS; i++ {
		for j := 0; j < BASE; j++ {
			if len(*(table.rows[i][j])) > 1 {
				t.Errorf("Nodes were not deleted from table.")
			}
			if len(*(table.rows[i][j])) == 1 &&
				!equal_ids(me.Id, (*(table.rows[i][j]))[0].Id) {
				t.Errorf("Nodes were not deleted from table.")
			}
		}
	}
}
