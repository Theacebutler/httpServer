package server

import (
	"fmt"
	"httpServer/response"
	"io"
	"net"
)

type Server struct {
	Closed bool
}

func handle(conn io.WriteCloser) {
	defer conn.Close()
	response.WriteResponse(conn)
}

func listen(s *Server, listener net.Listener) {
	for !s.Closed {
		conn, err := listener.Accept()
		if s.Closed || err != nil {
			return
		}
		go handle(conn)
	}
}

func (s *Server) Close() {
	s.Closed = true
}

func (s *Server) Serve(port int) error {
	defer s.Close()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	defer listener.Close()
	listen(s, listener)
	return nil
}
