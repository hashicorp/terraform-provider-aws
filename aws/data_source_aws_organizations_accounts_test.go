package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccDataSourceAwsOrganizationsAccounts_basic(t *testing.T) {
	dataSourceName := "data.aws_organizations_accounts.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsOrganizationAccountResourceOnlyConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "accounts.#", "1"),
					resource.TestCheckResourceAttrSet(dataSourceName, "accounts.0.arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "accounts.0.joined_method"),
					testAccCheckResourceAttrRfc3339(dataSourceName, "accounts.0.joined_timestamp"),
					resource.TestCheckResourceAttrSet(dataSourceName, "accounts.0.name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "accounts.0.email"),
					resource.TestCheckResourceAttrSet(dataSourceName, "accounts.0.status"),
					resource.TestCheckResourceAttrSet(dataSourceName, "accounts.0.tags.%"),
				),
			},
		},
	})
}

func testAccCheckAwsOrganizationAccountResourceOnlyConfig() string {
	return fmt.Sprintf(`
data "aws_organizations_organization" "test" { }
data "aws_organizations_accounts" "test" { account_ids = [ data.aws_organizations_organization.test.accounts[0].id] }
`)
}
