package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsMacie2Invitation_basic(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_macie2_invitation.test"
	email := "required@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsMacie2InvitationDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieInvitationConfigBasic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2InvitationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "invited_at"),
				),
			},
			{
				Config:            testAccAwsMacieInvitationConfigBasic(email),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsMacie2Invitation_disappears(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_macie2_invitation.test"
	email := "required@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsMacie2InvitationDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieInvitationConfigBasic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2InvitationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsMacie2Invitation(), resourceName),
				),
			},
			{
				Config:            testAccAwsMacieInvitationConfigBasic(email),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsMacie2InvitationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource (%s) has empty ID", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).macie2conn
		exists := false
		err := conn.ListMembersPages(&macie2.ListMembersInput{OnlyAssociated: aws.String("false")}, func(page *macie2.ListMembersOutput, lastPage bool) bool {
			for _, member := range page.Members {
				if aws.StringValue(member.AdministratorAccountId) == rs.Primary.ID {
					exists = true
					return false
				}
			}
			return true
		})

		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("no administrator account found for: %s", resourceName)
		}

		return nil
	}
}

func testAccCheckAwsMacie2InvitationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).macie2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie2_invitation" {
			continue
		}

		empty := true
		err := conn.ListMembersPages(&macie2.ListMembersInput{OnlyAssociated: aws.String("false")}, func(page *macie2.ListMembersOutput, lastPage bool) bool {
			for _, member := range page.Members {
				if aws.StringValue(member.AdministratorAccountId) == rs.Primary.ID {
					empty = false
					return false
				}
			}
			return true
		})

		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
			continue
		}

		if err != nil {
			return err
		}

		if !empty {
			return fmt.Errorf("macie Invitation %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccAwsMacieInvitationConfigBasic(email string) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
data "aws_caller_identity" "inviter" {
  provider = "awsalternate"
}

resource "aws_macie2_account" "test" {}

resource "aws_macie2_member" "test" {
  account_id = data.aws_caller_identity.inviter.account_id
  email      = %[1]q
  depends_on = [aws_macie2_account.test]
}

resource "aws_macie2_invitation" "test" {
  account_id = data.aws_caller_identity.inviter.account_id
  depends_on = [aws_macie2_member.test]
}
`, email)
}
