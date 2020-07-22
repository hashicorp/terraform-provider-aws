package AWSAT008_test

import (
	"testing"

	"github.com/terraform-providers/terraform-provider-aws/awsproviderlint/passes/AWSAT008"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAWSAT008(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, AWSAT008.Analyzer, "a")
}
