package chord

import (
	"bytes"
	"fmt"
	"errors"
)

/* Validate that we're executing this RPC on the intended node */
func validateRpc(node *Node, reqId []byte) error {
	if !bytes.Equal(node.Id, reqId) {
		errStr := fmt.Sprintf("Node ids do not match %v, %v", node.Id, reqId)
		return errors.New(errStr)
	}
	return nil
}

func (node *Node) GetPredecessorId(req *RemoteId, reply *IdReply) error {
	if err := validateRpc(node, req.Id); err != nil {
		return err
	}
	// Predecessor may be nil, which is okay.
	if node.Predecessor == nil {
		reply.Id = nil
		reply.Addr = ""
		reply.Valid = false
	} else {
		reply.Id = node.Predecessor.Id
		reply.Addr = node.Predecessor.Addr
		reply.Valid = true
	}
	return nil
}

func (node *Node) SetPredecessorId(req *UpdateReq, reply *RpcOkay) error {
	if err := validateRpc(node, req.FromId); err != nil {
		return err
	}
	node.Predecessor.Id = req.UpdateId
	node.Predecessor.Addr = req.UpdateAddr
	reply.Ok = true
	return nil
}

func (node *Node) GetSuccessorId(req *RemoteId, reply *IdReply) error {
	if err := validateRpc(node, req.Id); err != nil {
		return err
	}

	reply.Id = node.Successor.Id
	reply.Addr = node.Successor.Addr
	reply.Valid = true
	return nil
}

func (node *Node) SetSuccessorId(req *UpdateReq, reply *RpcOkay) error {
	if err := validateRpc(node, req.FromId); err != nil {
		return err
	}
	node.Successor.Id = req.UpdateId
	node.Successor.Addr = req.UpdateAddr
	reply.Ok = true
	return nil
}

func (node *Node) FindSuccessor(query *RemoteQuery, reply *IdReply) error {
	if err := validateRpc(node, query.FromId); err != nil {
		return err
	}
	remNode, err := node.findSuccessor(query.Id)
	if err != nil {
		reply.Valid = false
		return err
	}
	reply.Id = remNode.Id
	reply.Addr = remNode.Addr
	reply.Valid = true
	return nil
}

func (node *Node) ClosestPrecedingFinger(query *RemoteQuery, reply *IdReply) error {
	if err := validateRpc(node, query.FromId); err != nil {
		return err
	}
	//remoteId and fromId
	for i := KEY_LENGTH - 1; i >= 0; i-- {
		if BetweenRightIncl(node.FingerTable[i].Node.Id, node.Id, query.Id) {
			reply.Id = node.FingerTable[i].Node.Id
			reply.Addr = node.FingerTable[i].Node.Addr
			reply.Valid = true
			return nil
		}
	}

	reply.Valid = false
	return errors.New("There is no closest preceding finger")
}

func (node *Node) Notify(req *NotifyReq, reply *RpcOkay) error {
	if err := validateRpc(node, req.NodeId); err != nil {
		reply.Ok = false
		return err
	}
	remote_node := new(RemoteNode)
	remote_node.Id = req.UpdateId
	remote_node.Addr = req.UpdateAddr
	node.notify(remote_node)
	reply.Ok = true
	return nil
}
