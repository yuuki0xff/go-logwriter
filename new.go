package logwriter

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type OpenOption struct {
	// Path to file or directory.
	// If empty string specified, discard all writes operations.
	// If "-" is specified, send to stderr.
	// If existing directory path specified, file name is generated automatically.
	FileOrDir string
	// Prefix for file name.
	// This option only affect if FileOrDir points to a directory.
	Prefix string
	// Additional file extensions.
	// This option only affect if FileOrDir points to a directory.
	//
	// Supported extensions list:
	//	".zst"
	//	".gz"
	//	"" (without compression)
	Suffix string
	// Flag for open a file.
	Flag int
	// Default file mode.
	Mode os.FileMode
	// BufferSize specifies the size of the buffer used for data compression and writing.
	// If BufferSize is not a positive value, buffering is disabled.
	BufferSize int
	// FlushInterval specifies the interval to flush the buffer.
	// If FlushInterval is not a positive value, buffering is disabled.
	FlushInterval time.Duration
}

var DefaultOpenOption = OpenOption{
	FileOrDir:     "-",
	Prefix:        filepath.Base(os.Args[0]),
	Suffix:        ".zst",
	Flag:          os.O_WRONLY | os.O_APPEND | os.O_CREATE,
	Mode:          0666,
	BufferSize:    os.Getpagesize(),
	FlushInterval: time.Second,
}

type TearDown func() error

func Setup(option OpenOption) (TearDown, error) {
	w, err := Open(option)
	if err != nil {
		return nil, err
	}

	old := log.Writer()
	log.SetOutput(w)
	return func() error {
		log.SetOutput(old)
		return w.Close()
	}, nil
}

func Open(option OpenOption) (io.WriteCloser, error) {
	w, ok := openFast(option)
	if ok {
		return w, nil
	}
	return openSlow(option)
}

func openFast(option OpenOption) (io.WriteCloser, bool) {
	p := option.FileOrDir
	if p == "" || p == "/dev/null" {
		// Drop all logs.
		return Discard, true
	}
	if p == "-" {
		// Send to stderr.
		// To prevent closing stderr unexpectedly, wraps os.Stderr by nopCloserFile.
		return &nopCloserFile{os.Stderr}, true
	}
	return nil, false
}

func openSlow(opt OpenOption) (io.WriteCloser, error) {
	filePath := opt.FileOrDir
	stat, err := os.Stat(opt.FileOrDir)
	if err != nil && !os.IsNotExist(err) {
		// Unexpected error occurred.
		return nil, err
	}
	if err == nil && stat.IsDir() {
		dir := opt.FileOrDir
		filename := fmt.Sprintf(
			"%s.%s-%d.log%s",
			opt.Prefix,
			time.Now().Format(time.RFC3339Nano), os.Getpid(),
			opt.Suffix,
		)
		filePath = path.Join(dir, filename)
	}
	// Ignore os.ErrNotExist.

	return openSuitableLogger(filePath, opt)
}

func openSuitableLogger(filePath string, opt OpenOption) (w io.WriteCloser, err error) {
	w, err = os.OpenFile(filePath, opt.Flag, opt.Mode)
	if err != nil {
		return nil, err
	}

	// Select compression algorithm
	if strings.HasSuffix(filePath, ".gz") {
		w = NewCompressedWriter(w, &GzipAlgorithm{})
	} else if strings.HasSuffix(filePath, ".zst") {
		w = NewCompressedWriter(w, &ZstdAlgorithm{})
	}

	if 0 < opt.BufferSize && 0 < opt.FlushInterval {
		// Add write buffer to improve compression efficiency.
		w = NewBuffer(opt.BufferSize, opt.FlushInterval, w)

		// Add tick writer to flush buffer periodically and protect the thread-unsafe WriteCloser object.
		w = NewTickWriter(w, opt.FlushInterval)
	} else {
		// No buffering.
		// Add tick writer to protect the thread-unsafe WriteCloser object.
		w = NewTickWriter(w, 0)
	}
	return
}
