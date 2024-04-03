package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"sync"
)

func proxyLoop(conn *net.TCPConn, remote *net.TCPConn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		if _, err := io.Copy(conn, remote); err != nil && !errors.Is(err, io.ErrClosedPipe) {
			log.Printf("error copying from ORPort %v", err)
		}
		remote.Close()
		conn.Close()
		wg.Done()
	}()
	go func() {
		if _, err := io.Copy(remote, conn); err != nil && !errors.Is(err, io.EOF) {
			log.Printf("error copying to ORPort %v", err)
		}
		remote.Close()
		conn.Close()
		wg.Done()
	}()

	wg.Wait()

}

func main() {
	proxy := flag.String("addr", "", "proxy address to bind to ")
	server := flag.String("server", "", "server address to connect to ")
	flag.Parse()
	if *proxy == "" || *server == "" {
		log.Fatalf("Specify a local and destination address")
	}
	saddr, err := net.ResolveTCPAddr("tcp", *server)
	if err != nil {
		log.Fatal(err.Error())
	}
	paddr, err := net.ResolveTCPAddr("tcp", *proxy)
	if err != nil {
		log.Fatal(err.Error())
	}
	ln, err := net.ListenTCP("tcp", paddr)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer ln.Close()
	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Temporary() {
				continue
			}
			log.Fatalf("Accept error: %s", err.Error())
		}
		go func() {
			remote, err := net.DialTCP("tcp", nil, saddr)
			if err != nil {
				log.Fatal(err.Error())
			}
			proxyLoop(conn, remote)
			conn.Close()
		}()
	}
}
