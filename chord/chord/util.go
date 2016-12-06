/*  Purpose: Utility functions to help with dealing with ID hashes in Chord. */

package chord

import (
	"crypto/sha1"
	"math/big"
	"bytes"
)

/* Hash a string to its appropriate size */
func HashKey(key string) []byte {
	h := sha1.New()
	h.Write([]byte(key))
	v := h.Sum(nil)
	return v[:KEY_LENGTH/8]
}

/* Convert a []byte to a big.Int string, useful for debugging/logging */
func HashStr(keyHash []byte) string {
	keyInt := big.Int{}
	keyInt.SetBytes(keyHash)
	return keyInt.String()
}

//On the Chord ring, X is between (A : B)
func Between(nodeX, nodeA, nodeB []byte) bool {

	xInt := big.Int{}
	xInt.SetBytes(nodeX)

	aInt := big.Int{}
	aInt.SetBytes(nodeA)

	bInt := big.Int{}
	bInt.SetBytes(nodeB)

	var result bool
	if aInt.Cmp(&bInt) == 0 {
		result = false
	} else if aInt.Cmp(&bInt) < 0 {
		result = (xInt.Cmp(&aInt) == 1 && xInt.Cmp(&bInt) == -1)
	} else {
		result = !(xInt.Cmp(&bInt) == 1 && xInt.Cmp(&aInt) == -1)
	}

	return result
}

func EqualIds(a, b []byte) bool {
	return bytes.Equal(a, b)
}

func CreateNNodes(n int) ([]*Node, error) {
	if n == 0 {
		return nil, nil
	}
	nodes := make([]*Node, n)

	id := []byte{byte(0)}
	curr, err := CreateDefinedNode(nil, id)
	nodes[0] = curr
	if err != nil {
		return nil, err
	}

	for i := 1; i < n; i++ {
		id := []byte{byte(i * 10)}
		curr, err := CreateDefinedNode(nodes[0].RemoteSelf, id)
		nodes[i] = curr
		if err != nil {
			return nil, err
		}
	}

	return nodes, nil
}

/* Is X between (A : B] */
func BetweenRightIncl(nodeX, nodeA, nodeB []byte) bool {

	xInt := big.Int{}
	xInt.SetBytes(nodeX)

	aInt := big.Int{}
	aInt.SetBytes(nodeA)

	bInt := big.Int{}
	bInt.SetBytes(nodeB)

	var result bool
	if aInt.Cmp(&bInt) == 0 {
		result = true
	} else if aInt.Cmp(&bInt) < 0 {
		result = (xInt.Cmp(&aInt) == 1 && xInt.Cmp(&bInt) <= 0)
	} else {
		result = !(xInt.Cmp(&bInt) == 1 && xInt.Cmp(&aInt) <= 0)
	}

	return result
}