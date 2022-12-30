package logwriter

import (
	"bytes"
	"io"
)

func NewCompressedWriter(w io.WriteCloser, a Algorithm) io.WriteCloser {
	return &CompressedWriter{
		w: w,
		a: a,
	}
}

// CompressedWriter compress data before writing data.
type CompressedWriter struct {
	w   io.WriteCloser
	a   Algorithm
	buf bytes.Buffer
}

func (c *CompressedWriter) Write(p []byte) (n int, err error) {
	err = c.a.Compress(p, &c.buf)
	if err == nil {
		_, err = c.w.Write(c.buf.Bytes())
	}
	c.buf.Reset()
	if err == nil {
		n = len(p)
	}
	return
}

func (c *CompressedWriter) Close() error {
	return c.w.Close()
}
