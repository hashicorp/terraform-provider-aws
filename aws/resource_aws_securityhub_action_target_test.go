package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsSecurityHubActionTarget_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityHubAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecurityHubActionTargetConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubAccountExists("aws_securityhub_account.example"),
					testAccCheckAwsSecurityHubActionTargetExists("aws_securityhub_action_target.example"),
				),
			},
			{
				ResourceName:      "aws_securityhub_action_target.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Check Destroy - but only target the specific resource (otherwise Security Hub
				// will be disabled and the destroy check will fail)
				Config: testAccAwsSecurityHubActionTargetConfig_empty,
				Check:  testAccCheckAwsSecurityHubActionTargetDestroy,
			},
		},
	})
}

func testAccCheckAwsSecurityHubActionTargetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Security Hub custom action ARN is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).securityhubconn

		action, err := resourceAwsSecurityHubActionTargetCheckExists(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if action == nil {
			return fmt.Errorf("Security Hub custom action %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAwsSecurityHubActionTargetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).securityhubconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_securityhub_action_target" {
			continue
		}

		action, err := resourceAwsSecurityHubActionTargetCheckExists(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if action == nil {
			return fmt.Errorf("Security Hub custom action %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccAwsSecurityHubActionTargetConfig_empty = `
resource "aws_securityhub_account" "example" {}
`

const testAccAwsSecurityHubActionTargetConfig = `
resource "aws_securityhub_account" "example" {}

data "aws_region" "current" {}

resource "aws_securityhub_action_target" "example" {
  depends_on  = ["aws_securityhub_account.example"]
  name        = "Test action"
  identifier  = "testaction"
  description = "This is a test custom action"
}
`
