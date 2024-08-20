// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecretsmanager "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecretsManagerSecretPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var policy secretsmanager.GetResourcePolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretPolicyExists(ctx, resourceName, &policy),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy,
						regexache.MustCompile(`{"Action":"secretsmanager:GetSecretValue".+`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"block_public_policy"},
			},
			{
				Config: testAccSecretPolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretPolicyExists(ctx, resourceName, &policy),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy,
						regexache.MustCompile(`{"Action":"secretsmanager:\*".+`)),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecretPolicy_blockPublicPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var policy secretsmanager.GetResourcePolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretPolicyConfig_block(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"block_public_policy"},
			},
			{
				Config: testAccSecretPolicyConfig_block(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", acctest.CtFalse),
				),
			},
			{
				Config: testAccSecretPolicyConfig_block(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecretPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var policy secretsmanager.GetResourcePolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretPolicyExists(ctx, resourceName, &policy),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsecretsmanager.ResourceSecretPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSecretsManagerSecretPolicy_Disappears_secret(t *testing.T) {
	ctx := acctest.Context(t)
	var policy secretsmanager.GetResourcePolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_policy.test"
	secretResourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretPolicyExists(ctx, resourceName, &policy),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsecretsmanager.ResourceSecret(), secretResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSecretPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_secretsmanager_secret_policy" {
				continue
			}

			output, err := tfsecretsmanager.FindSecretPolicyByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if aws.ToString(output.ResourcePolicy) == "" {
				continue
			}

			return fmt.Errorf("Secrets Manager Secret Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSecretPolicyExists(ctx context.Context, n string, v *secretsmanager.GetResourcePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerClient(ctx)

		output, err := tfsecretsmanager.FindSecretPolicyByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSecretPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_policy" "test" {
  secret_arn = aws_secretsmanager_secret.test.arn

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
	{
	  "Sid": "EnableAllPermissions",
	  "Effect": "Allow",
	  "Principal": {
		"AWS": "${aws_iam_role.test.arn}"
	  },
	  "Action": "secretsmanager:GetSecretValue",
	  "Resource": "*"
	}
  ]
}
POLICY
}
`, rName)
}

func testAccSecretPolicyConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_policy" "test" {
  secret_arn = aws_secretsmanager_secret.test.arn

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
	{
	  "Sid": "EnableAllPermissions",
	  "Effect": "Allow",
	  "Principal": {
		"AWS": "*"
	  },
	  "Action": "secretsmanager:*",
	  "Resource": "*"
	}
  ]
}
POLICY
}
`, rName)
}

func testAccSecretPolicyConfig_block(rName string, block bool) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_policy" "test" {
  secret_arn          = aws_secretsmanager_secret.test.arn
  block_public_policy = %[2]t

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
	{
	  "Sid": "EnableAllPermissions",
	  "Effect": "Allow",
	  "Principal": {
		"AWS": "${aws_iam_role.test.arn}"
	  },
	  "Action": "secretsmanager:GetSecretValue",
	  "Resource": "*"
	}
  ]
}
POLICY
}
`, rName, block)
}
