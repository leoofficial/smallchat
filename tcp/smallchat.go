package main

import (
	"flag"
	"log"
)

func main() {
	port := flag.Int("p", 3000, "port")
	maxClients := flag.Int("c", 3000, "maxClients")
	flag.Parse()
	chatState, err := NewChatState(*port, *maxClients)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := chatState.Ln.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		cli, err := chatState.CreateClient(conn)
		if err != nil {
			if err := conn.Close(); err != nil {
				log.Print(err)
				continue
			}
		}
		go chatState.ReceiveMessageStream(cli)
		go cli.BroadcastMessageStream()
	}
}
