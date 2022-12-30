package logwriter

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_discard_Write(t *testing.T) {
	n, err := Discard.Write(nil)
	assert.Equal(t, 0, n)
	assert.NoError(t, err)

	n, err = Discard.Write([]byte(""))
	assert.Equal(t, 0, n)
	assert.NoError(t, err)

	n, err = Discard.Write([]byte("hello"))
	assert.Equal(t, 5, n)
	assert.NoError(t, err)

	assert.NoError(t, Discard.Close())
	n, err = Discard.Write([]byte("write to closed file"))
	assert.Equal(t, 0, n)
	assert.Equal(t, os.ErrClosed, err)

}
