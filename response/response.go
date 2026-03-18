package response

/* This is the package that creates the response and gives it to the server
* to serve it back over the line, it has handler and writer methods, the handler is
* implemented by the end user - it tells the writer what to write based on the
* request it gets. And the writer function - it takes in a io writer/ a []byte
* buff to write the request line, headers and body and returns the response
* to the server package so that is can send it down the line. */

import (
	"bytes"
	"io"

	"httpServer/parser"
)

type WriterState string

const (
	WriteError   WriterState = "writer error"
	WriteInit    WriterState = "writer init"
	WriteRL      WriterState = "write request line"
	WriteHeaders WriterState = "write headers"
	Writebody    WriterState = "write body"
	WriteDone    WriterState = "writer done"
)

type Writer struct {
	State  WriterState
	Writer io.Writer
}

type Response struct {
	RL      parser.RequestLine
	Headers parser.Headers
	Body    parser.Body
	Writer  *Writer
}

func newResponse() *Response {
	var buff bytes.Buffer
	w := &Writer{
		State:  WriteInit,
		Writer: &buff,
	}
	return &Response{Writer: w}
}
