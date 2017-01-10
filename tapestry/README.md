# Tapestry

Tapestry is a distributed object location are retrieval (DOLR) system. Its design is motivated from <a href="http://www.srhea.net/papers/tapestry_jsac.pdf">Tapestry: A Resilient Global-Scale Overlay for
Service Deployment</a> paper. It is an overlay network that implements simple key based routing.

# Identifying Nodes and Objects
Nodes and objects in the Tapestry network are each assigned a sequence of n base-16 digits globally unique identifier. 

# Root Nodes
In order to make it possible for any node in the network to find the location of an object, a single node is appointed as the “root” node for that object. A root node is the one which shares the same hash value as the object.

# Surrogate Nodes
There to be fewer nodes in the network than possible values in the space of hash values so a “surrogate” node for an object is chosen to be the one with a hash value that shares as many prefix digits in the object’s hash value as possible.

# Selecting the Surrogate Node
Starting at the leftmost digit d, we take the set of nodes that have d as the leftmost digit of their hashes as well. If no such set of nodes exists, it is necessary to deterministically choose another set. To do this, we can try to find a set of nodes that share the digit d + 1 as their leftmost hash digit. Until a non-empty set of nodes is found, the value of the digit we are searching with increases (modulo the base of the hash-value). Once the set has been found the same logic can be applied for the next digit in the hash, choosing from the set of nodes we identified with the previous digit. When this algorithm has been applied for every digit, only one node will be left and that node is the surrogate.

# Routing Tables
In order to allow nodes to locate objects stored at other nodes, each node maintains a routing table that stores references to a subset of the nodes in the network.

# Backpointer Tables
Backpointers are references to every node in the network which refers to the local node in their own routing tables. These are useful in maintaining routing tables in a dynamic network. When the local node adds or removes a remote node from its routing table, it notifies the remote node, who will then update its backpointer table.

# Prefix Routing
A node that matches some number of digits from the object’s hash may be chosen from the routing table. In turn, the selected node’s routing table is inspected and the next node in the route to the surrogate is chosen. At each successive node in the route, the number of digits that match the object’s hash value increases until the last digit has been matched and the surrogate node has been reached. This type of routing is called “prefix routing”. [findRoot](tapestry/tapestry-local.go#L94) has this logic.

# Publishing and Retrieving Objects
When an object is “published” by a node, that node routes towards the root node for the key, then registers itself on that node as a location of the key. Multiple nodes can publish the same object. A tapestry client wishing to lookup the object will first route to the root node of the object. The root node then informs the client of which Tapestry nodes are the ones that have published the object. The client then directly contacts one or more of those publishers to retrieve the actual object data.

# Adding Tapestry Nodes
The new node is assigned its ID and then routes towards the root node for that id. The root node initiates the transfer of all keys that should now be stored on the new node. The new node then iteratively traverses backpointers, starting from the root node, to populate its own routing table.

# Acknowledged Multicast
If the new node has a shared prefix of length n with its root, then any other node that also has a shared prefix of length n is called <b><i><u>need-to-know node</u></i></b>. The root node performs an <i><u>acknowledged multicast</u></i> when it is contacted by the new node. The multicast eventually returns the full set of need-to-know nodes from the Tapestry. The multicast is a recursive call — the root node contacts all nodes on levels >= n of its routing table; those nodes contact all nodes on levels >= n + 1 of their routing tables; and so on. A node that is contacted during the multicast will initiate background transfer of relevant object references to the new node, trigger multicast to the next level of its routing table, then merge and return the resulting lists of nodes (removing duplicates). [AddNodeMulticast](tapestry/tapestry-local.go#L321) has this logic.

# Backpointer Traversal
Once the multicast has completed, the root node returns the list of need-to-know nodes to the new node. The new node uses this list as an initial neighbor set to populate its routing table. The node iteratively contacts the nodes, asking for their backpointers.

# Graceful Exit

# Fault Tolerance
* Errors while routing
* Loss of Root Node
* Loss of replicas
* Miscellaneous
