package oceanstore

import (
	"testing"
	"strings"
	"time"
	"fmt"
)

func TestRaftMap(t *testing.T) {
	ocean, err := Start()
	if err != nil {
		t.Errorf("Could not init puddlestore: %v", err)
		return
	}
	time.Sleep(time.Millisecond * 1000)

	fmt.Println(ocean.Local)
	client := ocean.raftClient

	err = ocean.setRaftVguid("DEAD", "BEEF", client.Id)
	if err != nil {
		t.Errorf("Could set raft vguid: %v", err)
		return
	}

	response, err := ocean.getRaftVguid("DEAD", client.Id)
	if err != nil {
		t.Errorf("Could get raft vguid: %v", err)
		return
	}

	ok := strings.Split(string(response), ":")[0]
	aguid := strings.Split(string(response), ":")[1]

	if ok != "SUCCESS" {
		t.Errorf("Could not get raft vguid: %v", response)
	}

	if aguid != "BEEF" {
		t.Errorf("Raft didn't return the correct vguid. BEEF != %d", aguid)
	}

	// Reset aguid to another vguid
	err = ocean.setRaftVguid("DEAD", "B004", client.Id)
	if err != nil {
		t.Errorf("Could set raft vguid: %v", err)
		return
	}

	response, err = ocean.getRaftVguid("DEAD", client.Id)
	if err != nil {
		t.Errorf("Could get raft vguid: %v", err)
		return
	}

	ok = strings.Split(string(response), ":")[0]
	aguid = strings.Split(string(response), ":")[1]

	if ok != "SUCCESS" {
		t.Errorf("Could not get raft vguid: %v", response)
	}

	if aguid != "B004" {
		t.Errorf("Raft didn't return the correct vguid. B004 != %d", aguid)
	}
}