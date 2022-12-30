package logwriter

import (
	"bytes"
	"io"
	"os"
	"time"
)

func NewBuffer(size int, interval time.Duration, w io.WriteCloser) io.WriteCloser {
	return newBuffer(size, interval, w, time.Now)
}

func newBuffer(size int, interval time.Duration, w io.WriteCloser, now func() time.Time) io.WriteCloser {
	return &Buffer{
		Size:      size,
		Interval:  interval,
		Now:       now,
		w:         w,
		lastFlush: now(),
	}
}

// Buffer implements buffered io.Writer like bufio.Writer, but specializing in log writer.
type Buffer struct {
	Size     int
	Interval time.Duration
	// Now() is a function returns current time like time.Time().
	// This function used to inject time from outside.
	Now       func() time.Time
	w         io.WriteCloser
	buf       bytes.Buffer
	err       error
	lastFlush time.Time
	closed    bool
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	n, err = b.write(p)
	return
}

func (b *Buffer) Close() (err error) {
	err = b.close()
	return
}

func (b *Buffer) flush() {
	if b.err != nil {
		return
	}
	n := b.buf.Len()
	var wrote int
	if n > 0 {
		wrote, b.err = b.w.Write(b.buf.Bytes())
	}
	b.lastFlush = b.Now()
	b.buf.Reset()
	if b.err == nil && wrote < n {
		b.err = io.ErrShortWrite
	}
	return
}

func (b *Buffer) needFlush() bool {
	expired := b.Now().Sub(b.lastFlush).Abs().Nanoseconds() >= b.Interval.Nanoseconds()
	overflow := b.Size <= b.buf.Len()
	return expired || overflow
}

func (b *Buffer) write(p []byte) (int, error) {
	if b.closed {
		return 0, os.ErrClosed
	}
	if b.Size <= len(p) {
		return b.largeWrite(p)
	}
	return b.smallWrite(p)
}

func (b *Buffer) smallWrite(p []byte) (int, error) {
	var n int
	if b.err == nil && 0 < len(p) {
		// This operation may require a buffer space of twice b.Size.
		n, b.err = b.buf.Write(p)
	}
	if b.err == nil && b.needFlush() {
		b.flush()
	}
	return n, b.err
}

func (b *Buffer) largeWrite(p []byte) (int, error) {
	b.flush()
	if b.err == nil && b.buf.Len() != 0 {
		panic("bug: internal buffer is not empty even though it was flushed")
	}
	var wrote int
	if b.err == nil {
		wrote, b.err = b.w.Write(p)
		b.lastFlush = b.Now()
	}
	if b.err == nil && wrote < len(p) {
		b.err = io.ErrShortWrite
	}
	return wrote, b.err
}

func (b *Buffer) close() error {
	if b.closed {
		return os.ErrClosed
	}
	b.flush()
	if b.err == nil {
		b.err = b.w.Close()
		b.closed = true
	}
	return b.err
}
