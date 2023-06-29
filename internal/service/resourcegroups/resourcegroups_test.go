package resourcegroups_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccResourceGroups_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Resource": {
			"basic":      testAccResource_basic,
			"disappears": testAccResource_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
