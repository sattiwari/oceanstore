package raft

import (
	"fmt"
	"bytes"
	"encoding/gob"
)

func shutdown(off bool) {
	if (off) {
		return nil
	}
}

//TODO FIXME
func (r *RaftNode) getLogTerm(index uint64) uint64 {
	return 0
}

func (r *RaftNode) handleCompetingRequests(msg RequestVoteMsg) bool {
	request := msg.request
	reply := msg.reply
	currentTerm := r.currentTerm
	prevIndex := r.commitIndex
	prevTerm := r.getLogTerm(prevIndex)

	if prevTerm > request.CandidateLastLogTerm {
		reply <- RequestVoteReply{currentTerm, false}
		return false
	} else if prevTerm < request.CandidateLastLogTerm {
		reply <- RequestVoteReply{currentTerm, true}
		return true
	} else {
		if prevIndex > request.CandidateLastLogIndex {
			reply <- RequestVoteReply{currentTerm, false}
			return false
		} else if prevIndex < request.CandidateLastLogIndex {
			reply <- RequestVoteReply{currentTerm, true}
			return true
		} else {
			if currentTerm > request.Term {
				reply <- RequestVoteReply{currentTerm, false}
				return false
			} else if currentTerm < request.Term {
				reply <- RequestVoteReply{currentTerm, true}
				return true
			} else {
				if r.votedFor != "" && r.state == LEADERSTATE {
					reply <- RequestVoteReply{currentTerm, false}
					return false
				} else {
					reply <- RequestVoteReply{currentTerm, true}
					return true
				}
			}
		}
	}
}

func (r *RaftNode) sendNoop() {
	entries := make([]LogEntry, 1)
	entries[0] = LogEntry{r.getLastLogIndex() + 1, r.currentTerm, make([]byte, 0)}
	fmt.Println("NOOP logged with log index %d and term index %d", r.getLastLogIndex() + 1, r.currentTerm)
	r.appendEntries
}

func (r *RaftNode) appendLogEntry(entry LogEntry) error {
	err := AppendLogEntry(&r.logFileDescriptor, &entry)
	if err != nil {
		return err
	}
//	update entry in cache
	return nil
}

func getSizeBytes(size int) ([]byte, error) {
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	err  := e.Encode(size)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func getLogEntryBytes(entry *LogEntry) ([]byte, error) {
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	err  := e.Encode(*entry)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func AppendLogEntry(fileData *FileData, entry *LogEntry) error {
	logBytes, err := getLogEntryBytes(entry)
	if err != nil {
		return err
	}
	size , err := getSizeBytes(logBytes)
	if err != nil {
		return err
	}
	numOfBytesWritten, err := fileData.fileDescriptor.Write(size)
	if err != nil {
		return err
	}
	if numOfBytesWritten != len(logBytes) {
		panic("did not write correct number of bytes")
	}
	fileData.sizeOfFile += numOfBytesWritten
	err = fileData.fileDescriptor.Sync()
	if err != nil {
		return err
	}
	fileData.logEntryIdxToFileSizeMap[entry.Index] = fileData.sizeOfFile
	return nil
}