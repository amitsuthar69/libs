/*
HTTP/1.1 request parser

format:

	request-line CRLF
	\*( field-line CRLF )
	CRLF
	[ message-body ]

eg:

	GET /goodies HTTP/1.1    # request-line CRLF
	Host: localhost:42069    # *( field-line CRLF )
	User-Agent: curl/7.81.0  # *( field-line CRLF )
	Accept:  /               # *( field-line CRLF )
	# CRLF
	# [ message-body ] (empty)
*/
package request

import (
	"fmt"
	"io"
	"strings"
)

// ParserState represents the current state of the parser.
type ParserState int

const (
	Initialized ParserState = iota
	Done
)

// RequestLine represents the first line of an HTTP request message.
type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

type Request struct {
	RequestLine RequestLine
	state       ParserState
}

// parse processes incoming data and attempts to parse the request line.
// It returns the number of bytes consumed from the data slice.
// If the CRLF is not found, it returns 0, nil to indicate that more data is needed.
func (r *Request) parse(data []byte) (int, error) {
	if r.state == Done {
		return 0, fmt.Errorf("error: trying to read data in a done state")
	}

	dataStr := string(data)

	// looking for CRLF to determine if we have a full line.
	index := strings.Index(dataStr, "\r\n")
	if index == -1 {
		// incomplete request line; need more data.
		return 0, nil
	}

	// extracting the line (excluding CRLF).
	line := dataStr[:index]
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid request line format: %s", line)
	}

	// method has to be in capital letters.
	for _, ch := range parts[0] {
		if ch < 'A' || ch > 'Z' {
			return 0, fmt.Errorf("invalid method: %s", parts[0])
		}
	}

	// explicit version matching.
	if parts[2] != "HTTP/1.1" {
		return 0, fmt.Errorf("invalid version: %s", parts[2])
	}

	// Update the RequestLine field.
	r.RequestLine = RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   parts[2],
	}

	r.state = Done

	// returning number of bytes consumed (line length + 2 for CRLF).
	return index + 2, nil
}

const bufferSize = 8

// RequestFromReader reads from the io.Reader in small chunks
// and uses the parse method to build the request.
func RequestFromReader(r io.Reader) (*Request, error) {
	buffer := make([]byte, bufferSize)
	idx := 0
	req := &Request{state: Initialized}

	// while we are not done reading
	for req.state != Done {
		// if buffer gets full, grow it.
		if idx == len(buffer) {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		// read data into the buffer.
		n, err := r.Read(buffer[idx:]) // from idx to end (inclusive)
		if err != nil && err != io.EOF {
			return nil, err
		}
		idx += n

		// attempt to parse the available data.
		consumed, parseErr := req.parse(buffer[:idx]) // from start to idx (exclusive)
		if parseErr != nil {
			return nil, parseErr
		}
		if consumed > 0 {
			// shift the buffer to remove the parsed data.
			copy(buffer, buffer[consumed:idx])
			idx -= consumed
		}

		if err == io.EOF && req.state != Done {
			return nil, fmt.Errorf("incomplete request, reached EOF")
		}
	}

	return req, nil
}
