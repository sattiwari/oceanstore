package chord

import (
	"testing"
)

func TestSimple(t *testing.T) {
	_, err := CreateNode(nil)
	if err != nil {
		t.Errorf("Unable to create node, received error:%v\n", err)
	}
}
