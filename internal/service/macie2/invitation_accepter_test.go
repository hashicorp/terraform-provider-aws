package macie2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccInvitationAccepter_basic(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_macie2_invitation_accepter.member"
	email := conns.SkipIfEnvVarEmpty(t, EnvVarPrincipalEmail, EnvVarPrincipalEmailMessageError)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckInvitationAccepterDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccInvitationAccepterConfig_basic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvitationAccepterExists(resourceName),
				),
			},
			{
				Config:            testAccInvitationAccepterConfig_basic(email),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckInvitationAccepterExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource (%s) has empty ID", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Conn
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

func testAccCheckInvitationAccepterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie2_invitation_accepter" {
			continue
		}

		input := &macie2.GetAdministratorAccountInput{}
		output, err := conn.GetAdministratorAccount(input)
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

func testAccInvitationAccepterConfig_basic(email string) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
data "aws_caller_identity" "admin" {
  provider = "awsalternate"
}

data "aws_caller_identity" "member" {}

resource "aws_macie2_account" "admin" {
  provider = "awsalternate"
}

resource "aws_macie2_account" "member" {}

resource "aws_macie2_member" "member" {
  provider           = "awsalternate"
  account_id         = data.aws_caller_identity.member.account_id
  email              = %[1]q
  invite             = true
  invitation_message = "This is a message of the invite"
  depends_on         = [aws_macie2_account.admin]
}

resource "aws_macie2_invitation_accepter" "member" {
  administrator_account_id = data.aws_caller_identity.admin.account_id
  depends_on               = [aws_macie2_member.member]
}
`, email)
}
