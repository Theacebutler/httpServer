package parser

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"strconv"
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
	ContentLength int
	Body          []byte
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

var ERROR_PREQUEST_LINE_NO_INPUT = fmt.Errorf("Request line parser error: no input")
var ERROR_BAD_PREQUEST_LINE = fmt.Errorf("Request line parser error: bad request line")
var ERROR_NO_EMPTY_SPACE_IN_REQUEST_LINE = fmt.Errorf("Request line parser Error: No empty space found in request line")
var ERROR_NO_EMPTY_SPACE_ALLOWED_BEFORE_METHOD = fmt.Errorf("Request line parser Error: No empty space allowed before the METHOD")
var ERROR_NO_EMPTY_SPACE_ALLOWED_BEFORE_TARGET = fmt.Errorf("Request line parser Error: No empty space allowed before the TARGET")
var ERROR_INVALID_HTTP_VERSION = fmt.Errorf("Request line parser Error: Invalid version")
var ERROR_INVALID_HEADER_KEY = fmt.Errorf("Header parser Error: Invalid header key")
var ERROR_HEADER_NO_COLON = fmt.Errorf("Header parser error: No colon found in header")

var RN = []byte("\r\n")
var RNRN = []byte("\r\n\r\n")
var SP = []byte(" ")

func newRequest() *Request {
	return &Request{
		State: parserInit,
	}
}

// request-line   = method SP request-target SP HTTP-version
func (r *Request) ParseRequestLine(rl []byte) (*RequestLine, int, error) {
	read := 0
	// get the method view
	n := bytes.Index(rl, SP)
	if n == -1 {
		r.State = ParserError
		return nil, 0, ERROR_NO_EMPTY_SPACE_IN_REQUEST_LINE
	}
	if n == 0 {
		r.State = ParserError
		return nil, 0, ERROR_NO_EMPTY_SPACE_ALLOWED_BEFORE_METHOD
	}
	// where can the method be found
	method := rl[:n]
	if bytes.HasPrefix(method, SP) {
		r.State = ParserError
		return nil, 0, ERROR_NO_EMPTY_SPACE_ALLOWED_BEFORE_METHOD
	}
	method = bytes.ToUpper(method)
	read += n + len(SP)

	n = bytes.Index(rl[read:], SP)
	if n == -1 {
		r.State = ParserError
		return nil, 0, ERROR_NO_EMPTY_SPACE_IN_REQUEST_LINE
	}
	target := rl[read : read+n]
	if len(target) < 1 {
		r.State = ParserError
		return nil, 0, ERROR_NO_EMPTY_SPACE_ALLOWED_BEFORE_TARGET
	}
	read += n + len(SP)
	n = bytes.Index(rl[read:], RN)
	if n == -1 {
		r.State = ParserError
		return nil, 0, ERROR_INVALID_HTTP_VERSION
	}
	version, err := r.RequestLine.ParseVersion(rl[read : read+n])
	if err != nil {
		r.State = ParserError
		return nil, 0, err
	}
	read += len(version)
	return &RequestLine{
		Method:  method,
		Target:  target,
		Version: version,
	}, read, nil
}

