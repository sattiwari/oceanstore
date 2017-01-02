package raft

import (
	"crypto/md5"
	"fmt"
	"strings"
)

func (r *RaftNode) processLog(entry LogEntry) ClientReply {
	Out.Printf("%v\n", entry)
	status := OK
	response := ""
	switch entry.Command {
	case HASH_CHAIN_INIT:
		if r.hash == nil {
			r.hash = entry.Data
			response = fmt.Sprintf("%v", r.hash)
		} else {
			status = REQ_FAILED
			response = "The hash chain should only be initialized once!"
		}
	case HASH_CHAIN_ADD:
		if r.hash == nil {
			status = REQ_FAILED
			response = "The hash chain hasn't been initialized yet"
		} else {
			sum := md5.Sum(r.hash)
			fmt.Printf("hash is changing from %v to %v\n", r.hash, sum)
			r.hash = sum[:]
			response = fmt.Sprintf("%v", r.hash)
		}
	// For each of the following idk what to do with the hash chain
	//TODO: Do the byte[] and string casting for entry.Data
	case REMOVE:
		//So by now we have received consensus, we need to delete
		r.requestMutex.Lock()
		key := string(entry.Data)
		if entry.Data == nil {
			response = "FAIL:The key cannot be nil"
		} else if val, ok := r.fileMap[key]; ok {
			delete(r.fileMap, key)
			response = "SUCCESS:" + val
		} else {
			response = "FAIL:The key does not exist"
		}
		r.requestMutex.Unlock()
	case SET:
		r.requestMutex.Lock()
		if entry.Data == nil {
			response = "FAIL:The key cannot be nil"
		} else {
			keyVal := string(entry.Data)
			keyValAr := strings.Split(keyVal, ":")
			r.fileMap[keyValAr[0]] = keyValAr[1]
			response = "SUCCESS:" + keyValAr[1]
		}
		r.requestMutex.Unlock()

	case LOCK:
		r.lockMapMtx.Lock()
		key := string(entry.Data)
		if entry.Data == nil {
			response = "FAIL:The key cannot be nil"
		} else if _, ok := r.lockMap[key]; ok {
			//means its locked --
			response = "FAIL:The key is locked is locked"
		} else {
			//means its unlocked, so we lock
			r.lockMap[key] = true
			response = "SUCCESS:Key " + key + "is now locked"
		}
		r.lockMapMtx.Unlock()

	case UNLOCK:
		r.lockMapMtx.Lock()
		key := string(entry.Data)
		if entry.Data == nil {
			response = "FAIL:The key cannot be nil"
		} else {
			//We dont care and we unlock its for the user not to unlock something of
			//someone else
			delete(r.lockMap, key)
			response = "SUCCESS:Key " + key + "is now unlocked"
		}
		r.lockMapMtx.Unlock()

	default:
		response = "Success!"
	}

	reply := ClientReply{
		Status:     status,
		Response:   response,
		LeaderHint: *r.GetLocalAddr(),
	}

	if entry.CacheId != "" {
		r.AddRequest(entry.CacheId, reply)
	}

	r.requestMutex.Lock()
	msg, exists := r.requestMap[entry.Index]
	if exists {
		msg.reply <- reply
		delete(r.requestMap, entry.Index)
	}
	r.requestMutex.Unlock()

	return reply
}
