package main

import (
	"fmt"
	"httpServer/parser"
	"log"
	"net"
)

func main() {
	fmt.Println("started...")
	listnr, err := net.Listen("tcp", fmt.Sprintf(":%d", 3030))
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listnr.Accept()
		if err != nil {
			fmt.Printf("error, closing connection...\r\n")
			break
		}
		req, err := parser.ParseRequest(conn)
		if err != nil {
			log.Fatal(err)
			break
		}

		fmt.Println("Request:")
		fmt.Printf("Request line\r\n")
		fmt.Printf("Method: %s\r\n", req.RequestLine.Method)
		fmt.Printf("Target: %s\r\n", req.RequestLine.Target)
		fmt.Printf("Version: %s\r\n", req.RequestLine.Version)
		for key, val := range req.Headers.Pairs() {
			fmt.Printf("%s: %s\r\n", key, val)
		}
		fmt.Printf("Body:\r\n%s", req.Body.Body)
	}
}
