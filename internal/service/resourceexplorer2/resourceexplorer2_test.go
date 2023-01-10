package resourceexplorer2_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccResourceExplorer2_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Index": {
			"basic":      testAccIndex_basic,
			"disappears": testAccIndex_disappears,
			"tags":       testAccIndex_tags,
			"type":       testAccIndex_type,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
