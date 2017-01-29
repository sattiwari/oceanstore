package oceanstore

import (
	"../../raft/raft"
	"../../tapestry/tapestry"
	"math/rand"
)

const TAPESTRY_NODES = 3
const RAFT_NODES = 1

type OceanAddr struct {
	Addr string
}

type Vguid string
type Aguid string
type Guid string

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

	// RPC server --------------------------------------
	ocean.server = newOceanstoreRPCServer(p)
	ocean.Local = OceanAddr{ocean.server.listener.Addr().String()}
	// -------------------------------------------------

	// Create ocean raft client. Persist until raft is settled
	client, err := CreateClient(ocean.Local)
	for err != nil {
		client, err = CreateClient(ocean.Local)
	}

	ocean.raftClient = ocean.clients[client.Id]
	if ocean.raftClient == nil {
		panic("Could not retrieve puddle raft client.")
	}

	// Create the root node ----------------------------
	_, err = ocean.mkdir(&MkdirRequest{ocean.raftClient.Id, "/"})
	if err != nil {
		panic("Could not create root node")
	}

	return
}

func (ocean *OceanNode) getCurrentDir(id uint64) string {
	curdir, ok := ocean.clientPaths[id]
	if !ok {
		panic("Did not found the current path of a client that is supposed to be registered")
	}
	return curdir
}

func randSeq(n int) string {
	var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (puddle *OceanNode) getRandomRaftNode() *raft.RaftNode {
	index := rand.Int() % RAFT_NODES
	return puddle.rnodes[index]
}