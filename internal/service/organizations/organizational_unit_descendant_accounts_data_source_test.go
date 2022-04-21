package organizations_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccOrganizationalUnitDescendantAccountsDataSource_basic(t *testing.T) {
	resourceName1 := "aws_organizations_account.test1"
	resourceName2 := "aws_organizations_account.test2"
	resourceName3 := "aws_organizations_account.test3"
	dataSourceName := "data.aws_organizations_organizational_unit_descendant_accounts.test"

	domain := acctest.RandomDomainName()
	address1 := acctest.RandomEmailAddress(domain)
	address2 := acctest.RandomEmailAddress(domain)
	address3 := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck: acctest.ErrorCheck(t, organizations.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationalUnitChildAccountsDataSourceConfig(address1, address2, address3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "accounts.#", "3"),
					resource.TestCheckResourceAttrPair(resourceName1, "arn", dataSourceName, "accounts.0.arn"),
					resource.TestCheckResourceAttrPair(resourceName1, "email", dataSourceName, "accounts.0.email"),
					resource.TestCheckResourceAttrPair(resourceName1, "id", dataSourceName, "accounts.0.id"),
					resource.TestCheckResourceAttrPair(resourceName1, "joined_method", dataSourceName, "accounts.0.joined_method"),
					resource.TestCheckResourceAttrPair(resourceName1, "joined_timestamp", dataSourceName, "accounts.0.joined_timestamp"),
					resource.TestCheckResourceAttrPair(resourceName1, "name", dataSourceName, "accounts.0.name"),
					resource.TestCheckResourceAttrPair(resourceName1, "status", dataSourceName, "accounts.0.status"),

					resource.TestCheckResourceAttrPair(resourceName2, "arn", dataSourceName, "accounts.1.arn"),
					resource.TestCheckResourceAttrPair(resourceName2, "email", dataSourceName, "accounts.1.email"),
					resource.TestCheckResourceAttrPair(resourceName2, "id", dataSourceName, "accounts.1.id"),
					resource.TestCheckResourceAttrPair(resourceName2, "joined_method", dataSourceName, "accounts.1.joined_method"),
					resource.TestCheckResourceAttrPair(resourceName2, "joined_timestamp", dataSourceName, "accounts.1.joined_timestamp"),
					resource.TestCheckResourceAttrPair(resourceName2, "name", dataSourceName, "accounts.1.name"),
					resource.TestCheckResourceAttrPair(resourceName2, "status", dataSourceName, "accounts.1.status"),

					resource.TestCheckResourceAttrPair(resourceName3, "arn", dataSourceName, "accounts.2.arn"),
					resource.TestCheckResourceAttrPair(resourceName3, "email", dataSourceName, "accounts.2.email"),
					resource.TestCheckResourceAttrPair(resourceName3, "id", dataSourceName, "accounts.2.id"),
					resource.TestCheckResourceAttrPair(resourceName3, "joined_method", dataSourceName, "accounts.2.joined_method"),
					resource.TestCheckResourceAttrPair(resourceName3, "joined_timestamp", dataSourceName, "accounts.2.joined_timestamp"),
					resource.TestCheckResourceAttrPair(resourceName3, "name", dataSourceName, "accounts.2.name"),
					resource.TestCheckResourceAttrPair(resourceName3, "status", dataSourceName, "accounts.2.status"),
				),
			},
		},
	})
}

func testAccOrganizationalUnitChildAccountsDataSourceConfig(rAddress1 string, rAddress2 string, rAddress3 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_account" "test1" {
  name  = "test1"
  email = %[1]q
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_organizational_unit" "test1" {
  name      = "test1"
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_account" "test2" {
  name  = "test2"
  email = %[2]q
  parent_id = aws_organizations_organizational_unit.test1.id
}

resource "aws_organizations_organizational_unit" "test2" {
  name      = "test2"
  parent_id = aws_organizations_organizational_unit.test1.id
}

resource "aws_organizations_account" "test3" {
  name  = "test3"
  email = %[3]q
  parent_id = aws_organizations_organizational_unit.test2.id
}

data "aws_organizations_organizational_unit_descendant_accounts" "test" {
  parent_id = aws_organizations_organization.test.roots[0].id
}
`, rAddress1, rAddress2, rAddress3)
}
