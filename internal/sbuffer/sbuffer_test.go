package sbuffer

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"time"
)

// mockReader implements ConnReadDeadline for testing.
type mockReader struct {
	data     []byte
	pos      int
	eofAfter int // return EOF after this many bytes (-1 = never)
	errAfter int // return custom error after this many bytes (-1 = never)
	err      error
	deadline time.Time
}

func (m *mockReader) Read(p []byte) (int, error) {
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}
	n := copy(p, m.data[m.pos:])
	m.pos += n

	if m.eofAfter >= 0 && m.pos >= m.eofAfter {
		return n, io.EOF
	}
	if m.errAfter >= 0 && m.pos >= m.errAfter {
		return n, m.err
	}
	return n, nil
}

func (m *mockReader) SetReadDeadline(t time.Time) error {
	m.deadline = t
	return nil
}

func TestNextBasicRead(t *testing.T) {
	data := []byte("hello world")
	reader := &mockReader{data: data, eofAfter: -1, errAfter: -1}
	buf := NewBuffer(reader, 64)

	ctx := context.Background()
	out, err := buf.Next(ctx, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "hello" {
		t.Errorf("expected 'hello', got %q", string(out))
	}
}

func TestNextMultipleReads(t *testing.T) {
	data := []byte("hello world test data")
	reader := &mockReader{data: data, eofAfter: -1, errAfter: -1}
	buf := NewBuffer(reader, 64)

	ctx := context.Background()

	// First read
	out, err := buf.Next(ctx, 6)
	if err != nil {
		t.Fatalf("unexpected error on first read: %v", err)
	}
	if string(out) != "hello " {
		t.Errorf("expected 'hello ', got %q", string(out))
	}
	buf.Used(6)

	// Second read
	out, err = buf.Next(ctx, 6)
	if err != nil {
		t.Fatalf("unexpected error on second read: %v", err)
	}
	if string(out) != "world " {
		t.Errorf("expected 'world ', got %q", string(out))
	}
	buf.Used(6)

	// Third read
	out, err = buf.Next(ctx, 4)
	if err != nil {
		t.Fatalf("unexpected error on third read: %v", err)
	}
	if string(out) != "test" {
		t.Errorf("expected 'test', got %q", string(out))
	}
}

func TestNextUsedPartially(t *testing.T) {
	data := []byte("hello world")
	reader := &mockReader{data: data, eofAfter: -1, errAfter: -1}
	buf := NewBuffer(reader, 64)

	ctx := context.Background()

	// Request 8 bytes
	out, err := buf.Next(ctx, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "hello wo" {
		t.Errorf("expected 'hello wo', got %q", string(out))
	}

	// Only use 5 bytes
	buf.Used(5)

	// Next read should start from position 5
	out, err = buf.Next(ctx, 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != " world" {
		t.Errorf("expected ' world', got %q", string(out))
	}
}

func TestNextBufferWrapAround(t *testing.T) {
	data := []byte("abcdefghijklmnopqrstuvwxyz")
	reader := &mockReader{data: data, eofAfter: -1, errAfter: -1}
	buf := NewBuffer(reader, 16) // Small buffer to force wraparound

	ctx := context.Background()

	// Read and use 10 bytes
	out, err := buf.Next(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "abcdefghij" {
		t.Errorf("expected 'abcdefghij', got %q", string(out))
	}
	buf.Used(10)

	// Now tail=10, head=16 (or more). Request 10 more bytes.
	// This should trigger a copy since tail+needed >= len(backer)
	out, err = buf.Next(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "klmnopqrst" {
		t.Errorf("expected 'klmnopqrst', got %q", string(out))
	}
}

func TestNextShortRead(t *testing.T) {
	data := []byte("short")
	reader := &mockReader{data: data, eofAfter: -1, errAfter: -1}
	buf := NewBuffer(reader, 64)

	ctx := context.Background()

	// Request more than available
	out, err := buf.Next(ctx, 10)
	if err == nil {
		t.Fatal("expected error for short read")
	}
	if len(out) != 5 {
		t.Errorf("expected 5 bytes, got %d", len(out))
	}
	if string(out) != "short" {
		t.Errorf("expected 'short', got %q", string(out))
	}
}

func TestNextWithEOF(t *testing.T) {
	data := []byte("hello")
	reader := &mockReader{data: data, eofAfter: 5, errAfter: -1}
	buf := NewBuffer(reader, 64)

	ctx := context.Background()

	// Read exactly 5 bytes - should succeed with EOF
	out, err := buf.Next(ctx, 5)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "hello" {
		t.Errorf("expected 'hello', got %q", string(out))
	}
}

func TestNextContextCanceled(t *testing.T) {
	// Create a reader that blocks
	data := []byte("hello")
	reader := &mockReader{data: data, eofAfter: -1, errAfter: -1}
	buf := NewBuffer(reader, 64)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Request more data than available - reader will return EOF and context is canceled
	_, err := buf.Next(ctx, 10)
	if err == nil {
		t.Fatal("expected error due to short read/context")
	}
}

func TestUsedPanicsOnOveruse(t *testing.T) {
	data := []byte("hello world")
	reader := &mockReader{data: data, eofAfter: -1, errAfter: -1}
	buf := NewBuffer(reader, 64)

	ctx := context.Background()

	// Read 5 bytes
	_, err := buf.Next(ctx, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Try to use more than we have
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on overuse")
		}
	}()
	buf.Used(100)
}

func TestNextPanicsOnOversizedRequest(t *testing.T) {
	data := []byte("hello")
	reader := &mockReader{data: data, eofAfter: -1, errAfter: -1}
	buf := NewBuffer(reader, 8) // Small buffer

	ctx := context.Background()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on oversized request")
		}
	}()
	buf.Next(ctx, 100) // Request more than buffer size
}

func TestUsedZero(t *testing.T) {
	data := []byte("hello world")
	reader := &mockReader{data: data, eofAfter: -1, errAfter: -1}
	buf := NewBuffer(reader, 64)

	ctx := context.Background()

	// Read 5 bytes
	out1, err := buf.Next(ctx, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Use zero bytes
	buf.Used(0)

	// Should get the same data again
	out2, err := buf.Next(ctx, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out1) != string(out2) {
		t.Errorf("expected same data, got %q and %q", string(out1), string(out2))
	}
}

func TestMultipleWrapArounds(t *testing.T) {
	// Create enough data for multiple wraparounds
	data := make([]byte, 100)
	for i := range data {
		data[i] = byte('a' + (i % 26))
	}
	reader := &mockReader{data: data, eofAfter: -1, errAfter: -1}
	buf := NewBuffer(reader, 20)

	ctx := context.Background()

	// Read in chunks that will cause multiple wraparounds
	totalRead := 0
	for i := range 10 {
		out, err := buf.Next(ctx, 8)
		if err != nil {
			t.Fatalf("unexpected error at iteration %d: %v", i, err)
		}
		if len(out) != 8 {
			t.Errorf("expected 8 bytes at iteration %d, got %d", i, len(out))
		}
		// Verify data is correct
		for j, b := range out {
			expected := byte('a' + ((totalRead + j) % 26))
			if b != expected {
				t.Errorf("at position %d, expected %c, got %c", totalRead+j, expected, b)
			}
		}
		buf.Used(8)
		totalRead += 8
	}
}

func TestReadExactBufferSize(t *testing.T) {
	data := []byte("12345678")
	reader := &mockReader{data: data, eofAfter: -1, errAfter: -1}
	buf := NewBuffer(reader, 8)

	ctx := context.Background()

	// Read exactly the buffer size
	out, err := buf.Next(ctx, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "12345678" {
		t.Errorf("expected '12345678', got %q", string(out))
	}
}

func TestReadWithCustomError(t *testing.T) {
	customErr := errors.New("custom read error")
	data := []byte("hello")
	reader := &mockReader{data: data, eofAfter: -1, errAfter: 3, err: customErr}
	buf := NewBuffer(reader, 64)

	ctx := context.Background()

	// Request 5 bytes - should get error after reading 3
	out, err := buf.Next(ctx, 5)
	// Depending on timing, we might get partial data with error
	if len(out) < 3 {
		t.Errorf("expected at least 3 bytes, got %d", len(out))
	}
	if err == nil {
		t.Error("expected an error")
	}
}

// deadlineReader returns os.ErrDeadlineExceeded for the first N reads,
// then returns data normally.
type deadlineReader struct {
	data            []byte
	pos             int
	deadlineReturns int // how many times to return ErrDeadlineExceeded
	readCount       int
}

func (d *deadlineReader) Read(p []byte) (int, error) {
	d.readCount++
	if d.readCount <= d.deadlineReturns {
		return 0, os.ErrDeadlineExceeded
	}
	if d.pos >= len(d.data) {
		return 0, io.EOF
	}
	n := copy(p, d.data[d.pos:])
	d.pos += n
	return n, nil
}

func (d *deadlineReader) SetReadDeadline(t time.Time) error {
	return nil
}

func TestNextDeadlineExceededRetry(t *testing.T) {
	data := []byte("hello world")
	reader := &deadlineReader{data: data, deadlineReturns: 3}
	buf := NewBuffer(reader, 64)

	ctx := context.Background()

	// Should succeed despite first 3 reads returning ErrDeadlineExceeded
	out, err := buf.Next(ctx, 11)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(out))
	}
	if reader.readCount != 4 {
		t.Errorf("expected 4 read attempts, got %d", reader.readCount)
	}
}
