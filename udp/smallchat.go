package main

import (
	"flag"
	"log"
)

func main() {
	port := flag.Int("p", 3000, "port")
	ttl := flag.Int("t", 120, "ttl")
	flag.Parse()
	chatState, err := NewChatState(*port, *ttl)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		close(chatState.MStream)
		if err := chatState.Conn.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	go chatState.ReceiveMessageStream()
	for {
		buf := make([]byte, 1024)
		n, addr, err := chatState.Conn.ReadFromUDP(buf)
		if err != nil {
			log.Print(err)
			continue
		}
		chatState.MStream <- Message{
			Addr: addr,
			Buf:  buf[:n],
		}
	}
}
