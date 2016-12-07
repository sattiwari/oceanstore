package tapestry

import (
	 "math/rand"
	"math/big"
	"bytes"
	"fmt"
)

/*
	An ID is just a typedef'ed digit array
*/
type ID [DIGITS]Digit

/*
	A digit is just a typedef'ed uint8
*/
type Digit uint8

/*
	Generates a random ID
*/
func RandomID() ID {
	var id ID
	for i := range id {
		id[i] = Digit(rand.Intn(BASE))
	}
	return id
}

/*
	Returns the length of the prefix that is shared by the two IDs
*/
func SharedPrefixLength(a ID, b ID) (i int) {
	count := 0
	for i := 0; i < DIGITS && a[i] == b[i]; i++ {
		count++
	}
	return count
}

/*
	Used by Tapestry's surrogate routing.  Given IDs first and second, which is the better choice?
	The "better choice" is the ID that:
	- has the longest shared prefix with id
	- if both have prefix of length n, which id has a better (n+1)th digit?
	- if both have the same (n+1)th digit, consider (n+2)th digit, etc.
	Returns true if the first ID is the better choice.  Returns false if second ID is closer or if first==second
*/
func (id ID) BetterChoice(first ID, second ID) bool {
	fPrefix := SharedPrefixLength(first, id)
	sPrefix := SharedPrefixLength(second, id)
	if fPrefix != sPrefix || (sPrefix == DIGITS && fPrefix == DIGITS) {
		//If they are not the same or if they are the same in all the numbers then we return
		return fPrefix > sPrefix
	}
	//So they are the same, but not as long as DIGITS we need to figure out which one is better
	index := sPrefix
	start := id[index]
	target := id[index]
	madeAChoice := false
	for !madeAChoice {
		//If it stays in -1 then the digit of the first is > the digit in ID
		fDigit := first[index] % BASE
		sDigit := second[index] % BASE
		fDistance := 0
		sDistance := 0
		for sDigit != target {
			sDistance++
			target++
			target = target % BASE
		}
		target = start
		for fDigit != target {
			fDistance++
			target++
			target = target % BASE
		}


		if fDistance == sDistance {
			if index == DIGITS-1 {
				return false
			} else {
				index++
				target = id[index]
				start = id[index]
			}
		} else {
			//fmt.Printf("fDistance: %v, sDistance: %v, target: %v, fDigit: %v, sDigit: %v\n", fDistance, sDistance, target, fDigit, sDigit)
			return fDistance < sDistance
		}
	}
	return false
}

/*
	Used when inserting nodes into Tapestry's routing table.  If the routing table has multiple candidate nodes for a slot,
	then it chooses the node that is closer to the local node.
	In a production Tapestry implementation, closeness is determined by looking at the round-trip-times (RTTs) between (a, id) and (b, id),
	and the node with the shorter RTT is closer.
	In my implementation, I have decided to define closeness as the absolute value of the difference between a and b.
	This is NOT the same as	the implementation of BetterChoice.
	Returns true if a is closer than b.  Returns false if b is closer than a, or if a == b.
*/
func (id ID) Closer(first ID, second ID) bool {

	firstNum := first.big()
	secondNum := second.big()
	idNum := id.big()

	difF := big.NewInt(0)
	difS := big.NewInt(0)

	difF.Sub(firstNum, idNum)
	difS.Sub(secondNum, idNum)
	difF.Abs(difF)
	difS.Abs(difS)

	if difF.Cmp(difS) == -1 {
		return true
	} else {
		return false
	}
}

/*
	Helper function: convert an ID to a big int.
*/
func (id ID) big() (b *big.Int) {
	b = big.NewInt(0)
	base := big.NewInt(BASE)
	for _, digit := range id {
		b.Mul(b, base)
		b.Add(b, big.NewInt(int64(digit)))
	}
	return b
}

/*
	String representation of an ID is hexstring of each digit
*/
func (id ID) String() string {
	var buf bytes.Buffer
	for _, d := range id {
		buf.WriteString(d.String())
	}
	return buf.String()
}

/*
	String representation of a digit is its hex value
*/
func (digit Digit) String() string {
	return fmt.Sprintf("%X", byte(digit))
}