package passes

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"golang.org/x/tools/go/analysis"
)

func TestValidateAllChecks(t *testing.T) {
	err := analysis.Validate(AllChecks)

	if err != nil {
		t.Fatal(err)
	}
}
