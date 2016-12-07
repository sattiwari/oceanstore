package tapestry

import (
	 "math/rand"
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
