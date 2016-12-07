package tapestry

import (
	"sync"
	"time"
)

/*
	Objects advertised to the tapestry get stored in the object store of the object's root node.
	An object can be advertised by multiple nodes
	Objects time out after some amount of time if the advertising node is not heard from
*/
type ObjectStore struct {
	mutex sync.Mutex                      // to manage concurrent access to the object store
	data  map[string]map[Node]*time.Timer // multimap: stores multiple nodes per key, and each node has a timeout
}

/*
	Create a new objectstore
*/
func NewObjectStore() *ObjectStore {
	m := new(ObjectStore)
	m.data = make(map[string]map[Node]*time.Timer)
	return m
}
