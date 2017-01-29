package oceanstore

import (
	"fmt"
	"strings"
	"../../tapestry/tapestry"
)

func (ocean *OceanNode) mkdir(req *MkdirRequest) (MkdirReply, error) {
	fmt.Println("Entered mkdir")
	reply := MkdirReply{}

	path := req.Path
	length := len(path)
	clientId := req.ClientId

	if path[0] != '/' {
		path = ocean.getCurrentDir(clientId) + "/" + path
	}
	path = removeExcessSlashes(path)

	if length == 0 {
		return reply, fmt.Errorf("Empty path")
	}
	if (length > 2 && path[length-1] == '.' && path[length-2] == '.') ||
		path[length-1] == '.' {
		return reply, fmt.Errorf("There already exists a file/dir with that name.")
	}

	dirInode, name, fullPath, dirPath, err := ocean.dir_namev(path, clientId)
	if err != nil {
		fmt.Println(err)
		return reply, err
	}

	// File we are about to make should not exist.
	_, err = ocean.getInode(fullPath, clientId)
	if err == nil {
		return reply, fmt.Errorf("There already exists a file/dir with that name.")
	}

	// This is the root node creation.
	if dirInode == nil {

		// Create the root Inode and its block
		newDirInode := CreateDirInode(name)
		newDirBlock := CreateBlock()

		// Set block paths for the indirect block and dot references
		blockPath := fmt.Sprintf("%v:%v", fullPath, "indirect") // this will be '/:indirect'

		// Hash the dot references to put them on the indirect block.
		blockHash := tapestry.Hash(blockPath)

		// Save the root Inode indirect block in tapestry
		ocean.storeIndirectBlock(fullPath, newDirBlock.bytes, clientId)

		newDirInode.indirect = hashToGuid(blockHash)
		fmt.Println(blockHash, "->", newDirInode.indirect)

		// Save the root Inode
		ocean.storeInode(fullPath, newDirInode, clientId)

	} else {
		// Get indirect block from the directory that is going to create
		// the node
		dirBlock, err := ocean.getInodeBlock(dirPath, clientId)
		if err != nil {
			fmt.Println(err)
			return reply, err
		}

		// Create new inode and block
		newDirInode := CreateDirInode(name)
		newDirBlock := CreateBlock()

		// Declare block paths
		blockPath := fmt.Sprintf("%v:%v", fullPath, "indirect")

		// Get hashes
		newDirInodeHash := tapestry.Hash(fullPath)

		fmt.Println("Dirpath: %v", dirPath)
		fmt.Println("Fullpath: %v", fullPath)
		fmt.Println("blockPath: %v", blockPath)
		fmt.Println("newDirInodeHAsh: %v", newDirInodeHash)

		// Write the new dir to the old dir and increase its size
		IdIntoByte(dirBlock, &newDirInodeHash, int(dirInode.size))
		dirInode.size += tapestry.DIGITS

		bytes := make([]byte, tapestry.DIGITS)
		IdIntoByte(bytes, &newDirInodeHash, 0)
		newDirInode.indirect = Guid(ByteIntoAguid(bytes, 0))
		fmt.Println("\n\n\n\n\n\n", newDirInodeHash, "->", newDirInode.indirect)

		// Save both blocks in tapestry
		ocean.storeIndirectBlock(fullPath, newDirBlock.bytes, clientId)
		ocean.storeIndirectBlock(dirPath, dirBlock, clientId)

		// Encode both inodes
		ocean.storeInode(dirPath, dirInode, clientId)
		ocean.storeInode(fullPath, newDirInode, clientId)
	}

	reply.Ok = true
	return reply, nil
}

func (ocean *OceanNode) dir_namev(pathname string, id uint64) (*Inode, string, string, string, error) {

	path := removeExcessSlashes(pathname)
	lastSlash := strings.LastIndex(path, "/")
	var dirPath, name string

	fmt.Println("Last slash:", lastSlash)

	if lastSlash == 0 && len(path) != 1 {
		return ocean.getRootInode(id), pathname[1:], pathname, "/", nil
	} else if lastSlash == 0 {
		return nil, "/", "/", "", nil
	} else if lastSlash != -1 && len(path) != 1 { // K. all good
		dirPath = path[:lastSlash]
		name = path[lastSlash+1:]
	} else if lastSlash == -1 { // No slashes at all (relative path probably)
		dirPath = ocean.getCurrentDir(id)
		name = path
	} else {
		panic("What should go here?")
	}

	path = removeExcessSlashes(path)

	if dirPath[0] != '/' {
		dirPath = ocean.getCurrentDir(id) + "/" + dirPath
	}

	dirInode, err := ocean.getInode(dirPath, id)
	if err != nil { // Dir path does not exist
		fmt.Println(err)
		return nil, "", "", "", err
	}

	dirPath = removeExcessSlashes(dirPath)
	fullPath := removeExcessSlashes(dirPath + "/" + name)

	return dirInode, name, fullPath, dirPath, nil
}

func (ocean *OceanNode) getRootInode(id uint64) *Inode {
	inode, err := ocean.getInode("/", id)
	if err != nil {
		panic("Root inode not found!")
	}
	return inode
}