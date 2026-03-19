package main

import (
	"fmt"

	"httpServer/server"
)

func newServer() *server.Server {
	return &server.Server{Closed: false}
}

func main() {
	fmt.Println("started...")
	s := newServer()
	s.Serve(3030)
}
