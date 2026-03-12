package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var r = newRequest()

func TestRequestLineParsing(t *testing.T) {
	rl, err := r.ParseRequestLine([]byte("GET /foobar HTTP/1.1\r\n"))
	require.NoError(t, err)
	require.NotNil(t, rl)
	t.FailNow()
	assert.Equal(t, "/foobar", string(rl.Target))
	assert.Equal(t, "HTTP/1.1", string(rl.Version))
}

// func TestHeadersParsing(t *testing.T)
// func TestRequestParsing(t *testing.T)
