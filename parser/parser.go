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
	if n == 0 {
		r.State = ParserError
		return nil, ERROR_NO_EMPTY_SPACE_ALLOWED_BEFORE_METHOD
	}
	method := rl[:n]
	n += len(SP)
	if bytes.HasPrefix(method, SP) {
		r.State = ParserError
		return nil, ERROR_NO_EMPTY_SPACE_ALLOWED_BEFORE_METHOD
	}
	method = bytes.ToUpper(method)

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
	key, value := string(k), string(v)
	key = strings.ToLower(key)
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

// set the content_length
func (r *Request) ParseHeaders(b []byte) (*Headers, error) {
	// field-line   = field-name ":" OWS field-value OWS
	read := 0
	idx := 0
	header := []byte{}
	headers := newHerders()
	var err error = nil

	for {
		idx = bytes.Index(b[read:], RN)
		if idx == -1 {
			err = fmt.Errorf("idx == -1")
			break
		}
		if idx == 0 {
			break
		}
		header = b[read : read+idx]
		key, value, ok := bytes.Cut(header, []byte(":"))
		if !ok {
			err = ERROR_HEADER_NO_SEMICOLON
			break
		}

		if bytes.Contains(key, SP) {
			r.State = ParserError
			err = ERROR_HEADER_KEY_WITH_WHITESPACE
			break
		}
		headers.Set(key, value)
		read += idx + len(RN)
	}
	if err != nil {
		return nil, err
	}

	return headers, err
}

func (r *Request) ParseBody(b []byte) (*Body, error) {
	ln := len(b)
	return &Body{
		Body:          b,
		ContentLength: ln,
	}, nil
}

func toInt(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return -1, err
	}
	return i, nil
}

func ParseRequest(req io.Reader) (*Request, error) {
	// TODO: as of now, we are reading all the data at once, we
	// can be more efficient by reading in chunks until we hit a  or a "\r\n\r\n"
	r := newRequest()
	buff, err := io.ReadAll(req)

	if err != nil {
		r.State = ParserError
		return nil, err
	}

	n := bytes.Index(buff, RN)
	if n == -1 {
		r.State = ParserError
		return nil, ERROR_NO_EMPTY_SPACE_IN_REQUEST_LINE
	}
	n += len(RN)
	rl, err := r.ParseRequestLine(buff[:n])
	if err != nil {
		r.State = ParserError
		return nil, err
	}
	r.State = parserRL
	r.RequestLine = *rl
	// Headers
	h_idx := bytes.Index(buff[n:], RNRN)
	if h_idx == -1 {
		r.State = ParserError
		return nil, fmt.Errorf("headers not terminated by RNRN")
	}

	headers, err := r.ParseHeaders(buff[n : n+h_idx+len(RNRN)])
	if err != nil {
		r.State = ParserError
		return nil, err
	}
	r.State = parserHeaders

	cl, err := headers.Get("content-length")
	if err == nil {
		i, converr := toInt(cl)
		if converr != nil {
			r.State = ParserError
			return nil, fmt.Errorf("Cant get content length: %s", converr)
		}
		r.Body.ContentLength = i
	} else {
		r.Body.ContentLength = 0
	}
	r.Headers = *headers
	// Body
	if r.Body.ContentLength != 0 {
		bodyStart := n + h_idx + len(RNRN)
		body, err := r.ParseBody(buff[bodyStart:])
		if err != nil {
			r.State = ParserError
			return nil, err
		}
		r.Body = *body
	}
	r.State = parserDone

	return r, nil
}
