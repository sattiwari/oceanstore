package raft

import (
	"fmt"
	"errors"
)

var ErrorTestingPolicyDenied = errors.New("testing policy has denied this communication")

type TestingPolicy struct {
	pauseWorld bool
	rpcPolicy  map[string]bool
}

func NewTesting() *TestingPolicy {
	var tp TestingPolicy
	tp.rpcPolicy = make(map[string]bool)
	return &tp
}

func (tp *TestingPolicy) IsDenied(a, b NodeAddr) bool {
	if tp.PauseWorld {
		return true
	}
	commStr := getCommId(a, b)
	denied, exists := tp.rpcPolicy[commStr]
	return exists && denied
}

func getCommId(a, b NodeAddr) string {
	return fmt.Sprintf("%v_%v", a.Id, b.Id)
}

func (tp *TestingPolicy) RegisterPolicy(a, b NodeAddr, allowed bool) {
	commStr := getCommId(a, b)
	tp.rpcPolicy[commStr] = allowed
}

func (tp *TestingPolicy) PauseWorld(on bool) {
	tp.pauseWorld = on
}