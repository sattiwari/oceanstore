# Chord

Chord is a <a href="https://en.wikipedia.org/wiki/Distributed_hash_table">distributed hash table (DHT)</a> protocol. Its design is motivated from <a href="https://pdos.csail.mit.edu/papers/chord:sigcomm01/chord_sigcomm.pdf">Chord: A Scalable Peer-to-peer Lookup Service for Internet Applications</a> paper. Chord distributes objects over a dynamic network of nodes, and implements a protocol for finding these objects once they have been placed in the network.

# Usage Example
[cli](cli.go) serves as a console for interacting with chord, creating nodes and querying state on the local nodes. It provides the following commands:
* <b>node</b> display node ID, successor, and predecessor
* <b>table</b> display finger table information for node(s)
* <b>addr</b> display node listener address(es)
* <b>data</b> display datastore(s) for node(s)
* <b>get</b> get value from Chord ring associated with this key
* <b>put</b> put key/value into Chord ring
* <b>quit</b> quit node(s)

# Keys
The hashed value of the key takes the form of an m-bit unsigned integer. Thus, the keyspace for the DHT resides between 0 and 2<sup>m</sup> - 1, inclusive. Current implementation uses SHA-1 for hashing.

# The Ring
Each node in the system also has a hash value (hash of its name - ip address + port). Chord orders the nodes in a circular fashion, in which each node’s successor is the node with the next highest hash.

# Overlay Network (Finger Table)
To locate the node at which a particular key-value pair is stored, we need to find the successor to the hash value of the key. Linear search would be very slow on large network of nodes so chord uses an overlay network. It maintains a finger table at each node. The number of entries in the finger table is equal to m, where m is the number of bits representing a hash in the keyspace of the DHT (e.g., 128). Entry i in the table, with 0 <= i < m, is the node which the owner of the table believes is the successor for the hash h + 2i (h is the current node’s hash). 

# Lookup in Chord
When node A services a request to find the successor of the key k, it first determines whether its own successor is the owner of k (the successor is simply entry 0 in the finger table). If it is, then A returns its successor in response to the request. Otherwise, node A finds node B in its finger table such that B has the largest hash smaller than the hash of k, and forwards the request to B.

# Dynamics
Chord supports the dynamic addition and removal of nodes from the network. Each node calls [stabilize](chord/node_local_impl.go#L32) and [fixNextFinger](chord/finger.go#L34) functions periodically to determine the successor and predecessor relationship between nodes as they are added to the network.

# Future Work
Support fault tolerance by maintaining a list of successors. This would need the keys to be replicated across a number of nodes. 
