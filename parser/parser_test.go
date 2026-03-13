package parser

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var r = newRequest()

func TestRequestLineParsing(t *testing.T) {
	rl, err := r.ParseRequestLine([]byte("GET /foobar HTTP/1.1\r\n"))
	require.NoError(t, err)
	require.NotNil(t, rl)
	assert.Equal(t, "GET", string(rl.Method))
	assert.Equal(t, "/foobar", string(rl.Target))
	assert.Equal(t, "HTTP/1.1", string(rl.Version))

	rl, err = r.ParseRequestLine([]byte("GET /foobar HTTP/1.1"))
	require.Error(t, err)
	assert.Equal(t, ERROR_INVALID_HTTP_VERSION, err)
	require.Nil(t, rl)

	rl, err = r.ParseRequestLine([]byte("GET /foobar HTTP/1.\r\n"))
	require.Error(t, err)
	assert.Equal(t, ERROR_INVALID_HTTP_VERSION, err)
	require.Nil(t, rl)

	rl, err = r.ParseRequestLine([]byte("GET HTTP/1.1\r\n"))
	require.Error(t, err)
	require.Nil(t, rl)

}

func TestHeadersParsing(t *testing.T) {
	headers, err := r.ParseHeaders([]byte("Content-legth: 18   \r\nContent-type: image/mp4\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, headers)
	cl, err := headers.Get("Content-legth")
	ct, err := headers.Get("Content-type")
	require.NoError(t, err)
	assert.Equal(t, "18", cl)
	assert.Equal(t, "image/mp4", ct)
	headers, err = r.ParseHeaders([]byte("Content-legth : 18   \r\nContent-type: image/mp4\r\n\r\n"))
	assert.Error(t, err)
	assert.Equal(t, ERROR_HEADER_KEY_WITH_WHITESPACE, err)

	headers, err = r.ParseHeaders([]byte("Content-legth: 18   \r\nContent-type: image/mp4\r\nContent-type: application/json\r\n\r\n"))
	ct, err = headers.Get("Content-type")
	assert.Equal(t, "image/mp4,application/json", ct)

	headers.Set([]byte("foo"), []byte("bar"))
	ct, err = headers.Get("foo")
	assert.Equal(t, "bar", ct)
	headers.Replace("foo", "zap")
	ct, err = headers.Get("foo")
	assert.Equal(t, "zap", ct)
	headers.Delete("foo")
	ct, err = headers.Get("foo")
	assert.Error(t, err)
	assert.Equal(t, "", ct)

}

func TestRequestParsing(t *testing.T) {
	body, err := r.ParseBody([]byte("this is some request body to test..."))
	require.NoError(t, err)
	assert.Equal(t, len([]byte("this is some request body to test...")), body.ContentLength)
	assert.Equal(t, string("this is some request body to test..."), string(body.Body))
}

func TestMain(t *testing.T) {
	req := []byte("GET /foobar HTTP/1.1\r\ncontent-length: 18   \r\nContent-type: image/mp4\r\nfoo:bar\r\n\r\nthis is some request body to test...if it fails, sing somebody save me, it will help(your mood, not the test 😭  \r\n")

	buff := bytes.NewReader(req)
	rq, err := ParseRequest(buff)
	require.NoError(t, err)
	assert.Equal(t, "GET", string(rq.RequestLine.Method))
	assert.Equal(t, "HTTP/1.1", string(rq.RequestLine.Version))
	cl, _ := rq.Headers.Get("content-length")
	foo, _ := rq.Headers.Get("foo")
	assert.Equal(t, "bar", string(foo))
	assert.NotEqual(t, 0, string(cl))
}
