package AWSV001

import (
	"testing"

	_ "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"golang.org/x/tools/go/analysis/analysistest"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAWSV001(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "a")
}
