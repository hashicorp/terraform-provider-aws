package organizations_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccOrganizationalUnitDescendantAccountsDataSource_basic(t *testing.T) {
	firstResourceName := "aws_organizations_account.test1"
	secondResourceName := "aws_organizations_account.test2"
	thirdResourceName := "aws_organizations_account.test3"
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
					resource.TestCheckResourceAttrPair(firstResourceName, "arn", dataSourceName, "accounts.0.arn"),
					resource.TestCheckResourceAttrPair(firstResourceName, "email", dataSourceName, "accounts.0.email"),
					resource.TestCheckResourceAttrPair(firstResourceName, "id", dataSourceName, "accounts.0.id"),
					resource.TestCheckResourceAttrPair(firstResourceName, "joined_method", dataSourceName, "accounts.0.joined_method"),
					resource.TestCheckResourceAttrPair(firstResourceName, "joined_timestamp", dataSourceName, "accounts.0.joined_timestamp"),
					resource.TestCheckResourceAttrPair(firstResourceName, "name", dataSourceName, "accounts.0.name"),
					resource.TestCheckResourceAttrPair(firstResourceName, "status", dataSourceName, "accounts.0.status"),

					resource.TestCheckResourceAttrPair(secondResourceName, "arn", dataSourceName, "accounts.1.arn"),
					resource.TestCheckResourceAttrPair(secondResourceName, "email", dataSourceName, "accounts.1.email"),
					resource.TestCheckResourceAttrPair(secondResourceName, "id", dataSourceName, "accounts.1.id"),
					resource.TestCheckResourceAttrPair(secondResourceName, "joined_method", dataSourceName, "accounts.1.joined_method"),
					resource.TestCheckResourceAttrPair(secondResourceName, "joined_timestamp", dataSourceName, "accounts.1.joined_timestamp"),
					resource.TestCheckResourceAttrPair(secondResourceName, "name", dataSourceName, "accounts.1.name"),
					resource.TestCheckResourceAttrPair(secondResourceName, "status", dataSourceName, "accounts.1.status"),

					resource.TestCheckResourceAttrPair(thirdResourceName, "arn", dataSourceName, "accounts.2.arn"),
					resource.TestCheckResourceAttrPair(thirdResourceName, "email", dataSourceName, "accounts.2.email"),
					resource.TestCheckResourceAttrPair(thirdResourceName, "id", dataSourceName, "accounts.2.id"),
					resource.TestCheckResourceAttrPair(thirdResourceName, "joined_method", dataSourceName, "accounts.2.joined_method"),
					resource.TestCheckResourceAttrPair(thirdResourceName, "joined_timestamp", dataSourceName, "accounts.2.joined_timestamp"),
					resource.TestCheckResourceAttrPair(thirdResourceName, "name", dataSourceName, "accounts.2.name"),
					resource.TestCheckResourceAttrPair(thirdResourceName, "status", dataSourceName, "accounts.2.status"),
				),
			},
		},
	})
}

const testAccOrganizationalUnitChildAccountsDataSourceConfig = `
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_account" "test1" {
  name  = "test1"
  email = "ahubler+test1@rapidsos.org"
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_organizational_unit" "test1" {
  name      = "test1"
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_account" "test2" {
  name  = "test2"
  email = "ahubler+test2@rapidsos.org"
  parent_id = aws_organizations_organizational_unit.test1.id
}

resource "aws_organizations_organizational_unit" "test2" {
  name      = "test2"
  parent_id = aws_organizations_organizational_unit.test1.id
}

resource "aws_organizations_account" "test3" {
  name  = "test3"
  email = "ahubler+test3@rapidsos.org"
  parent_id = aws_organizations_organizational_unit.test2.id
}

data "aws_organizations_organizational_unit_descendant_accounts" "test" {
  parent_id = aws_organizations_organization.test.roots[0].id
}
`
