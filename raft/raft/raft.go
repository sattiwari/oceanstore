package raft

//import "fmt"

type RaftNode struct {
	leaderAddress *NodeAddr
}

type NodeAddr struct {
	address string
	id      string
}

//TODO not yet complete
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
		nodes[i], err = createNode(0, nodes[0].leaderAddress, conf)
		if err != nil {
			return nil, err
		}
	}
	return nodes, nil
}


