package raft

type UInt64Slice []uint64

func (p UInt64Slice) Len() int {
	return len(p)
}

func (p UInt64Slice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p UInt64Slice) Less(i, j int) bool {
	return p[i] < p[j]
}

func (r *RaftNode) hasMajority(N uint64) bool {
	numNodes := len(r.GetOtherNodes())
	sum := 1
	for k, v := range r.matchIndex {
		if k != r.Id && v >= N {
			sum++
		}
	}
	if sum > numNodes/2 {
		return true
	}
	return false
}