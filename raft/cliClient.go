package main

import "./raft"

func clientInit(shell *Shell, args []string) error {
	return shell.c.SendRequest(raft.HASH_CHAIN_INIT, []byte(args[1]))
}

func clientHash(shell *Shell, args []string) error {
	return shell.c.SendRequest(raft.HASH_CHAIN_ADD, []byte{})
}