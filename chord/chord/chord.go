/*  Purpose: Chord struct and related functions to create new nodes, etc.    */
/*                                                                           */

package chord

import (
	"../utils"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

// Number of bits (i.e. M value), assumes <= 128 and divisible by 8
const KEY_LENGTH = 8

/* Non-local node representation */
type RemoteNode struct {
	Id   []byte
	Addr string
}

/* Local node representation */
type Node struct {
	Id          []byte            /* Unique Node ID */
	Listener    net.Listener      /* Node listener socket */
	Addr        string            /* String of listener address */
	Successor   *RemoteNode       /* This Node's successor */
	Predecessor *RemoteNode       /* This Node's predecessor */
	RemoteSelf  *RemoteNode       /* Remote node of our self */
	IsShutdown  bool              /* Is node in process of shutting down? */
	FingerTable []FingerEntry     /* Finger table entries */
	ftLock      sync.RWMutex      /* RWLock for finger table */
	dataStore   map[string]string /* Local datastore for this node */
	dsLock      sync.RWMutex      /* RWLock for datastore */
}

/* Create Chord node with random ID based on listener address */
func CreateNode(parent *RemoteNode) (*Node, error) {
	node := new(Node)
	err := node.init(parent, nil)
	if err != nil {
		return nil, err
	}
	return node, err
}

/* Initailize a Chord node, start listener, rpc server, and go routines */
func (node *Node) init(parent *RemoteNode, definedId []byte) error {
	if KEY_LENGTH > 128 || KEY_LENGTH%8 != 0 {
		log.Fatal(fmt.Sprintf("KEY_LENGTH of %v is not supported! Must be <= 128 and divisible by 8", KEY_LENGTH))
	}

	listener, _, err := utils.OpenListener()
	if err != nil {
		return err
	}

	node.Id = HashKey(listener.Addr().String())
	if definedId != nil {
		node.Id = definedId
	}

	node.Listener = listener
	node.Addr = listener.Addr().String()
	node.IsShutdown = false
	node.dataStore = make(map[string]string)

	// Populate RemoteNode that points to self
	node.RemoteSelf = new(RemoteNode)
	node.RemoteSelf.Id = node.Id
	node.RemoteSelf.Addr = node.Addr

	// Join this node to the same chord ring as parent
	err = node.join(parent)
	if err != nil {
		return err
	}

	// Populate finger table
	node.initFingerTable()

	// Thread 1: start RPC server on this connection
	rpc.RegisterName(node.Addr, node)
	go node.startRpcServer()

	// Thread 2: kick off timer to stabilize periodically
	ticker1 := time.NewTicker(time.Millisecond * 100) //freq
	go node.stabilize(ticker1)

	// Thread 3: kick off timer to fix finger table periodically
	ticker2 := time.NewTicker(time.Millisecond * 90) //freq
	go node.fixNextFinger(ticker2)

	return err
}

/* Go routine to accept and process RPC requests */
func (node *Node) startRpcServer() {
	for {
		if node.IsShutdown {
			fmt.Printf("[%v] Shutting down RPC server\n", HashStr(node.Id))
			return
		}
		if conn, err := node.Listener.Accept(); err != nil {
			log.Fatal("accept error: " + err.Error())
		} else {
			go rpc.ServeConn(conn)
		}
	}
}