package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

type (
	HttpMethod   string
	RequestState int

	// GET /index.html HTTP/1.1
	RequestLine struct {
		Method        HttpMethod
		HttpVersion   string
		RequestTarget string
	}

	Request struct {
		RequestLine RequestLine
		Headers     map[string]string
		Body        []byte
		state       RequestState
	}
)

const (
	stateInitialized RequestState = iota
	stateDone
)

const (
	crlf       = "\r\n"
	bufferSize = 8
)

const (
	GET     HttpMethod = "GET"
	POST    HttpMethod = "POST"
	PUT     HttpMethod = "PUT"
	PATCH   HttpMethod = "PATCH"
	DELETE  HttpMethod = "DELETE"
	OPTIONS HttpMethod = "OPTIONS"
	TRACE   HttpMethod = "TRACE"
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	r := &Request{
		Headers: nil,
		Body:    []byte{},
	}
	buff := make([]byte, bufferSize)
	var fullReqData []byte

	for {
		n, err := reader.Read(buff)
		if n > 0 {
			fullReqData = append(fullReqData, buff[:n]...)
		}

		bytesConsumed, err := r.parse(fullReqData)
		if err != nil {
			return nil, err
		}

		if bytesConsumed > 0 {
			fullReqData = fullReqData[bytesConsumed:]
		}

		if r.state == stateDone {
			return r, nil
		}

		if err == io.EOF {
			if len(fullReqData) > 0 {
				return nil, errors.New("unexpected EOF: incomplete request")
			}
			break
		}

		if err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesConsumeds := 0

	switch r.state {
	case stateInitialized:
		rl, bytesConsumed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if bytesConsumed == 0 {
			return 0, nil
		}

		r.RequestLine = *rl
		r.state = stateDone
		totalBytesConsumeds = bytesConsumed
	}

	return totalBytesConsumeds, nil
}

func parseRequestLine(rawReq []byte) (*RequestLine, int, error) {
	idx := bytes.Index(rawReq, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}

	line := string(rawReq[:idx])
	rlParts := strings.Split(line, " ")
	if len(rlParts) != 3 {
		return nil, 0, errors.New("invalid request line")
	}

	method := rlParts[0]
	target := rlParts[1]
	versionStr := rlParts[2]

	if !isValidHttpMethod(method) {
		return nil, 0, errors.New("invalid http method")
	}

	version, err := parseHttpVerion(versionStr)
	if err != nil {
		return nil, 0, err
	}

	// Plus two bytes for the \r\n
	return &RequestLine{
		Method:        HttpMethod(method),
		HttpVersion:   version,
		RequestTarget: target,
	}, idx + 2, nil
}

func isValidHttpMethod(v string) bool {
	switch HttpMethod(v) {
	case GET, POST, PUT, PATCH, DELETE, OPTIONS, TRACE:
		return true
	default:
		return false
	}
}

func parseHttpVerion(v string) (string, error) {
	parts := strings.Split(v, "/")
	if len(parts) != 2 {
		return "", errors.New("invalid http version formats")
	}

	if strings.TrimSpace(parts[1]) != "1.1" {
		return "", fmt.Errorf("unsupported http version")
	}

	return parts[1], nil
}
