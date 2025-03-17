// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfkms "github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKMSKeyPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"
	attachmentResourceName := "aws_kms_key_policy.test"
	expectedPolicyText := fmt.Sprintf(`{"Version":"2012-10-17","Id":%[1]q,"Statement":[{"Sid":"Enable IAM User Permissions","Effect":"Allow","Principal":{"AWS":"*"},"Action":"kms:*","Resource":"*"}]}`, rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, keyResourceName, &key),
					testAccCheckKeyHasPolicy(ctx, keyResourceName, expectedPolicyText),
				),
			},
			{
				ResourceName:            attachmentResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bypass_policy_lockout_safety_check"},
			},
			{
				Config: testAccKeyPolicyConfig_removedPolicy(keyResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, keyResourceName, &key),
				),
			},
		},
	})
}

func TestAccKMSKeyPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	attachmentResourceName := "aws_kms_key_policy.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, attachmentResourceName, &key),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkms.ResourceKey(), attachmentResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKMSKeyPolicy_bypass(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"
	attachmentResourceName := "aws_kms_key_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config:      testAccKeyPolicyConfig_policyBypass(rName, false),
				ExpectError: regexache.MustCompile(`The new key policy will not allow you to update the key policy in the future`),
			},
			{
				Config: testAccKeyPolicyConfig_policyBypass(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, keyResourceName, &key),
					resource.TestCheckResourceAttr(attachmentResourceName, "bypass_policy_lockout_safety_check", acctest.CtTrue),
				),
			},
			{
				ResourceName:            attachmentResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bypass_policy_lockout_safety_check"},
			},
		},
	})
}

func TestAccKMSKeyPolicy_bypassUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"
	attachmentResourceName := "aws_kms_key_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, keyResourceName, &before),
					resource.TestCheckResourceAttr(attachmentResourceName, "bypass_policy_lockout_safety_check", acctest.CtFalse),
				),
			},
			{
				Config: testAccKeyPolicyConfig_policyBypass(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, keyResourceName, &after),
					resource.TestCheckResourceAttr(attachmentResourceName, "bypass_policy_lockout_safety_check", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccKMSKeyPolicy_keyIsEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyConfig_keyIsEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, keyResourceName, &before),
				),
			},
			{
				Config: testAccKeyPolicyConfig_keyIsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, keyResourceName, &after),
				),
			},
		},
	})
}

func TestAccKMSKeyPolicy_iamRole(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyConfig_policyIAMRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, keyResourceName, &key),
				),
			},
		},
	})
}

func TestAccKMSKeyPolicy_iamRoleUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, keyResourceName, &key),
				),
			},
			{
				Config: testAccKeyPolicyConfig_policyIAMRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, keyResourceName, &key),
				),
			},
		},
	})
}

// // Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11801
func TestAccKMSKeyPolicy_iamRoleOrder(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyConfig_policyIAMMultiRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, keyResourceName, &key),
				),
			},
			{
				Config: testAccKeyPolicyConfig_policyIAMMultiRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, keyResourceName, &key),
				),
				PlanOnly: true,
			},
			{
				Config: testAccKeyPolicyConfig_policyIAMMultiRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, keyResourceName, &key),
				),
				PlanOnly: true,
			},
			{
				Config: testAccKeyPolicyConfig_policyIAMMultiRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, keyResourceName, &key),
				),
				PlanOnly: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7646.
func TestAccKMSKeyPolicy_iamServiceLinkedRole(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyConfig_policyIAMServiceLinkedRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, keyResourceName, &key),
				),
			},
		},
	})
}

func TestAccKMSKeyPolicy_booleanCondition(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyConfig_policyBooleanCondition(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, keyResourceName, &key),
				),
			},
		},
	})
}

func testAccKeyPolicyConfig_policy(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}
resource "aws_kms_key_policy" "test" {
  key_id = aws_kms_key.test.id
  policy = jsonencode({
    Id = %[1]q
    Statement = [{
      Sid    = "Enable IAM User Permissions"
      Effect = "Allow"
      Principal = {
        "AWS" : "*"
      }
      Action   = "kms:*"
      Resource = "*"
    }]
    Version = "2012-10-17"
  })
}
`, rName)
}

func testAccKeyPolicyConfig_policyBypass(rName string, bypassFlag bool) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_key_policy" "test" {
  key_id                             = aws_kms_key.test.id
  bypass_policy_lockout_safety_check = %[2]t

  policy = jsonencode({
    Id = %[1]q
    Statement = [
      {
        Action = [
          "kms:CreateKey",
          "kms:DescribeKey",
          "kms:ScheduleKeyDeletion",
          "kms:Describe*",
          "kms:Get*",
          "kms:List*",
          "kms:TagResource",
          "kms:UntagResource",
        ]
        Effect = "Allow"
        Principal = {
          AWS = data.aws_caller_identity.current.arn
        }
        Resource = "*"
        Sid      = "Enable IAM User Permissions"
      },
    ]
    Version = "2012-10-17"
  })
}
`, rName, bypassFlag)
}

