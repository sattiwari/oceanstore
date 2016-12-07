package tapestry

func equal_ids(id1, id2 ID) bool {
	if SharedPrefixLength(id1, id2) == DIGITS {
		return true
	}
	return false
}
