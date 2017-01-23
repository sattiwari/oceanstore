package oceanstore

import (
	"../../tapestry/tapestry"
	"fmt"
	"bytes"
	"encoding/gob"
	"strings"
)

type Filetype int

type Inode struct {
	name     string
	filetype Filetype
	size     uint32
	indirect Guid
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