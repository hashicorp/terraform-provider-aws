// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMAccountPasswordPolicy_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccAccountPasswordPolicy_basic,
		acctest.CtDisappears: testAccAccountPasswordPolicy_disappears,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccAccountPasswordPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.PasswordPolicy
	resourceName := "aws_iam_account_password_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPasswordPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPasswordPolicyConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPasswordPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "minimum_password_length", "8"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountPasswordPolicyConfig_modified,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPasswordPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "minimum_password_length", "7"),
				),
			},
		},
	})
}

func testAccAccountPasswordPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.PasswordPolicy
	resourceName := "aws_iam_account_password_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPasswordPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPasswordPolicyConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPasswordPolicyExists(ctx, resourceName, &policy),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceAccountPasswordPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAccountPasswordPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_account_password_policy" {
				continue
			}

			_, err := tfiam.FindAccountPasswordPolicy(ctx, conn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM Account Password Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccountPasswordPolicyExists(ctx context.Context, n string, v *awstypes.PasswordPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IAM Account Password Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		output, err := tfiam.FindAccountPasswordPolicy(ctx, conn)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

const testAccAccountPasswordPolicyConfig_basic = `
resource "aws_iam_account_password_policy" "test" {
  allow_users_to_change_password = true
  minimum_password_length        = 8
  require_numbers                = true
}
`

const testAccAccountPasswordPolicyConfig_modified = `
resource "aws_iam_account_password_policy" "test" {
  allow_users_to_change_password = true
  minimum_password_length        = 7
  require_numbers                = false
  require_symbols                = true
  require_uppercase_characters   = true
}
`
