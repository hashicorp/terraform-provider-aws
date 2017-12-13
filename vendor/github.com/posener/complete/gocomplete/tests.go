package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/posener/complete"
)

var (
	predictBenchmark = funcPredict(regexp.MustCompile("^Benchmark"))
	predictTest      = funcPredict(regexp.MustCompile("^(Test|Example)"))
)

// predictTest predict test names.
// it searches in the current directory for all the go test files
// and then all the relevant function names.
// for test names use prefix of 'Test' or 'Example', and for benchmark
// test names use 'Benchmark'
func funcPredict(funcRegexp *regexp.Regexp) complete.Predictor {
	return complete.PredictFunc(func(a complete.Args) []string {
		return funcNames(funcRegexp)
	})
}

// get all test names in current directory
func funcNames(funcRegexp *regexp.Regexp) (tests []string) {
	filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
		// if not a test file, skip
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		// inspect test file and append all the test names
		tests = append(tests, functionsInFile(path, funcRegexp)...)
		return nil
	})
	return
}
