package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsOrganizationsAccount_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccDataSourceAwsOrganizationsAccountConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsOrganizationsAccountCheck("data.aws_organizations_account.test"),
				),
			},
		},
	})
}

func testAccDataSourceAwsOrganizationsAccountCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Can't find Account data source: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Account data source ID not set")
		}

		return nil
	}
}

const testAccDataSourceAwsOrganizationsAccountConfig = `
resource "aws_organizations_organization" "test" {}

data "aws_organizations_account_ids" "test" {
  depends_on = ["aws_organizations_organization.test"]
}

data "aws_organizations_account" "test" {
  account_id = "${data.aws_organizations_account_ids.test.ids[0]}"
}
`
