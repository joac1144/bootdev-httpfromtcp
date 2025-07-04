package request

import (
	"errors"
	"io"
	"regexp"
	"strings"
)

const bufferSize int = 8

type requestState int

type Request struct {
	RequestLine RequestLine
	state       requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const (
	initialized requestState = iota
	done
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	buffer := make([]byte, bufferSize)
	readToIndex := 0
	request := &Request{
		state: initialized,
	}

	for request.state != done {
		if readToIndex >= len(buffer) {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		bytesReadFromReader, err := reader.Read(buffer[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				request.state = done
				break
			}
			return nil, err
		}

		readToIndex += bytesReadFromReader

		bytesParsed, err := request.parse(buffer[:readToIndex])
		if err != nil {
			return nil, err
		}
		if bytesParsed > 0 {
			newBuffer := make([]byte, len(buffer)-bytesParsed)
			copy(newBuffer, buffer[bytesParsed:readToIndex])
			buffer = newBuffer
			readToIndex -= bytesParsed
		}
	}

	return request, nil
}

func parseRequestLine(data []byte) (RequestLine, int, error) {
	lines := strings.Split(string(data), "\r\n")

	if len(lines) < 2 {
		return RequestLine{}, 0, nil
	}

	startLine := lines[0]

	parts := strings.Split(startLine, " ")
	if len(parts) != 3 {
		return RequestLine{}, 0, errors.New("invalid request line format")
	}

	method := parts[0]
	methodRegex := `^[A-Z]+$`
	if matched, err := regexp.MatchString(methodRegex, method); err != nil || !matched {
		return RequestLine{}, 0, errors.New("invalid HTTP method")
	}

	httpVersion := parts[2]
	if httpVersion != "HTTP/1.1" {
		return RequestLine{}, 0, errors.New("unsupported HTTP version, only HTTP/1.1 is supported")
	}

	return RequestLine{
		Method:        method,
		RequestTarget: parts[1],
		HttpVersion:   strings.Split(httpVersion, "/")[1],
	}, len(startLine) + 2, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case initialized:
		requestLine, bytesConsumed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if bytesConsumed == 0 {
			return 0, nil
		}

		r.RequestLine = requestLine
		r.state = done
		return bytesConsumed, nil
	case done:
		return 0, errors.New("error: trying to read data in a done state")
	}

	return 0, errors.New("error: unknown state")
}
