package securityhub_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccOrganizationConfiguration_basic(t *testing.T) {
	resourceName := "aws_securityhub_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil, //lintignore:AT001
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig(true),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfigurationConfig(false),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "false"),
				),
			},
		},
	})
}

func testAccOrganizationConfigurationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn

		_, err := conn.DescribeOrganizationConfiguration(&securityhub.DescribeOrganizationConfigurationInput{})

		return err
	}
}

func testAccOrganizationConfigurationConfig(autoEnable bool) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["securityhub.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_securityhub_account" "test" {}

data "aws_caller_identity" "current" {}

resource "aws_securityhub_organization_admin_account" "test" {
  admin_account_id = data.aws_caller_identity.current.account_id

  depends_on = [aws_organizations_organization.test, aws_securityhub_account.test]
}

resource "aws_securityhub_organization_configuration" "test" {
  auto_enable = %[1]t

  depends_on = [aws_securityhub_organization_admin_account.test]
}
`, autoEnable)
}
