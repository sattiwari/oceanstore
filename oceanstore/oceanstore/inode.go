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
func (puddle *PuddleNode) getInodeFromAguid(aguid Aguid, id uint64) (*Inode, error) {
	// Get the vguid using raft
	bytes, err := puddle.getTapestryData(aguid, id)

	inode := new(Inode)
	err = inode.GobDecode(bytes)
	if err != nil {
		fmt.Println(bytes)
		return nil, err
	}

	return inode, nil
}