package raft

import (
	"net"
	"sync"
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
	registerClient     chan ClientRegistration
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

func createNode(localPort int, remoteAddr *NodeAddr, conf *Config) (*RaftNode, error) {
	var r RaftNode
	node := &r

	return node, nil

}

func createCluster(conf *Config) ([] *RaftNode, error) {
	if conf == nil {
		conf = DefaultConfig()
	}
	err := CheckConfig(conf)
	if err != nil {
		return nil, err
	}
	nodes := make([] *RaftNode, conf.ClusterSize)
	nodes[0], err = createNode(0, nil, conf)
	for i := 1; i < conf.ClusterSize; i++ {
		nodes[i], err = createNode(0, nodes[0].LeaderAddress, conf)
		if err != nil {
			return nil, err
		}
	}
	return nodes, nil
}