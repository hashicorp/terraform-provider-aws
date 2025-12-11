// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package AWSR001

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAWSR001(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "testdata/src/a")
}
