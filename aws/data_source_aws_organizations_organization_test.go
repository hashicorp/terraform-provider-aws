package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAwsDataSourceOrganizationsOrganization_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceAccAwsOrganizationsOrganizationConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_organizations_organization.test", "arn"),
					resource.TestCheckResourceAttrSet("data.aws_organizations_organization.test", "master_account_arn"),
					resource.TestCheckResourceAttrSet("data.aws_organizations_organization.test", "master_account_email"),
					resource.TestCheckResourceAttrSet("data.aws_organizations_organization.test", "feature_set"),
					resource.TestCheckResourceAttrSet("data.aws_organizations_organization.test", "available_policy_types.#"),
				),
			},
		},
	})
}

const testDataSourceAccAwsOrganizationsOrganizationConfig = `data "aws_organizations_organization" "test" {}`
