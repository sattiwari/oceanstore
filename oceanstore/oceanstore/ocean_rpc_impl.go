package oceanstore

import (
	"net"
	"net/rpc"
)

type OceanRPCServer struct {
	node     *OceanNode
	listener net.Listener
	rpc      *rpc.Server
}