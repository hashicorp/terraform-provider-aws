package securityhub_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccAccount_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists("aws_securityhub_account.example"),
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

func testAccCheckAccountExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn

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

func testAccCheckAccountDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn

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

func testAccAccountConfig_basic() string {
	return `
resource "aws_securityhub_account" "example" {}
`
}
