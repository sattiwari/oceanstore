package chord

import (
	"testing"
	"strconv"
	"math/rand"
)

func TestRemotePutAndGetBundleRandom(t *testing.T) {
	nNodes := 10
	numRange := 100
	base := make(map[int]int64, numRange)
	result := make(map[int]int64, numRange)
	nodes, _ := CreateNNodesRandom(nNodes)

	for i := 0; i < numRange; i++ {
		base[i] = int64(i * i)
		//Now we randomly pick a node and put the value in it
		nodeIndex := rand.Intn(9)
		Put(nodes[nodeIndex], strconv.Itoa(i), strconv.Itoa(i*i))
	}

	for i := 0; i < numRange; i++ {
		nodeIndex := rand.Intn(9)
		val, _ := Get(nodes[nodeIndex], strconv.Itoa(i))
		result[i], _ = strconv.ParseInt(val, 10, 32)
	}

	equal := true
	for i := 0; i < numRange; i++ {
		if result[i] != base[i] {
			equal = false
		}
	}
	if !equal {
		t.Errorf("TestRemotePutAndGetBundleRandom: result")
	}
}
