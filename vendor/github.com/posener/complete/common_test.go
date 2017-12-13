package complete

import (
	"os"
	"sync"
	"testing"
)

var once = sync.Once{}

func initTests() {
	once.Do(func() {
		// Set debug environment variable so logs will be printed
		if testing.Verbose() {
			os.Setenv(envDebug, "1")
			// refresh the logger with environment variable set
			Log = getLogger()
		}

		// Change to tests directory for testing completion of files and directories
		err := os.Chdir("./tests")
		if err != nil {
			panic(err)
		}
	})
}
