package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
)

type parsingState string

type Request struct {
	RequestLine  RequestLine
	parsingState parsingState
}

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

const (
	parsingStateInitialized parsingState = "initialized"
	parsingStateDone        parsingState = "done"
)

const BUFFERSIZE = 8

const SEPERATOR = "\r\n"

func (r *Request) parse(data []byte) (int, error) {
	switch r.parsingState {
	case parsingStateInitialized:
		requestLine, bytesConsumed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if bytesConsumed == 0 {
			return 0, nil
		}

		r.parsingState = parsingStateDone
		r.RequestLine = *requestLine

		return bytesConsumed, nil
	case parsingStateDone:
		return 0, errors.New("error: trying to read data in a done state")
	default:

		return 0, errors.New("error: unknown state")
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buffer := make([]byte, BUFFERSIZE)
	r := &Request{
		parsingState: parsingStateInitialized,
	}

	readToIdx := 0

	for r.parsingState != parsingStateDone {

		// if our buffer size is full, copy to a larger array
		if len(buffer) == cap(buffer) {
			newBuffer := make([]byte, 2*cap(buffer))
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		readCount, err := reader.Read(buffer[readToIdx:])

		if err == io.EOF {
			r.parsingState = parsingStateDone
			break
		}

		if err != nil {
			return nil, fmt.Errorf("error reading to buffer: %w", err)
		}

		readToIdx += readCount

		log.Printf("buffer: %s\n", string(buffer))
		log.Printf("Read from reader: %s\n", string(buffer[:readCount]))

		parsedCount, err := r.parse(buffer[:readToIdx])
		if err != nil {
			return nil, fmt.Errorf("error parsing the request: %w", err)
		}

		if parsedCount > 0 {
			removeParsedBuffer := make([]byte, cap(buffer))
			copy(removeParsedBuffer, buffer[parsedCount:])
			buffer = removeParsedBuffer
			readToIdx -= parsedCount
		}

	}

	return r, nil
}

// GET /coffee HTTP/1.1 - example valid request line
func parseRequestLine(requestData []byte) (*RequestLine, int, error) {
	idx := bytes.Index(requestData, []byte(SEPERATOR))

	if idx == -1 {
		return nil, 0, nil // still needs more data before we can continue
	}

	startLine := requestData[:idx]

	fmt.Printf("startLine: %s", string(startLine))

	requestLineParts := strings.Split(string(startLine), " ")

	if len(requestLineParts) != 3 {
		return nil, 0, errors.New("incorrect number of parts in the request line")
	}

	method := requestLineParts[0]

	// Validate the the method
	if bytes.ContainsAny([]byte(method), "abcdefghijklmnopqrstuvwxyz") {
		return nil, 0, errors.New("method part in the request line cannot contains any lowercase")
	}

	requestTarget := requestLineParts[1]

	// Velidate the version
	httpVersionPart := requestLineParts[2]

	httpVersionParts := strings.Split(httpVersionPart, "/")

	if len(httpVersionParts) != 2 {
		return nil, 0, errors.New("invalid http version format in the request line")
	}

	version := httpVersionParts[1]
	if version != "1.1" {
		return nil, 0, errors.New("only http 1.1 is supported for now")
	}

	return &RequestLine{
		Method:        method,
		HTTPVersion:   version,
		RequestTarget: requestTarget,
	}, len(startLine), nil
}
