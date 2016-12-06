package chord

import (
	"errors"
	"net/rpc"
	"fmt"
)

type RemoteId struct {
	Id []byte
}

type RemoteQuery struct {
	FromId []byte
	Id     []byte
}

type IdReply struct {
	Id    []byte
	Addr  string
	Valid bool
}

type KeyValueReq struct {
	NodeId []byte
	Key    string
	Value  string
}

type KeyValueReply struct {
	Key   string
	Value string
}

/* RPC connection map cache */
var connMap = make(map[string]*rpc.Client)

/* Find the successor node of a given ID in the entire ring */
func FindSuccessor_RPC(remoteNode *RemoteNode, id []byte) (*RemoteNode, error) {
	if remoteNode == nil {
		return nil, errors.New("RemoteNode is empty!")
	}
	var reply IdReply
	err := makeRemoteCall(remoteNode, "FindSuccessor", RemoteQuery{remoteNode.Id, id}, &reply)

	rNode := new(RemoteNode)
	rNode.Id = reply.Id
	rNode.Addr = reply.Addr
	return rNode, err
}

/* Helper function to make a call to a remote node */
func makeRemoteCall(remoteNode *RemoteNode, method string, req interface{}, rsp interface{}) error {
	// Dial the server if we don't already have a connection to it
	remoteNodeAddrStr := remoteNode.Addr
	var err error
	client, ok := connMap[remoteNodeAddrStr]
	if !ok {
		client, err = rpc.Dial("tcp", remoteNode.Addr)
		if err != nil {
			return err
		}
		connMap[remoteNodeAddrStr] = client
	}

	// Make the request
	uniqueMethodName := fmt.Sprintf("%v.%v", remoteNodeAddrStr, method)
	err = client.Call(uniqueMethodName, req, rsp)
	if err != nil {
		return err
	}

	return nil
}

/* Get the successor ID of a remote node */
func GetSuccessorId_RPC(remoteNode *RemoteNode) (*RemoteNode, error) {
	var reply IdReply
	err := makeRemoteCall(remoteNode, "GetSuccessorId", RemoteId{remoteNode.Id}, &reply)
	if err != nil {
		return nil, err
	}
	rNode := new(RemoteNode)
	rNode.Id = reply.Id
	rNode.Addr = reply.Addr
	return rNode, err
}

/* Get a value from a remote node's datastore for a given key */
func Get_RPC(locNode *RemoteNode, key string) (string, error) {
	if locNode == nil {
		return "", errors.New("RemoteNode is empty!")
	}

	var reply KeyValueReply
	req := KeyValueReq{locNode.Id, key, ""}
	err := makeRemoteCall(locNode, "GetLocal", &req, &reply)

	return reply.Value, err
}

/* Put a key/value into a datastore on a remote node */
func Put_RPC(locNode *RemoteNode, key string, value string) error {
	if locNode == nil {
		return errors.New("RemoteNode is empty!")
	}

	var reply KeyValueReply
	req := KeyValueReq{locNode.Id, key, value}
	err := makeRemoteCall(locNode, "PutLocal", &req, &reply)

	return err
}
