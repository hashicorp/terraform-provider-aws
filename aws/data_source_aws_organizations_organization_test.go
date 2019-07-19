package aws

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsOrganizationsOrganization_basic(t *testing.T) {
	resourceName := "data.aws_organizations_organization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsOrganizationConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "arn",
						regexp.MustCompile(`^arn:aws:organizations::\d{12}:organization\/o-[a-z0-9]{10,32}$`),
					),
					resource.TestMatchResourceAttr(resourceName, "id",
						regexp.MustCompile(`^o-[a-z0-9]{10,32}$`),
					),
					resource.TestMatchResourceAttr(resourceName, "master_account_arn",
						regexp.MustCompile(`^arn:aws:organizations::\d{12}:account\/o-[a-z0-9]{10,32}\/\d{12}$`),
					),
					resource.TestMatchResourceAttr(resourceName, "master_account_email",
						regexp.MustCompile(`[^\s@]+@[^\s@]+\.[^\s@]+`),
					),
					resource.TestMatchResourceAttr(resourceName, "master_account_id",
						regexp.MustCompile(`^\d{12}$`),
					),
					resource.TestCheckResourceAttrSet(resourceName, "feature_set"),
				),
			},
		},
	})
}

const testAccCheckAwsOrganizationConfig = `
data "aws_organizations_organization" "test" {}
`
