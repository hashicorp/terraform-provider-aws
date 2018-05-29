package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsOrganizationsAccountIDs_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccDataSourceAwsOrganizationsAccountIDsConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsOrganizationsAccountIDsCheck("data.aws_organizations_account_ids.test"),
				),
			},
		},
	})
}

func testAccDataSourceAwsOrganizationsAccountIDsCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Can't find Account IDs data source: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Account IDs data source ID not set")
		}

		return nil
	}
}

const testAccDataSourceAwsOrganizationsAccountIDsConfig = `
resource "aws_organizations_organization" "test" {}

data "aws_organizations_account_ids" "test" {
  depends_on = ["aws_organizations_organization.test"]
}
`
