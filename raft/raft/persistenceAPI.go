package raft

import (
	"os"
	"fmt"
	"errors"
)

/*
Raft divides time into terms of arbitrary length. Terms act as a logical clock in Raft, they allow servers to detect obsolete information such as stale leaders.
 */

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
	Command FsmCommand

	//this Id is used when caching the response after processing the command. Empty string means no caching.
	CacheId string
}

type FileData struct {
	fileDescriptor *os.File
	sizeOfFile uint64
	fileName string
	logEntryIdxToFileSizeMap map[uint64]uint64
	isFileDescriptorOpen bool
}

func (r *RaftNode) GetLastLogIndex() uint64 {
	return uint64(len(r.logCache) - 1)
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

func CreateMetaLog(FileData *FileData) error {
	fd, err := os.OpenFile(FileData.fileName, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0600)
	FileData.fileDescriptor = fd
	FileData.isFileDescriptorOpen = true
	return err
}

func (r *RaftNode) initStableStore() (bool, error) {
	freshNode := false
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

	logSize, raftLogExists  := getFileStats(r.logFileDescriptor.fileDescriptor)
	r.logFileDescriptor.sizeOfFile = logSize

	metaSize, raftMetaExists := false
	r.metaFileDescriptor.sizeOfFile = metaSize

	if raftLogExists && raftMetaExists {
		Out.Println("Previous stable state exists, repopulate everything")
		entries, err := ReadRaftLog(&r.logFileDescriptor)
		if err != nil {
			return freshNode, err
		}
		r.logCache = entries
		err = openRaftLogForWrite(&r.logFileDescriptor)
		if err != nil {
			return freshNode, err
		}
		ss, err := ReadStableState(&r.metaFileDescriptor)
		if err != nil {
			return freshNode, err
		}
		r.stableState = *ss
	} else if (!raftLogExists && raftMetaExists) || (raftLogExists && !raftMetaExists) {
		Error.Println("Both log and meta files should exist together")
		return errors.New("Both log and meta files should exist together")
		return freshNode, err
	} else {
		freshNode = true
		fmt.Println("Creating new raft node with meta and log files")

		err := CreateRaftLog(&r.logFileDescriptor)
		if err !=  nil {
			return freshNode, err
		}

		err = CreateStableState(&r.metaFileDescriptor)
		if err != nil {
			return freshNode, err
		}

		r.stableState.OtherNodes = make([]NodeAddr, 0)
		r.stableState.ClientRequestSequences = make(map[string]ClientReply)
		r.logCache = make([]LogEntry, 0)

		initEntry := LogEntry{Index: 0, Term: 0, Command: INIT, Data: []byte{0}}
		r.appendLogEntry(initEntry)
		r.setCurrentTerm(0)
	}
	return freshNode, nil
}

func (r *RaftNode) GetCurrentTerm() uint64 {
	return r.stableState.CurrentTerm
}

func (r *RaftNode) setCurrentTerm(newTerm uint64) {
	r.ssMutex.Lock()
	defer r.ssMutex.Unlock()

	if r.stableState.CurrentTerm != newTerm {
		Out.Println("Changing current term from %v to %v", r.stableState.CurrentTerm, newTerm)
	}
	r.stableState.CurrentTerm = newTerm
	err := WriteStableState(&r.metaFileDescriptor, r.stableState)
	if err != nil {
		Error.Println("Unable to flush new term to disk %v", err)
		panic(err)
	}
}

func (r *RaftNode) GetVotedFor() string {
	return r.stableState.VotedFor
}

func (r *RaftNode) setVotedFor(candidateId string) {
	r.ssMutex.Lock()
	defer r.ssMutex.Unlock()
	r.stableState.VotedFor = candidateId
	err := WriteStableState(&r.metaFileDescriptor, r.stableState)
	if err != nil {
		Error.Println("unable to flush newly voted for to disk %v", err)
		panic(err)
	}
}


func (r *RaftNode) GetLocalAddr() *NodeAddr {
	return &r.stableState.LocalAddr
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

func (r *RaftNode) GetOtherNodes() []NodeAddr {
	return r.stableState.OtherNodes
}

func (r *RaftNode) setOtherNodes(nodes []NodeAddr) {
	r.ssMutex.Lock()
	defer r.ssMutex.Unlock()
	r.stableState.OtherNodes = nodes
	err := WriteStableState(&r.metaFileDescriptor, r.stableState)
	if err != nil {
		Error.Printf("unable to flush new other nodes to disk: %v", err)
		panic(err)
	}
}


func (r *RaftNode) CheckClientRequestCache(clientReq ClientRequest) (*ClientReply, bool) {
	uniqueId := fmt.Sprintf("%v-%v", clientReq.ClientId, clientReq.SequenceNumber)
	val, ok := r.stableState.ClientRequestSequences[uniqueId]
	if ok {
		return &val, ok
	} else {
		return nil, ok
	}
}

func (r *RaftNode) AddRequest(req ClientRequest, rep ClientReply) error {
	r.ssMutex.Lock()
	defer r.ssMutex.Unlock()

	uniqueID := fmt.Sprintf("%v-%v", req.ClientId, req.SequenceNumber)
	_, ok := r.stableState.ClientRequestSequences[uniqueID]
	if ok {
		return errors.New("Request with same client and sequence number exists")
	}
	r.stableState.ClientRequestSequences[uniqueID] = rep
	err := WriteStableState(&r.metaFileDescriptor, r.stableState)
	if err != nil {
		Error.Println("Unable to flush new client request to disk %v", err)
		panic(err)
	}
	return nil
}

func (r *RaftNode) getLogEntry(index uint64) *LogEntry {
	if index < uint64(len(r.logCache)) {
		return &r.logCache(index)
	} else {
		return nil
	}
}

func (r *RaftNode) getLastLogEntry() *LogEntry {
	return r.getLogEntry(r.getLastLogIndex())
}

func (r *RaftNode) getLogEntries(start uint64, end uint64) []LogEntry {
	if start < uint64(len(r.logCache)) {
		if end > uint64(len(r.logCache)) {
			end = uint64(len(r.logCache))
		} else {
			end++
		}
		return r.logCache[start:end]
	} else {
		return make([]LogEntry, 0)
	}
}

func (r *RaftNode) getLastLogIndex() uint64 {
	return uint64(len(r.logCache)) - 1
}

func (r *RaftNode) getLastLogTerm() uint64 {
	return r.getLastLogEntry().Term
}

func (r *RaftNode) getLogTerm(index uint64) uint64  {
	return r.getLogEntry(index).Term
}

func (r *RaftNode) appendLogEntry(entry LogEntry) error {
	err := AppendLogEntry(&r.logFileDescriptor, &entry)
	if err != nil {
		return err
	}
	r.logCache = append(r.logCache, entry)
	return nil
}

// truncate file to remove everything at index and after it
func (r *RaftNode) truncateLog(index uint64) error {
	err := TruncateLog(&r.logFileDescriptor, index)
	if err != nil {
		return err
	}
	r.logCache = r.logCache[:index]
	return nil
}

func (r *RaftNode) RemoveLogs() error {
	r.logFileDescriptor.fileDescriptor.Close()
	r.logFileDescriptor.isFileDescriptorOpen = false
	err := os.Remove(r.logFileDescriptor.fileName)
	if err != nil {
		Error.Println("unable to remove raft log file")
		return err
	}
	r.metaFileDescriptor.fileDescriptor.Close()
	r.metaFileDescriptor.isFileDescriptorOpen = false
	err = os.Remove(r.metaFileDescriptor.fileName)
	if err != nil {
		Error.Println("unable to remove meta file")
		return err
	}
	return nil
}