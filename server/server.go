package server

// TODO: create a response struct, it will have all the same parts as a
// request, it should take in a handler function that will tell the server
// what the status code, headers and body are

type StatusCode int
type Headers map[string]string
type Body []byte

type Response struct {
	Status  StatusCode
	Headers Headers
	Body    Body
}
