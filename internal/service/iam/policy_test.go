// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var out awstypes.Policy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_policy.test"
	expectedPolicyText := `{"Statement":[{"Action":["ec2:Describe*"],"Effect":"Allow","Resource":"*"}],"Version":"2012-10-17"}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &out),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "iam", fmt.Sprintf("policy/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "attachment_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPath, "/"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, expectedPolicyText),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "policy_id"),
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

func TestAccIAMPolicy_description(t *testing.T) {
	ctx := acctest.Context(t)
	var out awstypes.Policy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
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

func TestAccIAMPolicy_whitespace(t *testing.T) {
	ctx := acctest.Context(t)
	var out awstypes.Policy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_whitespace(rName, "", "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &out),
				),
			},
			{
				Config:   testAccPolicyConfig_whitespace(rName, " ", "", ""),
				PlanOnly: true,
			},
			{
				Config:   testAccPolicyConfig_whitespace(rName, " ", "\n", ""),
				PlanOnly: true,
			},
			{
				Config:   testAccPolicyConfig_whitespace(rName, " ", "\n", " "),
				PlanOnly: true,
			},
			{
				Config:   testAccPolicyConfig_whitespace(rName, " \n", "\n", "\t "),
				PlanOnly: true,
			},
		},
	})
}

func TestAccIAMPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var out awstypes.Policy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &out),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMPolicy_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var out awstypes.Policy
	resourceName := "aws_iam_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_namePrefix(acctest.ResourcePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &out),
					resource.TestMatchResourceAttr(resourceName, names.AttrName, regexache.MustCompile(fmt.Sprintf("^%s", acctest.ResourcePrefix))),
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

func TestAccIAMPolicy_path(t *testing.T) {
	ctx := acctest.Context(t)
	var out awstypes.Policy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_path(rName, "/path1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, names.AttrPath, "/path1/"),
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

func TestAccIAMPolicy_policy(t *testing.T) {
	ctx := acctest.Context(t)
	var out awstypes.Policy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_policy.test"
	policy1 := `{"Statement":[{"Action":["ec2:Describe*"],"Effect":"Allow","Resource":"*"}],"Version":"2012-10-17"}`
	policy2 := `{"Statement":[{"Action":["ec2:*"],"Effect":"Allow","Resource":"*"}],"Version":"2012-10-17"}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccPolicyConfig_basic(rName, "not-json"),
				ExpectError: regexache.MustCompile("invalid JSON"),
			},
			{
				Config: testAccPolicyConfig_basic(rName, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, policy1),
				),
			},
			{
				Config: testAccPolicyConfig_basic(rName, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, policy2),
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

// https://github.com/hashicorp/terraform-provider-aws/issues/28833
func TestAccIAMPolicy_diffs(t *testing.T) {
	ctx := acctest.Context(t)
	var out awstypes.Policy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_diffs(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPolicyConfig_diffs(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config:   testAccPolicyConfig_diffs(rName, ""),
				PlanOnly: true,
			},
			{
				Config:   testAccPolicyConfig_diffs(rName, ""),
				PlanOnly: true,
			},
			{
				Config:   testAccPolicyConfig_diffs(rName, ""),
				PlanOnly: true,
			},
			{
				Config:   testAccPolicyConfig_diffs(rName, ""),
				PlanOnly: true,
			},
			{
				Config: testAccPolicyConfig_diffs(rName, "tags = {}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccPolicyConfig_diffs(rName, "tags = {}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config:   testAccPolicyConfig_diffs(rName, "tags = {}"),
				PlanOnly: true,
			},
			{
				Config:   testAccPolicyConfig_diffs(rName, "tags = {}"),
				PlanOnly: true,
			},
			{
				Config:   testAccPolicyConfig_diffs(rName, "tags = {}"),
				PlanOnly: true,
			},
			{
				Config:   testAccPolicyConfig_diffs(rName, "tags = {}"),
				PlanOnly: true,
			},
			{
				Config: testAccPolicyConfig_diffs(rName, "tags = {}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccIAMPolicy_policyDuplicateKeys(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccPolicyConfig_policyDuplicateKeys(rName),
				ExpectError: regexache.MustCompile(`"policy" contains duplicate JSON keys: duplicate key "Statement.0.Condition.StringEquals"`),
			},
		},
	})
}

func testAccCheckPolicyExists(ctx context.Context, n string, v *awstypes.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IAM Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		output, err := tfiam.FindPolicyByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_policy" {
				continue
			}

			_, err := tfiam.FindPolicyByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPolicyConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  description = %q
  name        = %q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}
