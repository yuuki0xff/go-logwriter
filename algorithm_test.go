package logwriter

import (
	"bytes"
	"crypto/sha1"
	"github.com/stretchr/testify/assert"
	"testing"
)

func testAlgorithm(t *testing.T, a Algorithm, plain []byte) {
	hash := sha1.Sum(plain)

	compressedBuf := bytes.Buffer{}
	compressedBuf.Grow(len(plain))
	assert.NoError(t, a.Compress(plain, &compressedBuf))

	plainBuf := bytes.Buffer{}
	assert.NoError(t, a.Decompress(compressedBuf.Bytes(), &plainBuf))
	plain2 := plainBuf.Bytes()
	hash2 := sha1.Sum(plain2)

	assert.Equal(t, hash, hash2)
}

func fuzzAlgorithm(f *testing.F, a Algorithm) {
	f.Add([]byte(""))
	f.Add([]byte("aaaa"))
	f.Fuzz(func(t *testing.T, src []byte) {
		testAlgorithm(t, a, src)
	})
}

func FuzzNopAlgorithm_Compress(f *testing.F) {
	a := &NopAlgorithm{}
	fuzzAlgorithm(f, a)
}

func FuzzGzipAlgorithm_Compress(f *testing.F) {
	a := &GzipAlgorithm{}
	fuzzAlgorithm(f, a)
}

func FuzzZstdAlgorithm_Compress(f *testing.F) {
	a := &ZstdAlgorithm{}
	fuzzAlgorithm(f, a)
}
