// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/efs"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfefs "github.com/hashicorp/terraform-provider-aws/internal/service/efs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEFSFileSystemPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var desc efs.DescribeFileSystemPolicyOutput
	resourceName := "aws_efs_file_system_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemPolicyExists(ctx, resourceName, &desc),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bypass_policy_lockout_safety_check"},
			},
			{
				Config: testAccFileSystemPolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemPolicyExists(ctx, resourceName, &desc),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
		},
	})
}

func TestAccEFSFileSystemPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var desc efs.DescribeFileSystemPolicyOutput
	resourceName := "aws_efs_file_system_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemPolicyExists(ctx, resourceName, &desc),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfefs.ResourceFileSystemPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEFSFileSystemPolicy_policyBypass(t *testing.T) {
	ctx := acctest.Context(t)
	var desc efs.DescribeFileSystemPolicyOutput
	resourceName := "aws_efs_file_system_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemPolicyExists(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bypass_policy_lockout_safety_check"},
			},
			{
				Config: testAccFileSystemPolicyConfig_bypass(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemPolicyExists(ctx, resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", acctest.CtTrue),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/21968
func TestAccEFSFileSystemPolicy_equivalentPolicies(t *testing.T) {
	ctx := acctest.Context(t)
	var desc efs.DescribeFileSystemPolicyOutput
	resourceName := "aws_efs_file_system_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemPolicyConfig_firstEquivalent(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemPolicyExists(ctx, resourceName, &desc),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
			{
				Config:   testAccFileSystemPolicyConfig_secondEquivalent(rName),
				PlanOnly: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19245
func TestAccEFSFileSystemPolicy_equivalentPoliciesIAMPolicyDoc(t *testing.T) {
	ctx := acctest.Context(t)
	var desc efs.DescribeFileSystemPolicyOutput
	resourceName := "aws_efs_file_system_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemPolicyConfig_equivalentIAMDoc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemPolicyExists(ctx, resourceName, &desc),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
			{
				Config:   testAccFileSystemPolicyConfig_equivalentIAMDoc(rName),
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckFileSystemPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EFSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_efs_file_system_policy" {
				continue
			}

			_, err := tfefs.FindFileSystemPolicyByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EFS File System Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFileSystemPolicyExists(ctx context.Context, n string, v *efs.DescribeFileSystemPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EFSClient(ctx)

		output, err := tfefs.FindFileSystemPolicyByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccFileSystemPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_file_system_policy" "test" {
  file_system_id = aws_efs_file_system.test.id

  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Id": "ExamplePolicy01",
    "Statement": [
        {
            "Sid": "ExampleStatement01",
            "Effect": "Allow",
            "Principal": {
                "AWS": "*"
            },
            "Resource": "${aws_efs_file_system.test.arn}",
            "Action": [
                "elasticfilesystem:ClientMount",
                "elasticfilesystem:ClientWrite"
            ],
            "Condition": {
                "Bool": {
                    "aws:SecureTransport": "true"
                }
            }
        }
    ]
}
POLICY
}
`, rName)
}

func testAccFileSystemPolicyConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_file_system_policy" "test" {
  file_system_id = aws_efs_file_system.test.id

  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Id": "ExamplePolicy01",
    "Statement": [
        {
            "Sid": "ExampleStatement01",
            "Effect": "Allow",
            "Principal": {
                "AWS": "*"
            },
            "Resource": "${aws_efs_file_system.test.arn}",
            "Action": "elasticfilesystem:ClientMount",
            "Condition": {
                "Bool": {
                    "aws:SecureTransport": "true"
                }
            }
        }
    ]
}
POLICY
}
`, rName)
}

func testAccFileSystemPolicyConfig_bypass(rName string, bypass bool) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_file_system_policy" "test" {
  file_system_id = aws_efs_file_system.test.id

  bypass_policy_lockout_safety_check = %[2]t

  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Id": "ExamplePolicy01",
    "Statement": [
        {
            "Sid": "ExampleStatement01",
            "Effect": "Allow",
            "Principal": {
                "AWS": "*"
            },
            "Resource": "${aws_efs_file_system.test.arn}",
            "Action": [
                "elasticfilesystem:ClientMount",
                "elasticfilesystem:ClientWrite"
            ],
            "Condition": {
                "Bool": {
                    "aws:SecureTransport": "true"
                }
            }
        }
    ]
}
POLICY
}
`, rName, bypass)
}

func testAccFileSystemPolicyConfig_firstEquivalent(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_file_system_policy" "test" {
  file_system_id = aws_efs_file_system.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "ExamplePolicy01"
    Statement = [{
      Sid    = "ExampleStatement01"
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Resource = aws_efs_file_system.test.arn
      Action = [
        "elasticfilesystem:ClientMount",
        "elasticfilesystem:ClientWrite",
      ]
      Condition = {
        Bool = {
          "aws:SecureTransport" = "true"
        }
      }
    }]
  })
}
`, rName)
}

func testAccFileSystemPolicyConfig_secondEquivalent(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_file_system_policy" "test" {
  file_system_id = aws_efs_file_system.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "ExamplePolicy01"
    Statement = [{
      Sid    = "ExampleStatement01"
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Resource = aws_efs_file_system.test.arn
      Action = [
        "elasticfilesystem:ClientWrite",
        "elasticfilesystem:ClientMount",
      ]
      Condition = {
        Bool = {
          "aws:SecureTransport" = ["true"]
        }
      }
    }]
  })
}
`, rName)
}

func testAccFileSystemPolicyConfig_equivalentIAMDoc(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_file_system_policy" "test" {
  file_system_id = aws_efs_file_system.test.id
  policy         = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  version = "2012-10-17"

  statement {
    sid = "Allow mount and write"

    actions = [
      "elasticfilesystem:ClientWrite",
      "elasticfilesystem:ClientRootAccess",
      "elasticfilesystem:ClientMount",
    ]

    resources = [aws_efs_file_system.test.arn]

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
  }

  statement {
    sid       = "Enforce in-transit encryption for all clients"
    effect    = "Deny"
    actions   = ["*"]
    resources = [aws_efs_file_system.test.arn]

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["false"]
    }
  }
}
`, rName)
}
