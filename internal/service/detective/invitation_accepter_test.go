// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package detective_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdetective "github.com/hashicorp/terraform-provider-aws/internal/service/detective"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccInvitationAccepter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_detective_invitation_accepter.test"
	email := testAccMemberFromEnv(t)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckInvitationAccepterDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.DetectiveServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccInvitationAccepterConfig_basic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvitationAccepterExists(ctx, t, resourceName),
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

func testAccCheckInvitationAccepterExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DetectiveClient(ctx)

		_, err := tfdetective.FindInvitationByGraphARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckInvitationAccepterDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DetectiveClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_detective_invitation_accepter" {
				continue
			}

			_, err := tfdetective.FindInvitationByGraphARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Detective Invitation Accepter %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccInvitationAccepterConfig_basic(email string) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
data "aws_caller_identity" "member" {
  provider = "awsalternate"
}

resource "aws_detective_graph" "test" {}

resource "aws_detective_member" "test" {
  account_id    = data.aws_caller_identity.member.account_id
  graph_arn     = aws_detective_graph.test.graph_arn
  email_address = %[1]q
  message       = "This is a message of the invite"
}

resource "aws_detective_invitation_accepter" "test" {
  provider  = "awsalternate"
  graph_arn = aws_detective_member.test.graph_arn

  depends_on = [aws_detective_member.test]
}
`, email)
}
