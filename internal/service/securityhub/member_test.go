// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccMember_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var member types.Member
	resourceName := "aws_securityhub_member.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_basic("111111111111"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName, &member),
					resource.TestCheckResourceAttr(resourceName, names.AttrAccountID, "111111111111"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEmail, ""),
					resource.TestCheckResourceAttr(resourceName, "invite", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "member_status", "Created"),
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

func testAccMember_invite(t *testing.T) {
	ctx := acctest.Context(t)
	var member types.Member
	resourceName := "aws_securityhub_member.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_invite("111111111111", acctest.DefaultEmailAddress, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName, &member),
					resource.TestCheckResourceAttr(resourceName, names.AttrAccountID, "111111111111"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEmail, acctest.DefaultEmailAddress),
					resource.TestCheckResourceAttr(resourceName, "invite", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "member_status", "Invited"),
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

func testAccCheckMemberExists(ctx context.Context, n string, v *types.Member) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		output, err := tfsecurityhub.FindMemberByAccountID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckMemberDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_member" {
				continue
			}

			_, err := tfsecurityhub.FindMemberByAccountID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Hub Member %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccMemberConfig_basic(accountID string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_member" "test" {
  depends_on = [aws_securityhub_account.test]
  account_id = %[1]q
}
`, accountID)
}

func testAccMemberConfig_invite(accountID, email string, invite bool) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_member" "test" {
  depends_on = [aws_securityhub_account.test]
  account_id = %[1]q
  email      = %[2]q
  invite     = %[3]t
}
`, accountID, email, invite)
}