`, description, rName)
}

func testAccPolicyConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name = %q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}
`, rName)
}

func testAccPolicyConfig_whitespace(rName, ws1, ws2, ws3 string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name = %[1]q

  policy = <<EOF
%[2]s{
  "Version": "2012-10-17",%[3]s
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": %[3]s"Allow",
      "Resource": "*"
    }
  ]
}
EOF
}
`, rName, ws1, ws2, ws3)
}

func testAccPolicyConfig_namePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name_prefix = %q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}
`, namePrefix)
}

func testAccPolicyConfig_path(rName, path string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name = %q
  path = %q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}
`, rName, path)
}

func testAccPolicyConfig_basic(rName, policy string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name   = %q
  policy = %q
}
`, rName, policy)
}

func testAccPolicyConfig_diffs(rName string, tags string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name = %[1]q

  policy = jsonencode({
    Id      = %[1]q
    Version = "2012-10-17"
    Statement = [{
      Sid    = "60c9d11f"
      Effect = "Allow"
      Action = [
        "kms:Describe*",
        "kms:Get*",
        "kms:List*",
        "kms:CreateKey",
        "kms:DescribeKey",
        "kms:ScheduleKeyDeletion",
        "kms:TagResource",
        "kms:UntagResource",
        "s3:ListMultipartUploadParts",
        "s3:ListBucketVersions",
        "s3:ListBucketMultipartUploads",
        "s3:ListBucket",
        "s3:GetReplicationConfiguration",
        "s3:GetObjectVersionTorrent",
        "s3:GetObjectVersionTagging",
        "s3:GetObjectVersionForReplication",
        "s3:GetObjectVersionAcl",
        "s3:GetObjectVersion",
        "s3:GetObjectTorrent",
        "s3:GetObjectTagging",
        "s3:GetObjectRetention",
        "s3:GetObjectLegalHold",
        "s3:GetObjectAcl",
        "s3:GetObject",
        "s3:GetMetricsConfiguration",
        "s3:GetLifecycleConfiguration",
        "s3:GetJobTagging",
        "s3:GetInventoryConfiguration",
        "s3:GetEncryptionConfiguration",
        "s3:GetBucketWebsite",
        "s3:GetBucketVersioning",
        "s3:GetBucketTagging",
        "s3:GetBucketRequestPayment",
        "s3:GetBucketPublicAccessBlock",
        "s3:GetBucketPolicyStatus",
        "s3:GetBucketPolicy",
        "s3:GetBucketOwnershipControls",
        "s3:GetBucketObjectLockConfiguration",
        "s3:GetBucketNotification",
        "s3:GetBucketLogging",
        "s3:GetBucketLocation",
        "s3:GetBucketCORS",
        "s3:GetBucketAcl",
        "s3:GetAnalyticsConfiguration",
        "s3:GetAccessPointPolicyStatus",
        "s3:GetAccessPointPolicy",
        "s3:GetAccelerateConfiguration",
        "s3:DescribeJob",
      ]
      Resource = "*"
    }]
  })

  %[2]s
}
`, rName, tags)
}

func testAccPolicyConfig_policyDuplicateKeys(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name = %q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "s3:PutObject",
      "Resource": "*",
      "Condition": {
        "StringEquals": {
          "s3:prefix": ["one/", "two/"]
        },
        "StringEquals": {
          "s3:versionid": "abc123"
        }
      }
    }
  ]
}
EOF
}
`, rName)
}
