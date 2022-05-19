package detective_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdetective "github.com/hashicorp/terraform-provider-aws/internal/service/detective"
)

func testAccMember_basic(t *testing.T) {
	var providers []*schema.Provider
	var detectiveOutput detective.MemberDetail
	resourceName := "aws_detective_member.test"
	dataSourceAlternate := "data.aws_caller_identity.member"
	email := testAccMemberFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckMemberDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, detective.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_basic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(resourceName, &detectiveOutput),
					acctest.CheckResourceAttrAccountID(resourceName, "administrator_id"),
					resource.TestCheckResourceAttrPair(resourceName, "account_id", dataSourceAlternate, "account_id"),
					acctest.CheckResourceAttrRFC3339(resourceName, "invited_time"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_time"),
					resource.TestCheckResourceAttr(resourceName, "status", detective.MemberStatusInvited),
				),
			},
			{
				Config:                  testAccMemberConfig_basic(email),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"disable_email_notification"},
			},
		},
	})
}

func testAccMember_disappears(t *testing.T) {
	var providers []*schema.Provider
	var detectiveOutput detective.MemberDetail
	resourceName := "aws_detective_member.test"
	email := testAccMemberFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckMemberDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, detective.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_basic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(resourceName, &detectiveOutput),
					acctest.CheckResourceDisappears(acctest.Provider, tfdetective.ResourceMember(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccMember_message(t *testing.T) {
	var detectiveOutput detective.MemberDetail
	var providers []*schema.Provider
	resourceName := "aws_detective_member.test"
	dataSourceAlternate := "data.aws_caller_identity.member"
	email := testAccMemberFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckInvitationAccepterDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, detective.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_invitationMessage(email, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(resourceName, &detectiveOutput),
					acctest.CheckResourceAttrAccountID(resourceName, "administrator_id"),
					resource.TestCheckResourceAttrPair(resourceName, "account_id", dataSourceAlternate, "account_id"),
					acctest.CheckResourceAttrRFC3339(resourceName, "invited_time"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_time"),
					resource.TestCheckResourceAttr(resourceName, "status", detective.MemberStatusInvited),
				),
			},
			{
				Config: testAccMemberConfig_invitationMessage(email, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(resourceName, &detectiveOutput),
					acctest.CheckResourceAttrAccountID(resourceName, "administrator_id"),
					resource.TestCheckResourceAttrPair(resourceName, "account_id", dataSourceAlternate, "account_id"),
					acctest.CheckResourceAttrRFC3339(resourceName, "invited_time"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_time"),
					resource.TestCheckResourceAttr(resourceName, "status", detective.MemberStatusInvited),
				),
			},
			{
				Config:                  testAccMemberConfig_invitationMessage(email, true),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"message", "disable_email_notification"},
			},
		},
	})
}

func testAccCheckMemberExists(resourceName string, detectiveSession *detective.MemberDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DetectiveConn

		graphArn, accountId, err := tfdetective.DecodeMemberID(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := tfdetective.FindMemberByGraphArnAndAccountID(context.Background(), conn, graphArn, accountId)
		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("detective Member %q does not exist", rs.Primary.ID)
		}

		*detectiveSession = *resp

		return nil
	}
}

func testAccCheckMemberDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DetectiveConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_detective_member" {
			continue
		}

		graphArn, accountId, err := tfdetective.DecodeMemberID(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := tfdetective.FindMemberByGraphArnAndAccountID(context.Background(), conn, graphArn, accountId)
		if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil {
			return fmt.Errorf("detective Member %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccMemberConfig_basic(email string) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
data "aws_caller_identity" "member" {
  provider = "awsalternate"
}

resource "aws_detective_graph" "test" {}

resource "aws_detective_member" "test" {
  account_id    = data.aws_caller_identity.member.account_id
  graph_arn     = aws_detective_graph.test.id
  email_address = %[1]q
}
`, email)
}

func testAccMemberConfig_invitationMessage(email string, invite bool) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
data "aws_caller_identity" "member" {
  provider = "awsalternate"
}

resource "aws_detective_graph" "test" {}

resource "aws_detective_member" "test" {
  account_id    = data.aws_caller_identity.member.account_id
  graph_arn     = aws_detective_graph.test.id
  email_address = %[1]q
  message       = "This is a message of the invitation"
}
`, email, invite)
}
