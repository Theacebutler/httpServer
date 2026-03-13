package parser

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"strings"
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

var ERROR_INVALID_HTTP_VERSION = fmt.Errorf("Request line parser Error: Invelid version")
var ERROR_HEADER_KEY_WITH_WHITESPACE = fmt.Errorf("Header parser Error: No white space allowed in header key")
var ERROR_HEADER_NO_SEMICOLON = fmt.Errorf("Header parser error: No semicolon found in header")
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
		return nil, ERROR_NO_EMPTY_SPACE_IN_REQUEST_LINE
	}
	// DEBUG: read up to n, where method_idx is the idx of the first SP
	method := rl[:n]
	n += len(SP)
	if bytes.HasPrefix(method, SP) {
		r.State = ParserError
		return nil, ERROR_NO_EMPTY_SPACE_ALLOWED_BEFORE_METHOD
	}
	method = bytes.ToUpper(method)

	// DEBUG: read from n to the end and find the next idx of SP, target_idx is the idx of the next SP
	target_idx := bytes.Index(rl[n:], SP)
	if target_idx == -1 {
		r.State = ParserError
		return nil, ERROR_NO_EMPTY_SPACE_IN_REQUEST_LINE
	}
	target := rl[n : target_idx+n]
	if len(target) < 1 {
		r.State = ParserError
		return nil, ERROR_NO_EMPTY_SPACE_ALLOWED_BEFORE_TARGET
	}
	after_target_idx := target_idx + n + len(SP)

	version_idx := bytes.Index(rl[after_target_idx:], RN)
	if version_idx == -1 {
		r.State = ParserError
		return nil, ERROR_INVALID_HTTP_VERSION
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
		return nil, ERROR_INVALID_HTTP_VERSION
	}

	for _, x := range v {
		if x == '.' && v[len(v)-1] != '.' {
			continue
		}
		if x > '9' || x < '0' {
			return nil, ERROR_INVALID_HTTP_VERSION
		}
	}

	return fmt.Appendf(nil, "%s/%s", http, v), nil
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
