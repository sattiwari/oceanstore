package oceanstore

import (
	"../../raft/raft"
	"../../tapestry/tapestry"
)

const TAPESTRY_NODES = 3
const RAFT_NODES = 1

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
	server     *OceanRPCServer
}

func Start() (p *OceanNode, err error) {
	var ocean OceanNode
	p = &ocean
	ocean.tnodes = make([]*tapestry.Tapestry, TAPESTRY_NODES)
	ocean.rnodes = make([]*raft.RaftNode, RAFT_NODES)
	ocean.clientPaths = make(map[uint64]string)
	ocean.clients = make(map[uint64]*raft.Client)

	// Start runnning the tapestry nodes. --------------
	t, err := tapestry.Start(0, "")
	if err != nil {
		panic(err)
	}

	ocean.tnodes[0] = t
	for i := 1; i < TAPESTRY_NODES; i++ {
		t, err = tapestry.Start(0, ocean.tnodes[0].GetLocalAddr())
		if err != nil {
			panic(err)
		}
		ocean.tnodes[i] = t
	}

	ocean.rnodes, err = raft.CreateLocalCluster(raft.DefaultConfig())
	if err != nil {
		panic(err)
	}

	return
}