package raft

import (
	"errors"
	"fmt"
	"os"
)

type NodeStableState struct {
	/* Latest term the server has seen (initialized */
	/* to 0 on first boot, increases monotonically) */
	CurrentTerm uint64

	/* The candidate Id that received our vote in   */
	/* the current term (or "" if none).            */
	VotedFor string

	/* Our local listening address and Id           */
	LocalAddr NodeAddr

	/* The addresses of everyone in our cluster     */
	OtherNodes []NodeAddr

	/* Client request cache, maps a client request  */
	/* to the response that was sent to them.       */
	ClientRequestSequences map[string]ClientReply
}

type LogEntry struct {
	/* Index of log entry (first index = 1)         */
	Index uint64

	/* The term that this entry was in when added   */
	TermId uint64

	/* Command associated with this log entry in    */
	/* the user's finite-state-machine.             */
	Command FsmCommand

	/* Data associated with this log entry in the   */
	/* user's finite-state-machine.                 */
	Data []byte

	/* After processing this log entry, what ID to  */
	/* use when caching the response. Use an empty  */
	/* string to not cache at all                   */
	CacheId string
}

type FileData struct {
	/* Active file descriptor of to file */
	fd *os.File

	/* Size of file after reading it in and after writes */
	size int64

	/* Filename of file */
	filename string

	/* Map from LogEntry index to size of file before that index starts */
	idxMap map[uint64]int64

	/* Is the fd open or not? */
	open bool
}

func (r *RaftNode) initStableStore() (bool, error) {
	freshNode := false
	// Create log path directory if it doesn't already exist
	err := os.Mkdir(r.conf.LogPath, 0777)
	if err == nil {
		Out.Printf("Created log directory: %v\n", r.conf.LogPath)
	}
	if err != nil && !os.IsExist(err) {
		Error.Printf("error creating dir %v\n", err)
		return freshNode, err
	}

	r.logFileDescriptor = FileData{
		fd:       nil,
		size:     0,
		filename: fmt.Sprintf("%v/%d_raftlog.dat", r.conf.LogPath, r.listenPort),
	}
	r.metaFileDescriptor = FileData{
		fd:       nil,
		size:     0,
		filename: fmt.Sprintf("%v/%d_raftmeta.dat", r.conf.LogPath, r.listenPort),
	}
	raftLogSize, raftLogExists := getFileInfo(r.logFileDescriptor.filename)
	r.logFileDescriptor.size = raftLogSize

	raftMetaSize, raftMetaExists := getFileInfo(r.metaFileDescriptor.filename)
	r.metaFileDescriptor.size = raftMetaSize

	// Previous state exists, re-populate everything
	if raftLogExists && raftMetaExists {
		fmt.Printf("Reloading previous raftlog (%v) and raftmeta (%v)\n",
			r.logFileDescriptor.filename, r.metaFileDescriptor.filename)
		// Read in previous log and populate index mappings
		entries, err := ReadRaftLog(&r.logFileDescriptor)
		if err != nil {
			Error.Printf("Error reading in raft log: %v\n", err)
			return freshNode, err
		}
		r.logCache = entries

		// Create append-only file descriptor for later writing out of log entries.
		err = openRaftLogForWrite(&r.logFileDescriptor)
		if err != nil {
			Error.Printf("Error opening raftlog for write: %v\n", err)
			return freshNode, err
		}

		// Read in previous metalog and set cache
		ss, err := ReadStableState(&r.metaFileDescriptor)
		if err != nil {
			Error.Printf("Error reading stable state: %v\n", err)
			return freshNode, err
		}
		r.stableState = *ss

	} else if (!raftLogExists && raftMetaExists) || (raftLogExists && !raftMetaExists) {
		Error.Println("Both raftlog and raftmeta files must exist to proceed!")
		err = errors.New("Both raftlog and raftmeta files must exist to start this node")
		return freshNode, err

	} else {
		// We now assume neither file exists, so let's create new ones
		freshNode = true
		Out.Printf("Creating new raftlog and raftmeta files")
		err := CreateRaftLog(&r.logFileDescriptor)
		if err != nil {
			Error.Printf("Error creating new raftlog: %v\n", err)
			return freshNode, err
		}
		err = CreateStableState(&r.metaFileDescriptor)
		if err != nil {
			Error.Printf("Error creating new stable state: %v\n", err)
			return freshNode, err
		}

		// Init other nodes to zero, this will become populated
		r.stableState.OtherNodes = make([]NodeAddr, 0)

		// Init client request cache
		r.stableState.ClientRequestSequences = make(map[string]ClientReply)

		// No previous log cache exists, so a fresh one must be created.
		r.logCache = make([]LogEntry, 0)

		// If the log is empty we need to bootstrap it by adding the first committed entry.
		initEntry := LogEntry{
			Index:   0,
			TermId:  r.GetCurrentTerm(),
			Command: INIT,
			Data:    []byte{0},
		}
		r.appendLogEntry(initEntry)
		r.setCurrentTerm(0)
	}

	return freshNode, nil
}

/* Raft metadata setters/getters */
func (r *RaftNode) setCurrentTerm(newTerm uint64) {
	r.ssMutex.Lock()
	defer r.ssMutex.Unlock()
	if newTerm != r.stableState.CurrentTerm {
		Out.Printf("(%v) Setting current term from %v -> %v", r.Id, r.stableState.CurrentTerm, newTerm)
	}
	r.stableState.CurrentTerm = newTerm
	err := WriteStableState(&r.metaFileDescriptor, r.stableState)
	if err != nil {
		Error.Printf("Unable to flush new term to disk: %v\n", err)
		panic(err)
	}
}

