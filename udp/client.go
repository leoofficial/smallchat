package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

type Client struct {
	Addr    *net.UDPAddr
	Conn    *net.UDPConn
	MStream chan Message
	Nick    string
	Timer   *time.Timer
}

func NewClient(addr *net.UDPAddr, conn *net.UDPConn, ttl time.Duration) *Client {
	return &Client{
		Addr:    addr,
		Conn:    conn,
		MStream: make(chan Message),
		Nick:    addr.String(),
		Timer:   time.NewTimer(ttl),
	}
}

func (c *Client) BroadcastMessageStream() {
	for m := range c.MStream {
		prefix := fmt.Sprintf("%v> ", m.Nick)
		buf := append([]byte(prefix), m.Buf...)
		buf = append(buf, '\n')
		_, err := c.Conn.WriteToUDP(buf, c.Addr)
		if err != nil {
			log.Print(err)
			continue
		}
	}
}
