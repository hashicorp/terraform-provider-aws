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

func testAccDetectiveMember_basic(t *testing.T) {
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
		CheckDestroy:      testAccCheckDetectiveMemberDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, detective.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectiveMemberConfigBasic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectiveMemberExists(resourceName, &detectiveOutput),
					acctest.CheckResourceAttrAccountID(resourceName, "administrator_id"),
					resource.TestCheckResourceAttrPair(resourceName, "account_id", dataSourceAlternate, "account_id"),
					acctest.CheckResourceAttrRFC3339(resourceName, "invited_time"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_time"),
					resource.TestCheckResourceAttr(resourceName, "status", detective.MemberStatusInvited),
				),
			},
			{
				Config:                  testAccDetectiveMemberConfigBasic(email),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"disable_email_notification"},
			},
		},
	})
}

func testAccDetectiveMember_disappears(t *testing.T) {
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
		CheckDestroy:      testAccCheckDetectiveMemberDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, detective.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectiveMemberConfigBasic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectiveMemberExists(resourceName, &detectiveOutput),
					acctest.CheckResourceDisappears(acctest.Provider, tfdetective.ResourceMember(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDetectiveMember_message(t *testing.T) {
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
		CheckDestroy:      testAccCheckDetectiveInvitationAccepterDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, detective.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectiveMemberConfigInvitationMessage(email, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectiveMemberExists(resourceName, &detectiveOutput),
					acctest.CheckResourceAttrAccountID(resourceName, "administrator_id"),
					resource.TestCheckResourceAttrPair(resourceName, "account_id", dataSourceAlternate, "account_id"),
					acctest.CheckResourceAttrRFC3339(resourceName, "invited_time"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_time"),
					resource.TestCheckResourceAttr(resourceName, "status", detective.MemberStatusInvited),
				),
			},
			{
				Config: testAccDetectiveMemberConfigInvitationMessage(email, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectiveMemberExists(resourceName, &detectiveOutput),
					acctest.CheckResourceAttrAccountID(resourceName, "administrator_id"),
					resource.TestCheckResourceAttrPair(resourceName, "account_id", dataSourceAlternate, "account_id"),
					acctest.CheckResourceAttrRFC3339(resourceName, "invited_time"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_time"),
					resource.TestCheckResourceAttr(resourceName, "status", detective.MemberStatusInvited),
				),
			},
			{
				Config:                  testAccDetectiveMemberConfigInvitationMessage(email, true),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"message", "disable_email_notification"},
			},
		},
	})
}

func testAccCheckDetectiveMemberExists(resourceName string, detectiveSession *detective.MemberDetail) resource.TestCheckFunc {
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

func testAccCheckDetectiveMemberDestroy(s *terraform.State) error {
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

func testAccDetectiveMemberConfigBasic(email string) string {
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

func testAccDetectiveMemberConfigInvitationMessage(email string, invite bool) string {
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
