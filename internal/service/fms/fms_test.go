package fms_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccFMS_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"AdminAccount": {
			"basic": testAccAdminAccount_basic,
		},
		"Policy": {
			"basic":                  TestAccFMSPolicy_basic,
			"cloudfrontDistribution": TestAccFMSPolicy_cloudFrontDistribution,
			"includeMap":             TestAccFMSPolicy_includeMap,
			"update":                 TestAccFMSPolicy_update,
			"resourceTags":           TestAccFMSPolicy_resourceTags,
			"tags":                   TestAccFMSPolicy_tags,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
