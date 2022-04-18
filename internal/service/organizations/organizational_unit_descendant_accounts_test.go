package organizations_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccOrganizationalUnitDescendantAccountsDataSource_basic(t *testing.T) {
	resourceName := "aws_organizations_account.test0"
	dataSourceName := "data.aws_organizations_organizational_unit_descendant_accounts.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck: acctest.ErrorCheck(t, organizations.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationalUnitChildAccountsDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "accounts.#", "3"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "accounts.0.arn"),
					resource.TestCheckResourceAttrPair(resourceName, "email", dataSourceName, "accounts.0.email"),
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "accounts.0.id"),
					resource.TestCheckResourceAttrPair(resourceName, "joined_method", dataSourceName, "accounts.0.joined_method"),
					resource.TestCheckResourceAttrPair(resourceName, "joined_timestamp", dataSourceName, "accounts.0.joined_timestamp"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "accounts.0.name"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceName, "accounts.0.status"),
				),
			},
		},
	})
}

const testAccOrganizationalUnitChildAccountsDataSourceConfig = `
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_account" "test0" {
  name  = "test0"
  email = "john@doe.org"
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_organizational_unit" "test0" {
  name      = "test0"
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_account" "test1" {
  name  = "test1"
  email = "john@doe.org"
  parent_id = aws_organizations_organizational_unit.test0.id
}

resource "aws_organizations_organizational_unit" "test1" {
  name      = "test1"
  parent_id = aws_organizations_organizational_unit.test0.id
}

resource "aws_organizations_account" "test2" {
  name  = "test2"
  email = "john@doe.org"
  parent_id = aws_organizations_organizational_unit.test1.id
}

data "aws_organizations_organizational_unit_descendant_accounts" "test" {
  parent_id = aws_organizations_organization.test.roots[0].id
}
`