package raft

import (
	"math/big"
	"crypto/sha1"
	"net/rpc"
	"fmt"
	"os"
)

var connectionMap = make(map[string] *rpc.Client)

func AddrToId(addr string, length int) string {
	h := sha1.New()
	h.Write([]byte(addr))
	v := h.Sum(nil)
	keyInt := big.Int{}
	keyInt.SetBytes(v[:length])
	return keyInt.String()
}

func makeRemoteCall(remoteNode *NodeAddr, method string, req interface{}, resp interface{}) error {
	var err error
	remoteAddStr := remoteNode.Address
	client, ok := connectionMap[remoteAddStr]
	if !ok {
		client, err = rpc.Dial("tcp", remoteAddStr)
		if err != nil {
			return nil
		}
		connectionMap[remoteAddStr] = client
	}
	methodPath := fmt.Sprintf("%v.%v", remoteAddStr, method)
	err = client.Call(methodPath, req, resp)
	if err != nil {
		return err
	}
	return nil
}

func getFileStats(filename string) (uint64, bool) {
	stats, err := os.Stat(filename)
	if err == nil {
		return uint64(stats.Size()), true
	} else if os.IsNotExist(err) {
		return 0, false
	} else {
		panic(err)
	}
}

func Exit()  {
	Out.Println("abruptly shutting down node")
	os.Exit(0)
}

func (r *RaftNode) hasMajority(N uint64) bool {
	return true
}

type UInt64Slice []uint64

func (p UInt64Slice) Len() uint64 {
	return uint64(len(p))
}

func (p UInt64Slice) Swap(i, j uint64) {
	p[i], p[j] = p[j], p[i]
}