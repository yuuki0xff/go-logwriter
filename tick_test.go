package logwriter

import (
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

func newMockWriter() *mockWriter {
	return &mockWriter{
		wroteCh: make(chan struct{}, 1),
	}
}

type mockWriter struct {
	wroteCh chan struct{}
	// 0 - no method called
	// 1 - Write() called
	// 2 - Close() called
	flag atomic.Uint32
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	m.flag.CompareAndSwap(0, 1)
	select {
	case m.wroteCh <- struct{}{}:
		return
	default:
		return
	}
}

func (m *mockWriter) Close() error {
	m.flag.CompareAndSwap(0, 2)
	return nil
}

func TestTickWriter_worker(t1 *testing.T) {
	t1.Run("close before write", func(t *testing.T) {
		mw := newMockWriter()
		tw := NewTickWriter(mw, time.Second)
		assert.NoError(t, tw.Close())
		assert.Equal(t, uint32(2), mw.flag.Load())
	})
	t1.Run("write and close", func(t *testing.T) {
		mw := newMockWriter()
		tw := NewTickWriter(mw, time.Second)
		_, err := tw.Write([]byte("foo"))
		assert.NoError(t, err)
		assert.NoError(t, tw.Close())
		assert.Equal(t, uint32(1), mw.flag.Load())
	})
	t1.Run("wait for write", func(t *testing.T) {
		mw := newMockWriter()
		tw := NewTickWriter(mw, time.Second)
		timeout := time.NewTimer(2 * time.Second)
		select {
		case <-mw.wroteCh:
		// ok
		case <-timeout.C:
			assert.FailNow(t, "timeout exceeded")
		}
		assert.NoError(t, tw.Close())
		assert.Equal(t, uint32(1), mw.flag.Load())
	})
	t1.Run("should not write after close", func(t *testing.T) {
		mw := newMockWriter()
		tw := NewTickWriter(mw, time.Second)
		assert.NoError(t, tw.Close())
		assert.Equal(t, uint32(2), mw.flag.Load())
		timeout := time.NewTimer(2 * time.Second)
		select {
		case <-mw.wroteCh:
			assert.FailNow(t, "write() called after close")
		case <-timeout.C:
			// ok
		}
	})
}
