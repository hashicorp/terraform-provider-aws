package AWSAT005_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/awsproviderlint/passes/AWSAT005"
	"golang.org/x/tools/go/analysis/analysistest"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAWSAT005(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, AWSAT005.Analyzer, "a")
}
