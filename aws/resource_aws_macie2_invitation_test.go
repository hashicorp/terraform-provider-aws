package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsMacie2Invitation_basic(t *testing.T) {
	accountID, email := testAccAWSMacie2MemberFromEnv(t)
	accountIds := []string{accountID}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2InvitationDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieInvitationConfigBasic(accountID, email, accountIds),
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func testAccAwsMacie2Invitation_disappears(t *testing.T) {
	resourceName := "aws_macie2_invitation.test"
	accountID, email := testAccAWSMacie2MemberFromEnv(t)
	accountIds := []string{accountID}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2InvitationDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieInvitationConfigBasic(accountID, email, accountIds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceDisappears(testAccProvider, resourceAwsMacie2Invitation(), resourceName),
				),
			},
		},
	})
}

func testAccCheckAwsMacie2InvitationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).macie2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie2_invitation" {
			continue
		}

		empty := true
		err := conn.ListInvitationsPages(&macie2.ListInvitationsInput{}, func(page *macie2.ListInvitationsOutput, lastPage bool) bool {
			if len(page.Invitations) > 0 {
				empty = false
				return false
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

func testAccAwsMacieInvitationConfigBasic(accountID, email string, accountIDs []string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {}

resource "aws_macie2_member" "test" {
  account_id = %[1]q
  email      = %[2]q
  depends_on = [aws_macie2_account.test]
}

resource "aws_macie2_invitation" "test" {
  account_ids = %[3]q
  depends_on  = [aws_macie2_member.test]
}
`, accountID, email, accountIDs)
}
