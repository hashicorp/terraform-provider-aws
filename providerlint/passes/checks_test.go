package passes

import (
	"testing"

	"golang.org/x/tools/go/analysis"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestValidateAllChecks(t *testing.T) {
	err := analysis.Validate(AllChecks)

	if err != nil {
		t.Fatal(err)
	}
}
