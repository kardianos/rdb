// Single-Buffer backing for readers.
//
// Use when reading from a large Reader when only small defined
// sequential slices are needed. Uses a single buffer for reading.
package sbuffer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

var (
	ErrNeedCap     = errors.New("Requested more than buffer size.")
	ErrUsedTooMuch = errors.New("Used more than requested.")
)

type buffer struct {
	ctxCheck time.Duration
	reader   ConnReadDeadline
	backer   []byte

	tail, head int
}

type Buffer interface {
	// Get the next slice of data. Will return at least "needed", but may return more.
	// The returned slice should not be used after Next or Used is called.
	// The returned slice may be shorter then needed if the inner read returns a short
	// read (for example, returns an io.EOF).
	Next(ctx context.Context, needed int) ([]byte, error)

	// Called after using the slice returned from Next. This frees
	// the underlying buffer for more data.
	Used(used int)
}

type ConnReadDeadline interface {
	io.Reader
	SetReadDeadline(t time.Time) error
}

const readContextCheckPeriod = time.Millisecond * 120

// Read from read for more data.
// The bufferSize should be several times the max read size to prevent excessive copying.
func NewBuffer(read ConnReadDeadline, bufferSize int) Buffer {
	return &buffer{
		reader: read,
		backer: make([]byte, bufferSize),
	}
}

func (b *buffer) Next(ctx context.Context, needed int) ([]byte, error) {
	if needed > len(b.backer) {
		panic(ErrNeedCap)
	}
	if b.tail+needed >= len(b.backer) {
		// Copy end of tail to beginning of buffer.
		b.head = copy(b.backer, b.backer[b.tail:b.head])
		b.tail = 0
	}
	var err error
	var n int
	for b.tail+needed > b.head {
		// Read more data.
		b.reader.SetReadDeadline(time.Now().Add(readContextCheckPeriod))
		n, err = b.reader.Read(b.backer[b.head:])
		b.head += n
		if err != nil {
			if cerr := ctx.Err(); cerr != nil {
				err = fmt.Errorf("Next read: %w & %w", err, cerr)
			} else if errors.Is(err, os.ErrDeadlineExceeded) {
				err = nil
				continue
			}
			break
		}
	}
	out := b.backer[b.tail:min(b.tail+needed, b.head)]
	if len(out) != needed {
		return out, fmt.Errorf("requested %d, but received %d bytes", needed, len(out))
	}
	return out, err
}

func (b *buffer) Used(used int) {
	b.tail += used
	if b.tail > b.head {
		panic(ErrUsedTooMuch)
	}
}
