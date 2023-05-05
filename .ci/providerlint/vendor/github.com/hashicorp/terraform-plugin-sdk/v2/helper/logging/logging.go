package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/hashicorp/logutils"
	testing "github.com/mitchellh/go-testing-interface"
)

// These are the environmental variables that determine if we log, and if
// we log whether or not the log should go to a file.
const (
	EnvLog        = "TF_LOG"          // See ValidLevels
	EnvLogFile    = "TF_LOG_PATH"     // Set to a file
	EnvAccLogFile = "TF_ACC_LOG_PATH" // Set to a file
	// EnvLogPathMask splits test log files by name.
	EnvLogPathMask = "TF_LOG_PATH_MASK"
)

var ValidLevels = []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"}

// LogOutput determines where we should send logs (if anywhere) and the log
// level. This only effects this log.Print* functions called in the provider
// under test. Dependency providers for the provider under test will have their
// logging controlled by Terraform itself and managed with the TF_ACC_LOG_PATH
// environment variable. Calls to tflog.* will have their output managed by the
// tfsdklog sink.
func LogOutput(t testing.T) (logOutput io.Writer, err error) {
	logOutput = io.Discard

	logLevel := LogLevel()
	if logLevel == "" {
		if os.Getenv(EnvAccLogFile) != "" {
			// plugintest defaults to TRACE when TF_ACC_LOG_PATH is
			// set for Terraform and dependency providers of the
			// provider under test. We should do the same for the
			// provider under test.
			logLevel = "TRACE"
		} else {
			return
		}
	}

	logOutput = os.Stderr
	if logPath := os.Getenv(EnvLogFile); logPath != "" {
		var err error
		logOutput, err = os.OpenFile(logPath, syscall.O_CREAT|syscall.O_RDWR|syscall.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
	}

	if logPath := os.Getenv(EnvAccLogFile); logPath != "" {
		var err error
		logOutput, err = os.OpenFile(logPath, syscall.O_CREAT|syscall.O_RDWR|syscall.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
	}

	if logPathMask := os.Getenv(EnvLogPathMask); logPathMask != "" {
		// Escape special characters which may appear if we have subtests
		testName := strings.Replace(t.Name(), "/", "__", -1)

		logPath := fmt.Sprintf(logPathMask, testName)
		var err error
		logOutput, err = os.OpenFile(logPath, syscall.O_CREAT|syscall.O_RDWR|syscall.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
	}

	// This was the default since the beginning
	logOutput = &logutils.LevelFilter{
		Levels:   ValidLevels,
		MinLevel: logutils.LogLevel(logLevel),
		Writer:   logOutput,
	}

	return
}

// SetOutput checks for a log destination with LogOutput, and calls
// log.SetOutput with the result. If LogOutput returns nil, SetOutput uses
// io.Discard. Any error from LogOutout is fatal.
func SetOutput(t testing.T) {
	out, err := LogOutput(t)
	if err != nil {
		log.Fatal(err)
	}

	if out == nil {
		out = io.Discard
	}

	log.SetOutput(out)
}

// LogLevel returns the current log level string based the environment vars
func LogLevel() string {
	envLevel := os.Getenv(EnvLog)
	if envLevel == "" {
		return ""
	}

	logLevel := "TRACE"
	if isValidLogLevel(envLevel) {
		// allow following for better ux: info, Info or INFO
		logLevel = strings.ToUpper(envLevel)
	} else {
		log.Printf("[WARN] Invalid log level: %q. Defaulting to level: TRACE. Valid levels are: %+v",
			envLevel, ValidLevels)
	}

	return logLevel
}

// IsDebugOrHigher returns whether or not the current log level is debug or trace
func IsDebugOrHigher() bool {
	level := LogLevel()
	return level == "DEBUG" || level == "TRACE"
}

func isValidLogLevel(level string) bool {
	for _, l := range ValidLevels {
		if strings.ToUpper(level) == string(l) {
			return true
		}
	}

	return false
}
