package logwriter_test

import (
	"fmt"
	"github.com/yuuki0xff/go-logwriter"
	"io"
	"os"
)

// An example using Open() with application log without standard logging system.
func Example_open() {
	opt := logwriter.DefaultOpenOption
	opt.FileOrDir = os.TempDir()     // Please change path to correct directory where you want log files to be placed.
	opt.Prefix = "logwriter-example" // Alternatively, you can use auto-detected name.
	w, err := logwriter.Open(opt)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	// w implemented io.WriteCloser interface.
	// Note that w is buffered. If the application exits without calling w.Close(), some data will be lost.

	// third_party_logging_framework.SetOutput(w)
	//   or
	// use w directly like this:
	w.Write([]byte("This message is written to /tmp/logwriter-example.*.log.zst within a second.\n"))
	fmt.Fprintf(w, "pid is %d.\n", os.Getpid())
	io.WriteString(w, "Application going to shutdown.\n")
	// Output:
}
