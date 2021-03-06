package tapestry

import "fmt"

/*
	This is a utility class tacked on to the tapestry DOLR.
*/
type BlobStore struct {
	blobs map[string]Blob
}

type Blob struct {
	bytes []byte
	done  chan bool
}

type BlobStoreRPC struct {
	store *BlobStore
}

/*
	Create a new blobstore
*/
func NewBlobStore() *BlobStore {
	bs := new(BlobStore)
	bs.blobs = make(map[string]Blob)
	return bs
}

/*
	For RPC server registration
*/
func NewBlobStoreRPC(store *BlobStore) *BlobStoreRPC {
	rpc := new(BlobStoreRPC)
	rpc.store = store
	return rpc
}

/*
   Remove all blobs and unregister them all
*/
func (bs *BlobStore) DeleteAll() {
	// unregister every blob
	for _, blob := range bs.blobs {
		blob.done <- true
	}
	// clear the map
	bs.blobs = make(map[string]Blob)
}

/*
	Remove the blob and unregister it
*/
func (bs *BlobStore) Delete(key string) bool {
	// If a previous blob exists, unregister it
	previous, exists := bs.blobs[key]
	if exists {
		previous.done <- true
	}
	delete(bs.blobs, key)
	return exists
}

/*
	Store bytes in the blobstore
*/
func (bs *BlobStore) Put(key string, blob []byte, unregister chan bool) {
	// If a previous blob exists, delete it
	bs.Delete(key)

	// Register the new one
	bs.blobs[key] = Blob{blob, unregister}
}

/*
	Get bytes from the blobstore
*/
func (bs *BlobStore) Get(key string) ([]byte, bool) {
	blob, exists := bs.blobs[key]
	if exists {
		return blob.bytes, true
	} else {
		return nil, false
	}
}

/*
	Fetches the specified blob from the remote node
*/
func FetchRemoteBlob(remote Node, key string) (blob *[]byte, err error) {
	fmt.Printf("FetchRemoteBlob %v %v", key, remote)
	err = makeRemoteCall(remote.Address, "BlobStoreRPC", "Fetch", key, &blob)
	return
}

/*
   Invoked over RPC to fetch bytes from the blobstore
*/
func (rpc *BlobStoreRPC) Fetch(key string, blob *[]byte) error {
	b, exists := rpc.store.blobs[key]
	if exists {
		*blob = b.bytes
	}
	return nil
}