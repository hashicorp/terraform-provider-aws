package organizations_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDataSourceAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	orgsEmailDomain := acctest.SkipIfEnvVarNotSet(t, "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN")

	rInt := sdkacctest.RandInt()
	nameRoot := fmt.Sprintf("tf_acctest_root_%d", rInt)
	emailRoot := fmt.Sprintf("tf-acctest+root+%d@%s", rInt, orgsEmailDomain)

	nameChild := fmt.Sprintf("tf_acctest_child_%d", rInt)
	emailChild := fmt.Sprintf("tf-acctest+child+%d@%s", rInt, orgsEmailDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsEnabled(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Root Account
			{
				Config: testAccDataSourceAccountConfig_root(nameRoot, emailRoot),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_organizations_account.test_root", "id"),
					resource.TestCheckResourceAttrSet("data.aws_organizations_account.test_root", "arn"),
					resource.TestCheckResourceAttrSet("data.aws_organizations_account.test_root", "principal_org_path"),
				),
			},
			// Step 2: Simple Account
			{
				Config: testAccDataSourceAccountConfig_simple(nameChild, emailChild),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_organizations_account.test_simple", "id"),
					resource.TestCheckResourceAttrSet("data.aws_organizations_account.test_simple", "arn"),
					resource.TestCheckResourceAttrSet("data.aws_organizations_account.test_simple", "principal_org_path"),
				),
			},
			// Step 3: Nested Account
			{
				Config: testAccDataSourceAccountConfig_nested(nameChild, emailChild),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_organizations_account.test_nested", "id"),
					resource.TestCheckResourceAttrSet("data.aws_organizations_account.test_nested", "arn"),
					resource.TestCheckResourceAttrSet("data.aws_organizations_account.test_nested", "principal_org_path"),
				),
			},
		},
	})
}

func testAccDataSourceAccountConfig_root(name, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_account" "root" {
  name  = %[1]q
  email = %[2]q
}

data "aws_organizations_account" "test_root" {
  id = aws_organizations_account.root.id
}
`, name, email)
}

func testAccDataSourceAccountConfig_simple(name, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_account" "simple" {
  name  = %[1]q
  email = %[2]q
}

data "aws_organizations_account" "test_simple" {
  id = aws_organizations_account.simple.id
}
`, name, email)
}

func testAccDataSourceAccountConfig_nested(name, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organizational_unit" "ou1" {
  name      = "ou1"
  parent_id = "r-examplerootid"
}

resource "aws_organizations_organizational_unit" "ou2" {
  name      = "ou2"
  parent_id = aws_organizations_organizational_unit.ou1.id
}

resource "aws_organizations_account" "nested" {
  name      = %[1]q
  email     = %[2]q
  parent_id = aws_organizations_organizational_unit.ou2.id
}

data "aws_organizations_account" "test_nested" {
  id = aws_organizations_account.nested.id
}
`, name, email)
}
