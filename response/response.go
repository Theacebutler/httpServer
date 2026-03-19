package response

/* This is the package that creates the response and gives it to the server
* to serve it back over the line, it has handler and writer methods, the handler is
* implemented by the end user - it tells the writer what to write based on the
* request it gets. And the writer function - it takes in a io writer/ a []byte
* buff to write the request line, headers and body and returns the response
* to the server package so that is can send it down the line. */

import (
	"fmt"
	"io"
	"time"

	"httpServer/parser"
)

const (
	WriteError   WriterState = "writer error"
	WriteInit    WriterState = "writer init"
	WriteRL      WriterState = "write request line"
	WriteHeaders WriterState = "write headers"
	Writebody    WriterState = "write body"
	WriteDone    WriterState = "writer done"
)

type StatusCode int

const (
	Ok                  StatusCode = 200
	BadRequest          StatusCode = 400
	NotFound            StatusCode = 404
	InternalServerError StatusCode = 500
)

type WriterState string

const (
	WriterInit    WriterState = "init"
	WriterRL      WriterState = "RL"
	WriterHeaders WriterState = "Headers"
	WriterBody    WriterState = "Body"
	WriterDone    WriterState = "Done"
)

type Writer struct {
	State  WriterState
	Writer io.Writer
}

type Response struct {
	RL             parser.RequestLine
	Headers        parser.Headers
	Body           parser.Body
	responseWriter io.Writer
}

func (s *Response) writeRL(status StatusCode) error {
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
	_, err := s.responseWriter.Write(rl)
	if err != nil {
		return err
	}
	return nil
}

func newHeaders() *parser.Headers {
	return &parser.Headers{Headers: map[string]string{}}
}

func (s *Response) writeDefaultHeaders(h *parser.Headers, n int) {
	h.Set([]byte("Content-Type"), []byte("text/html"))
	h.Set([]byte("Connection"), []byte("open"))
	h.Set([]byte("Content-length"), fmt.Appendf(nil, "%d", n))
	t, err := time.Now().MarshalText()
	if err == nil {
		h.Set([]byte("date"), []byte(t))
	}
}

func (s *Response) writeHaders(content_length int) error {
	header := []byte{}
	h := newHeaders()
	s.writeDefaultHeaders(h, content_length)
	for key, val := range h.Pairs() {
		header = fmt.Appendf(header, "%s: %s\r\n", key, val)
	}
	header = fmt.Appendf(header, "\r\n")
	_, err := s.responseWriter.Write(header)
	if err != nil {
		return err
	}
	return nil
}

// Write the body as a byte slice to the conn
func (s *Response) writeBody(b []byte) error {
	_, err := s.responseWriter.Write(b)
	if err != nil {
		return err
	}
	return nil
}

func WriteResponse(conn io.Writer) {
	res := &Response{responseWriter: conn}
	res.writeRL(Ok)
	b := Respons200()
	res.writeHaders(len(b))
	res.writeBody(b)
}
