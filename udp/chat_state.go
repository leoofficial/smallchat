package main

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"
)

type ChatState struct {
	Conn    *net.UDPConn
	Cli     *sync.Map
	MStream chan Message
	TTL     time.Duration
}

type Message struct {
	Addr *net.UDPAddr
	Buf  []byte
	Nick string
}

func NewChatState(port int, ttl int) (*ChatState, error) {
	addrStr := fmt.Sprintf(":%v", port)
	addr, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	return &ChatState{
		Conn:    conn,
		Cli:     new(sync.Map),
		MStream: make(chan Message),
		TTL:     time.Duration(ttl) * time.Second,
	}, nil
}

func (cs *ChatState) ReceiveMessageStream() {
	for m := range cs.MStream {
		cli, ok := cs.CreateClient(m.Addr)
		if ok {
			go cli.BroadcastMessageStream()
			cli.MStream <- Message{
				Buf: []byte("Welcome to Simple Chat! Use /nick <nick> to set your nick."),
			}
		}
		buf := bytes.TrimSpace(m.Buf)
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
					Nick: cli.Nick,
				}
			}
			return true
		})
	}
}

func (cs *ChatState) CreateClient(addr *net.UDPAddr) (*Client, bool) {
	cli, ok := cs.Cli.Load(addr.AddrPort())
	if ok {
		cli.(*Client).Timer.Reset(cs.TTL)
	} else {
		cli = NewClient(addr, cs.Conn, cs.TTL)
		cs.Cli.Store(addr.AddrPort(), cli)
		go func() {
			<-cli.(*Client).Timer.C
			cs.FreeClient(addr)
		}()
	}
	return cli.(*Client), !ok
}

func (cs *ChatState) FreeClient(addr *net.UDPAddr) {
	cli, ok := cs.Cli.LoadAndDelete(addr.AddrPort())
	if ok {
		close(cli.(*Client).MStream)
	}
}