func (r *RaftNode) GetCurrentTerm() uint64 {
	return r.stableState.CurrentTerm
}

func (r *RaftNode) setVotedFor(candidateId string) {
	r.ssMutex.Lock()
	defer r.ssMutex.Unlock()
	r.stableState.VotedFor = candidateId
	err := WriteStableState(&r.metaFileDescriptor, r.stableState)
	if err != nil {
		Error.Printf("Unable to flush new votedFor to disk: %v\n", err)
		panic(err)
	}
}

func (r *RaftNode) GetVotedFor() string {
	return r.stableState.VotedFor
}

func (r *RaftNode) setLocalAddr(localAddr *NodeAddr) {
	r.ssMutex.Lock()
	defer r.ssMutex.Unlock()
	r.stableState.LocalAddr = *localAddr
	err := WriteStableState(&r.metaFileDescriptor, r.stableState)
	if err != nil {
		Error.Printf("Unable to flush new localaddr to disk: %v\n", err)
		panic(err)
	}
}

func (r *RaftNode) GetLocalAddr() *NodeAddr {
	return &r.stableState.LocalAddr
}

func (r *RaftNode) GetOtherNodes() []NodeAddr {
	return r.stableState.OtherNodes
}

func (r *RaftNode) SetOtherNodes(nodes []NodeAddr) {
	r.ssMutex.Lock()
	defer r.ssMutex.Unlock()
	r.stableState.OtherNodes = nodes
	err := WriteStableState(&r.metaFileDescriptor, r.stableState)
	if err != nil {
		Error.Printf("Unable to flush new other nodes to disk: %v\n", err)
		panic(err)
	}
}

func (r *RaftNode) AppendOtherNodes(other NodeAddr) {
	r.ssMutex.Lock()
	defer r.ssMutex.Unlock()
	r.stableState.OtherNodes = append(r.stableState.OtherNodes, other)
	err := WriteStableState(&r.metaFileDescriptor, r.stableState)
	if err != nil {
		Error.Printf("Unable to flush new other nodes to disk: %v\n", err)
		panic(err)
	}
}

func (r *RaftNode) CheckRequestCache(clientReq ClientRequest) (*ClientReply, bool) {
	uniqueId := fmt.Sprintf("%v-%v", clientReq.ClientId, clientReq.SequenceNum)
	val, ok := r.stableState.ClientRequestSequences[uniqueId]
	if ok {
		return &val, ok
	} else {
		return nil, ok
	}
}

func (r *RaftNode) AddRequest(uniqueId string, reply ClientReply) error {
	r.ssMutex.Lock()
	defer r.ssMutex.Unlock()
	_, ok := r.stableState.ClientRequestSequences[uniqueId]
	if ok {
		return errors.New("Request with the same clientId and seqNum already exists!")
	}
	r.stableState.ClientRequestSequences[uniqueId] = reply

	err := WriteStableState(&r.metaFileDescriptor, r.stableState)
	if err != nil {
		Error.Printf("Unable to flush new client request to disk: %v\n", err)
		panic(err)
	}

	return nil
}

/* Raft log setters/getters */
func (r *RaftNode) getLogEntry(index uint64) *LogEntry {
	if index < uint64(len(r.logCache)) {
		return &r.logCache[index]
	} else {
		return nil
	}
}

func (r *RaftNode) getLastLogEntry() *LogEntry {
	return r.getLogEntry(r.getLastLogIndex())
}

func (r *RaftNode) getLogEntries(start, end uint64) []LogEntry {
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
	return uint64(len(r.logCache) - 1)
}

func (r *RaftNode) getLastLogTerm() uint64 {
	return r.getLogEntry(r.getLastLogIndex()).TermId
}

func (r *RaftNode) getLogTerm(index uint64) uint64 {
	return r.getLogEntry(index).TermId
}

func (r *RaftNode) appendLogEntry(entry LogEntry) error {
	// write entry to disk
	err := AppendLogEntry(&r.logFileDescriptor, &entry)
	if err != nil {
		return err
	}
	// update entry in cache
	r.logCache = append(r.logCache, entry)
	return nil
}

// Truncate file to remove everything at index and after it (an inclusive truncation!)
func (r *RaftNode) truncateLog(index uint64) error {
	err := TruncateLog(&r.logFileDescriptor, index)
	if err != nil {
		return err
	}

	// Truncate cache as well
	r.logCache = r.logCache[:index]
	return nil
}

func CreateFileData(filename string) FileData {
	fileData := FileData{}
	fileData.filename = filename
	return fileData
}

func (r *RaftNode) RemoveLogs() error {
	r.logFileDescriptor.fd.Close()
	r.logFileDescriptor.open = false
	err := os.Remove(r.logFileDescriptor.filename)
	if err != nil {
		r.Error("Unable to remove raftlog file")
		return err
	}

	r.metaFileDescriptor.fd.Close()
	r.metaFileDescriptor.open = false
	err = os.Remove(r.metaFileDescriptor.filename)
	if err != nil {
		r.Error("Unable to remove raftmeta file")
		return err
	}

	return nil
}