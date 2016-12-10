package main

import (
	"./tapestry"
	"bufio"
	"fmt"
	"os"
	"strings"
	"flag"
)

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println(" - help                    Prints this help message")
	fmt.Println(" - table                   Prints this node's routing table")
	fmt.Println(" - backpointers            Prints this node's backpointers")
	fmt.Println(" - objects                 Prints the advertised objects that are registered to this node")
	fmt.Println(" - put <key> <value>       Stores the provided key-value pair on the local node and advertises the key to the tapestry")
	fmt.Println(" - lookup <key>            Looks up the specified key in the tapestry and prints its location")
	fmt.Println(" - get <key>               Looks up the specified key in the tapestry, then fetches the value from one of the replicas")
	fmt.Println(" - remove <key>            Remove the specified key from the tapestry")
	fmt.Println(" - list                    List the blobs being stored and advertised by the local node")
	fmt.Println(" - leave                   Instructs the local node to gracefully leave the tapestry")
	fmt.Println(" - kill                    Leaves the tapestry without graceful exit")
	fmt.Println(" - exit                    Quit this CLI")
}

func CLI(t *tapestry.Tapestry, done chan bool) {

	printHelp()
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		splits := strings.Split(text, " ")
		command := strings.ToLower(splits[0])
		switch command {
		case "quit", "exit":
			{
				done <- true
				return
			}
		case "table":
			{
				t.PrintRoutingTable()
			}
		case "backpointers":
			{
				t.PrintBackpointers()
			}
		case "replicas", "data", "objects":
			{
				t.PrintObjectStore()
			}
		case "leave":
			{
				t.Leave()
			}
		case "put", "add", "store":
			{
				if len(splits) < 3 {
					fmt.Printf("Insufficient arguments for %s, expect %s <key> <value>\n", command, command)
				} else {
					key := splits[1]
					bytes := []byte(splits[2])
					err := t.Store(key, bytes)
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		case "list", "listblobs":
			{
				t.PrintBlobStore()
			}
		case "lookup", "find":
			{
				if len(splits) < 2 {
					fmt.Printf("Insufficient arguments for %s, expect %s <key>\n", command, command)
				} else {
					key := splits[1]
					replicas, err := t.Lookup(key)
					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Printf("%v: %v\n", key, replicas)
					}
				}
			}
		case "get":
			{
				if len(splits) < 2 {
					fmt.Printf("Insufficient arguments for %s, expect %s <key>\n", command, command)
				} else {
					key := splits[1]
					bytes, err := t.Get(key)
					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Printf("%v: %v\n", key, string(bytes))
					}
				}
			}
		case "remove":
			{
				if len(splits) < 2 {
					fmt.Printf("Insufficient arguments for %s, expect %s <key>\n", command, command)
				} else {
					key := splits[1]
					exists := t.Remove(key)
					if !exists {
						fmt.Printf("This node is not advertising %v\n", key)
					}
				}
			}
		case "help", "commands":
			{
				printHelp()
			}
		case "kill":
			{
				t.Kill()
			}
		default:
			{
				fmt.Printf("Unknown command %s\n", text)
			}
		}
	}
}

func main() {
	var port int
	var addr string

	flag.IntVar(&port, "port", 0, "The server port to bind to. Defaults to a random port.")
	flag.StringVar(&addr, "connect", "", "An existing node to connect to. If left blank, does not attempt to connect to another node.")
	flag.Parse()

	switch {
	case port != 0 && addr != "":
		{
			fmt.Printf("Starting a node on port %v and connecting to %v\n", port, addr)
		}
	case port != 0:
		{
			fmt.Printf("Starting a standalone node on port %v\n", port)
		}
	case addr != "":
		{
			fmt.Printf("Starting a node on a random port and connecting to %v\n", addr)
		}
	default:
		{
			fmt.Printf("Starting a standalone node on a random port\n")
		}
	}

	t, err := tapestry.Start(port, addr)

	if err != nil {
		fmt.Printf("Error starting tapestry node: %v\n", err)
		return
	}

	fmt.Printf("Successfully started: %v\n", t)

	// Kick off CLI, await exit
	done := make(chan bool)
	go CLI(t, done)

	for !(<-done) {
	}

	fmt.Println("Closing tapestry")
	t.Leave()
}