func (l *RequestLine) ParseVersion(v []byte) ([]byte, error) {
	http, v, ok := bytes.Cut(v, []byte("/"))
	if !ok || len(http) == 0 || len(v) == 0 {
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

func newHerders() *Headers {
	return &Headers{
		Headers: map[string]string{},
	}
}

func (h *Headers) Get(key string) (string, error) {
	v := h.Headers[strings.ToLower(key)]
	if len(v) == 0 {
		return "", fmt.Errorf("Key '%s' not found", key)
	}
	return v, nil
}

func (h *Headers) Set(k []byte, v []byte) {
	key := strings.ToLower(string(k))
	value := string(v)
	value = strings.TrimSpace(value)
	old, err := h.Get(key)
	if err != nil {
		h.Headers[key] = value
		return
	}
	h.Headers[key] = fmt.Sprintf("%s,%s", old, value)
}

func (h *Headers) Delete(key string) {
	key = strings.ToLower(key)
	_, err := h.Get(key)
	if err != nil {
		return
	}
	delete(h.Headers, key)
}

func (h *Headers) Replace(key string, value string) {
	key = strings.ToLower(key)
	value = strings.TrimSpace(value)
	h.Headers[key] = value
}
func (h *Headers) Pairs() map[string]string {
	m := map[string]string{}
	maps.Copy(m, h.Headers)
	return m
}

func validKey(b []byte) error {
	good := false
	if len(b) < 2 {
		return ERROR_INVALID_HEADER_KEY
	}
	for _, char := range b {
		good = false
		switch char {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			good = true
		}
		if char >= 'A' && char <= 'Z' || char >= 'a' && char <= 'z' || char >= '0' && char <= '9' {
			good = true
		}
		if !good {
			return ERROR_INVALID_HEADER_KEY
		}
	}
	return nil
}

func (r *Request) ParseHeaders(b []byte) (*Headers, int, error) {
	// field-line   = field-name ":" OWS field-value OWS
	read := 0
	n := 0
	header := []byte{}
	headers := newHerders()
	var err error = nil

	for {
		n = bytes.Index(b[read:], RNRN)
		if n == -1 {
			err = fmt.Errorf("idx == -1")
			break
		}
		if n == 0 {
			break
		}
		header = b[:n]
		key, value, ok := bytes.Cut(header, []byte(":"))
		if !ok {
			err = ERROR_HEADER_NO_COLON
			break
		}

		if bytes.Contains(key, SP) {
			r.State = ParserError
			err = ERROR_INVALID_HEADER_KEY
			break
		}
		headers.Set(key, value)
		err = validKey(key)
		if err != nil {
			return nil, 0, ERROR_INVALID_HEADER_KEY
		}
		read = read + n
	}
	if err != nil {
		return nil, 0, err
	}
	return headers, read, err
}

func (r *Request) ParseBody(b []byte) (*Body, int, error) {
	ln := len(b)
	return &Body{
		Body:          b,
		ContentLength: ln,
	}, ln, nil
}

func toInt(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return -1, err
	}
	return i, nil
}

func (req *Request) parse(data []byte) (int, error) {
	read := 0
	var err error = nil
outer:
	for {
		current := data[read:]
		switch req.State {
		case parserInit:
			// go parse the rl
			rl, n, err := req.ParseRequestLine(current)
			if err != nil {
				req.State = ParserError
			}
			read = read + n
			read += len(RN)
			req.RequestLine = *rl
			req.State = parserHeaders
		case parserHeaders:
			headers, n, err := req.ParseHeaders(current)
			if err != nil {
				req.State = ParserError
				break
			}
			read = read + n + len(RNRN)
			req.Headers = *headers
			req.State = parserBody

		case parserBody:
			cl, err := req.Headers.Get("content-length")
			if err != nil || cl == "0" {
				req.State = parserDone
				break
			}
			body, n, err := req.ParseBody(current)
			if err != nil {
				req.State = ParserError
				break
			}
			read = read + n
			req.Body = *body
			req.State = parserDone
		case parserDone:
			return read, err
		case ParserError:
			break outer
		}
	}
	return 0, err
}

func ParseRequest(reader io.Reader) (*Request, error) {
	// take in a reader and send it to a parder method
	var err error = nil
	req := newRequest()
	bufflen := 0
	buff := make([]byte, 1024)
	for {
		// read into the buff, staring from the bufflen
		n, err := reader.Read(buff[bufflen:])
		if err != nil {
			break
		}
		bufflen += n
		nParse, err := req.parse(buff[:bufflen])
		if err != nil {
			break
		}
		copy(buff, buff[nParse:bufflen])
		bufflen = bufflen - nParse
	}

	return req, err
}
