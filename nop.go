package logwriter

import (
	"os"
)

type nopCloserFile struct {
	*os.File
}

func (s nopCloserFile) Close() error {
	// No operation.
	return nil
}
