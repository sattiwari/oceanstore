package raft

import (
	"../../tapestry/tapestry"
	"crypto/sha1"
	"math/big"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

/* Node's can be in three possible states */
type NodeState int

// Tapestry's id
type ID tapestry.ID

const (
	FOLLOWER_STATE NodeState = iota
	CANDIDATE_STATE
	LEADER_STATE
	JOIN_STATE
)

type RaftNode struct {
	Id                 string
	Listener           net.Listener
	listenPort         int

	//At any given time each server is in one of three states: leader, follower, or candidate.
	State              NodeState
	LeaderAddress      *NodeAddr

	conf               *Config
	IsShutDown bool
	RPCServer *RaftRPCServer
	mutex sync.Mutex
	Testing *TestingPolicy

	logCache           []LogEntry

	//file descriptors and values for persistent state
	logFileDescriptor  FileData
	metaFileDescriptor FileData
	stableState NodeStableState
	ssMutex sync.Mutex

	//leader specific volatile state
	commitIndex uint64
	lastApplied uint64
	leaderMutex map[string]uint64
	nextIndex map[string]uint64
	matchIndex map[string]uint64

	// channels to send and rcv RPC messages
	appendEntries      chan AppendEntriesMsg
	requestVote        chan RequestVoteMsg
	clientRequest      chan ClientRequestMsg
	registerClient     chan RegisterClientMsg
	gracefulExit       chan bool

//	the replicated state machine
	hash []byte
	requestMutex sync.Mutex
	requestMap map[uint64]ClientRequestMsg

	fileMap    map[string]string
	fileMapMtx sync.Mutex
	lockMap    map[string]bool
	lockMapMtx sync.Mutex
}

type NodeAddr struct {
	Id   string
	Addr string
}

func CreateNode(localPort int, leaderAddr *NodeAddr, conf *Config) (rp *RaftNode, err error) {
	var r RaftNode
	rp = &r
	var conn net.Listener

	r.IsShutDown = false
	r.conf = conf

	//init rpc channels
	r.appendEntries = make(chan AppendEntriesMsg)
	r.requestVote = make(chan RequestVoteMsg)
	r.clientRequest = make(chan ClientRequestMsg)
	r.registerClient = make(chan RegisterClientMsg)
	r.gracefulExit = make(chan bool)

	r.hash = nil
	r.requestMap = make(map[uint64]ClientRequestMsg)

	r.commitIndex = 0
	r.lastApplied = 0
	r.nextIndex = make(map[string]uint64)
	r.matchIndex = make(map[string]uint64)

	r.fileMap = make(map[string]string)
	r.lockMap = make(map[string]bool)
	r.Testing = NewTesting()
	r.Testing.PauseWorld(false)

	if localPort != 0 {
		conn, err = OpenPort(localPort)
	} else {
		conn, localPort, err = OpenListener()
	}

	if err != nil {
		return nil, err
	}

	// create node id based on listener address
	r.Id = AddrToId(conn.Addr().String(), conf.NodeIdSize)

	r.Listener = conn
	r.listenPort = localPort
	Out.Printf("started node with id %v, listening at %v", r.Id, conn.Addr().String())

	freshNode, err := r.initStableStore()
	if err != nil {
		Error.Printf("Error initializing the stable store: %v \n", err)
		return nil, err
	}

	r.setLocalAddr(&NodeAddr{Id: r.Id, Addr: conn.Addr().String()})

	// Start RPC server
	r.RPCServer = &RaftRPCServer{rp}
	rpc.RegisterName(r.GetLocalAddr().Addr, r.RPCServer)
	go r.RPCServer.startRpcServer()

	if freshNode {
		r.State = JOIN_STATE
		if leaderAddr != nil {
			err = JoinRPC(leaderAddr, r.GetLocalAddr())
		} else {
			Out.Printf("Waiting to start nodes until all have joined\n")
			go r.startNodes()
		}
	} else {
		r.State = FOLLOWER_STATE
		go r.run()
	}

	return
}

func (r *RaftNode) startNodes() {
	r.mutex.Lock()
	r.AppendOtherNodes(*r.GetLocalAddr())
	r.mutex.Unlock()

	for len(r.GetOtherNodes()) < r.conf.ClusterSize {
		time.Sleep(time.Millisecond * 100)
	}

	for _, otherNode := range r.GetOtherNodes() {
		if r.Id != otherNode.Id {
			Out.Printf("(%v) Starting node-%v\n", r.Id, otherNode.Id)
			StartNodeRPC(otherNode, r.GetOtherNodes())
		}
	}

	// Start the Raft finite-state-machine, initially in follower state
	go r.run()
}

func CreateLocalCluster(config *Config) ([]*RaftNode, error) {
	if config == nil {
		config = DefaultConfig()
	}
	err := CheckConfig(config)
	if err != nil {
		return nil, err
	}

	nodes := make([]*RaftNode, config.ClusterSize)

	nodes[0], err = CreateNode(0, nil, config)
	for i := 1; i < config.ClusterSize; i++ {
		nodes[i], err = CreateNode(0, nodes[0].GetLocalAddr(), config)
		if err != nil {
			return nil, err
		}
	}
	return nodes, nil
}



func AddrToId(addr string, length int) string {
	h := sha1.New()
	h.Write([]byte(addr))
	v := h.Sum(nil)
	keyInt := big.Int{}
	keyInt.SetBytes(v[:length])
	return keyInt.String()
}

func (r *RaftNode) Exit() {
	Out.Printf("Abruptly shutting down node!")
	os.Exit(0)
}

func (r *RaftNode) GracefulExit() {
	r.Testing.PauseWorld(true)
	Out.Println("gracefully shutting down the node %v", r.Id)
	r.gracefulExit <- true
}

func (r *RaftNode) GetConfig() *Config {
	return r.conf
}

func (r *RaftNode) run() {
	curr := r.doFollower
	for curr != nil {
		curr = curr()
	}
}