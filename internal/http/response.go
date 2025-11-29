package http

import "fmt"

var statusTextMap = map[int]string{
	200: "OK",
	400: "Bad Request",
	403: "Forbidden",
	404: "Not Found",
	500: "Internal Server Error", // origin server down?
}

// NewResponse initializes a Response with the given status code and the appropriate status text.
// The Headers and Body are initially empty.
func NewResponse(status int) *Response {
	return &Response{
		Version:    "HTTP/1.0",
		Status:     status,
		StatusText: statusTextMap[status],
		Headers:    make(map[string]string),
	}
}

// BuildResponse builds and returns a Response with the given status code, file content-type, and body.
func BuildResponse(status int, contentType string, body []byte) *Response {
	resp := NewResponse(status)
	resp.Body = body
	resp.Headers["Content-Type"] = contentType
	resp.Headers["Content-Length"] = fmt.Sprint(len(body))
	return resp
}

// WithHeader sets a header with the given key–value pair on the response
// and returns the modified Response to allow for fluent chaining.
// NOTE: add last-modified for better caching? or no cuz it's just fifo?
func (r *Response) WithHeader(key, value string) *Response {
	r.Headers[key] = value
	return r
}

// WithBody sets the body on the respones and returns the
// modified Response to allow for fluent chaining.
func (r *Response) WithBody(b []byte) *Response {
	r.Body = b
	r.Headers["Content-Length"] = fmt.Sprint(len(b)) // update content length as needed
	return r
}
