package raft

import (
	"os"
	"fmt"
)

type NodeStableState struct {
	currentTerm uint64
}

type LogEntry struct {
	Index uint64
	Term uint64
	Data []byte
}

type FileData struct {
	fileDescriptor *os.File
	sizeOfFile uint64
	fileName string
	logEntryIdxToFileSizeMap map[uint64]uint64
	isFileDescriptorOpen bool
}

func (r *RaftNode) getLastLogIndex() uint64 {
	return uint64(len(r.logCache) - 1)
}

func (r *RaftNode) initStableStore() (bool, error) {

	err := os.Mkdir(r.conf.LogPath, 0777)
	if err != nil {
		fmt.Println("Error in creating the log directory")
	} else {
		fmt.Println("Created log directory %v\n", r.conf.LogPath)
	}
	r.logFile = fmt.Sprint("%v/%d_raft_log.dat", r.conf.LogPath, r.port)
	r.metaFile = fmt.Sprint("%v/%d_raft_meta.dat", r.conf.LogPath, r.port)
	_, raftLogExists  := getFileStats(r.logFile)
	_, raftMetaExists := getFileStats(r.logFile)

	if raftLogExists && raftMetaExists {

	} else if (!raftLogExists && raftMetaExists) || (raftLogExists && !raftMetaExists) {

	} else {
		fres
	}

}
