package server

import (
	"fmt"
	"io"
	"net"
	"time"

	"httpServer/parser"
)

type WriterState string

const (
	WriterInit    WriterState = "init"
	WriterRL      WriterState = "RL"
	WriterHeaders WriterState = "Headers"
	WriterBody    WriterState = "Body"
	WriterDone    WriterState = "Done"
)

type StatusCode int

const (
	Ok                  StatusCode = 200
	BadRequest          StatusCode = 400
	NotFound            StatusCode = 404
	InternalServerError StatusCode = 500
)

type ResponseWriter struct {
	State  WriterState
	Writer io.Writer
}

type Server struct {
	Closed         bool
	responseWriter ResponseWriter
}

func (s *Server) writeRL(status StatusCode) error {
	s.responseWriter.State = WriterRL
	rl := []byte{}
	switch status {
	case Ok:
		rl = []byte("HTTP/1.1 200 OK")
	case BadRequest:
		rl = []byte("HTTP/1.1 400 Bad Request")
	case NotFound:
		rl = []byte("HTTP/1.1 404 not Found")
	case InternalServerError:
		rl = []byte("HTTP/1.1 500 Internal server error")
	}
	rl = fmt.Appendf(rl, "\r\n")
	_, err := s.responseWriter.Writer.Write(rl)
	if err != nil {
		return err
	}
	return nil
}

func newHeaders() *parser.Headers {
	return &parser.Headers{Headers: map[string]string{}}
}

func (s *Server) writeDefaultHeaders(h *parser.Headers, n int) {
	h.Set([]byte("Content-Type"), []byte("text/plain"))
	h.Set([]byte("Connection"), []byte("open"))
	h.Set([]byte("Content-length"), fmt.Appendf(nil, "%d", n))
	t, err := time.Now().MarshalText()
	if err == nil {
		h.Set([]byte("date"), []byte(t))
	}
}

func (s *Server) writeHaders(content_length int) error {
	s.responseWriter.State = WriterHeaders
	header := []byte{}
	h := newHeaders()
	s.writeDefaultHeaders(h, content_length)
	for key, val := range h.Pairs() {
		header = fmt.Appendf(header, "%s: %s\r\n", key, val)
	}
	header = fmt.Appendf(header, "\r\n")
	_, err := s.responseWriter.Writer.Write(header)
	if err != nil {
		return err
	}
	return nil
}

// Write the body as a byte slice to the conn
func (s *Server) writeBody(b []byte) error {
	s.responseWriter.State = WriterBody
	_, err := s.responseWriter.Writer.Write(b)
	if err != nil {
		return err
	}
	return nil

}

func handle500(s *Server) {
	s.writeRL(InternalServerError)
	s.writeHaders(len(Respons500()))
	s.writeBody(Respons500())
}

func handle(s *Server) {
	s.writeRL(Ok)
	s.writeHaders(len(Respons200()))
	s.writeBody(Respons200())
}

func listen(s *Server, listener net.Listener) {
	for !s.Closed {
		conn, err := listener.Accept()
		s.responseWriter.Writer = conn
		if s.Closed || err != nil {
			return
		}
		go handle(s)
	}
}

func (s *Server) Close() {
	s.Closed = true
}

func (s *Server) Serve(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	defer listener.Close()
	listen(s, listener)
	return nil
}
