package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type Message struct {
	From string
	Text string
}

type Client struct {
	Name string
	Conn net.Conn
	Send chan string
}

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan Message
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Message, 128),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = true
			h.sendAllExcept(fmt.Sprintf("%s вошёл в чат", c.Name), c)
		case c := <-h.unregister:
			if h.clients[c] {
				delete(h.clients, c)
				close(c.Send)
				h.sendAllExcept(fmt.Sprintf("%s покинул чат", c.Name), c)
			}
		case m := <-h.broadcast:
			line := fmt.Sprintf("[%s] %s", m.From, m.Text)
			for c := range h.clients {
				select {
				case c.Send <- line:
				default:
					delete(h.clients, c)
					close(c.Send)
				}
			}
		}
	}
}

func (h *Hub) sendAllExcept(text string, except *Client) {
	for c := range h.clients {
		if c != except {
			select {
			case c.Send <- text:
			default:
				delete(h.clients, c)
				close(c.Send)
			}
		}
	}
}

func handleConn(h *Hub, conn net.Conn) {
	name := conn.RemoteAddr().String()
	client := &Client{
		Name: name,
		Conn: conn,
		Send: make(chan string, 32),
	}

	h.register <- client

	go func() {
		w := bufio.NewWriter(conn)
		for msg := range client.Send {
			_, _ = io.WriteString(w, msg+"\n")
			_ = w.Flush()
		}
	}()

	r := bufio.NewScanner(conn)
	for r.Scan() {
		line := strings.TrimSpace(r.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "/name ") {
			newName := strings.TrimSpace(strings.TrimPrefix(line, "/name "))
			if newName != "" {
				old := client.Name
				client.Name = newName
				client.Send <- fmt.Sprintf("Имя обновлено: %s → %s", old, client.Name)
				h.sendAllExcept(fmt.Sprintf("%s теперь %s", old, client.Name), client)
			}
			continue
		}
		h.broadcast <- Message{From: client.Name, Text: line}
	}

	h.unregister <- client
	_ = conn.Close()
}

func runServer(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Println("TCP chat server listening on", addr)

	h := NewHub()
	go h.Run()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		log.Println("Shutting down...")
		_ = ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			log.Println("accept error:", err)
			continue
		}
		go handleConn(h, conn)
	}
}

func main() {
	addr := ":8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	if err := runServer(addr); err != nil {
		log.Fatal(err)
	}
}
