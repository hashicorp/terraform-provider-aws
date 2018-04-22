package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_fms_admin_account", &resource.Sweeper{
		Name: "aws_fms_admin_account",
		F:    testSweepFmsAdminAccount,
	})
}

func testSweepFmsAdminAccount(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*AWSClient).fmsconn

	_, err = conn.GetAdminAccount(&fms.GetAdminAccountInput{})
	if err != nil {
		// FMS returns an AccessDeniedException if no account is associated,
		// but does not define this in its error codes
		if isAWSErr(err, "AccessDeniedException", "is not currently delegated by AWS FM") {
			log.Print("[DEBUG] No associated firewall manager admin account to sweep")
			return nil
		}
		return fmt.Errorf("Error retrieving firewall manager admin account: %s", err)
	}

	_, err = conn.DisassociateAdminAccount(&fms.DisassociateAdminAccountInput{})
	if err != nil {
		// FMS returns an AccessDeniedException if no account is associated,
		// but does not define this in its error codes
		if isAWSErr(err, "AccessDeniedException", "is not currently delegated by AWS FM") {
			log.Print("[DEBUG] No associated firewall manager admin account to sweep")
			return nil
		}
		return fmt.Errorf("Error disassociating firewall manager admin account: %s", err)
	}

	return nil
}

func TestAccFmsAdminAccount_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFmsAdminAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsAdminAccountConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_fms_admin_account.example", "account_id", regexp.MustCompile("^\\d{12}$")),
				),
			},
		},
	})
}

func testAccCheckFmsAdminAccountDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).fmsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fms_admin_account" {
			continue
		}

		res, err := conn.GetAdminAccount(&fms.GetAdminAccountInput{})
		if err != nil {
			// FMS returns an AccessDeniedException if no account is associated,
			// but does not define this in its error codes
			if isAWSErr(err, "AccessDeniedException", "is not currently delegated by AWS FM") {
				log.Print("[DEBUG] No associated firewall manager admin account")
				return nil
			}
		}

		return fmt.Errorf("Firewall manager admin account still exists: %v", res.AdminAccount)
	}

	return nil
}

const testAccFmsAdminAccountConfig_basic = `
provider "aws" {
  region = "us-east-1"
}

resource "aws_fms_admin_account" "example" {
  depends_on = ["aws_organizations_organization.example"]
  account_id = "${data.aws_caller_identity.current.account_id}" # Required
}

resource "aws_organizations_organization" "example" {
  feature_set = "ALL"
}

data "aws_caller_identity" "current" {}
`
