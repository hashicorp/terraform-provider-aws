package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccDataSourceAwsOrganizationsAccounts_basic(t *testing.T) {
	TestAccSkip(t, "AWS Organizations Account testing is not currently automated due to manual account deletion steps.")

	resourceName := "data.aws_organizations_organization.test"
	dataSourceName := "data.aws_organizations_accounts.test1"

	orgsEmailDomain, ok := os.LookupEnv("TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN")

	if !ok {
		TestAccSkip(t, "'TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN' not set, skipping test.")
	}

	name1 := fmt.Sprintf("tf_acctest_%d", acctest.RandInt())
	email1 := fmt.Sprintf("tf-acctest+%d@%s", acctest.RandInt(), orgsEmailDomain)
	name2 := fmt.Sprintf("tf_acctest_%d", acctest.RandInt())
	email2 := fmt.Sprintf("tf-acctest+%d@%s", acctest.RandInt(), orgsEmailDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOrganizationsAccountPreCheck(t)
		},
		ErrorCheck: testAccErrorCheck(t, organizations.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsOrganizationsAccountsConfig(name1, email1, name2, email2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "children.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "non_master_accounts.0.arn", dataSourceName, "children.0.arn"),
					resource.TestCheckResourceAttrPair(resourceName, "non_master_accounts.0.email", dataSourceName, "children.0.email"),
					resource.TestCheckResourceAttrPair(resourceName, "non_master_accounts.0.name", dataSourceName, "children.0.name"),
					resource.TestCheckResourceAttrPair(resourceName, "non_master_accounts.0.id", dataSourceName, "children.0.id"),
					resource.TestCheckResourceAttrPair(resourceName, "non_master_accounts.0.status", dataSourceName, "children.0.status"),
				),
			},
		},
	})
}

func testAccDataSourceAwsOrganizationsAccountsConfig(name1, email1, name2, email2 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = "test"
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_account" "test1" {
  depends_on = [aws_organizations_organization.test]

  name      = %[1]q
  email     = %[2]q
}

resource "aws_organizations_account" "test2" {
  name      = %[3]q
  email     = %[4]q
  parent_id = aws_organizations_organizational_unit.test.parent_id
}

data "aws_organizations_accounts" "test1" {
  depends_on = [aws_organizations_account.test1, aws_organizations_account.test2]

  parent_id = aws_organizations_organization.test.roots[0].id
}

data "aws_organizations_accounts" "test2" {
  depends_on = [aws_organizations_account.test1, aws_organizations_account.test2]

  parent_id = aws_organizations_organizational_unit.test.parent_id
}
`, name1, email1, name2, email2)
}
