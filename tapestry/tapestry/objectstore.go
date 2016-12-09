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

/*
	Get the nodes that are advertising a given key
*/
func (store *ObjectStore) Get(key string) (replicas []Node) {
	store.mutex.Lock()

	replicas = slice(store.data[key])

	store.mutex.Unlock()

	return
}

// Utility function to get the keys of a map
func slice(valmap map[Node]*time.Timer) (values []Node) {
	for value, _ := range valmap {
		values = append(values, value)
	}
	return
}
