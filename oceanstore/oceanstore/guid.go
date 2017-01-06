package oceanstore

import (
	"../../raft/raft"
	"fmt"
)

func (puddle *OceanNode) getRaftVguid(aguid Aguid, id uint64) (Vguid, error) {
	// Get the raft client struct
	c, ok := puddle.clients[id]
	if !ok {
		panic("Attempted to get client from id, but not found.")
	}

	res, err := c.SendRequestWithResponse(raft.GET, []byte(aguid))
	if err != nil {
		return "", err
	}
	if res.Status != raft.OK {
		return "", fmt.Errorf("Could not get response from raft.")
	}

	return Vguid(res.Response), nil
}