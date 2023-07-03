// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
)

func TestAccIAMGroupPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var groupPolicy1, groupPolicy2 iam.GetGroupPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(ctx, "aws_iam_group.group",
						"aws_iam_group_policy.foo",
						&groupPolicy1,
					),
				),
			},
			{
				ResourceName:      "aws_iam_group_policy.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroupPolicyConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(ctx, "aws_iam_group.group",
						"aws_iam_group_policy.bar",
						&groupPolicy2,
					),
					testAccCheckGroupPolicyNameChanged(&groupPolicy1, &groupPolicy2),
				),
			},
			{
				ResourceName:      "aws_iam_group_policy.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIAMGroupPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var out iam.GetGroupPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(ctx, "aws_iam_group.group",
						"aws_iam_group_policy.foo",
						&out,
					),
					testAccCheckGroupPolicyDisappears(ctx, &out),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMGroupPolicy_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var groupPolicy1, groupPolicy2 iam.GetGroupPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyConfig_namePrefix(rName, "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(ctx, "aws_iam_group.test",
						"aws_iam_group_policy.test",
						&groupPolicy1,
					),
				),
			},
			{
				Config: testAccGroupPolicyConfig_namePrefix(rName, "ec2:*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(ctx, "aws_iam_group.test",
						"aws_iam_group_policy.test",
						&groupPolicy2,
					),
					testAccCheckGroupPolicyNameMatches(&groupPolicy1, &groupPolicy2),
				),
			},
			{
				ResourceName:            "aws_iam_group_policy.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccIAMGroupPolicy_generatedName(t *testing.T) {
	ctx := acctest.Context(t)
	var groupPolicy1, groupPolicy2 iam.GetGroupPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyConfig_generatedName(rName, "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(ctx, "aws_iam_group.test",
						"aws_iam_group_policy.test",
						&groupPolicy1,
					),
				),
			},
			{
				Config: testAccGroupPolicyConfig_generatedName(rName, "ec2:*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(ctx, "aws_iam_group.test",
						"aws_iam_group_policy.test",
						&groupPolicy2,
					),
					testAccCheckGroupPolicyNameMatches(&groupPolicy1, &groupPolicy2),
				),
			},
			{
				ResourceName:      "aws_iam_group_policy.test",
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
	var gp iam.GetGroupPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policy.test"
	groupName := "aws_iam_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRolePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyConfig_unknowns(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyExists(ctx, groupName, resourceName, &gp),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
		},
	})
}

func testAccCheckGroupPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_group_policy" {
				continue
			}

			group, name, err := tfiam.GroupPolicyParseID(rs.Primary.ID)
			if err != nil {
				return err
			}

			request := &iam.GetGroupPolicyInput{
				PolicyName: aws.String(name),
				GroupName:  aws.String(group),
			}

			getResp, err := conn.GetGroupPolicyWithContext(ctx, request)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
					// none found, that's good
					continue
				}
				return fmt.Errorf("Error reading IAM policy %s from group %s: %s", name, group, err)
			}

			if getResp != nil {
				return fmt.Errorf("Found IAM group policy, expected none: %s", getResp)
			}
		}

		return nil
	}
}

func testAccCheckGroupPolicyDisappears(ctx context.Context, out *iam.GetGroupPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)

		params := &iam.DeleteGroupPolicyInput{
			PolicyName: out.PolicyName,
			GroupName:  out.GroupName,
		}

		_, err := conn.DeleteGroupPolicyWithContext(ctx, params)
		return err
	}
}

func testAccCheckGroupPolicyExists(ctx context.Context,
	iamGroupResource string,
	iamGroupPolicyResource string,
	groupPolicy *iam.GetGroupPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[iamGroupResource]
		if !ok {
			return fmt.Errorf("Not Found: %s", iamGroupResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		policy, ok := s.RootModule().Resources[iamGroupPolicyResource]
		if !ok {
			return fmt.Errorf("Not Found: %s", iamGroupPolicyResource)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)
		group, name, err := tfiam.GroupPolicyParseID(policy.Primary.ID)
		if err != nil {
			return err
		}

		output, err := conn.GetGroupPolicyWithContext(ctx, &iam.GetGroupPolicyInput{
			GroupName:  aws.String(group),
			PolicyName: aws.String(name),
		})

		if err != nil {
			return err
		}

		*groupPolicy = *output

		return nil
	}
}

func testAccCheckGroupPolicyNameChanged(i, j *iam.GetGroupPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.PolicyName) == aws.StringValue(j.PolicyName) {
			return errors.New("IAM Group Policy name did not change")
		}

		return nil
	}
}

func testAccCheckGroupPolicyNameMatches(i, j *iam.GetGroupPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.PolicyName) != aws.StringValue(j.PolicyName) {
			return errors.New("IAM Group Policy name did not match")
		}

		return nil
	}
}

func testAccGroupPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
  name = %[1]q
  path = "/"
}

resource "aws_iam_group_policy" "foo" {
  name  = %[1]q
  group = aws_iam_group.group.name

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
`, rName)
}

func testAccGroupPolicyConfig_namePrefix(rName, policyAction string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "test" {
  name = %[1]q
  path = "/"
}

resource "aws_iam_group_policy" "test" {
  name_prefix = %[1]q
  group       = aws_iam_group.test.name

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
`, rName, policyAction)
}

func testAccGroupPolicyConfig_generatedName(rName, policyAction string) string {
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
    "Action": %[2]q,
    "Resource": "*"
  },
  "Version": "2012-10-17"
}
EOF
}
`, rName, policyAction)
}

func testAccGroupPolicyConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
  name = %[1]q
  path = "/"
}

resource "aws_iam_group_policy" "foo" {
  name  = %[1]q
  group = aws_iam_group.group.name

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

resource "aws_iam_group_policy" "bar" {
  name  = "%[1]s-2"
  group = aws_iam_group.group.name

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
`, rName)
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

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
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
