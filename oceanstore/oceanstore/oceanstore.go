package oceanstore

import (
	"../../raft/raft"
	"../../tapestry/tapestry"
)

type OceanAddr struct {
	Addr string
}

type OceanNode struct {
	tnodes      []*tapestry.Tapestry
	rnodes      []*raft.RaftNode
	rootV       uint32
	clientPaths map[uint64]string       // client id -> curpath
	clients     map[uint64]*raft.Client // client id -> client

	Local      OceanAddr
	raftClient *raft.Client
	server     *PuddleRPCServer
}