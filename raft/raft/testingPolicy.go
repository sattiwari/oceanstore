package raft

import (
	"fmt"
	"errors"
)

var ErrorTestingPolicyDenied = errors.New("testing policy has denied this communication")

type TestingPolicy struct {
	PauseWorld bool
	RpcPolicy  map[string]bool
}

func NewTesting() *TestingPolicy {
	var tp TestingPolicy
	tp.RpcPolicy = make(map[string]bool)
	return &tp
}

func (tp *TestingPolicy) IsDenied(a, b NodeAddr) bool {
	if tp.PauseWorld {
		return true
	}
	commStr := getCommId(a, b)
	denied, exists := tp.RpcPolicy[commStr]
	return exists && denied
}

func getCommId(a, b NodeAddr) string {
	if a.Id < b.Id {
		fmt.Sprintf("%v_%v", a.Id, b.Id)
	} else {
		fmt.Sprintf("%v_%v", b.Id, a.Id)
	}
}