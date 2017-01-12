# Oceanstore

Oceanstore is a distributed file system. Its design is motivated from <a href="http://www.cs.cornell.edu/~hweather/Oceanstore/asplos00.pdf">OceanStore: An Architecture for Global-Scale Persistent Storage</a> paper. 

# Usage Example
WIP

# File System - Abstractions
The two primitive file system objects are <b>files</b> and <b>directories</b>. A file is a single collection of sequential bytes and directories provide a way to hierarchically organize files. 
* <b>Data block</b> represents fixed length array of bytes. Files consist of a number of data blocks.
* <b>Indirect block</b> stores references to ordered list of data blocks that make up a file.
* <b>Inode</b> maintains the metadata associated with a file. Inode of a file points to direct blocks or indirect blocks. Inode of a directory points to inode of files or other directories.

# File System Operations

## Lookup
Find the inode of the root. Traverse the directories/files in its indirect block to find the first directory/file in the path. Repeat the search until we reach end of path.

## Reading and Writing
To write or read from file, we need:
* <b>Location</b> This tells the starting location in the file for reading or writing.
* <b>Buffer</b> While reading, contents of the file are put in buffer and while writing, contents of the buffer are put into the file. <br/>

If we pass the end of the file while writing then we need to add new data blocks and add their refernces in the indirect block. We also choose the block size. Given the starting location and number of bytes to read or write, it is easy to find the relevant blocks with start position in the first block and end position in the last block.




