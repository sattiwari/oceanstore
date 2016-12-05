/*  Purpose: Finger table related functions for a given Chord node.          */

package chord

import (
	"time"
)

/* A single finger table entry */
type FingerEntry struct {
	Start []byte       /* ID hash of (n + 2^i) mod (2^m)  */
	Node  *RemoteNode  /* RemoteNode that Start points to */
}

/* Create initial finger table that only points to itself, will be fixed later */
func (node *Node) initFingerTable() {
	//TODO implement this method
}

/* Called periodically (in a seperate go routine) to fix entries in our finger table. */
func (node *Node) fixNextFinger(ticker *time.Ticker) {
	for _ = range ticker.C {
		//TODO implement this method
	}
}
