package raft

import (
	"fmt"
	"crypto/md5"
)

func (r *RaftNode) processLog(entry LogEntry) ClientReply {
	Out.Println("%v", entry)

	response := ""
	status := OK

	switch entry.Command {
	case HASH_CHAIN_INIT:
		if r.hash == nil {
			r.hash = entry.Data
			response = fmt.Sprintf("%v", r.hash)
		} else {
			status = REQUEST_FAILED
			response = "hash chain should only be initialized once"
		}
	case HASH_CHAIN_ADD:
		if r.hash == nil {
			status = REQUEST_FAILED
			response = "hash chain is not yeat initialized"
		} else {
			sum := md5.Sum(r.hash)
			Out.Println("hash is being changed from %v to %v", r.hash, sum)
			r.hash = sum[:]
			response = fmt.Sprintf("%v", r.hash)
		}
	default:
	}

	reply := ClientReply{Status: status, Response: response, LeaderHint: r.GetLocalAddr()}

	r.requestMutex.Lock()

	msg, exists := r.requestMap[entry.Index]
	if exists {
		msg.reply <-reply
		r.AddRequest(*msg.request, reply)
		delete(r.requestMap, entry.Index)
	}

	r.requestMutex.Unlock()

	return reply
}
