package main

import "./raft"

type Shell struct {
	r *raft.RaftNode
	c *raft.Client
	done chan bool
}
