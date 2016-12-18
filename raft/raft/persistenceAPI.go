package raft

import (
	"os"
	"fmt"
)

const INT_GOB_SIZE uint64 = 5

type NodeStableState struct {
	//latest term the server has been assigned, starts with 0 on boot and then increases
	CurrentTerm uint64

	//the candidate id that recieved our vote in the current term. "" if none
	VotedFor string

	//local listening address and id
	LocalAddr NodeAddr

	//addresses of everyone in the cluster
	OtherNodes []NodeAddr

	//client request cache, maps a client request to the response that was sent to them
	ClientRequestSequences map[string]ClientReply
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

func (r *RaftNode) setLocalAddr(addr *NodeAddr) {
	r.ssMutex.Lock()
	defer r.ssMutex.Unlock()
	r.stableState.LocalAddr = *addr
	err := WriteStableState(&r.metaFileDescriptor, r.stableState)
	if err != nil {
		Error.Printf("Unable to flush new local address to disk: %v\n", err)
		panic(err)
	}
}

func (r *RaftNode) GetLocalAddr() *NodeAddr {
	return &r.stableState.LocalAddr
}

func (r *RaftNode) GetOtherNodes() []NodeAddr {
	return r.stableState.OtherNodes
}

func (r *RaftNode) AppendOtherNodes(other NodeAddr) {
	r.ssMutex.Lock()
	defer r.ssMutex.Unlock()
	r.stableState.OtherNodes = append(r.stableState.OtherNodes, other)
	err := WriteStableState(&r.metaFileDescriptor, r.stableState)
	if err != nil {
		Error.Printf("Unable to flush new nodes to disk")
		panic(err)
	}
}

func CreateRaftLog(fileData *FileData) error {
	fd, err := os.OpenFile(fileData.fileName, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0600)
	fileData.fileDescriptor = fd
	fileData.sizeOfFile = uint64(0)
	fileData.logEntryIdxToFileSizeMap = make(map[uint64]uint64)
	fileData.isFileDescriptorOpen = true
	return err
}

func CreateMetaLog(FileData *FileData) error {
	fd, err := os.OpenFile(FileData.fileName, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0600)
	FileData.fileDescriptor = fd
	FileData.isFileDescriptorOpen = true
	return err
}

func (r *RaftNode) initStableStore() (bool, error) {
	freshnode := false
	err := os.Mkdir(r.conf.LogPath, 0777)
	if err != nil {
		fmt.Println("Error in creating the log directory")
	} else {
		fmt.Println("Created log directory %v\n", r.conf.LogPath)
	}

	logFileName  := fmt.Sprint("%v/%d_raft_log.dat", r.conf.LogPath, r.listenPort)
	metaFileName := fmt.Sprint("%v/%d_raft_meta.dat", r.conf.LogPath, r.listenPort)

	r.logFileDescriptor  = FileData{fileName: logFileName}
	r.metaFileDescriptor = FileData{fileName: metaFileName}

	_, raftLogExists  := getFileStats(r.logFileDescriptor.fileDescriptor)

	_, raftMetaExists := false

	if raftLogExists && raftMetaExists {

	} else if (!raftLogExists && raftMetaExists) || (raftLogExists && !raftMetaExists) {

	} else {
		freshnode = true
		fmt.Println("Creating new raft node with meta and log files")

		err := CreateRaftLog(&r.logFileDescriptor)
		if err !=  nil {
			return freshnode, err
		}

		err = CreateMetaLog(&r.metaFileDescriptor)
		if err != nil {
			return freshnode, err
		}

		initEntry := LogEntry{Index: 0, Term: 0, Data: []byte{0}}
		r.appendLogEntry(initEntry)
		r.currentTerm = 0
	}
	return freshnode, nil
}
