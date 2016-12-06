/*  Purpose: Finger table related functions for a given Chord node.          */

package chord

import (
	"time"
	"math/big"
	"log"
)

/* A single finger table entry */
type FingerEntry struct {
	Start []byte       /* ID hash of (n + 2^i) mod (2^m)  */
	Node  *RemoteNode  /* RemoteNode that Start points to */
}

/* Create initial finger table that only points to itself, will be fixed later */
func (node *Node) initFingerTable() {
	// Create an array of FingerEntries of length KEY_LENGTH
	node.FingerTable = make([]FingerEntry, KEY_LENGTH)

	for i := range node.FingerTable {
		// FingerEntry pointing to node
		newEntry := new(FingerEntry)
		newEntry.Start = fingerMath(node.Id, i, KEY_LENGTH)
		newEntry.Node = node.RemoteSelf
		node.FingerTable[i] = *newEntry
	}
	node.Successor = node.RemoteSelf
}

/* Called periodically (in a seperate go routine) to fix entries in our finger table. */
func (node *Node) fixNextFinger(ticker *time.Ticker) {
	for _ = range ticker.C {
		for _ = range ticker.C {
			next_hash := fingerMath(node.Id, node.next, KEY_LENGTH)
			successor, err := node.findSuccessor(next_hash)
			if err != nil {
				log.Fatal(err)
			}
			node.ftLock.Lock()
			node.FingerTable[node.next].Node = successor
			node.ftLock.Unlock()
			node.next += 1
			if node.next >= KEY_LENGTH {
				node.next = 1
			}
		}
	}
}

/* (n + 2^i) mod (2^m) */
func fingerMath(n []byte, i int, m int) []byte {
	two := &big.Int{}
	two.SetInt64(2)

	N := &big.Int{}
	N.SetBytes(n)

	// 2^i
	I := &big.Int{}
	I.SetInt64(int64(i))
	I.Exp(two, I, nil)

	// 2^m
	M := &big.Int{}
	M.SetInt64(int64(m))
	M.Exp(two, M, nil)

	result := &big.Int{}
	result.Add(N, I)
	result.Mod(result, M)

	// Big int gives an empty array if value is 0.
	// Here is a way for us to still return a 0 byte
	zero := &big.Int{}
	zero.SetInt64(0)
	if result.Cmp(zero) == 0 {
		return []byte{0}
	}

	return result.Bytes()
}
