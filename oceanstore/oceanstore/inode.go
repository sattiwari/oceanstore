package oceanstore

import (
	"../../tapestry/tapestry"
	"fmt"
	"bytes"
	"encoding/gob"
	"strings"
	"strconv"
)

type Filetype int

const (
	DIR Filetype = iota
	FILE
)

const BLOCK_SIZE = uint32(4096)
const FILES_PER_INODE = 4

type Inode struct {
	name     string
	filetype Filetype
	size     uint32
	indirect Guid
}

type Block struct {
	bytes []byte
}

func CreateDirInode(name string) *Inode {
	inode := new(Inode)
	inode.name = name
	inode.filetype = DIR
	inode.size = 0
	inode.indirect = ""
	return inode
}

func CreateFileInode(name string) *Inode {
	inode := new(Inode)
	inode.name = name
	inode.filetype = FILE
	inode.size = 0
	inode.indirect = ""
	return inode
}

func CreateBlock() *Block {
	block := new(Block)
	block.bytes = make([]byte, BLOCK_SIZE)
	return block
}

// Gets the inode that has a given path
func (ocean *OceanNode) getInode(path string, id uint64) (*Inode, error) {

	hash := tapestry.Hash(path)

	aguid := Aguid(hashToGuid(hash))

	// Get the vguid using raft
	bytes, err := ocean.getTapestryData(aguid, id)

	inode := new(Inode)
	err = inode.GobDecode(bytes)
	if err != nil {
		fmt.Println(bytes)
		return nil, err
	}

	return inode, nil
}

func (d *Inode) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(d.name)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(d.filetype)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(d.size)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(d.indirect)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (d *Inode) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&d.name)
	if err != nil {
		return err
	}
	err = decoder.Decode(&d.filetype)
	if err != nil {
		return err
	}
	err = decoder.Decode(&d.size)
	if err != nil {
		return err
	}
	return decoder.Decode(&d.indirect)
}

// Generic method. Gets data given an aguid.
func (ocean *OceanNode) getTapestryData(aguid Aguid, id uint64) ([]byte, error) {
	tapestryNode := ocean.getRandomTapestryNode()
	response, err := ocean.getRaftVguid(aguid, id)
	if err != nil {
		return nil, err
	}

	ok := strings.Split(string(response), ":")[0]
	vguid := strings.Split(string(response), ":")[1]
	if ok != "SUCCESS" {
		return nil, fmt.Errorf("Could not get raft vguid: %v", response)
	}

	data, err := tapestry.TapestryGet(tapestryNode, string(vguid))
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Gets the inode that has a given aguid
func (ocean *OceanNode) getInodeFromAguid(aguid Aguid, id uint64) (*Inode, error) {
	// Get the vguid using raft
	bytes, err := ocean.getTapestryData(aguid, id)

	inode := new(Inode)
	err = inode.GobDecode(bytes)
	if err != nil {
		fmt.Println(bytes)
		return nil, err
	}

	return inode, nil
}

func (ocean *OceanNode) getFileBlock(key string, blockno uint32, id uint64) ([]byte, error) {
	blockPath := fmt.Sprintf("%v:%v", key, blockno)
	hash := tapestry.Hash(blockPath)
	aguid := Aguid(hashToGuid(hash))

	return ocean.getTapestryData(aguid, id)
}

// Gets the block of the inode of the specified key/path
func (ocean *OceanNode) getInodeBlock(key string, id uint64) ([]byte, error) {
	blockPath := fmt.Sprintf("%v:%v", key, "indirect")
	hash := tapestry.Hash(blockPath)
	aguid := Aguid(hashToGuid(hash))

	return ocean.getTapestryData(aguid, id)
}


// Stores inode as data
func (ocean *OceanNode) storeInode(path string, inode *Inode, id uint64) error {

	hash := tapestry.Hash(path)

	aguid := Aguid(hashToGuid(hash))
	vguid := Vguid(randSeq(tapestry.DIGITS))

	// Encode the inode
	bytes, err := inode.GobEncode()
	if err != nil {
		return err
	}

	// Set the new aguid -> vguid pair with raft
	err = ocean.setRaftVguid(aguid, vguid, id)
	if err != nil {
		return err
	}

	// Store data in tapestry with key: vguid
	err = tapestry.TapestryStore(ocean.getRandomTapestryNode(), string(vguid), bytes)
	if err != nil {
		return err
	}

	return nil
}

func (ocean *OceanNode) storeIndirectBlock(inodePath string, block []byte,
id uint64) error {

	blockPath := fmt.Sprintf("%v:%v", inodePath, "indirect")
	hash := tapestry.Hash(blockPath)

	aguid := Aguid(hashToGuid(hash))
	vguid := Vguid(randSeq(tapestry.DIGITS))

	// Set the new aguid -> vguid pair with raft
	err := ocean.setRaftVguid(aguid, vguid, id)
	if err != nil {
		return err
	}

	err = tapestry.TapestryStore(ocean.getRandomTapestryNode(), string(vguid), block)
	if err != nil {
		return fmt.Errorf("Tapestry error")
	}

	return nil
}

func (ocean *OceanNode) storeFileBlock(inodePath string, blockno uint32,
block []byte, id uint64) error {

	blockPath := fmt.Sprintf("%v:%v", inodePath, blockno)
	hash := tapestry.Hash(blockPath)

	aguid := Aguid(hashToGuid(hash))
	vguid := Vguid(randSeq(tapestry.DIGITS))

	// Set the new aguid -> vguid pair with raft
	err := ocean.setRaftVguid(aguid, vguid, id)
	if err != nil {
		return err
	}

	err = tapestry.TapestryStore(ocean.getRandomTapestryNode(), string(vguid), block)
	if err != nil {
		return err
	}

	return nil
}

// Removes an entry from a directory block. If it not the last entry,
// It moves and replaces the last entry with the removing entry.
func (ocean *OceanNode) removeEntryFromBlock(bytes []byte, vguid Vguid,
size uint32, id uint64) error {

	start, err := ocean.lookupInode(bytes, vguid, size, id)
	if err != nil {
		return err
	}
	if start == size-tapestry.DIGITS { // Last one
		// MakeZeros(bytes, start)
	} else {
		for i := uint32(0); i < tapestry.DIGITS; i++ {
			bytes[start+i] = bytes[size-tapestry.DIGITS+i]
		}
	}
	return nil
}

// Get the inode that has a specific vuid from a directory block.
func (puddle *OceanNode) lookupInode(block []byte, vguid Vguid,
size uint32, id uint64) (uint32, error) {
	length := size / tapestry.DIGITS
	for i := uint32(0); i < length; i++ {
		curAguid := ByteIntoAguid(block, i*tapestry.DIGITS)
		res, err := puddle.getRaftVguid(curAguid, id)
		curVguid := Vguid(strings.Split(string(res), ":")[1])
		if err != nil {
			return 0, err
		}
		if curVguid == vguid {
			fmt.Println("Found:", curAguid, curVguid)
			return i, nil
		}
	}

	return 0, fmt.Errorf("Not found!")
}

func ByteIntoAguid(bytes []byte, start uint32) Aguid {
	aguid := ""
	for i := uint32(0); i < tapestry.DIGITS; i++ {
		aguid += strconv.FormatUint(uint64(bytes[start+i]), tapestry.BASE)
	}
	return Aguid(strings.ToUpper(aguid))
}