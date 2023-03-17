package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
)

func testAccMember_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var member securityhub.Member
	resourceName := "aws_securityhub_member.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_basic("111111111111"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName, &member),
					resource.TestCheckResourceAttr(resourceName, "account_id", "111111111111"),
					resource.TestCheckResourceAttr(resourceName, "email", ""),
					resource.TestCheckResourceAttr(resourceName, "invite", "false"),
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
	var member securityhub.Member
	resourceName := "aws_securityhub_member.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberConfig_invite("111111111111", acctest.DefaultEmailAddress, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberExists(ctx, resourceName, &member),
					resource.TestCheckResourceAttr(resourceName, "account_id", "111111111111"),
					resource.TestCheckResourceAttr(resourceName, "email", acctest.DefaultEmailAddress),
					resource.TestCheckResourceAttr(resourceName, "invite", "true"),
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

func testAccCheckMemberExists(ctx context.Context, n string, member *securityhub.Member) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn()

		resp, err := conn.GetMembersWithContext(ctx, &securityhub.GetMembersInput{
			AccountIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if len(resp.Members) == 0 {
			return fmt.Errorf("Security Hub member %s not found", rs.Primary.ID)
		}

		member = resp.Members[0]

		return nil
	}
}

func testAccCheckMemberDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_member" {
				continue
			}

			resp, err := conn.GetMembersWithContext(ctx, &securityhub.GetMembersInput{
				AccountIds: []*string{aws.String(rs.Primary.ID)},
			})

			if tfawserr.ErrCodeEquals(err, tfsecurityhub.ErrCodeBadRequestException) {
				continue
			}

			if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error getting Security Hub Member (%s): %w", rs.Primary.ID, err)
			}

			if len(resp.Members) != 0 {
				return fmt.Errorf("Security Hub member still exists")
			}

			return nil
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
