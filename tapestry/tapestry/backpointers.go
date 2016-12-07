package tapestry

import "sync"

/*
	Backpointers are stored by level, like the routing table
	A backpointer at level n indicates that the backpointer shares a prefix of length n with this node
	Access to the backpointers is managed by a lock
*/
type Backpointers struct {
	local Node             // the local tapestry node
	sets  [DIGITS]*NodeSet // backpointers
}

/*
	Represents a set of nodes.  The implementation is just a wrapped map, and access is controlled with a mutex.
*/
type NodeSet struct {
	mutex sync.Mutex
	data  map[Node]bool
}

/*
	Creates and returns a new backpointer set
*/
func NewBackpointers(me Node) *Backpointers {
	b := new(Backpointers)
	b.local = me
	for i := 0; i < DIGITS; i++ {
		b.sets[i] = NewNodeSet()
	}
	return b
}

/*
	Create a new node set
*/
func NewNodeSet() *NodeSet {
	s := new(NodeSet)
	s.data = make(map[Node]bool)
	return s
}
