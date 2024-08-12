package logwriter

import (
	"context"
	"io"
	"os"
	"sync"
	"time"
)

// NewTickWriter wraps the provided io.WriteCloser with a TickWriter.
// It calls w.Write(nil) at the specified interval to flush the buffer.
// Additionally, it protects w from concurrent Write and Close method calls.
func NewTickWriter(w io.WriteCloser, interval time.Duration) io.WriteCloser {
	ctx, cancel := context.WithCancel(context.Background())
	tw := &TickWriter{
		w:        w,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
	}
	if 0 < interval {
		go tw.worker()
	}
	return tw
}

// TickWriter calls w.Write(nil) every specified interval.
// Main use case is flushing write buffer in the Buffer.
//
//	var file io.WriteCloser
//	w = NewBuffer(0, time.Second, file)
//	w = TickWriter(w, time.Second)
//	defer w.Close()
type TickWriter struct {
	w        io.WriteCloser
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	mux      sync.Mutex
	closed   bool
}

func (t *TickWriter) Write(p []byte) (int, error) {
	t.mux.Lock()
	defer t.mux.Unlock()
	if t.closed {
		return 0, os.ErrClosed
	}
	return t.w.Write(p)
}

func (t *TickWriter) Close() error {
	t.mux.Lock()
	defer t.mux.Unlock()
	if t.closed {
		return os.ErrClosed
	}
	t.closed = true
	t.cancel()
	return t.w.Close()
}

func (t *TickWriter) worker() {
	timer := time.NewTimer(t.interval)
	for {
		select {
		case <-t.ctx.Done():
			goto stop
		case <-timer.C:
			timer.Reset(t.interval)
			t.Write(nil)
		}
	}

stop:
	if !timer.Stop() {
		<-timer.C
	}
}
