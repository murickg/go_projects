package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

type ClientConfig struct {
	Addr string
	Name string
}

func runClient(cfg ClientConfig) error {
	conn, err := net.Dial("tcp", cfg.Addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	log.Println("Connected to", cfg.Addr)

	w := bufio.NewWriter(conn)
	r := bufio.NewScanner(conn)

	if cfg.Name != "" {
		_, _ = io.WriteString(w, "/name "+cfg.Name+"\n")
		_ = w.Flush()
	}

	go func() {
		for r.Scan() {
			fmt.Println(r.Text())
		}
		os.Exit(0)
	}()

	stdin := bufio.NewScanner(os.Stdin)
	for {
		if !stdin.Scan() {
			return stdin.Err()
		}
		line := stdin.Text()
		_, _ = io.WriteString(w, line+"\n")
		if err := w.Flush(); err != nil {
			return err
		}
	}
}

func main() {
	addr := flag.String("addr", "localhost:8080", "server address")
	name := flag.String("name", "", "nickname (optional)")
	flag.Parse()

	if err := runClient(ClientConfig{Addr: *addr, Name: *name}); err != nil {
		log.Fatal(err)
	}
}
