package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func testAccAWSSecurityHubProductSubscription_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityHubAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityHubProductSubscriptionConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubProductSubscriptionExists("aws_securityhub_product_subscription.example"),
				),
			},
			{
				ResourceName:            "aws_securityhub_product_subscription.example",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"product_arn"},
			},
			{
				// Check Destroy - but only target the specific resource (otherwise Security Hub
				// will be disabled and the destroy check will fail)
				Config: testAccAWSSecurityHubProductSubscriptionConfig_empty,
				Check:  testAccCheckAWSSecurityHubProductSubscriptionDestroy,
			},
		},
	})
}

func testAccCheckAWSSecurityHubProductSubscriptionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).securityhubconn

		exists, err := resourceAwsSecurityHubProductSubscriptionCheckExists(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("Security Hub product subscription %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAWSSecurityHubProductSubscriptionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).securityhubconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_securityhub_product_subscription" {
			continue
		}

		exists, err := resourceAwsSecurityHubProductSubscriptionCheckExists(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if exists {
			return fmt.Errorf("Security Hub product subscription %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccAWSSecurityHubProductSubscriptionConfig_empty = `
resource "aws_securityhub_account" "example" {}
`

const testAccAWSSecurityHubProductSubscriptionConfig_basic = `
resource "aws_securityhub_account" "example" {}

data "aws_region" "current" {}

resource "aws_securityhub_product_subscription" "example" {
  depends_on  = ["aws_securityhub_account.example"]
  product_arn = "arn:aws:securityhub:${data.aws_region.current.name}:733251395267:product/alertlogic/althreatmanagement"
}
`
