package kms_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkms "github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccKMSKeyPolicyAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"
	attachmentResourceName := "aws_kms_key_policy_attachment.test"
	expectedPolicyText := fmt.Sprintf(`{"Version":"2012-10-17","Id":%[1]q,"Statement":[{"Sid":"Enable IAM User Permissions","Effect":"Allow","Principal":{"AWS":"*"},"Action":"kms:*","Resource":"*"}]}`, rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyAttachmentConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyPolicyAttachmentExists(ctx, keyResourceName, &key),
					testAccCheckKeyPolicyAttachmentHasPolicy(ctx, keyResourceName, expectedPolicyText),
				),
			},
			{
				ResourceName:            attachmentResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bypass_policy_lockout_safety_check"},
			},
			{
				Config: testAccKeyPolicyAttachmentConfig_removedPolicy(keyResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyPolicyAttachmentExists(ctx, keyResourceName, &key),
				),
			},
		},
	})
}

func TestAccKMSKeyPolicyAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	attachmentResourceName := "aws_kms_key_policy_attachment.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyAttachmentConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyPolicyAttachmentExists(ctx, attachmentResourceName, &key),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkms.ResourceKey(), attachmentResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKMSKeyPolicyAttachment_bypass(t *testing.T) {
	ctx := acctest.Context(t)
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"
	attachmentResourceName := "aws_kms_key_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccKeyPolicyAttachmentConfig_policyBypass(rName, false),
				ExpectError: regexp.MustCompile(`The new key policy will not allow you to update the key policy in the future`),
			},
			{
				Config: testAccKeyPolicyAttachmentConfig_policyBypass(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyPolicyAttachmentExists(ctx, keyResourceName, &key),
					resource.TestCheckResourceAttr(attachmentResourceName, "bypass_policy_lockout_safety_check", "true"),
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

func TestAccKMSKeyPolicyAttachment_bypassUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"
	attachmentResourceName := "aws_kms_key_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyAttachmentConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyPolicyAttachmentExists(ctx, keyResourceName, &before),
					resource.TestCheckResourceAttr(attachmentResourceName, "bypass_policy_lockout_safety_check", "false"),
				),
			},
			{
				Config: testAccKeyPolicyAttachmentConfig_policyBypass(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyPolicyAttachmentExists(ctx, keyResourceName, &after),
					resource.TestCheckResourceAttr(attachmentResourceName, "bypass_policy_lockout_safety_check", "true"),
				),
			},
		},
	})
}

func TestAccKMSKeyPolicyAttachment_keyIsEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyAttachmentConfig_keyIsEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyPolicyAttachmentExists(ctx, keyResourceName, &before),
				),
			},
			{
				Config: testAccKeyPolicyAttachmentConfig_keyIsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyPolicyAttachmentExists(ctx, keyResourceName, &after),
				),
			},
		},
	})
}

func TestAccKMSKeyPolicyAttachment_iamRole(t *testing.T) {
	ctx := acctest.Context(t)
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyAttachmentConfig_policyIAMRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyPolicyAttachmentExists(ctx, keyResourceName, &key),
				),
			},
			{
				ResourceName:            keyResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_window_in_days", "bypass_policy_lockout_safety_check"},
			},
		},
	})
}

func TestAccKMSKeyPolicyAttachment_iamRoleUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyAttachmentConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyPolicyAttachmentExists(ctx, keyResourceName, &key),
				),
			},
			{
				Config: testAccKeyPolicyAttachmentConfig_policyIAMRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyPolicyAttachmentExists(ctx, keyResourceName, &key),
				),
			},
		},
	})
}

// // Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11801
func TestAccKMSKeyPolicyAttachment_iamRoleOrder(t *testing.T) {
	ctx := acctest.Context(t)
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyAttachmentConfig_policyIAMMultiRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyPolicyAttachmentExists(ctx, keyResourceName, &key),
				),
			},
			{
				Config: testAccKeyPolicyAttachmentConfig_policyIAMMultiRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyPolicyAttachmentExists(ctx, keyResourceName, &key),
				),
				PlanOnly: true,
			},
			{
				Config: testAccKeyPolicyAttachmentConfig_policyIAMMultiRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyPolicyAttachmentExists(ctx, keyResourceName, &key),
				),
				PlanOnly: true,
			},
			{
				Config: testAccKeyPolicyAttachmentConfig_policyIAMMultiRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyPolicyAttachmentExists(ctx, keyResourceName, &key),
				),
				PlanOnly: true,
			},
		},
	})
}

// // Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7646
func TestAccKMSKeyPolicyAttachment_iamServiceLinkedRole(t *testing.T) {
	ctx := acctest.Context(t)
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyAttachmentConfig_policyIAMServiceLinkedRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyPolicyAttachmentExists(ctx, keyResourceName, &key),
				),
			},
			{
				ResourceName:            keyResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_window_in_days", "bypass_policy_lockout_safety_check"},
			},
		},
	})
}

func TestAccKMSKeyPolicyAttachment_booleanCondition(t *testing.T) {
	ctx := acctest.Context(t)
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyAttachmentConfig_policyBooleanCondition(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyPolicyAttachmentExists(ctx, keyResourceName, &key),
				),
			},
		},
	})
}

func testAccCheckKeyPolicyAttachmentHasPolicy(ctx context.Context, name string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No KMS Key ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn()

		out, err := conn.GetKeyPolicyWithContext(ctx, &kms.GetKeyPolicyInput{
			KeyId:      aws.String(rs.Primary.ID),
			PolicyName: aws.String("default"),
		})
		if err != nil {
			return err
		}

		actualPolicyText := aws.StringValue(out.Policy)

		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicyText)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n",
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccCheckKeyPolicyAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kms_key" {
				continue
			}

			_, err := tfkms.FindKeyByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("KMS Key %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckKeyPolicyAttachmentExists(ctx context.Context, name string, key *kms.KeyMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No KMS Key ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn()

		outputRaw, err := tfresource.RetryWhenNotFound(ctx, tfkms.PropagationTimeout, func() (interface{}, error) {
			return tfkms.FindKeyByID(ctx, conn, rs.Primary.ID)
		})

		if err != nil {
			return err
		}

		*key = *(outputRaw.(*kms.KeyMetadata))

		return nil
	}
}

func testAccKeyPolicyAttachmentConfig_policy(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}
resource "aws_kms_key_policy_attachment" "test" {
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

func testAccKeyPolicyAttachmentConfig_policyBypass(rName string, bypassFlag bool) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_key_policy_attachment" "test" {
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

func testAccKeyPolicyAttachmentConfig_policyIAMRole(rName string) string {
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

resource "aws_kms_key_policy_attachment" "test" {
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

func testAccKeyPolicyAttachmentConfig_policyIAMMultiRole(rName string) string {
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

resource "aws_kms_key_policy_attachment" "test" {
  key_id = aws_kms_key.test.id
  policy = data.aws_iam_policy_document.test.json
}
`, rName)
}

func testAccKeyPolicyAttachmentConfig_policyIAMServiceLinkedRole(rName string) string {
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

resource "aws_kms_key_policy_attachment" "test" {
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

func testAccKeyPolicyAttachmentConfig_policyBooleanCondition(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}
resource "aws_kms_key_policy_attachment" "test" {
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

func testAccKeyPolicyAttachmentConfig_removedPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}
`, rName)
}

func testAccKeyPolicyAttachmentConfig_keyIsEnabled(rName string, isEnabled bool) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  is_enabled              = %[2]t
}
resource "aws_kms_key_policy_attachment" "test" {
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
