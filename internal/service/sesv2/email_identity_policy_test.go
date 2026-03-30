// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2EmailIdentityPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_email_identity_policy.test"
	emailIdentity := acctest.RandomEmailAddress(acctest.RandomDomainName())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityPolicyConfig_basic(emailIdentity, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityPolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "email_identity", emailIdentity),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
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

func TestAccSESV2EmailIdentityPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_email_identity_policy.test"
	emailIdentity := acctest.RandomEmailAddress(acctest.RandomDomainName())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityPolicyConfig_basic(emailIdentity, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityPolicyExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsesv2.ResourceEmailIdentityPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEmailIdentityPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_email_identity_policy" {
				continue
			}

			_, err := tfsesv2.FindEmailIdentityPolicyByTwoPartKey(ctx, conn, rs.Primary.Attributes["email_identity"], rs.Primary.Attributes["policy_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SESv2 Email Identity Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckEmailIdentityPolicyExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		_, err := tfsesv2.FindEmailIdentityPolicyByTwoPartKey(ctx, conn, rs.Primary.Attributes["email_identity"], rs.Primary.Attributes["policy_name"])

		return err
	}
}

func testAccEmailIdentityPolicyConfig_basic(emailIdentity, rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_sesv2_email_identity" "test" {
  email_identity = %[1]q
}

resource "aws_sesv2_email_identity_policy" "test" {
  email_identity = aws_sesv2_email_identity.test.email_identity
  policy_name    = %[2]q

  policy = <<EOF
{
  "Version":"2012-10-17",
  "Statement":[
    {
      "Effect":"Allow",
      "Resource":"${aws_sesv2_email_identity.test.arn}",
      "Principal":{
        "AWS":"arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action":"ses:SendEmail"
    }
  ]
}
EOF
}
`, emailIdentity, rName)
}
