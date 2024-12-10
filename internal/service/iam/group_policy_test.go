// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMGroupPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var groupPolicy string
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyConfig_basic(rName, "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(ctx, resourceName, &groupPolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
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

func TestAccIAMGroupPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var groupPolicy string
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyConfig_basic(rName, "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(ctx, resourceName, &groupPolicy),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceGroupPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMGroupPolicy_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var groupPolicy string
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyConfig_nameGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(ctx, resourceName, &groupPolicy),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, id.UniqueIdPrefix),
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

func TestAccIAMGroupPolicy_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var groupPolicy string
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyConfig_namePrefix(rName, "tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(ctx, resourceName, &groupPolicy),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
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

// When there are unknowns in the policy (interpolation), TF puts a
// random GUID (e.g., 14730d5f-efa3-5a5e-94b5-f8bad6f88282) in state
// at first for the policy which, obviously, behaves differently than
// a JSON policy. This test checks to make sure nothing goes wrong
// during that step.
func TestAccIAMGroupPolicy_unknownsInPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var groupPolicy string
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRolePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyConfig_unknowns(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(ctx, resourceName, &groupPolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
		},
	})
}

func TestAccIAMGroupPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var groupPolicy string
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyConfig_basic(rName, "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(ctx, resourceName, &groupPolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
				),
			},
			{
				Config: testAccGroupPolicyConfig_basic(rName, "ec2:*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(ctx, resourceName, &groupPolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
				),
			},
		},
	})
}

func testAccCheckGroupPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_group_policy" {
				continue
			}

			groupName, policyName, err := tfiam.GroupPolicyParseID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfiam.FindGroupPolicyByTwoPartKey(ctx, conn, groupName, policyName)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM Group Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckGroupPolicyExists(ctx context.Context, n string, v *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		groupName, policyName, err := tfiam.GroupPolicyParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		output, err := tfiam.FindGroupPolicyByTwoPartKey(ctx, conn, groupName, policyName)

		if err != nil {
			return err
		}

		*v = output

		return nil
	}
}

func testAccGroupPolicyConfig_basic(rName, action string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "test" {
  name = %[1]q
  path = "/"
}

resource "aws_iam_group_policy" "test" {
  name  = %[1]q
  group = aws_iam_group.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": %[2]q,
    "Resource": "*"
  }
}
EOF
}
`, rName, action)
}

func testAccGroupPolicyConfig_nameGenerated(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "test" {
  name = %[1]q
  path = "/"
}

resource "aws_iam_group_policy" "test" {
  group = aws_iam_group.test.name

  policy = <<EOF
{
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  },
  "Version": "2012-10-17"
}
EOF
}
`, rName)
}

func testAccGroupPolicyConfig_namePrefix(rName, namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "test" {
  name = %[1]q
  path = "/"
}

resource "aws_iam_group_policy" "test" {
  name_prefix = %[2]q
  group       = aws_iam_group.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}
`, rName, namePrefix)
}

func testAccGroupPolicyConfig_unknowns(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_group" "test" {
  name = %[1]q
  path = "/"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_iam_group_policy" "test" {
  name  = %[1]q
  group = aws_iam_group.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = ""
      Effect = "Allow"
      Action = [
        "s3:AbortMultipartUpload",
        "s3:GetBucketLocation",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads",
        "s3:PutObject",
      ]
      Resource = [
        aws_s3_bucket.test.arn,
        "${aws_s3_bucket.test.arn}/*",
      ]
    }]
  })
}
`, rName)
}
