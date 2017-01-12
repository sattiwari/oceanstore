# Oceanstore

Oceanstore is a distributed file system. Its design is motivated from <a href="http://www.cs.cornell.edu/~hweather/Oceanstore/asplos00.pdf">OceanStore: An Architecture for Global-Scale Persistent Storage</a> paper. 

# Usage Example
WIP

# File System - Abstractions
The two primitive file system objects are <b>files</b> and <b>directories</b>. A file is a single collection of sequential bytes and and directories provide a way to hierarchically organize files. 
* <b>Data block</b> represents fixed length array of bytes. Files consist of a number of data blocks.
* <b>Indirect block</b> stores references to ordered list of data blocks that make up a file.
* <b>Inode</b> maintains the metadata associated with a file. Inode of a file points to direct blocks or indirect blocks. Inode of a directory points to inode of files or other directories.
