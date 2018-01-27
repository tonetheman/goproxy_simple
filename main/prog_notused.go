package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"strconv"
)

var verbose bool

func printBuf(msg string, buf []byte, nr int) {
	if !verbose {
		return
	}
	for i := 0; i < nr; i++ {
		fmt.Printf("%x ", buf[i])
	}
	fmt.Println()
}

// copied from io!
// added some logging
func copyBuffer(dst io.Writer, src io.Reader, buf []byte, msg string) (written int64, err error) {
	if buf == nil {
		buf = make([]byte, 1024*32)
	}
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			printBuf(msg+":src.Read", buf, nr)
			nw, ew := dst.Write(buf)
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				fmt.Println("err on write", ew)
				break
			}
			if nr != nw {
				//	err = io.ErrShortWrite
				//	fmt.Println("ErrShortWrite")
				// not really correct
				break
			}
		} // end of nr > 0
		if er != nil {
			if er != io.EOF {
				fmt.Println("err on read not EOF", er)
				err = er
			}
			break
		}
	}
	return written, err
}

func forward(conn net.Conn, dest string) {
	fmt.Println("forward started")
	// connect to where you are going now
	client, err := net.Dial("tcp", dest)
	if err != nil {
		fmt.Println("could not connect in forward", err)
	}

	// this is specific to HTTP proto here
	// read from requestor then wait for reply
	go func() {
		defer client.Close()
		defer conn.Close()
		fmt.Println("copy from client to conn...")
		//io.Copy(client, conn)
		copyBuffer(client, conn, nil, "client_to_conn")
		fmt.Println("copy from conn to client...")
		//io.Copy(conn, client)
		copyBuffer(conn, client, nil, "conn_to_client")

	}()
}

func Notmain() {
	listenAddress := flag.String("listenip", "localhost", "listen on this ip address")
	listenPort := flag.Int("listenport", 4444, "local port to listen on for connection")
	destAddress := flag.String("destaddress", "localhost", "dest address for tcp traffic")
	destPort := flag.Int("destport", 5555, "dest port for tcp traffic")
	flag.BoolVar(&verbose, "verbose", true, "verbose mode!")
	flag.Parse()
	fmt.Println("listenAddress:listenPort", *listenAddress, *listenPort)
	fmt.Println("destAddress:destPort", *destAddress, *destPort)
	listenString := *listenAddress + ":" + strconv.Itoa(*listenPort)
	destString := *destAddress + ":" + strconv.Itoa(*destPort)

	listener, err := net.Listen("tcp", listenString)
	if err != nil {
		fmt.Println("err on Listen", err)
	}
	fmt.Println(listener)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("err on accept", err)
		}
		go forward(conn, destString)
	}
}