func testAccKeyPolicyConfig_policyIAMRole(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_key_policy" "test" {
  key_id = aws_kms_key.test.id
  policy = jsonencode({
    Id = %[1]q
    Statement = [
      {
        Action = "kms:*"
        Effect = "Allow"
        Principal = {
          AWS = data.aws_caller_identity.current.arn
        }

        Resource = "*"
        Sid      = "Enable IAM User Permissions"
      },
      {
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:ReEncrypt*",
          "kms:GenerateDataKey*",
          "kms:DescribeKey",
        ]
        Effect = "Allow"
        Principal = {
          AWS = aws_iam_role.test.arn
        }

        Resource = "*"
        Sid      = "Enable IAM User Permissions"
      },
    ]
    Version = "2012-10-17"
  })
}
`, rName)
}

func testAccKeyPolicyConfig_policyIAMMultiRole(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test1" {
  name = "%[1]s-sultan"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role" "test2" {
  name = "%[1]s-shepard"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role" "test3" {
  name = "%[1]s-tritonal"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role" "test4" {
  name = "%[1]s-artlec"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role" "test5" {
  name = "%[1]s-cazzette"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

data "aws_iam_policy_document" "test" {
  policy_id = %[1]q
  statement {
    actions = [
      "kms:*",
    ]
    effect = "Allow"
    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
    resources = ["*"]
  }

  statement {
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
      "kms:DescribeKey",
    ]
    effect = "Allow"
    principals {
      identifiers = [
        aws_iam_role.test2.arn,
        aws_iam_role.test1.arn,
        aws_iam_role.test4.arn,
        aws_iam_role.test3.arn,
        aws_iam_role.test5.arn,
      ]
      type = "AWS"
    }
    resources = ["*"]
  }
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_key_policy" "test" {
  key_id = aws_kms_key.test.id
  policy = data.aws_iam_policy_document.test.json
}
`, rName)
}

func testAccKeyPolicyConfig_policyIAMServiceLinkedRole(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_iam_service_linked_role" "test" {
  aws_service_name = "autoscaling.${data.aws_partition.current.dns_suffix}"
  custom_suffix    = %[1]q
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_key_policy" "test" {
  key_id = aws_kms_key.test.id
  policy = jsonencode({
    Id = %[1]q
    Statement = [
      {
        Action = "kms:*"
        Effect = "Allow"
        Principal = {
          AWS = data.aws_caller_identity.current.arn
        }

        Resource = "*"
        Sid      = "Enable IAM User Permissions"
      },
      {
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:ReEncrypt*",
          "kms:GenerateDataKey*",
          "kms:DescribeKey",
        ]
        Effect = "Allow"
        Principal = {
          AWS = aws_iam_service_linked_role.test.arn
        }

        Resource = "*"
        Sid      = "Enable IAM User Permissions"
      },
    ]
    Version = "2012-10-17"
  })
}
`, rName)
}

func testAccKeyPolicyConfig_policyBooleanCondition(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}
resource "aws_kms_key_policy" "test" {
  key_id = aws_kms_key.test.id
  policy = jsonencode({
    Id = %[1]q
    Statement = [
      {
        Action = "kms:*"
        Effect = "Allow"
        Principal = {
          AWS = data.aws_caller_identity.current.arn
        }

        Resource = "*"
        Sid      = "Enable IAM User Permissions"
      },
      {
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:ReEncrypt*",
          "kms:GenerateDataKey*",
          "kms:DescribeKey",
        ]
        Effect = "Allow"
        Principal = {
          AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        }

        Resource = "*"
        Sid      = "Enable IAM User Permissions"

        Condition = {
          Bool = {
            "kms:GrantIsForAWSResource" = true
          }
        }
      },
    ]
    Version = "2012-10-17"
  })
}
`, rName)
}

func testAccKeyPolicyConfig_removedPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}
`, rName)
}

func testAccKeyPolicyConfig_keyIsEnabled(rName string, isEnabled bool) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  is_enabled              = %[2]t
}
resource "aws_kms_key_policy" "test" {
  key_id = aws_kms_key.test.id
  policy = jsonencode({
    Id = %[1]q
    Statement = [{
      Sid    = "Enable IAM User Permissions"
      Effect = "Allow"
      Principal = {
        AWS = data.aws_caller_identity.current.arn
      }
      Action   = "kms:*"
      Resource = "*"
    }]
    Version = "2012-10-17"
  })
}
`, rName, isEnabled)
}
