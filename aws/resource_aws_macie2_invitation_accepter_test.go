package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsMacie2InvitationAccepter_basic(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_macie2_invitation_accepter.test"
	adminAccountID := "124861550386"
	accountID, email := testAccAWSMacie2MemberFromEnv(t)
	accountIds := []string{accountID}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsMacie2InvitationAccepterDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieInvitationAccepterConfigBasic(accountID, email, adminAccountID, accountIds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2InvitationAccepterExists(resourceName),
				),
			},
			{
				Config:            testAccAwsMacieInvitationAccepterConfigBasic(accountID, email, adminAccountID, accountIds),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsMacie2InvitationAccepter_memberStatus(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_macie2_invitation_accepter.test"
	adminAccountID := "124861550386"
	accountID, email := testAccAWSMacie2MemberFromEnv(t)
	accountIds := []string{accountID}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsMacie2InvitationAccepterDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieInvitationAccepterConfigMemberStatus(accountID, email, adminAccountID, macie2.MacieStatusEnabled, accountIds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2InvitationAccepterExists(resourceName),
				),
			},
			{
				Config: testAccAwsMacieInvitationAccepterConfigMemberStatus(accountID, email, adminAccountID, macie2.MacieStatusPaused, accountIds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2InvitationAccepterExists(resourceName),
				),
			},
			{
				Config:            testAccAwsMacieInvitationAccepterConfigBasic(accountID, email, adminAccountID, accountIds),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsMacie2InvitationAccepterExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource (%s) has empty ID", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).macie2conn
		input := &macie2.GetAdministratorAccountInput{}
		output, err := conn.GetAdministratorAccount(input)

		if err != nil {
			return err
		}

		if output == nil || output.Administrator == nil || aws.StringValue(output.Administrator.AccountId) == "" {
			return fmt.Errorf("no administrator account found for: %s", resourceName)
		}

		return nil
	}
}

func testAccCheckAwsMacie2InvitationAccepterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).macie2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie2_invitation_accepter" {
			continue
		}

		input := &macie2.GetAdministratorAccountInput{}
		output, err := conn.GetAdministratorAccount(input)

		if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
			return nil
		}

		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
			continue
		}

		if output == nil || output.Administrator == nil || aws.StringValue(output.Administrator.AccountId) != rs.Primary.Attributes["administrator_account_id"] {
			continue
		}

		return fmt.Errorf("macie InvitationAccepter %q still exists", rs.Primary.ID)

	}

	return nil

}

func testAccAwsMacieInvitationAccepterConfigBasic(accountID, email, adminAccountID string, accountIDs []string) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
resource "aws_macie2_account" "primary" {
  provider = "awsalternate"
}

resource "aws_macie2_account" "member" {}

resource "aws_macie2_member" "primary" {
  provider   = "awsalternate"
  account_id = %[1]q
  email      = %[2]q
  depends_on = [aws_macie2_account.primary]
}

resource "aws_macie2_invitation" "primary" {
  provider    = "awsalternate"
  account_ids = %[3]q
  depends_on  = [aws_macie2_member.primary]
}

resource "aws_macie2_invitation_accepter" "test" {
  administrator_account_id = %[4]q
  depends_on               = [aws_macie2_invitation.primary]
}

`, accountID, email, accountIDs, adminAccountID)
}

func testAccAwsMacieInvitationAccepterConfigMemberStatus(accountID, email, adminAccountID, memberStatus string, accountIDs []string) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
resource "aws_macie2_account" "primary" {
  provider = "awsalternate"
}

resource "aws_macie2_account" "member" {}

resource "aws_macie2_member" "primary" {
  provider   = "awsalternate"
  account_id = %[1]q
  email      = %[2]q
  status     = %[5]q
  depends_on = [aws_macie2_account.primary]
}

resource "aws_macie2_invitation" "primary" {
  provider    = "awsalternate"
  account_ids = %[3]q
  depends_on  = [aws_macie2_member.primary]
}

resource "aws_macie2_invitation_accepter" "test" {
  administrator_account_id = %[4]q
  depends_on               = [aws_macie2_invitation.primary]
}

`, accountID, email, accountIDs, adminAccountID, memberStatus)
}
