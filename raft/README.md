# Raft

Raft is a consensus protocol. Its design is motivated from <a href="https://raft.github.io/raft.pdf">In Search of an understandable consensus algorithm</a> paper. I would highly recommend <a href="http://thesecretlivesofdata.com/raft/">this</a> visualization for understanding raft.

# Usage Example
Cli-node serves as a console for interacting with raft, creating nodes and querying state on the local nodes. It provides the following commands:
* Debug <on/off>
* Recv <addr> <on/off>
* Send <addr> <on/off>
* Disable
* Enable
* state
* Exit

Testing-policy simulates different network splits to ensure correct behaviour under partitions.

# State Machine
Softwares that make use of raft work by interpreting the entries in a log as input to a state machine. In this project, the [state machine](https://github.com/sattiwari/oceanstore/blob/master/raft/raft/machine.go) calculates the next step of a hash chain. Cli-client supports interaction with state machine. It provides following commands:
* Init (value) sends an initial value for hashing to the replicated state machine
* Hash instructs the state machine to perform another round of hashing

# Elections
Leader election consists of a raft cluster deciding which of the nodes in the cluster should be the leader for a given term. Raft_states contain the logic for raft node being in one of the three states: FOLLOWER, CANDIDATE, LEADER.

# Log Replication
Log replication consists of making sure that the raft state machine is up to date across a majority of nodes in the cluster. It is based on AppendEntries (heartbeat), periodically initiated by the leader.

# Client Interaction
Client sends request to the cluster and get replies with the results once the corresponding log entries have been committed and fed to the state machine. If the raft node a client connects to is not the leader node, the node returns a hint to the leader node.
