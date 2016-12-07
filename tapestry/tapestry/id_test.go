package tapestry

import "testing"

// This test makes sure that the prefix length is working
func TestSharedPrefixLength(t *testing.T) {
	a := ID{1,2,3,4,5,6,7,8,9,6,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	b := ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	count := SharedPrefixLength(a, b)
	if (count != 9) {
		t.Errorf("The SharedPrefixLength does not work")
	}
	a = ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	b = ID{2,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	count = SharedPrefixLength(a, b)
	if (count != 0) {
		t.Errorf("The SharedPrefixLength does not work")
	}
	a = ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	b = ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	count = SharedPrefixLength(a, b)
	if (count != 40) {
		t.Errorf("The SharedPrefixLength does not work")
	}
}


//This function tests several types of ID and makes sure that the output is the expected one.*/
func TestBetterChoice(t *testing.T) {
	a := ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	b := ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	id := ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	choice := id.BetterChoice(a, b)
	if (choice) {//choice should be false since they are the same
		t.Errorf("The BetterChoice does not work")
	}
	a = ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	b = ID{1,2,3,4,5,6,8,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	id = ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	choice = id.BetterChoice(a, b)
	if (!choice) {//choice should be true for the prefix
		t.Errorf("The BetterChoice does not work")
	}
	a = ID{1,2,3,4,5,6,7,6,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	b = ID{1,2,3,4,5,6,7,7,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	id =ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	choice = id.BetterChoice(a, b)
	if (!choice) {//choice should be true becuase we get to 6 from 8 faster than to 7
		t.Errorf("The BetterChoice does not work", choice, a, b)
	}
	a = ID{1,2,3,4,5,6,7,8,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	b = ID{1,2,3,4,5,6,7,7,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	id =ID{1,2,3,4,5,6,7,6,9,5,11,12,13,14,15,0,2,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	choice = id.BetterChoice(a, b)
	if (choice) {//choice should be false because it is closer to get to b (7) than 8
		t.Errorf("The BetterChoice does not work", choice, a, b)
	}
	a = ID{1,2,3,4,5,6,7,6,9,5,11,12,1,2,3,0,4,    4,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	b = ID{1,2,3,4,5,6,7,6,9,5,11,12,1,2,3,0,4,    5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	id = ID{1,2,3,4,5,6,7,6,9,5,10,12,13,13,15,0,2,2,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	choice = id.BetterChoice(a, b)
	if (!choice) {//choice should be true because it is faster to get to a (4) from 2 than to b(5)
		t.Errorf("The BetterChoice does not work", choice, a, b)
	}
	a = ID{13,2,3,4,5,6,7,6,9,5,11,12,1,2,3,0,4,4,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	b = ID{7,2,3,4,5,6,7,6,9,5,11,12,1,2,3,0,4,5,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	id = ID{1,2,3,4,5,6,7,6,9,5,10,12,13,13,15,0,2,2,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	choice = id.BetterChoice(a, b)
	if (choice) {//choice should be false at the very beginning, 7 (b)is closer to 1 than 13(a)
		t.Errorf("The BetterChoice does not work", choice, a, b)
	}
}

//test for the Closer function.
func TestCloser(t *testing.T) {

	a := ID{1,2,3,4,5,6,7,6,9,5,11,12,1,2,3,0,4,4,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	id := ID{1,2,3,4,5,6,7,6,9,5,10,12,13,13,15,0,2,2,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	b := ID{1,2,3,4,5,6,7,6,9,5,11,12,1,2,3,0,4,4,3,0,2,12,15,13,15,13,2,5,10,11,13,2,8,9,12,13,0,9,8,5}
	choice := id.Closer(a, b)
	if (choice) {//Answer should be false because they are the same ids
		t.Errorf("The Closer does not work", choice, a, b)
	}
	a = ID{1,2,3,4,5,6,7,6,9,5,11,12,  1,2,3,0,  4,4,3,0,2,12,15,13,15,13,2,5,10,11,13,  13,8,9,12,13,0,9,8,5}
	id = ID{1,2,3,4,5,6,7,6,9,5,10,12,13,13,15,0,2,2,3,0,2,12,15,13,15,13,2,5,10,11,11,   2,8,9,12,13,0,9,8,5}
	b = ID{1,2,3,4,5,6,7,6,9,5,11,12,  1,2,3,0,  4,4,3,0,2,12,15,13,15,13,2,5,10,11,13,  10,8,9,12,13,0,9,8,5}
	choice = id.Closer(a, b)
	if (choice) {//Answer should be false because b is closer in absolute value
		t.Errorf("The Closer does not work", choice, a, b)
	}
	a = ID{1,2,3,4,5,6,7,6,9,5,11,12,1,2,3,    0,4,4,3,0,2,12,15,13,15,13,2,5,10,11,13, 15,8,9,12,13,0,9,8,5}
	id = ID{1,2,3,4,5,6,7,6,9,5,10,12,13,13,15,0,2,2,3,0,2,12,15,13,15,13,2,5,10,11,13, 13,8,9,12,13,0,9,8,5}
	b = ID{1,2,3,4,5,6,7,6,9,5,11,12,1,2,3,    0,4,4,3,0,2,12,15,13,15,13,2,5,10,11,13, 12,8,9,12,13,0,9,8,5}
	choice = id.Closer(a, b)
	if (choice) {//Answer should be false because b is closer in absolute value
		t.Errorf("The Closer does not work", choice, a, b)
	}
	//some more obvious
	a = ID{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,2}
	id = ID{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}
	b = ID{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1}
	choice = id.Closer(a, b)
	if (choice) {//Answer should be false because b is closer in absolute value
		t.Errorf("The Closer does not work", choice, a, b)
	}
	//This one is tricky because it goes to the other digit
	a = ID {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,13}
	id = ID{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,15}
	b = ID {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0}
	choice = id.Closer(a, b)
	if (choice) {//Answer should be false because of the base change
		t.Errorf("The Closer does not work", choice, a, b)
	}
	a = ID{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,2}
	id = ID{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,5}
	b = ID{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1}
	choice = id.Closer(a, b)
	if (choice) {//Answer should be false because a has a 1
		t.Errorf("The Closer does not work", choice, a, b)
	}
	a = ID {1,5,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,2}
	id = ID{0,0,9,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,5}
	b = ID {1,2,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1}
	choice = id.Closer(a, b)
	if (choice) {//Answer should be false (b)
		t.Errorf("The Closer does not work", choice, a, b)
	}
	a = ID {1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,2}
	id = ID{0,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9}
	b = ID {0,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,9,0}
	choice = id.Closer(a, b)
	if (choice) {//Answer should be b
		t.Errorf("The Closer does not work", choice, a, b)
	}
}