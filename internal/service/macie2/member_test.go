// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/macie2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/macie2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/envvar"
	tfmacie2 "github.com/hashicorp/terraform-provider-aws/internal/service/macie2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	envVarPrincipalEmail             = "AWS_MACIE2_ACCOUNT_EMAIL"
	envVarAlternateEmail             = "AWS_MACIE2_ALTERNATE_ACCOUNT_EMAIL"
	envVarPrincipalEmailMessageError = "Environment variable AWS_MACIE2_ACCOUNT_EMAIL is not set. " +
		"To properly test inviting Macie member account must be provided."
	envVarAlternateEmailMessageError = "Environment variable AWS_MACIE2_ALTERNATE_ACCOUNT_EMAIL is not set. " +
		"To properly test inviting Macie member account must be provided."
)

func testAccMember_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetMemberOutput
	resourceName := "aws_macie2_member.member"
	dataSourceAlternate := "data.aws_caller_identity.member"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckMemberDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_basic(acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", string(awstypes.RelationshipStatusCreated)),
					acctest.CheckResourceAttrAccountID(resourceName, "administrator_account_id"),
					acctest.CheckResourceAttrAccountID(resourceName, "master_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAccountID, dataSourceAlternate, names.AttrAccountID),
					acctest.CheckResourceAttrRFC3339(resourceName, "invited_at"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.MacieStatusEnabled)),
				),
			},
			{
				Config:            testAccMemberConfig_basic(acctest.DefaultEmailAddress),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMember_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetMemberOutput
	resourceName := "aws_macie2_member.member"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckMemberDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_basic(acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfmacie2.ResourceMember(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccMember_invitationDisableEmailNotification(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetMemberOutput
	resourceName := "aws_macie2_member.member"
	email := envvar.SkipIfEmpty(t, envVarAlternateEmail, envVarAlternateEmailMessageError)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckInvitationAccepterDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_inviteInvitationDisableEmailNotification(email, acctest.CtTrue, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName, &macie2Output),
				),
			},
			{
				Config: testAccMemberConfig_inviteInvitationDisableEmailNotification(email, acctest.CtFalse, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName, &macie2Output),
				),
			},
			{
				Config:            testAccMemberConfig_inviteInvitationDisableEmailNotification(email, acctest.CtFalse, false),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"invitation_disable_email_notification",
					"invitation_message",
				},
			},
		},
	})
}

func testAccMember_invite(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetMemberOutput
	resourceName := "aws_macie2_member.member"
	dataSourceAlternate := "data.aws_caller_identity.member"
	email := envvar.SkipIfEmpty(t, envVarAlternateEmail, envVarAlternateEmailMessageError)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckInvitationAccepterDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_invite(email, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", string(awstypes.RelationshipStatusCreated)),
					resource.TestCheckResourceAttr(resourceName, "invite", acctest.CtFalse),
					acctest.CheckResourceAttrAccountID(resourceName, "administrator_account_id"),
					acctest.CheckResourceAttrAccountID(resourceName, "master_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAccountID, dataSourceAlternate, names.AttrAccountID),
					acctest.CheckResourceAttrRFC3339(resourceName, "invited_at"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.MacieStatusEnabled)),
				),
			},
			{
				Config: testAccMemberConfig_invite(email, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", string(awstypes.RelationshipStatusInvited)),
					resource.TestCheckResourceAttr(resourceName, "invite", acctest.CtTrue),
					acctest.CheckResourceAttrAccountID(resourceName, "administrator_account_id"),
					acctest.CheckResourceAttrAccountID(resourceName, "master_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAccountID, dataSourceAlternate, names.AttrAccountID),
					acctest.CheckResourceAttrRFC3339(resourceName, "invited_at"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.MacieStatusEnabled)),
				),
			},
			{
				Config:                  testAccMemberConfig_invite(email, true),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"invitation_message"},
			},
		},
	})
}

func testAccMember_inviteRemoved(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetMemberOutput
	resourceName := "aws_macie2_member.member"
	dataSourceAlternate := "data.aws_caller_identity.member"
	email := envvar.SkipIfEmpty(t, envVarAlternateEmail, envVarAlternateEmailMessageError)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckInvitationAccepterDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_invite(email, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", string(awstypes.RelationshipStatusInvited)),
					resource.TestCheckResourceAttr(resourceName, "invite", acctest.CtTrue),
					acctest.CheckResourceAttrAccountID(resourceName, "administrator_account_id"),
					acctest.CheckResourceAttrAccountID(resourceName, "master_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAccountID, dataSourceAlternate, names.AttrAccountID),
					acctest.CheckResourceAttrRFC3339(resourceName, "invited_at"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.MacieStatusEnabled)),
				),
			},
			{
				Config: testAccMemberConfig_invite(email, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", string(awstypes.RelationshipStatusRemoved)),
					resource.TestCheckResourceAttr(resourceName, "invite", acctest.CtFalse),
					acctest.CheckResourceAttrAccountID(resourceName, "administrator_account_id"),
					acctest.CheckResourceAttrAccountID(resourceName, "master_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAccountID, dataSourceAlternate, names.AttrAccountID),
					acctest.CheckResourceAttrRFC3339(resourceName, "invited_at"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.MacieStatusEnabled)),
				),
			},
			{
				Config:                  testAccMemberConfig_invite(email, false),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"invitation_message"},
			},
		},
	})
}

