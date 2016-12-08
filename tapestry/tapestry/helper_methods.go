package tapestry

import "testing"

var port int

func equal_ids(id1, id2 ID) bool {
	if SharedPrefixLength(id1, id2) == DIGITS {
		return true
	}
	return false
}

func makeTapestryNode(id ID, addr string, t *testing.T) *TapestryNode {

	tapestry, err := start(id, port, addr)

	if err != nil {
		t.Errorf("Error while making a tapestry %v", err)
	}

	port++
	return tapestry.local
}