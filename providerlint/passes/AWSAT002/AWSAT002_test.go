package AWSAT002_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/awsproviderlint/passes/AWSAT002"
	"golang.org/x/tools/go/analysis/analysistest"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAWSAT002(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, AWSAT002.Analyzer, "a")
}
