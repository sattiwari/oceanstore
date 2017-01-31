package oceanstore

import (
	"testing"
	"time"
)

func TestGobEncoding(t *testing.T) {
	inode := new(Inode)
	inode.name = "Test inode"
	inode.filetype = 1
	inode.size = 666
	inode.indirect = "F666"

	bytes, err := inode.GobEncode()
	if err != nil {
		t.Errorf("Gob encode didn't work.")
	}

	sameInode := new(Inode)
	sameInode.GobDecode(bytes)

	if inode.name != sameInode.name {
		t.Errorf("Name not the same\n\t%v != %v.", inode.name, sameInode.name)
	}

	if inode.filetype != sameInode.filetype {
		t.Errorf("Name not the same\n\t%v != %v.", inode.filetype, sameInode.filetype)
	}

	if inode.size != sameInode.size {
		t.Errorf("Name not the same\n\t%v != %v.", inode.size, sameInode.size)
	}

	if inode.indirect != sameInode.indirect {
		t.Errorf("Name not the same\n\t%v != %v.", inode.indirect, sameInode.indirect)
	}
}

func TestInodeStorage(t *testing.T) {
	ocean, err := Start()
	if err != nil {
		return
		t.Errorf("Could not init oceanstore: %v", err)
	}
	time.Sleep(time.Millisecond * 500)

	client := ocean.raftClient

	inode := new(Inode)
	inode.name = "Test inode"
	inode.filetype = 1
	inode.size = 666
	inode.indirect = "F666"

	inode2 := new(Inode)
	inode2.name = "Test inode2"
	inode2.filetype = 0
	inode2.size = 66
	inode2.indirect = "BEEF"


	err = ocean.storeInode("/path/one", inode, client.Id)
	if err != nil {
		t.Errorf("Error storing Inode: %v", err)
		return
	}
	err = ocean.storeInode("/second/path", inode2, client.Id)
	if err != nil {
		t.Errorf("Error storing Inode2: %v", err)
		return
	}

	sameInode, err := ocean.getInode("/path/one", client.Id)
	if err != nil {
		t.Errorf("Error geting Inode: %v", err)
		return
	}
	sameInode2, err := ocean.getInode("/second/path", client.Id)
	if err != nil {
		t.Errorf("Error geting Inode2: %v", err)
		return
	}

	if inode.name != sameInode.name {
		t.Errorf("Name not the same\n\t%v != %v.", inode.name, sameInode.name)
	}
	if inode.filetype != sameInode.filetype {
		t.Errorf("Name not the same\n\t%v != %v.", inode.filetype, sameInode.filetype)
	}
	if inode.size != sameInode.size {
		t.Errorf("Name not the same\n\t%v != %v.", inode.size, sameInode.size)
	}
	if inode.indirect != sameInode.indirect {
		t.Errorf("Name not the same\n\t%v != %v.", inode.indirect, sameInode.indirect)
	}

	if inode2.name != sameInode2.name {
		t.Errorf("Name not the same\n\t%v != %v.", inode2.name, sameInode2.name)
	}
	if inode2.filetype != sameInode2.filetype {
		t.Errorf("Name not the same\n\t%v != %v.", inode2.filetype, sameInode2.filetype)
	}
	if inode2.size != sameInode2.size {
		t.Errorf("Name not the same\n\t%v != %v.", inode2.size, sameInode2.size)
	}
	if inode2.indirect != sameInode2.indirect {
		t.Errorf("Name not the same\n\t%v != %v.", inode2.indirect, sameInode2.indirect)
	}
}

func TestInodeReplacement(t *testing.T) {
	puddle, err := Start()
	if err != nil {
		return
		t.Errorf("Could not init puddlestore: %v", err)
	}
	time.Sleep(time.Millisecond * 500)
	client := puddle.raftClient

	inode := new(Inode)
	inode.name = "Test inode"
	inode.filetype = 1
	inode.size = 666
	inode.indirect = "F666"

	err = puddle.storeInode("/path/one", inode, client.Id)
	if err != nil {
		t.Errorf("Error storing Inode: %v", err)
		return
	}

	/*
		err = puddle.removeKey("/path/one")
		if err != nil {
			t.Errorf("Error removing key \"/path/one\": %v", err)
			return
		}*/

	inode2 := new(Inode)
	inode2.name = "Imma replace u beaaach"
	inode2.filetype = 1
	inode2.size = 50
	inode2.indirect = "DEAD"

	err = puddle.storeInode("/path/one", inode2, client.Id)
	if err != nil {
		t.Errorf("Error storing Inode: %v", err)
		return
	}

	sameInode2, err := puddle.getInode("/path/one", client.Id)
	if err != nil {
		t.Errorf("Error geting Inode: %v", err)
		return
	}
	if sameInode2 == nil {
		t.Errorf("Something went wrong man")
		return
	}

	if inode2.name != sameInode2.name {
		t.Errorf("Name not the same\n\t%v != %v.", inode2.name, sameInode2.name)
	}
	if inode2.filetype != sameInode2.filetype {
		t.Errorf("Name not the same\n\t%v != %v.", inode2.filetype, sameInode2.filetype)
	}
	if inode2.size != sameInode2.size {
		t.Errorf("Name not the same\n\t%v != %v.", inode2.size, sameInode2.size)
	}
	if inode2.indirect != sameInode2.indirect {
		t.Errorf("Name not the same\n\t%v != %v.", inode2.indirect, sameInode2.indirect)
	}
}