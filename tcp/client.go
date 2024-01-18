package main

import (
	"fmt"
	"log"
	"net"
)

type Client struct {
	Conn    *net.TCPConn
	MStream chan Message
	Nick    string
}

func NewClient(conn *net.TCPConn) *Client {
	return &Client{
		Conn:    conn,
		MStream: make(chan Message),
		Nick:    conn.RemoteAddr().String(),
	}
}

func (c *Client) BroadcastMessageStream() {
	for m := range c.MStream {
		prefix := fmt.Sprintf("%v> ", m.Nick)
		buf := append([]byte(prefix), m.Buf...)
		buf = append(buf, '\n')
		_, err := c.Conn.Write(buf)
		if err != nil {
			log.Print(err)
			continue
		}
	}
}
