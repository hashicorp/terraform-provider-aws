// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package detective_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/detective/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdetective "github.com/hashicorp/terraform-provider-aws/internal/service/detective"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccMember_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var detectiveOutput awstypes.MemberDetail
	resourceName := "aws_detective_member.test"
	dataSourceAlternate := "data.aws_caller_identity.member"
	email := testAccMemberFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckMemberDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.DetectiveServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_basic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName, &detectiveOutput),
					acctest.CheckResourceAttrAccountID(resourceName, "administrator_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAccountID, dataSourceAlternate, names.AttrAccountID),
					acctest.CheckResourceAttrRFC3339(resourceName, "invited_time"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_time"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.MemberStatusInvited)),
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
	ctx := acctest.Context(t)
	var detectiveOutput awstypes.MemberDetail
	resourceName := "aws_detective_member.test"
	email := testAccMemberFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckMemberDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.DetectiveServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_basic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName, &detectiveOutput),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdetective.ResourceMember(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccMember_message(t *testing.T) {
	ctx := acctest.Context(t)
	var detectiveOutput awstypes.MemberDetail
	resourceName := "aws_detective_member.test"
	dataSourceAlternate := "data.aws_caller_identity.member"
	email := testAccMemberFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckInvitationAccepterDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.DetectiveServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_invitationMessage(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName, &detectiveOutput),
					acctest.CheckResourceAttrAccountID(resourceName, "administrator_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAccountID, dataSourceAlternate, names.AttrAccountID),
					acctest.CheckResourceAttrRFC3339(resourceName, "invited_time"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_time"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.MemberStatusInvited)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrMessage, "disable_email_notification"},
			},
		},
	})
}

func testAccCheckMemberExists(ctx context.Context, n string, v *awstypes.MemberDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DetectiveClient(ctx)

		graphARN, accountID, err := tfdetective.MemberParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		output, err := tfdetective.FindMemberByGraphByTwoPartKey(ctx, conn, graphARN, accountID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckMemberDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DetectiveClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_detective_member" {
				continue
			}

			graphARN, accountID, err := tfdetective.MemberParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfdetective.FindMemberByGraphByTwoPartKey(ctx, conn, graphARN, accountID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Detective Member %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccMemberConfig_basic(email string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "member" {
  provider = "awsalternate"
}

resource "aws_detective_graph" "test" {}

resource "aws_detective_member" "test" {
  account_id    = data.aws_caller_identity.member.account_id
  graph_arn     = aws_detective_graph.test.id
  email_address = %[1]q
}
`, email))
}

func testAccMemberConfig_invitationMessage(email string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
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
`, email))
}
