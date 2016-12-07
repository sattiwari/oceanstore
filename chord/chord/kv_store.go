package chord

import (
	"log"
	"fmt"
)

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

/* Print the contents of a node's data store */
func PrintDataStore(node *Node) {
	fmt.Printf("Node-%v datastore: %v\n", HashStr(node.Id), node.dataStore)
}

func (node *Node) GetLocal(req *KeyValueReq, reply *KeyValueReply) error {
	if err := validateRpc(node, req.NodeId); err != nil {
		return err
	}
	(&node.dsLock).RLock()
	key := req.Key
	val := node.dataStore[key]
	reply.Key = key
	reply.Value = val
	(&node.dsLock).RUnlock()
	return nil
}

func (node *Node) PutLocal(req *KeyValueReq, reply *KeyValueReply) error {
	if err := validateRpc(node, req.NodeId); err != nil {
		return err
	}
	(&node.dsLock).Lock()
	key := req.Key
	val := req.Value
	node.dataStore[key] = val
	reply.Key = key
	reply.Value = val
	(&node.dsLock).Unlock()
	return nil
}

/* When we discover a new predecessor we may need to transfer some keys to it */
/*Oh I think I get it, this one is to send
This was eliminated by the TAs because of its redundancy */
func (node *Node) obtainNewKeys() error {
	//lock the local db and get the keys
	(&node.dsLock).Lock()
	for key, val := range node.dataStore {
		keyByte := HashKey(key)
		if !BetweenRightIncl(keyByte, node.Predecessor.Id, node.Id) {
			//means we send it to the predecessor
			err := Put_RPC(node.Predecessor, key, val)
			if err != nil {
				(&node.dsLock).Unlock()
				return err
			}
			//then we delete it locally
			delete(node.dataStore, key)
		}
	}
	//unlock the db
	(&node.dsLock).Unlock()
	return nil
}

/* Find locally stored keys that are between (predId : fromId], any of
   these nodes should be moved to fromId */
func (node *Node) TransferKeys(req *TransferReq, reply *RpcOkay) error {
	if err := validateRpc(node, req.NodeId); err != nil {
		return err
	}
	(&node.dsLock).Lock()
	for key, val := range node.dataStore {
		keyByte := HashKey(key)
		pred := req.PredId
		if pred == nil {
			pred = node.Id
		}
		if BetweenRightIncl(keyByte, pred, req.FromId) {
			//means we send it to the requester, because it belongs to them
			err := Put_RPC(node.Predecessor, key, val)
			if err != nil {
				(&node.dsLock).Unlock()
				reply.Ok = false
				return err
			}
			//then we delete it locally
			delete(node.dataStore, key)
		}
	}
	(&node.dsLock).Unlock()
	reply.Ok = true
	return nil
}