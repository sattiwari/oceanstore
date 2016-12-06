package chord

import (
	"math"
	"testing"
	"time"
	"fmt"
)

func TestInitFingerTable(t *testing.T) {
	var res, expected []byte
	m := int(math.Pow(2, KEY_LENGTH))
	for i := 0; i < m; i++ {
		node, _ := CreateDefinedNode(nil, []byte{byte(i)})
		for j := 0; j < KEY_LENGTH; j++ {
			res = node.FingerTable[j].Start
			expected = []byte{byte((i + int(math.Pow(float64(2), float64(j)))) % m)}
			if !EqualIds(res, expected) {
				t.Errorf("[%v] BAD ENTRY %v: %v != %v", i, j, res, expected)
			}
		}
	}
}

/*
Makes 26 nodes, waits a few seconds, and checks that every entry points to its next multiple of 10 from "Start"
(ex: Start = 178 would should always point to 180)
 */
func TestFixNextFinger(t *testing.T) {
	nodes, _ := CreateNNodes(26)
	time.Sleep(time.Second * 5)
	for i := 0; i < 26; i++ {
		node := nodes[i]
		for j := 0; j < KEY_LENGTH; j++ {
			start := node.FingerTable[j].Start
			pointer := node.FingerTable[j].Node
			var expected []byte
			if start[0]%10 == 0 {
				expected = []byte{byte(start[0])}
			} else {
				expected = []byte{byte(((start[0]/10 + 1) % 26) * 10)}
			}

			if !EqualIds(pointer.Id, expected) {
				fmt.Printf("[%v] Error at\nStart: %v, Node: %v, expected: %v",
					node.Id, start, pointer.Id, expected)
			}
		}
	}
}