package aws

import (
	"io/ioutil"
	"log"
	"os"
)

func init() {
	// Disable logging unless debugging, otherwise resource configuration is written to the logs
	val, ok := os.LookupEnv("TF_LOG")
	if !ok || val == "" {
		log.SetOutput(ioutil.Discard)
	}
}
