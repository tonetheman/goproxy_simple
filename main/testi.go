package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"strconv"
)

var remoteHost = "localhost"
var remotePort = 9000
var remoteAddress = remoteHost + ":" + strconv.Itoa(remotePort)

func proxy(conn net.Conn) {
	rAddr, err := net.ResolveTCPAddr("tcp", remoteAddress)
	if err != nil {
		fmt.Println("err resolve remote", err)
		panic(err)
	}

	rConn, err := net.DialTCP("tcp", nil, rAddr)
	if err != nil {
		fmt.Println("err in dial tcp remote", err)
		panic(err)
	}

	defer rConn.Close()

	// sending side
	go func() {
		for {
			data := make([]byte, 32*1024)
			// read from local
			n, err := conn.Read(data)
			if err != nil {
				fmt.Println("err on read sending side", err)
				panic(err)
			}
			fmt.Printf("sent: %v\n", hex.Dump(data[:n]))
			// write the data to the remote
			// use up to n
			rConn.Write(data[:n])
		}
	}()

	// recv side
	for {
		data := make([]byte, 32*1024)
		// read from remote remote
		n, err := rConn.Read(data)
		if err != nil {
			fmt.Println("err on read from remote", err)
			panic(err)
		}
		fmt.Printf("recv: %v\n", hex.Dump(data[:n]))
		// write to local
		conn.Write(data[:n])
	}

}

func tonySimpleProxy() {

	listenHost := "localhost"
	listenPort := 8000
	listenAddress := listenHost + ":" + strconv.Itoa(listenPort)
	addr, err := net.ResolveTCPAddr("tcp", listenAddress)
	if err != nil {
		fmt.Println("err resolve listen", err)
		panic(err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println("err on listen tcp", err)
		panic(err)
	}

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println("err on accept", err)
			panic(err)
		}

		go proxy(conn)
	}
}

func socks5() {
	conf := &socks5.Config{}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}
	if err := server.ListenAndServe("tcp", "127.0.0.1:8000"); err != nil {
		panic(err)
	}
}

func main() {
	listenAddress := flag.String("listenip", "localhost", "listen on this ip address")
	listenPort := flag.Int("listenport", 4444, "local port to listen on for connection")
	destAddress := flag.String("destaddress", "localhost", "dest address for tcp traffic")
	destPort := flag.Int("destport", 5555, "dest port for tcp traffic")

}
