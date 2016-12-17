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

	logFileName  := fmt.Sprint("%v/%d_raft_log.dat", r.conf.LogPath, r.port)
	metaFileName := fmt.Sprint("%v/%d_raft_meta.dat", r.conf.LogPath, r.port)

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


	}

}
