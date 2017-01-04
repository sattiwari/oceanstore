package oceanstore

import (
	"testing"
)

func TestRaftMap(t *testing.T) {
	_, err := Start()
	if err != nil {
		t.Errorf("Could not init puddlestore: %v", err)
		return
	}
}