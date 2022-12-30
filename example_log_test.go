package logwriter_test

import (
	"github.com/yuuki0xff/go-logwriter"
	"log"
	"os"
)

// An example integrates log and logwriter.
func Example_log() {
	log.Println("This message is written to default destination.")
	defer log.Println("This message is written to default destination.")

	opt := logwriter.DefaultOpenOption
	opt.FileOrDir = os.TempDir()     // Please change path to correct directory where you want log files to be placed.
	opt.Prefix = "logwriter-example" // Alternatively, you can use auto-detected name.
	tearDownLogger, err := logwriter.Setup(opt)
	if err != nil {
		panic(err)
	}
	defer tearDownLogger()

	log.Print("This message is written to /tmp/logwriter-example.*.log.zst within a second.")
	// Output:
}
