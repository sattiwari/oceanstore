package raft

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
)

/*                                                                  */
/* Main functions to assist with interacting with log entries, etc. */
/*                                                                  */

func openRaftLogForWrite(fileData *FileData) error {
	if fileExists(fileData.filename) {
		fd, err := os.OpenFile(fileData.filename, os.O_APPEND|os.O_WRONLY, 0600)
		fileData.fd = fd
		fileData.open = true
		return err
	} else {
		return errors.New("Raftfile does not exist")
	}
}

func CreateRaftLog(fileData *FileData) error {
	fd, err := os.OpenFile(fileData.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	fileData.fd = fd
	fileData.size = int64(0)
	fileData.idxMap = make(map[uint64]int64)
	fileData.open = true
	return err
}

func ReadRaftLog(fileData *FileData) ([]LogEntry, error) {
	f, err := os.Open(fileData.filename)
	defer f.Close()
	fileData.idxMap = make(map[uint64]int64)

	entries := make([]LogEntry, 0)

	fileLocation := int64(0)
	for err != io.EOF {
		size, err := readStructSize(f)
		if err != nil {
			if err == io.EOF {
				break
			}
			Error.Printf("Error reading struct size: %v at loc: %v\n", err, fileLocation)
			fileData.open = false
			return entries, err
		}

		entry, err := readLogEntry(f, size)
		if err != nil {
			Error.Printf("Error reading log entry: %v at loc: %v\n", err, fileLocation)
			fileData.open = false
			return entries, err
		}
		fileData.idxMap[entry.Index] = fileLocation
		fileLocation += INT_GOB_SIZE + int64(size)
		entries = append(entries, *entry)
	}

	fileData.open = false
	return entries, nil
}

func AppendLogEntry(fileData *FileData, entry *LogEntry) error {
	sizeIdx := fileData.size

	logBytes, err := getLogEntryBytes(entry)
	if err != nil {
		return err
	}
	size, err := getSizeBytes(len(logBytes))
	if err != nil {
		return err
	}

	numBytesWritten, err := fileData.fd.Write(size)
	if err != nil {
		return err
	}
	if int64(numBytesWritten) != INT_GOB_SIZE {
		panic("int gob size is not correct, cannot proceed")
	}
	fileData.size += int64(numBytesWritten)

	err = fileData.fd.Sync()
	if err != nil {
		return err
	}

	numBytesWritten, err = fileData.fd.Write(logBytes)
	if err != nil {
		return err
	}
	if numBytesWritten != len(logBytes) {
		panic("did not write correct amount of bytes for some reason for log entry")
	}
	fileData.size += int64(numBytesWritten)

	err = fileData.fd.Sync()
	if err != nil {
		return err
	}

	// Update index mapping for this entry
	fileData.idxMap[entry.Index] = int64(sizeIdx)

	return nil
}

func TruncateLog(raftLogFd *FileData, index uint64) error {
	newFileSize, exist := raftLogFd.idxMap[index]
	if !exist {
		return fmt.Errorf("Truncation failed, log index %v doesn't exist\n", index)
	}

	// Windows does not allow truncation of open file, must close first
	raftLogFd.fd.Close()
	err := os.Truncate(raftLogFd.filename, newFileSize)
	if err != nil {
		return err
	}
	fd, err := os.OpenFile(raftLogFd.filename, os.O_APPEND|os.O_WRONLY, 0600)
	raftLogFd.fd = fd

	for i := index; i < uint64(len(raftLogFd.idxMap)); i++ {
		delete(raftLogFd.idxMap, i)
	}
	raftLogFd.size = newFileSize
	return nil
}

/*                                                                           */
/* Main functions to assist with interacting with stable state entries, etc. */
/*                                                                           */
func openStableStateForWrite(fileData *FileData) error {
	if fileExists(fileData.filename) {
		fd, err := os.OpenFile(fileData.filename, os.O_APPEND|os.O_WRONLY, 0600)
		fileData.fd = fd
		fileData.open = true
		return err
	} else {
		return errors.New("Stable state file does not exist")
	}
}

func CreateStableState(fileData *FileData) error {
	fd, err := os.OpenFile(fileData.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	fileData.fd = fd
	fileData.open = true
	return err
}

func ReadStableState(fileData *FileData) (*NodeStableState, error) {
	f, err := os.Open(fileData.filename)

	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	ss, err := readStableStateEntry(f, int(stat.Size()))
	f.Close()

	if err != nil {
		// for some reason we failed to read our stable state file, try backup file.
		backupFilename := fmt.Sprintf("%v.bak", fileData.filename)
		fbak, err := os.Open(backupFilename)

		stat, err := f.Stat()
		if err != nil {
			fbak.Close()
			return nil, err
		}

		ss, err := readStableStateEntry(f, int(stat.Size()))
		if err != nil {
			Error.Printf("we were unable to read stable storage or its backup: %v\n", err)
			fbak.Close()
			return nil, err
		}
		fbak.Close()

		// we were successful reading from backup, move to live copy
		err = os.Remove(fileData.filename)
		if err != nil {
			return nil, err
		}
		err = copyFile(backupFilename, fileData.filename)
		if err != nil {
			return nil, err
		}

		return ss, nil
	}

	return ss, nil
}

func WriteStableState(fileData *FileData, ss NodeStableState) error {
	// backup old stable state
	backupFilename := fmt.Sprintf("%v.bak", fileData.filename)
	err := backupStableState(fileData, backupFilename)
	if err != nil {
		return fmt.Errorf("Backup failed: %v", err)
	}

	// Windows does not allow truncation of open file, must close first
	fileData.fd.Close()

	// truncate live stable state
	err = os.Truncate(fileData.filename, 0)
	if err != nil {
		return fmt.Errorf("Truncation failed: %v", err)
	}
	fd, err := os.OpenFile(fileData.filename, os.O_APPEND|os.O_WRONLY, 0600)
	fileData.fd = fd

	// write out stable state to live version
	bytes, err := getStableStateBytes(ss)
	if err != nil {
		return err
	}

	numBytes, err := fileData.fd.Write(bytes)
	if numBytes != len(bytes) {
		panic("did not write correct amount of bytes for some reason for ss")
	}

	err = fileData.fd.Sync()
	if err != nil {
		return fmt.Errorf("Sync #2 failed: %v", err)
	}

	// remove backup file
	err = os.Remove(backupFilename)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Remove failed: %v", err)
	}

	return nil
}

func backupStableState(fileData *FileData, backupFilename string) error {
	if fileData.open && fileData.fd != nil {
		err := fileData.fd.Close()
		fileData.open = false
		if err != nil {
			return fmt.Errorf("Closing file failed: %v", err)
		}
	}

	err := os.Remove(backupFilename)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Remove failed: %v", err)
	}

	err = copyFile(fileData.filename, backupFilename)
	if err != nil {
		return fmt.Errorf("File copy failed: %v", err)
	}

	err = openStableStateForWrite(fileData)
	if err != nil {
		return fmt.Errorf("Opening stable state for writing failed: %v", err)
	}

	return nil
}

