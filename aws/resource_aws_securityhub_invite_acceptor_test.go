package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSSecurityHubInviteAccepter_basic(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_securityhub_invite_accepter.example"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
			testAccAccountEmailPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAWSSecurityHubInviteAccepterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityHubInviteAccepterConfig_basic(testAccGetAccountEmail()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubInviteAccepterExists(resourceName),
				),
			},
			{
				Config:            testAccAWSSecurityHubInviteAccepterConfig_basic(testAccGetAccountEmail()),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSSecurityHubInviteAccepterExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).securityhubconn

		resp, err := conn.GetMasterAccount(&securityhub.GetMasterAccountInput{})

		if err != nil {
			return err
		}

		if resp == nil || resp.Master == nil || aws.StringValue(resp.Master.AccountId) == "" {
			return fmt.Errorf("Security Hub master account found for: %s", resourceName)
		}

		return nil
	}
}

func testAccCheckAWSSecurityHubInviteAccepterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).securityhubconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_securityhub_invite_accepter" {
			continue
		}

		resp, err := conn.GetMasterAccount(&securityhub.GetMasterAccountInput{})

		if err != nil {
			if isAWSErr(err, securityhub.ErrCodeResourceNotFoundException, "") {
				continue
			}
			return err
		}

		if resp == nil || resp.Master == nil || aws.StringValue(resp.Master.AccountId) == "" {
			continue
		}

		return fmt.Errorf("Security Hub master account still configured: %s", aws.StringValue(resp.Master.AccountId))
	}
	return nil
}

func testAccAWSSecurityHubInviteAccepterConfig_basic(email string) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
resource "aws_securityhub_account" "example" {
  provider = "aws.alternate"
}

resource "aws_securityhub_member" "example" {
  provider   = "aws.alternate"
  depends_on = [aws_securityhub_account.example]
  account_id = data.aws_caller_identity.accepter.account_id
  email      = "%s"
  invite     = true
}

data "aws_caller_identity" "accepter" {}

resource "aws_securityhub_account" "accepter" {}

resource "aws_securityhub_invite_accepter" "example" {
  depends_on = [aws_securityhub_account.accepter]
  master_id  = aws_securityhub_member.example.master_id
}
`, email)
}
