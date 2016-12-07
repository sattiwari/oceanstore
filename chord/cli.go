package main

import (
	"./chord"
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
)

func NodeStr(node *chord.Node) string {
	var succ []byte
	var pred []byte
	if node.Successor != nil {
		succ = node.Successor.Id
	}
	if node.Predecessor != nil {
		pred = node.Predecessor.Id
	}

	return fmt.Sprintf("Node-%v: {succ:%v, pred:%v}", node.Id, succ, pred)
}

func main() {
	countPtr := flag.Int("count", 5, "Total number of Chord nodes to start up in this process")
	addrPtr := flag.String("addr", "", "Address of a node in the Chord ring you wish to join")
	idPtr := flag.String("id", "", "ID of a node in the Chord ring you wish to join")
	flag.Parse()

	var parent *chord.RemoteNode
	if *addrPtr == "" {
		parent = nil
	} else {
		parent = new(chord.RemoteNode)
		val := big.NewInt(0)
		val.SetString(*idPtr, 10)
		parent.Id = val.Bytes()
		parent.Addr = *addrPtr
		fmt.Printf("Attach this node to id:%v, addr:%v\n", parent.Id, parent.Addr)
	}

	var err error
	nodes := make([]*chord.Node, *countPtr)
	for i, _ := range nodes {
		nodes[i], err = chord.CreateNode(parent)
		if err != nil {
			fmt.Println("Unable to create new node!")
			log.Fatal(err)
		}
		if parent == nil {
			parent = nodes[i].RemoteSelf
		}
		fmt.Printf("Created -id %v -addr %v\n", chord.HashStr(nodes[i].Id), nodes[i].Addr)
	}

	for {
		fmt.Printf("quit|node|table|addr|data|get|put > ")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		args := strings.SplitN(line, " ", 3)

		switch args[0] {
		case "node":
			for _, node := range nodes {
				fmt.Println(NodeStr(node))
			}
		case "table":
			for _, node := range nodes {
				chord.PrintFingerTable(node)
			}
		case "addr":
			for _, node := range nodes {
				fmt.Println(node.Addr)
			}
		case "data":
			for _, node := range nodes {
				chord.PrintDataStore(node)
			}
		case "get":
			if len(args) > 1 {
				val, err := chord.Get(nodes[0], args[1])
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println(val)
				}
			}
		case "put":
			if len(args) > 2 {
				err := chord.Put(nodes[0], args[1], args[2])
				if err != nil {
					fmt.Println(err)
				}
			}
		case "quit":
			fmt.Println("goodbye")
			for _, node := range nodes {
				chord.ShutdownNode(node)
			}
			return
		default:
			continue
		}
	}
}