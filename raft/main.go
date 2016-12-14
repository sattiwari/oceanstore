package main

import (
	"./raft"
	"fmt"
)

func main() {
	id := raft.AddrToId("abc", 2)
	fmt.Println(id)
}