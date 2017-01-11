# Tapestry

Tapestry is a distributed object location are retrieval (DOLR) system. Its design is motivated from <a href="http://www.srhea.net/papers/tapestry_jsac.pdf">Tapestry: A Resilient Global-Scale Overlay for
Service Deployment</a> paper. It is an overlay network that implements simple key based routing.

# Usage Example
[cli](cli.go) serves as a console for interacting with chord, creating nodes and querying state on the local nodes. It provides the following commands:
table- Print this node’s routing table
* <b>backpointers</b> Print this node’s backpointers
* <b>objects</b> Print the object replicas stored on this node
* <b>put</b> Stores the provided key-value pair on the local node and advertises the key to the tapestry
* <b>lookup</b> Looks up the specified key in the tapestry and prints its location
* <b>get</b> Looks up the specified key in the tapestry, then fetches the value from one of the returned replicas
* <b>remove</b> Remove the value stored locally for the provided key and stops advertising the key to the tapestry
* <b>list</b> List the keys currently being advertised by the local node 
* <b>leave</b> Instructs the local node to gracefully leave the Tapestry
* <b>kill</b> Leaves the tapestry without graceful exit
* <b>exit</b> Quit the CLI

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
Tapestry is extremely fault tolerant, so a node could leave without notifying any other nodes. However, a node can gracefully exit the Tapestry, too. When a node gracefully exits, it notifies all of the nodes in its backpointer table of the leave. As part of this notification, it consults its own routing table to find a suitable replacement for the other node’s routing table.

# Fault Tolerance
Following mechanisms ensure that there is no single point of failure in the system:

* <b> Errors While Routing </b>
When routing towards a surrogate node, it is possible that a communication failure with any of the intermediate nodes could impede the search. For this reason, routing tables store lists of nodes rather than a single node at each slot. If a failed node is encountered, the node that is searching can request that the failed node be removed from any routing tables it encounters, and resume its search at the last node it communicated with successfully. If the last node it communicated with successfully is no longer responding, it should communicate with the last successful node before that.

* <b>Loss of Root Node</b>
Published objects continually republish themselves at regular intervals. This ensures that if a surrogate node goes down, a new surrogate node will eventually take its place. 

* <b>Loss of replicas</b>
Finally, applications built on top of Tapestry might wish to ensure that an object remains available at all times, even if the node that published it fails.
Multiple tapestry nodes can publish the same object. This means that client applications can learn of multiple locations of the object, so if the object becomes unavailable in one of these locations, the client can simply contact another of the nodes. This ensures that an object remains available at all times, even if the node that published it fails.

* <b>Miscellaneous</b>
The cases listed above are the common issues which can arise due to network errors. There are other more obscure ways in which surrogates may become unreachable for a short time when nodes join or fail in a certain order. Tapestry’s method for dealing with this is to assume that there are enough seeded hash values for a given object that not all seeds will become unreachable due to such errors, and those which do become unreachable will be corrected when the replica performs its periodic republishing.

# Future Work
Reduce the number of hops in object lookups. This can be done by caching the nodes encountered along the path to the surrogate node when the location of an object is published to the its surrogate node.
