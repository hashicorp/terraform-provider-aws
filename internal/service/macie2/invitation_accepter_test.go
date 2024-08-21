// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/macie2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/envvar"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccInvitationAccepter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_macie2_invitation_accepter.member"
	email := envvar.SkipIfEmpty(t, envVarPrincipalEmail, envVarPrincipalEmailMessageError)

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
				Config: testAccInvitationAccepterConfig_basic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvitationAccepterExists(ctx, resourceName),
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

func testAccCheckInvitationAccepterExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource (%s) has empty ID", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Client(ctx)
		input := &macie2.GetAdministratorAccountInput{}
		output, err := conn.GetAdministratorAccount(ctx, input)

		if err != nil {
			return err
		}

		if output == nil || output.Administrator == nil || aws.ToString(output.Administrator.AccountId) == "" {
			return fmt.Errorf("no administrator account found for: %s", resourceName)
		}

		return nil
	}
}

func testAccCheckInvitationAccepterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_macie2_invitation_accepter" {
				continue
			}

			input := &macie2.GetAdministratorAccountInput{}
			output, err := conn.GetAdministratorAccount(ctx, input)
			if errs.IsA[*awstypes.ResourceNotFoundException](err) ||
				errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Macie is not enabled") {
				continue
			}

			if output == nil || output.Administrator == nil || aws.ToString(output.Administrator.AccountId) != rs.Primary.Attributes["administrator_account_id"] {
				continue
			}

			return fmt.Errorf("macie InvitationAccepter %q still exists", rs.Primary.ID)
		}

		return nil
	}
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
