package AWSV001

import (
	"testing"

	_ "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAWSV001(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "a")
}
