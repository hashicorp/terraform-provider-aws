package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSShieldSubscription(t *testing.T) {
	// Prevent activation of Shield Advanced
	// resource.Test(t, resource.TestCase{
	// 	PreCheck:     func() { testAccPreCheck(t) },
	// 	Providers:    testAccProviders,
	// 	CheckDestroy: testAccCheckAWSShieldSubscriptionDestroy,
	// 	Steps: []resource.TestStep{
	// 		{
	// 			Config: testAccShieldSubscriptionConfig,
	// 			Check: resource.ComposeTestCheckFunc(
	// 				testAccCheckAWSShieldProtectionExists("aws_shield_subscription.hoge"),
	// 			),
	// 		},
	// 	},
	// })
}

func testAccCheckAWSShieldSubscriptionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).shieldconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_shield_subscription" {
			continue
		}

		input := &shield.DescribeSubscriptionInput{}

		_, err := conn.DescribeSubscription(input)
		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckAWSShieldSubscriptionExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

const testAccShieldSubscriptionConfig = `
  resource "aws_shield_subscription" "hoge" {
  }
`
