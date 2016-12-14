package raft

import (
	"math/big"
	"crypto/sha1"
	"time"
)

func AddrToId(addr string, length int) string {
	h := sha1.New()
	h.Write([]byte(addr))
	v := h.Sum(nil)
	keyInt := big.Int{}
	keyInt.SetBytes(v[:length])
	return keyInt.String()
}

func (r *RaftNode) electionTimeOut() <- chan time.Time {
	return time.After(time.Duration(r.conf.ElectionTimeout))
}