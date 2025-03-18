package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chunkReader simulates reading from a network connection by returning a fixed number of bytes per Read call.
type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read fills the provided p slice with data from the chunkReader,
// up to a maximum of numBytesPerRead bytes at a time.
// we are trying to mimic a streamer.
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		// 0 > 90
		// 3 > 90
		return 0, io.EOF
	}

	endIndex := min(cr.pos+cr.numBytesPerRead, len(cr.data))
	// eg: pos = 0, if nBytes = 3 and len(data) = 90, then:
	// endIndex = min(0+3, 90) = 3
	// endIndex = min(3+3, 90) = 6

	n = copy(p, cr.data[cr.pos:endIndex])
	// n = copy([], data[0:3]) hence, n = 3
	// n = copy([], data[3:6]) hance, n = 3 (overwrites)

	cr.pos += n
	// pos = pos + 3 = 0 + 3 = 3
	// pos = pos + 3 = 3 + 3 = 6

	return n, nil
}

func TestRequestLineParse(t *testing.T) {

	// numBytesPerRead = 3
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "HTTP/1.1", r.RequestLine.HttpVersion)
}
