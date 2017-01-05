package oceanstore

import (
	"net"
	"math/rand"
	"time"
	"os"
	"fmt"
	"syscall"
)

// ephemeral port range
const LOW_PORT int = 32768
const HIGH_PORT int = 61000

func OpenListener() (net.Listener, int, error) {
	rand.Seed(time.Now().UTC().UnixNano())
	port := rand.Intn(HIGH_PORT - LOW_PORT) + LOW_PORT
	conn, err := OpenPort(port)
	if err != nil {
		if addrInUse(err) {
			time.Sleep(time.Millisecond * 100)
			return OpenListener()
		} else {
			return nil, 0, err //TODO check if I should use -1 for invalid port
		}
	}
	return conn, port, err
}

func addrInUse(err error) bool {
	if opErr, ok := err.(*net.OpError); ok {
		if osErr, ok := opErr.Err.(*os.SyscallError); ok {
			return osErr.Err == syscall.EADDRINUSE
		}
	}
	return false
}

func OpenPort(port int) (net.Listener, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	addr := fmt.Sprintf("%v:%v", hostname, port)
	conn, err := net.Listen("tcp4", addr)
	return conn, err
}