package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccAWSSecurityHubAccount_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, securityhub.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSecurityHubAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityHubAccountConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubAccountExists("aws_securityhub_account.example"),
				),
			},
			{
				ResourceName:      "aws_securityhub_account.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSSecurityHubAccountExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*AWSClient).securityhubconn

		_, err := conn.GetEnabledStandards(&securityhub.GetEnabledStandardsInput{})

		if err != nil {
			// Can only read enabled standards if Security Hub is enabled
			if tfawserr.ErrMessageContains(err, "InvalidAccessException", "not subscribed to AWS Security Hub") {
				return fmt.Errorf("Security Hub account not found")
			}
			return err
		}

		return nil
	}
}

func testAccCheckAWSSecurityHubAccountDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*AWSClient).securityhubconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_securityhub_account" {
			continue
		}

		_, err := conn.GetEnabledStandards(&securityhub.GetEnabledStandardsInput{})

		if err != nil {
			// Can only read enabled standards if Security Hub is enabled
			if tfawserr.ErrMessageContains(err, "InvalidAccessException", "not subscribed to AWS Security Hub") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Security Hub account still exists")
	}

	return nil
}

func testAccAWSSecurityHubAccountConfig() string {
	return `
resource "aws_securityhub_account" "example" {}
`
}
