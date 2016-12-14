# Raft

Raft is a consensus protocol. Its design is motivated from <a href="https://raft.github.io/raft.pdf">In Search of an understandable consensus algorithm</a> paper. 

# Components
* Leader Election
* Log Replication
* Client Interaction

## Leader Election
Leader election consists of a Raft cluster deciding which of the nodes in its cluster should be the leader of a given term. A node starts out in Follower state and if it does not receive a heartbeat within a predefined timeout it transitions into Candidate state. Once in the Candidate state a node votes for itself and sends vote requests to everyone else in the cluster. If a node receives a vote request before casting a vote for someone else in that term, then they cast a vote for the requester. If the node receives a majority of votes, then it moves to the Leader state. Once in the Leader state, the node appends a no-op to the log and sends out heartbeats to everyone else, so that the other nodes know who the new leader is.

## Log Replication
Log replication consists of making sure that the Raft state machine is up to date across a majority of nodes in the cluster. It is based on the heartbeats, periodically initiated by the leader. The leader accepts requests from clients, adds entries to its log, replicates these entries to a majority of the nodes, commits the entry to the log to allow the received followers to feed the entry to their state machines, and then replies to the clients. Raft leaders are responsible for tracking what log entries have been successfully sent to each node, and don’t let the replicated state machines use log entries until they have been propagated to a majority of the nodes. Refer to Figure 2 in the <a href="https://raft.github.io/raft.pdf">paper</a> for details.

## Client Interaction
Followers, candidates, and a leader form up a Raft cluster, which serves Raft clients. A client establishes a session with the Raft clusters by sending a request with CLIENT_REGISTRATION operation, and the log index of the entry corresponding to this request will be the Client ID returned to the client after the entry has been committed. The client then sends requests to the clusters and gets replies with the results once the corresponding log entries have been committed and fed to the state machine. If the Raft node a client connects to is not the leader node, the node returns a hint of the leader node, and the client follows the hint to reach the leader, and this process will be tried multiple times in case of in-progress elections. A request is cached with the Client ID and a request serial number to avoid being executed twice due to retransmissions.


