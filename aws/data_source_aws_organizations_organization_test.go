package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsOrganizationsOrganization_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccDataSourceAwsOrganizationsOrganizationConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsOrganizationsOrganizationCheck("data.aws_organizations_organization.test"),
					resource.TestCheckResourceAttr("data.aws_organizations_organization.test", "feature_set", organizations.OrganizationFeatureSetAll),
					resource.TestCheckResourceAttrSet("data.aws_organizations_organization.test", "arn"),
					resource.TestCheckResourceAttrSet("data.aws_organizations_organization.test", "feature_set"),
					resource.TestCheckResourceAttrSet("data.aws_organizations_organization.test", "master_account_arn"),
					resource.TestCheckResourceAttrSet("data.aws_organizations_organization.test", "master_account_email"),
					resource.TestCheckResourceAttrSet("data.aws_organizations_organization.test", "master_account_id"),
				),
			},
		},
	})
}

func testAccDataSourceAwsOrganizationsOrganizationCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Can't find Organization data source: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Organization data source ID not set")
		}

		return nil
	}
}

const testAccDataSourceAwsOrganizationsOrganizationConfig = `
resource "aws_organizations_organization" "test" {}

data "aws_organizations_organization" "test" {
  depends_on = ["aws_organizations_organization.test"]
}
`
