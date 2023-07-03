// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package AWSAT006_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/ci/providerlint/passes/AWSAT006"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAWSAT006(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, AWSAT006.Analyzer, "a")
}
