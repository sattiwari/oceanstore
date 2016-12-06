package chord

import "log"

/* Get a value in the datastore, provided an abitrary node in the ring */
func Get(node *Node, key string) (string, error) {
	remNode, err := node.locate(key)
	if err != nil {
		log.Fatal(err)
	}
	return Get_RPC(remNode, key)
}

/* Put a key/value in the datastore, provided an abitrary node in the ring */
func Put(node *Node, key string, value string) error {
	remNode, err := node.locate(key)
	if err != nil {
		log.Fatal(err)
	}
	return Put_RPC(remNode, key, value)
}

/* Internal helper method to find the appropriate node in the ring */
func (node *Node) locate(key string) (*RemoteNode, error) {
	id := HashKey(key)
	return node.findSuccessor(id)
}
