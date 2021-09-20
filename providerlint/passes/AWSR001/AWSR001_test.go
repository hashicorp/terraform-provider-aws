package AWSR001

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAWSR001(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "a")
}
