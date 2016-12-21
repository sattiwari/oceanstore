package raft

import (
	"os"
	"bytes"
	"encoding/gob"
)

// functions to assist interaction with Log entries

func openRaftLogForWrite(fileData *FileData) error {

}

func CreateRaftLog(fileData *FileData) error {
	fd, err := os.OpenFile(fileData.fileName, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0600)
	fileData.fileDescriptor = fd
	fileData.sizeOfFile = uint64(0)
	fileData.logEntryIdxToFileSizeMap = make(map[uint64]uint64)
	fileData.isFileDescriptorOpen = true
	return err
}


func ReadRaftLog(fileData *FileData) error  {
	
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

func TruncateLog(raftLogFd *FileData, index uint64) error {

}

// functions to assist interaction with stable state entries

func openStableStateForWrite(fileData *FileData) error {

}

func CreateStateState(FileData *FileData) error {
	fd, err := os.OpenFile(FileData.fileName, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0600)
	FileData.fileDescriptor = fd
	FileData.isFileDescriptorOpen = true
	return err
}

func ReadStableState(fileData *FileData) error {

}

func WriteStableState(fileData *FileData, ss NodeStableState) error {

}

func backupStableState(fileData *FileData, backupFileName string) error {

}

func copyFile(src string, des string) error {

}

// helper functions to assist read / write log entries

const INT_GOB_SIZE uint64 = 5

func getStableStateBytes(ss NodeStableState) ([]byte, error) {

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

}

func readLogentry(f *os.File, size int) (*LogEntry, error) {

}

func readStableStateEntry(f *os.File, size int) (*NodeStableState, error) {

}

func fileExists(fileName bool) {

}

func getFileInfo(filename string) (int64, bool) {

}