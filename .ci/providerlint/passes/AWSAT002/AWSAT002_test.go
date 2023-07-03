// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package AWSAT002_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/ci/providerlint/passes/AWSAT002"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAWSAT002(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, AWSAT002.Analyzer, "a")
}
