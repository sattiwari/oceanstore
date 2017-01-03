package oceanstore

import (
	"fmt"
	"net/rpc"
)

var connMap = make(map[string]*rpc.Client)

type ConnectRequest struct {
	FromNode OceanAddr
}

type ConnectReply struct {
	Ok bool
	Id uint64
}

func ConnectRPC(remotenode *OceanAddr, request ConnectRequest) (*ConnectReply, error) {
	fmt.Println("(Oceanstore) RPC Connect to", remotenode.Addr)
	var reply ConnectReply

	err := makeRemoteCall(remotenode, "ConnectImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

type PwdRequest struct {
	ClientId uint64
}

type PwdReply struct {
	Ok   bool
	Path string
}

func pwdRPC(remotenode *OceanAddr, request PwdRequest) (*PwdReply, error) {
	var reply PwdReply

	err := makeRemoteCall(remotenode, "PwdImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

type LsRequest struct {
	ClientId uint64
	Path     string
}

type LsReply struct {
	Ok       bool
	Elements string
}

func lsRPC(remotenode *OceanAddr, request LsRequest) (*LsReply, error) {
	var reply LsReply

	err := makeRemoteCall(remotenode, "LsImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

type CdRequest struct {
	ClientId uint64
	Path     string
}

type CdReply struct {
	Ok bool
}

func cdRPC(remotenode *OceanAddr, request CdRequest) (*CdReply, error) {
	var reply CdReply

	err := makeRemoteCall(remotenode, "CdImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

type MvRequest struct {
	ClientId uint64
	Source   string
	Dest     string
}

type MvReply struct {
	Ok bool
}

func mvRPC(remotenode *OceanAddr, request MvRequest) (*MvReply, error) {
	var reply MvReply

	err := makeRemoteCall(remotenode, "MvImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

type CpRequest struct {
	ClientId uint64
	Source   string
	Dest     string
}

type CpReply struct {
	Ok bool
}

func cpRPC(remotenode *OceanAddr, request CpRequest) (*CpReply, error) {
	var reply CpReply

	err := makeRemoteCall(remotenode, "CpImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

type MkdirRequest struct {
	ClientId uint64
	Path     string
}

type MkdirReply struct {
	Ok bool
}

func mkdirRPC(remotenode *OceanAddr, request MkdirRequest) (*MkdirReply, error) {
	var reply MkdirReply

	err := makeRemoteCall(remotenode, "MkdirImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

type RmdirRequest struct {
	ClientId uint64
	Path     string
}

type RmdirReply struct {
	Ok bool
}

func rmdirRPC(remotenode *OceanAddr, request RmdirRequest) (*RmdirReply, error) {
	var reply RmdirReply

	err := makeRemoteCall(remotenode, "RmdirImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

type MkfileRequest struct {
	ClientId uint64
	Path     string
}

type MkfileReply struct {
	Ok bool
}

func mkfileRPC(remotenode *OceanAddr, request MkfileRequest) (*MkfileReply, error) {
	var reply MkfileReply

	err := makeRemoteCall(remotenode, "MkfileImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

type RmfileRequest struct {
	ClientId uint64
	Path     string
}

type RmfileReply struct {
	Ok bool
}

func rmfileRPC(remotenode *OceanAddr, request RmfileRequest) (*RmfileReply, error) {
	var reply RmfileReply

	err := makeRemoteCall(remotenode, "RmfileImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

type WritefileRequest struct {
	ClientId uint64
	Path     string
	Location uint32
	Buffer   []byte
}

type WritefileReply struct {
	Ok      bool
	Written uint32
}

func writefileRPC(remotenode *OceanAddr, request WritefileRequest) (*WritefileReply, error) {
	var reply WritefileReply

	err := makeRemoteCall(remotenode, "WritefileImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

type CatRequest struct {
	ClientId uint64
	Path     string
	Location uint32
	Count    uint32
}

type CatReply struct {
	Ok     bool
	Read   uint32
	Buffer []byte
}

func catRPC(remotenode *OceanAddr, request CatRequest) (*CatReply, error) {
	var reply CatReply

	err := makeRemoteCall(remotenode, "CatImpl", request, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

/* Helper function to make a call to a remote node */
func makeRemoteCall(remoteNode *OceanAddr, method string, req interface{}, rsp interface{}) error {
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
		client.Close()
		delete(connMap, remoteNodeAddrStr)
		return err
	}

	return nil
}