package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"strconv"

	"github.com/tonetheman/go-socks5"
)

//var remoteHost = "localhost"
//var remotePort = 9000
//var remoteAddress = remoteHost + ":" + strconv.Itoa(remotePort)

type proxyinfo struct {
	listenhost string
	listenport int
	desthost   string
	destport   int
}

func proxy(conn net.Conn, p proxyinfo) {
	remoteAddress := p.desthost + ":" + strconv.Itoa(p.destport)
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

func tonySimpleProxy(p proxyinfo) {

	//listenHost := "localhost"
	//listenPort := 8000
	listenAddress := p.listenhost + ":" + strconv.Itoa(p.listenport)
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

		go proxy(conn, p)
	}
}

func socksstuff(listenAddress *string, listenPort *int) {
	conf := &socks5.Config{}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}
	if err := server.ListenAndServe("tcp", *listenAddress+":"+strconv.Itoa(*listenPort)); err != nil {
		panic(err)
	}
}

func main() {
	listenAddress := flag.String("listenip", "localhost", "listen on this ip address")
	listenPort := flag.Int("listenport", 4444, "local port to listen on for connection")
	destAddress := flag.String("destaddress", "localhost", "dest address for tcp traffic")
	destPort := flag.Int("destport", 5555, "dest port for tcp traffic")
	useSocks5 := flag.Bool("socks5", false, "use socks5 for this proxy")

	fmt.Println("listen info:", *listenAddress, *listenPort)
	fmt.Println("dest info:", *destAddress, *destPort)
	fmt.Println("use socks?", *useSocks5)

	if *useSocks5 {
		fmt.Println("socks5 mode")
		socksstuff(listenAddress, listenPort)
		return
	}

	p := proxyinfo{*listenAddress, *listenPort,
		*destAddress, *destPort}

	fmt.Println("running in simple proxy mode")
	tonySimpleProxy(p)
}
