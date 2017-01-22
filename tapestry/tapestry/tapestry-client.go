package tapestry

import "fmt"

// Get data from a tapestry node.  Looks up key then fetches directly
func TapestryGet(remote Node, key string) ([]byte, error) {
	fmt.Printf("Making remote TapestryGet call\n")
	// Lookup the key
	replicas, err := TapestryLookup(remote, key)
	if err != nil {
		return nil, err
	}
	if len(replicas) == 0 {
		return nil, fmt.Errorf("No replicas returned for key %v", key)
	}

	// Contact replicas
	var errs []error
	for _, replica := range replicas {
		blob, err := FetchRemoteBlob(replica, key)
		if err != nil {
			errs = append(errs, err)
		}
		if blob != nil {
			return *blob, nil
		}
	}

	return nil, fmt.Errorf("Error contacting replicas, %v: %v", replicas, errs)
}

