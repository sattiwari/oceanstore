package raft

import (
	"os"
	"bytes"
	"encoding/gob"
	"errors"
	"io"
	"fmt"
)

// functions to assist interaction with Log entries

func openRaftLogForWrite(fileData *FileData) error {
	if fileExists(fileData.fileName) {
		fd, err := os.OpenFile(fileData.fileName, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0600)
		fileData.fileDescriptor = fd
		fileData.isFileDescriptorOpen = true
		return err
	} else {
		return errors.New("file does not exist")
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

func ReadRaftLog(fileData *FileData) ([]LogEntry, error) {
	f, err := os.Open(fileData.fileName)
	defer f.Close()
	fileData.logEntryIdxToFileSizeMap = make(map[uint64]int64)

	entries := make([]LogEntry, 0)

	fileLocation := int64(0)
	for err != io.EOF {
		size, err := readStructSize(f)
		if err != nil {
			if err == io.EOF {
				break
			}
			Error.Printf("Error reading struct size: %v at loc: %v\n", err, fileLocation)
			fileData.isFileDescriptorOpen = false
			return entries, err
		}

		entry, err := readLogEntry(f, size)
		if err != nil {
			Error.Printf("Error reading log entry: %v at loc: %v\n", err, fileLocation)
			fileData.isFileDescriptorOpen = false
			return entries, err
		}
		fileData.logEntryIdxToFileSizeMap[entry.Index] = fileLocation
		fileLocation += INT_GOB_SIZE + int64(size)
		entries = append(entries, *entry)
	}

	fileData.isFileDescriptorOpen = false
	return entries, nil
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

func TruncateLog(logFd *FileData, index uint64) error {
	fileSize, exist := logFd.logEntryIdxToFileSizeMap(index)
	if !exist {
		return fmt.Errorf("log entry does not exist")
	}
	logFd.fileDescriptor.Close()
	err := os.Truncate(logFd.fileName, fileSize)
	if err != nil {
		return nil
	}
	fd, err := os.OpenFile(logFd.fileName, os.O_APPEND | os.O_WRONLY, 0600)
	logFd.fileDescriptor = fd
	for i := index; i < uint64(len(logFd.logEntryIdxToFileSizeMap)); i++ {
		delete(logFd.logEntryIdxToFileSizeMap, i)
	}
	logFd.sizeOfFile = fileSize
	return nil
}

// functions to assist interaction with stable state entries

func openStableStateForWrite(fileData *FileData) error {
	if fileExists(fileData) {
		fd, err := os.OpenFile(fileData.fileName, os.O_APPEND | os.O_WRONLY, 0600)
		fileData.fileDescriptor = fd
		fileData.isFileDescriptorOpen = true
		return err
	} else {
		return errors.New("stable state file does not exist")
	}
}

func CreateStableState(fileData *FileData) error {
	fd, err := os.OpenFile(fileData.fileName, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0600)
	fileData.fileDescriptor = fd
	fileData.isFileDescriptorOpen = true
	return err
}


func ReadStableState(fileData *FileData) (NodeStableState, error) {
	return nil, nil
}

func WriteStableState(fileData *FileData, ss NodeStableState) error {
	return nil
}

func backupStableState(fileData *FileData, backupFileName string) error {
	return nil
}

func copyFile(src string, des string) error {
	return nil
}

// helper functions to assist read / write log entries

const INT_GOB_SIZE uint64 = 5

func getStableStateBytes(ss NodeStableState) ([]byte, error) {
	return nil, nil
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

func readStructSize(f *os.File) (int, error) {
	// Read bytes for size value
	b := make([]byte, INT_GOB_SIZE)
	sizeBytes, err := f.Read(b)
	if err != nil {
		return -1, err
	}
	if int64(sizeBytes) != INT_GOB_SIZE {
		panic("The raftlog may be corrupt, cannot proceed")
	}

	// Decode bytes as int, which is sizeof(LogEntry).
	buff := bytes.NewBuffer(b)
	var size int
	dataDecoder := gob.NewDecoder(buff)
	err = dataDecoder.Decode(&size)
	if err != nil {
		return -1, err
	}

	return size, nil
}

func readLogEntry(f *os.File, size int) (*LogEntry, error) {
	b := make([]byte, size)
	leSize, err := f.Read(b)
	if err != nil {
		return nil, err
	}
	if leSize != size {
		panic("The raftlog may be corrupt, cannot proceed")
	}

	buff := bytes.NewBuffer(b)
	var entry LogEntry
	dataDecoder := gob.NewDecoder(buff)
	err = dataDecoder.Decode(&entry)
	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func readStableStateEntry(f *os.File, size int) (*NodeStableState, error) {
	return nil, nil
}

func fileExists(fileName string) bool {
	_, exists := getFileInfo(fileName)
	return exists
}

func getFileInfo(fileName string) (int64, bool) {
	size, err := os.Stat(fileName)
	if err == nil {
		return size, true
	} else if os.IsNotExist(err) {
		return 0, false
	} else {
		panic(err)
	}
}