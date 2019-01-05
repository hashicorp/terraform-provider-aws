package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsOrganizationsAccounts_basic(t *testing.T) {
	orgsEmailDomain, ok := os.LookupEnv("TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN")

	if !ok {
		t.Skip("'TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN' not set, skipping test.")
	}

	rInt := acctest.RandInt()
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsAccountsConfig(name, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsAccountIds("data.aws_organizations_accounts.current"),
				),
			},
		},
	})
}

func testAccCheckAwsOrganizationsAccountIds(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find Organizations resource: %s", n)
		}

		if len(rs.Primary.Attributes["account_ids"]) == 0 {
			return fmt.Errorf("No account IDs found")
		}

		return nil
	}
}

func testAccAwsOrganizationsAccountsConfig(name, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_account" "test" {
  name = "%s"
  email = "%s"
}

data "aws_organizations_accounts" "current" { }
`, name, email)
}
