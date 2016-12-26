package raft

import (
	"net"
	"sync"
	"net/rpc"
	"time"
)

//import "fmt"

type NodeState int

const (
	FOLLOWER_STATE NodeState = iota
	CANDIDATE_STATE
	LEADER_STATE
	JOIN_STATE
)

type RaftNode struct {
	Id                 string
	Listener           net.Listener
	listenPort         uint64
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
}

type NodeAddr struct {
	Address string
	Id      string
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

	r.Testing = NewTesting()
	r.Testing.PauseWorld(false)

	if localPort != 0 {
		conn, err = OpenPort(localPort)
	} else {
		conn, localPort, err = OpenListener()
	}

	if err != nil {
		return
	}

	// create node id based on listener address
	r.Id = AddrToId(conn.Addr().String(), conf.NodeIdSize)

	r.Listener = conn
	r.listenPort = localPort
	Out.Printf("started node with id %v, listening at %v", r.Id, conn.Addr().String())

	freshNode, err := r.initStableStore()
	if err != nil {
		Error.Printf("Error initializing the stable store: %v", err)
		return nil, err
	}

	r.setLocalAddr(&NodeAddr{Id: r.Id, Address: conn.Addr().String()})

	//start rpc server
	r.RPCServer = & RaftRPCServer{rp}
	rpc.RegisterName(r.GetLocalAddr().Address, r.RPCServer)
	go r.RPCServer.startRPCServer()

	if freshNode {
		r.State = JOIN_STATE
		if leaderAddr != nil {
			err = JoinRPC(leaderAddr, r.GetLocalAddr())
		} else {
			Out.Printf("waiting to start node until all have joined\n")
			r.startNodes()
		}
	} else {
		r.State = FOLLOWER_STATE
		go r.run()
	}

	return

}

func (r *RaftNode) startNodes()  {
	r.mutex.Lock()
	r.AppendOtherNodes(r.GetLocalAddr())
	r.mutex.Unlock()

	for len(r.GetOtherNodes()) < r.conf.ClusterSize {
		time.Sleep(time.Millisecond * 100)
	}

	for _, otherNode := range r.GetOtherNodes() {
		if r.Id != otherNode.Id {
			Out.Print("%v starting node %v", r.Id, otherNode.Id)
			StartNodeRPC(otherNode, r.GetOtherNodes()[:])
		}
	}
}

func CreateCluster(conf *Config) ([] *RaftNode, error) {
	if conf == nil {
		conf = DefaultConfig()
	}
	err := CheckConfig(conf)
	if err != nil {
		return nil, err
	}
	nodes := make([] *RaftNode, conf.ClusterSize)
	nodes[0], err = CreateNode(0, nil, conf)
	for i := 1; i < conf.ClusterSize; i++ {
		nodes[i], err = CreateNode(0, nodes[0].LeaderAddress, conf)
		if err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (r *RaftNode) run() {
	curr := r.doFollower()
	for curr != nil {
		curr = curr()
	}
}

func (r *RaftNode) GracefulExit() {
	r.Testing.PauseWorld(true)
	Out.Println("gracefully shutting down the node %v", r.Id)
	r.gracefulExit <- true
}