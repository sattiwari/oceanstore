package raft

import (
	"fmt"
	"time"
)

type Config struct {
	ElectionTimeout    time.Duration
	HeartbeatFrequency time.Duration
	ClusterSize        int
	NodeIdSize         int
	LogPath            string
}

func DefaultConfig() *Config {
	config := new(Config)
	config.ClusterSize = 3
	config.ElectionTimeout = time.Millisecond * 150
	config.HeartbeatFrequency = time.Millisecond * 50
	config.NodeIdSize = 2
	config.LogPath = "raftlogs"
	return config
}

func CheckConfig(config *Config) error {
	if config.ElectionTimeout < config.HeartbeatFrequency {
		return fmt.Errorf("The election timeout (%v) is less than the heartbeat frequency (%v)", config.ElectionTimeout, config.HeartbeatFrequency)
	}
	return nil
}