func testAccMember_status(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetMemberOutput
	resourceName := "aws_macie2_member.member"
	dataSourceAlternate := "data.aws_caller_identity.member"
	email := envvar.SkipIfEmpty(t, envVarAlternateEmail, envVarAlternateEmailMessageError)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckInvitationAccepterDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_status(email, string(awstypes.MacieStatusEnabled), true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", string(awstypes.RelationshipStatusInvited)),
					resource.TestCheckResourceAttr(resourceName, "invite", acctest.CtTrue),
					acctest.CheckResourceAttrAccountID(resourceName, "administrator_account_id"),
					acctest.CheckResourceAttrAccountID(resourceName, "master_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAccountID, dataSourceAlternate, names.AttrAccountID),
					acctest.CheckResourceAttrRFC3339(resourceName, "invited_at"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.MacieStatusEnabled)),
				),
			},
			{
				Config: testAccMemberConfig_status(email, string(awstypes.MacieStatusPaused), true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", string(awstypes.RelationshipStatusPaused)),
					resource.TestCheckResourceAttr(resourceName, "invite", acctest.CtTrue),
					acctest.CheckResourceAttrAccountID(resourceName, "administrator_account_id"),
					acctest.CheckResourceAttrAccountID(resourceName, "master_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAccountID, dataSourceAlternate, names.AttrAccountID),
					acctest.CheckResourceAttrRFC3339(resourceName, "invited_at"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.MacieStatusPaused)),
				),
			},
			{
				Config:                  testAccMemberConfig_status(email, string(awstypes.MacieStatusPaused), true),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"invitation_message"},
			},
		},
	})
}

func testAccMember_withTags(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetMemberOutput
	resourceName := "aws_macie2_member.member"
	dataSourceAlternate := "data.aws_caller_identity.member"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckMemberDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_tags(acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrRFC3339(resourceName, "invited_at"),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", names.AttrValue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", names.AttrValue),
					acctest.CheckResourceAttrAccountID(resourceName, "administrator_account_id"),
					acctest.CheckResourceAttrAccountID(resourceName, "master_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAccountID, dataSourceAlternate, names.AttrAccountID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckMemberExists(ctx context.Context, n string, v *macie2.GetMemberOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Client(ctx)

		output, err := tfmacie2.FindMemberByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckMemberDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_macie2_member" {
				continue
			}

			_, err := tfmacie2.FindMemberByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Macie Member %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccMemberConfig_basic(email string) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
data "aws_caller_identity" "member" {
  provider = "awsalternate"
}

resource "aws_macie2_account" "admin" {}

resource "aws_macie2_member" "member" {
  account_id = data.aws_caller_identity.member.account_id
  email      = %[1]q
  depends_on = [aws_macie2_account.admin]
}
`, email)
}

func testAccMemberConfig_tags(email string) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
data "aws_caller_identity" "member" {
  provider = "awsalternate"
}

resource "aws_macie2_account" "admin" {}

resource "aws_macie2_member" "member" {
  account_id = data.aws_caller_identity.member.account_id
  email      = %[1]q
  tags = {
    Key = "value"
  }
  depends_on = [aws_macie2_account.admin]
}
`, email)
}

func testAccMemberConfig_invite(email string, invite bool) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
data "aws_caller_identity" "member" {
  provider = "awsalternate"
}

data "aws_caller_identity" "admin" {}

resource "aws_macie2_account" "admin" {}

resource "aws_macie2_account" "member" {
  provider = "awsalternate"
}

resource "aws_macie2_member" "member" {
  account_id         = data.aws_caller_identity.member.account_id
  email              = %[1]q
  invite             = %[2]t
  invitation_message = "This is a message of the invitation"
  depends_on         = [aws_macie2_account.admin]
}
`, email, invite)
}

func testAccMemberConfig_inviteInvitationDisableEmailNotification(email, disable string, invite bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
data "aws_caller_identity" "member" {
  provider = "awsalternate"
}

data "aws_caller_identity" "admin" {}

resource "aws_macie2_account" "admin" {}

resource "aws_macie2_account" "member" {
  provider = "awsalternate"
}

resource "aws_macie2_member" "member" {
  account_id                            = data.aws_caller_identity.member.account_id
  email                                 = %[1]q
  invitation_disable_email_notification = %[2]q
  invitation_message                    = "This is a message of the invitation"
  invite                                = %[3]t
  depends_on                            = [aws_macie2_account.admin]
}
`, email, disable, invite))
}

func testAccMemberConfig_status(email, memberStatus string, invite bool) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
data "aws_caller_identity" "member" {
  provider = "awsalternate"
}

data "aws_caller_identity" "admin" {}

resource "aws_macie2_account" "admin" {}

resource "aws_macie2_account" "member" {
  provider = "awsalternate"
}

resource "aws_macie2_member" "member" {
  account_id         = data.aws_caller_identity.member.account_id
  email              = %[1]q
  status             = %[2]q
  invite             = %[3]t
  invitation_message = "This is a message of the invitation"
  depends_on         = [aws_macie2_account.admin]
}

resource "aws_macie2_invitation_accepter" "member" {
  provider                 = "awsalternate"
  administrator_account_id = data.aws_caller_identity.admin.account_id
  depends_on               = [aws_macie2_member.member]
}
`, email, memberStatus, invite)
}
