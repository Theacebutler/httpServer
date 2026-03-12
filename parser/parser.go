package parser

import (
	"bytes"
	"fmt"
	"io"
)

type RequestLine struct {
	Method  []byte
	Target  []byte
	Version []byte
}

type Headers struct {
	Headers map[string]string
}

type Body struct {
	content_length int
	Body           []byte
}

type Request struct {
	State       ParserState
	RequestLine RequestLine
	Headers     Headers
	Body        Body
}

type ParserState string

const (
	ParserError   ParserState = "Error"
	parserInit    ParserState = "Init"
	parserDone    ParserState = "Done"
	parserRL      ParserState = "Request Line"
	parserHeaders ParserState = "Request Headers"
	parserBody    ParserState = "Request Body"
)

var RN = []byte("\r\n")
var RNRN = []byte("\r\n\r\n")
var SP = []byte(" ")

func newRequest() *Request {
	return &Request{}
}

// request-line   = method SP request-target SP HTTP-version
func (r *Request) ParseRequestLine(rl []byte) (*RequestLine, error) {
	n := bytes.Index(rl, SP)
	if n == -1 {
		r.State = ParserError
		return nil, fmt.Errorf("Request line parser Error: No empty space found in request line")
	}
	// DEBUG: read up to n, where method_idx is the idx of the first SP
	method := rl[:n]
	n += len(SP)
	if bytes.HasPrefix(method, SP) {
		r.State = ParserError
		return nil, fmt.Errorf("Request line parser Error: No empty space allowed before the METHOD")
	}
	method = bytes.ToUpper(method)

	// DEBUG: read from n to the end and find the next idx of SP, target_idx is the idx of the next SP
	target_idx := bytes.Index(rl[n:], SP)
	if target_idx == -1 {
		r.State = ParserError
		return nil, fmt.Errorf("Request line parser Error: No empty space found in request line")
	}
	target := rl[n : target_idx+n]
	if len(target) < 1 {
		r.State = ParserError
		return nil, fmt.Errorf("Request line parser Error: No empty space allowed before the request-target")
	}
	after_target_idx := target_idx + n + len(SP)

	version_idx := bytes.Index(rl[after_target_idx:], RN)
	if version_idx == -1 {
		r.State = ParserError
		return nil, fmt.Errorf("Request line parser Error: Invelid version")
	}
	version, err := r.RequestLine.ParseVersion(rl[after_target_idx : after_target_idx+version_idx])
	if err != nil {
		r.State = ParserError
		return nil, err
	}
	return &RequestLine{
		Method:  method,
		Target:  target,
		Version: version,
	}, nil
}

func (l *RequestLine) ParseVersion(v []byte) ([]byte, error) {
	http, v, ok := bytes.Cut(v, []byte("/"))
	if !ok || http == nil || v == nil {
		return nil, fmt.Errorf("Cant parse HTTP-version")
	}

	for _, x := range v {
		if x == '.' && v[len(v)-1] != '.' {
			continue
		}
		if x > '9' || x < '0' {
			return nil, fmt.Errorf("Cant parse HTTP-version number")
		}
	}

	return fmt.Appendf(nil, "%s/%d", http, v), nil
}

// set the content_length
func (r *Request) ParseHeaders([]byte) *Headers {
	return nil
}

func (r *Request) ParseBody([]byte) *Body {
	return nil
}

func ParseRequest(req io.Reader) (*Request, error) {
	// TODO: as of now, we are reading all the data at once, we
	// can be more efficient by reading in chunks until we hit a  or a "\r\n\r\n"
	r := newRequest()
	buff, err := io.ReadAll(req)

	// while the req has data in it, read data []byte slices in to it.
	if err != nil {
		r.State = ParserError
		return nil, fmt.Errorf("Error: %s", err)
	}

	n := bytes.Index(buff, RN)
	n += len(RN)
	rl, err := r.ParseRequestLine(buff[n:])
	if err != nil {
		r.State = ParserError
		return nil, fmt.Errorf("Error: %s", err)
	}
	r.State = parserRL

	n = bytes.Index(buff[n:], RNRN)
	n += len(RNRN)
	headers := r.ParseHeaders(buff[n:])
	r.State = parserHeaders

	n = bytes.Index(buff[n:], RN)
	n += len(RN)
	body := r.ParseBody(buff[n:])
	r.State = parserBody

	return &Request{
		RequestLine: *rl,
		Headers:     *headers,
		Body:        *body,
	}, nil
}
