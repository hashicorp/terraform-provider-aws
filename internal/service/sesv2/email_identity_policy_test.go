// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2EmailIdentityPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_email_identity_policy.test"
	emailIdentity := acctest.RandomEmailAddress(acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityPolicyConfig_basic(emailIdentity, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityPolicyExists(ctx, resourceName),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_email_identity_policy.test"
	emailIdentity := acctest.RandomEmailAddress(acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityPolicyConfig_basic(emailIdentity, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsesv2.ResourceEmailIdentityPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEmailIdentityPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_email_identity_policy" {
				continue
			}

			_, err := tfsesv2.FindEmailIdentityPolicyByID(ctx, conn, rs.Primary.ID)

			if err != nil {
				var nfe *types.NotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.SESV2, create.ErrActionCheckingDestroyed, tfsesv2.ResNameEmailIdentityPolicy, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckEmailIdentityPolicyExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameEmailIdentityPolicy, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameEmailIdentityPolicy, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

		_, err := tfsesv2.FindEmailIdentityPolicyByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameEmailIdentityPolicy, rs.Primary.ID, err)
		}

		return nil
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
