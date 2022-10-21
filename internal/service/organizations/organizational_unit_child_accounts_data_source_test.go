package organizations_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccOrganizationalUnitChildAccountsDataSource_basic(t *testing.T) {
	resourceName := "aws_organizations_account.test"
	dataSourceName := "data.aws_organizations_organizational_unit_child_accounts.test"

	domain := acctest.RandomDomainName()
	address := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationalUnitChildAccountsDataSourceConfig(address),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "accounts.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "accounts.0.arn"),
					resource.TestCheckResourceAttrPair(resourceName, "email", dataSourceName, "accounts.0.email"),
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "accounts.0.id"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "accounts.0.name"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceName, "accounts.0.status"),
				),
			},
		},
	})
}

func testAccOrganizationalUnitChildAccountsDataSourceConfig(rAddress string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_account" "test" {
  name  = "test"
  email = %[1]q
  parent_id = aws_organizations_organization.test.roots[0].id
}

data "aws_organizations_organizational_unit_child_accounts" "test" {
  parent_id = aws_organizations_organization.test.roots[0].id
}
`, rAddress)
}
