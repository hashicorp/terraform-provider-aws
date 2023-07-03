// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package AWSAT003_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/ci/providerlint/passes/AWSAT003"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAWSAT003(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, AWSAT003.Analyzer, "a")
}
