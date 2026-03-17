package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLineParsing(t *testing.T) {
	r := newRequest()
	tests := []struct {
		name     string
		input    string
		expected *RequestLine
		wantErr  error
	}{
		{
			name:  "Valid GET request",
			input: "GET /foobar HTTP/1.1\r\n",
			expected: &RequestLine{
				Method:  []byte("GET"),
				Target:  []byte("/foobar"),
				Version: []byte("HTTP/1.1"),
			},
			wantErr: nil,
		},
		{
			name:  "Valid POST request",
			input: "POST /api/v1/data HTTP/2.0\r\n",
			expected: &RequestLine{
				Method:  []byte("POST"),
				Target:  []byte("/api/v1/data"),
				Version: []byte("HTTP/2.0"),
			},
			wantErr: nil,
		},
		{
			name:    "Missing newline",
			input:   "GET /foobar HTTP/1.1",
			wantErr: ERROR_INVALID_HTTP_VERSION,
		},
		{
			name:    "Invalid version format",
			input:   "GET /foobar HTTP/1.\r\n",
			wantErr: ERROR_INVALID_HTTP_VERSION,
		},
		{
			name:    "Missing target",
			input:   "GET  HTTP/1.1\r\n",
			wantErr: ERROR_NO_EMPTY_SPACE_ALLOWED_BEFORE_TARGET,
		},
		{
			name:    "No space in request line",
			input:   "GET/foobarHTTP/1.1\r\n",
			wantErr: ERROR_NO_EMPTY_SPACE_IN_REQUEST_LINE,
		},
		{
			name:    "Leading space before method",
			input:   " GET /foobar HTTP/1.1\r\n",
			wantErr: ERROR_NO_EMPTY_SPACE_ALLOWED_BEFORE_METHOD,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl, err := r.ParseRequestLine([]byte(tt.input))
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, rl)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rl)
				assert.Equal(t, string(tt.expected.Method), string(rl.Method))
				assert.Equal(t, string(tt.expected.Target), string(rl.Target))
				assert.Equal(t, string(tt.expected.Version), string(rl.Version))
			}
		})
	}
}

func TestHeadersParsing(t *testing.T) {
	r := newRequest()
	tests := []struct {
		name     string
		input    string
		expected map[string]string
		wantErr  error
	}{
		{
			name:  "Basic headers",
			input: "Content-Length: 18\r\nContent-Type: text/plain\r\n\r\n",
			expected: map[string]string{
				"content-length": "18",
				"content-type":   "text/plain",
			},
		},
		{
			name:  "Case insensitivity and trimming",
			input: "HOST: localhost:8080  \r\nUSER-agent: Mozilla/5.0\r\n\r\n",
			expected: map[string]string{
				"host":       "localhost:8080",
				"user-agent": "Mozilla/5.0",
			},
		},
		{
			name:  "Multiple same headers (comma separated)",
			input: "Accept: text/html\r\nAccept: application/xhtml+xml\r\n\r\n",
			expected: map[string]string{
				"accept": "text/html,application/xhtml+xml",
			},
		},
		{
			name:    "Invalid header (space in key)",
			input:   "Content Length: 18\r\n\r\n",
			wantErr: ERROR_INVALID_HEADER_KEY,
		},
		{
			name:    "Invalid header (no colon)",
			input:   "InvalidHeaderLine\r\n\r\n",
			wantErr: ERROR_HEADER_NO_COLON,
		},
		{
			name:    "Invalid header (Invalid chars)",
			input:   "Conte🚫t-Length: 18\r\n\r\n",
			wantErr: ERROR_INVALID_HEADER_KEY,
		},
		{
			name:    "Invalid header (short header key)",
			input:   "C: 18\r\n\r\n",
			wantErr: ERROR_INVALID_HEADER_KEY,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers, err := r.ParseHeaders([]byte(tt.input))
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, headers)
			} else {
				require.NoError(t, err)
				require.NotNil(t, headers)
				for k, v := range tt.expected {
					val, err := headers.Get(k)
					require.NoError(t, err)
					assert.Equal(t, v, val)
				}
			}
		})
	}
}

func TestBodyParsing(t *testing.T) {
	r := newRequest()
	input := []byte("hello world")
	body, err := r.ParseBody(input)
	require.NoError(t, err)
	assert.Equal(t, 11, body.ContentLength)
	assert.Equal(t, input, body.Body)
}

func TestFullRequestParsing(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{
			name: "Simple GET request",
			raw:  "GET /index.html HTTP/1.1\r\nHost: example.com\r\n\r\n",
		},
		{
			name: "POST request with body",
			raw:  "POST /submit HTTP/1.1\r\nContent-Length: 11\r\nContent-Type: text/plain\r\n\r\nhello world",
		},
		{
			name: "Request with multiple headers",
			raw:  "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: test\r\nAccept: */*\r\n\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseRequest(strings.NewReader(tt.raw))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, req)

				// Verify first line
				parts := strings.Split(strings.Split(tt.raw, "\r\n")[0], " ")
				assert.Equal(t, parts[0], string(req.RequestLine.Method))
				assert.Equal(t, parts[1], string(req.RequestLine.Target))
				assert.Equal(t, parts[2], string(req.RequestLine.Version))

				// Verify body length if Content-Length present
				if strings.Contains(tt.raw, "Content-Length:") {
					cl, _ := req.Headers.Get("content-length")
					assert.NotEmpty(t, cl)
					expectedBody := tt.raw[strings.Index(tt.raw, "\r\n\r\n")+4:]
					assert.Equal(t, expectedBody, string(req.Body.Body))
				}
			}
		})
	}
}
func TestParseVersion(t *testing.T) {
	rl := &RequestLine{}
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"HTTP/1.1", "HTTP/1.1", false},
		{"HTTP/2.0", "HTTP/2.0", false},
		{"HTTP/1.0", "HTTP/1.0", false},
		{"INVALID", "", true},
		{"HTTP/", "", true},
		{"/1.1", "", true},
		{"HTTP/1.", "", true},
		{"HTTP/1.a", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := rl.ParseVersion([]byte(tt.input))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, string(got))
			}
		})
	}
}
