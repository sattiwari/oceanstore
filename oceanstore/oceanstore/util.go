package oceanstore

func removeExcessSlashes(path string) string {
	var firstNonSlash, lastNonSlash, start int

	onlySlashes := true
	str := path

	length := len(path)

	// Nothing to do
	if path[0] != '/' && path[length-1] != '/' {
		return str
	}

	// Get the first non slash
	for i := 0; i < length; i++ {
		if str[i] != '/' {
			onlySlashes = false
			firstNonSlash = i
			break
		}
	}

	// Get the last non slash
	for i := length - 1; i >= 0; i-- {
		if str[i] != '/' {
			lastNonSlash = i
			break
		}
	}

	// Guaranteed to be the root path
	if onlySlashes {
		str = "/"
		return str
	} else {
		length = lastNonSlash - firstNonSlash + 1
		if str[0] == '/' {
			start = firstNonSlash - 1
			length++
		} else {
			start = 0
		}

		str = path[start : start+length]
	}

	length = len(str)
	for i := 0; i < length; i++ {
		if i+1 == length {
			break
		}

		if str[i] == '/' && str[i+1] == '/' {
			str = str[:i] + str[i+1:]
			length -= 1
			i -= 1
		}
	}

	return str
}