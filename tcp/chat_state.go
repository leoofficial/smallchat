package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

type Message struct {
	Nick string
	Buf  []byte
}

type ChatState struct {
	Cli        *sync.Map
	Ln         *net.TCPListener
	MStream    chan Message
	MaxClients int
	NumClients int
}

func NewChatState(port, maxClients int) (*ChatState, error) {
	addrStr := fmt.Sprintf(":%v", port)
	addr, err := net.ResolveTCPAddr("tcp", addrStr)
	if err != nil {
		return nil, err
	}
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &ChatState{
		Cli:        new(sync.Map),
		Ln:         ln,
		MaxClients: maxClients,
	}, nil
}

func (cs *ChatState) ReceiveMessageStream(cli *Client) {
	cli.MStream <- Message{
		Buf: []byte("Welcome to Simple Chat! Use /nick <nick> to set your nick."),
	}
	buf := make([]byte, 1024)
	for {
		n, err := cli.Conn.Read(buf)
		if err != nil {
			log.Print(err)
			cs.FreeClient(cli.Conn)
			return
		}
		buf = bytes.TrimSpace(buf[:n])
		if len(buf) == 0 {
			continue
		}
		if buf[0] == '/' {
			line := bytes.SplitN(buf, []byte(" "), 2)
			if len(line) == 2 && bytes.Equal(line[0], []byte("/nick")) {
				cli.Nick = string(line[1])
			} else {
				cli.MStream <- Message{
					Buf: []byte("Unsupported command."),
				}
			}
			continue
		}
		cs.Cli.Range(func(_, val any) bool {
			if val.(*Client) != cli {
				val.(*Client).MStream <- Message{
					Buf:  buf,
					Nick: val.(*Client).Nick,
				}
			}
			return true
		})
	}
}

func (cs *ChatState) CreateClient(conn net.Conn) (*Client, error) {
	if cs.NumClients >= cs.MaxClients {
		return nil, errors.New("exceed max clients")
	}
	cli := NewClient(conn.(*net.TCPConn))
	cs.Cli.Store(conn.RemoteAddr(), cli)
	cs.NumClients++
	return cli, nil
}

func (cs *ChatState) FreeClient(conn *net.TCPConn) {
	cli, ok := cs.Cli.LoadAndDelete(conn.RemoteAddr())
	if ok {
		if err := conn.Close(); err != nil {
			log.Fatal(err)
		}
		close(cli.(*Client).MStream)
		cs.NumClients--
	}
}
