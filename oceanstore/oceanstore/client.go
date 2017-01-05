package oceanstore

import "fmt"

const MAX_RETRIES = 10

type Client struct {
	LocalAddr string
	Id        uint64
	OceanServ OceanAddr
}

func CreateClient(remoteAddr OceanAddr) (cp *Client, err error) {
	fmt.Println("Oceanstore Create client")
	cp = new(Client)

	request := ConnectRequest{}
	var reply *ConnectReply

	retries := 0
	for retries < MAX_RETRIES {
		reply, err = ConnectRPC(&remoteAddr, request)
		if err == nil || err.Error() != "EOF" {
			break
		}
		retries++
	}
	if err != nil {
		fmt.Println(err)
		if err.Error() == "EOF" {
			err = fmt.Errorf("Could not access the ocean server.")
		}
		return
	}

	if !reply.Ok {
		fmt.Errorf("Could not register Client.")
	}

	fmt.Println("Create client reply:", reply, err)
	cp.Id = reply.Id
	cp.OceanServ = remoteAddr

	return
}
