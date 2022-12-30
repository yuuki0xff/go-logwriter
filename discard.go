package logwriter

import (
	"io"
	"os"
)

var Discard io.WriteCloser = &discard{}

type discard struct {
	closed bool
}

func (d *discard) Write(p []byte) (n int, err error) {
	if d.closed {
		return 0, os.ErrClosed
	}
	return io.Discard.Write(p)
}

func (d *discard) Close() error {
	d.closed = true
	return nil
}