func copyFile(srcFile string, dstFile string) error {
	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}

	dst, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	err = src.Close()
	if err != nil {
		fmt.Errorf("Error closing src file")
		return err
	}

	err = dst.Close()
	if err != nil {
		fmt.Errorf("Error closing dst file")
		return err
	}
	return nil
}

/*                                                                */
/* Helper functions to assist with read/writing log entries, etc. */
/*                                                                */

const INT_GOB_SIZE int64 = 5

func getStableStateBytes(ss NodeStableState) ([]byte, error) {
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	err := e.Encode(ss)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func getSizeBytes(size int) ([]byte, error) {
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	err := e.Encode(size)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func getLogEntryBytes(entry *LogEntry) ([]byte, error) {
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	err := e.Encode(*entry)
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
	b := make([]byte, size)
	leSize, err := f.Read(b)
	if err != nil {
		return nil, err
	}
	if leSize != size {
		panic("The stable state log may be corrupt, cannot proceed")
	}

	buff := bytes.NewBuffer(b)
	var ss NodeStableState
	dataDecoder := gob.NewDecoder(buff)
	err = dataDecoder.Decode(&ss)
	if err != nil {
		return nil, err
	}

	return &ss, nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		panic(err)
	}
}

func getFileInfo(filename string) (int64, bool) {
	stat, err := os.Stat(filename)
	if err == nil {
		return stat.Size(), true
	} else if os.IsNotExist(err) {
		return 0, false
	} else {
		panic(err)
	}
}