package chord

import (
	"math"
	"testing"
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