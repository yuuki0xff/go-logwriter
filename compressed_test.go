package logwriter

import (
	"crypto/sha1"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"sync"
	"testing"
)

func FuzzCompressedWriter_Write(f *testing.F) {
	f.Add([]byte(nil))
	f.Add([]byte(""))
	f.Add([]byte("aaaa"))
	f.Fuzz(func(t *testing.T, data []byte) {
		r, w, err := os.Pipe()
		if !assert.NoError(t, err) {
			return
		}

		dataHash := sha1.Sum(data)
		wg := sync.WaitGroup{}
		defer wg.Wait()
		wg.Add(2)
		go func() {
			defer wg.Done()
			cw := NewCompressedWriter(w, &NopAlgorithm{})
			defer cw.Close()
			n, err := cw.Write(data)
			assert.NoError(t, err)
			assert.Equal(t, len(data), n)
			assert.NoError(t, cw.Close())
		}()
		go func() {
			defer wg.Done()
			received, err := io.ReadAll(r)
			defer r.Close()
			assert.NoError(t, err)
			receivedHash := sha1.Sum(received)
			assert.Equal(t, dataHash, receivedHash)
		}()
	})
}
