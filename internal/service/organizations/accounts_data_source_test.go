package organizations_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccOrganizationsAccountsDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_organizations_accounts.test"
	email_account1 := acctest.RandomEmailAddress("whatever.com")
	email_account2 := acctest.RandomEmailAddress("whatever.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountsDataSourceConfig_basic(email_account1, email_account2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "parent_id"),
					resource.TestCheckResourceAttr(dataSourceName, "accounts.#", "2"),
				),
			},
		},
	})
}

func testAccAccountsDataSourceConfig_basic(email_account1, email_account2 string) string {
	return fmt.Sprintf(`
data "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = "test"
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_account" "test1" {
  name              = "test1"
  email             = %[1]q
  parent_id         = aws_organizations_organizational_unit.test.id
  close_on_deletion = false
}

resource "aws_organizations_account" "test2" {
  name              = "test2"
  email             = %[2]q
  parent_id         = aws_organizations_organizational_unit.test.id
  close_on_deletion = false
}

data "aws_organizations_accounts" "test" {
  parent_id = aws_organizations_organizational_unit.test.id
}
	`, email_account1, email_account2)
}
