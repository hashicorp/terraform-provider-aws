// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package AWSV001

import (
	"testing"

	_ "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAWSV001(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "a")
}
