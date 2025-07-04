package main

import (
	"fmt"
	"log"
	"net"

	"github.com/joac1144/bootdev-httpfromtcp/internal/request"
)

func main() {
	tcpListener, err := net.Listen("tcp", ":42069")
	if err != nil {
		panic(err)
	}
	defer tcpListener.Close()

	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("New connection from %s\n", conn.RemoteAddr())

		go func(c net.Conn) {
			defer c.Close()
			request, err := request.RequestFromReader(c)
			if err != nil {
				log.Printf("Error reading request: %v\n", err)
				return
			}

			fmt.Println("Request line:")
			fmt.Printf("- Method: %s\n", request.RequestLine.Method)
			fmt.Printf("- Target: %s\n", request.RequestLine.RequestTarget)
			fmt.Printf("- Version: %s\n", request.RequestLine.HttpVersion)
		}(conn)
	}
}
