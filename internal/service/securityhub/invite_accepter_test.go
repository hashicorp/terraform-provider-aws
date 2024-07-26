// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccInviteAccepter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_invite_accepter.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckInviteAccepterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInviteAccepterConfig_basic(acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInviteAccepterExists(ctx, resourceName),
				),
			},
			{
				Config:            testAccInviteAccepterConfig_basic(acctest.DefaultEmailAddress),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckInviteAccepterExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		_, err := tfsecurityhub.FindMasterAccount(ctx, conn)

		return err
	}
}

func testAccCheckInviteAccepterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_invite_accepter" {
				continue
			}

			_, err := tfsecurityhub.FindMasterAccount(ctx, conn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Hub Master Account (%s) still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccInviteAccepterConfig_basic(email string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
resource "aws_securityhub_invite_accepter" "test" {
  master_id = aws_securityhub_member.source.master_id

  depends_on = [aws_securityhub_account.test]
}

resource "aws_securityhub_member" "source" {
  provider = awsalternate

  account_id = data.aws_caller_identity.test.account_id
  email      = %[1]q
  invite     = true

  depends_on = [aws_securityhub_account.source]
}

resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_account" "source" {
  provider = awsalternate
}

data "aws_caller_identity" "test" {}
`, email))
}
