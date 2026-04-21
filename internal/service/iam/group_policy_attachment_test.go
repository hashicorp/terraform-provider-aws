// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMGroupPolicyAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	groupName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	policyName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	policyName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	policyName3 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policy_attachment.test1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPolicyAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyAttachmentConfig_attach(groupName, policyName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyAttachmentExists(ctx, t, resourceName),
					testAccCheckGroupPolicyAttachmentCount(ctx, t, groupName, 1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccGroupPolicyAttachmentImportStateIdFunc(resourceName),
				// We do not have a way to align IDs since the Create function uses id.PrefixedUniqueId()
				// Failed state verification, resource with ID GROUP-POLICYARN not found
				// ImportStateVerify: true,
				ImportStateCheck: func(s []*terraform.InstanceState) error {
					if len(s) != 1 {
						return fmt.Errorf("expected 1 state: %#v", s)
					}
					rs := s[0]
					if !arn.IsARN(rs.Attributes["policy_arn"]) {
						return fmt.Errorf("expected policy_arn attribute to be set and begin with arn:, received: %s", rs.Attributes["policy_arn"])
					}
					return nil
				},
			},
			{
				Config: testAccGroupPolicyAttachmentConfig_attachUpdate(groupName, policyName1, policyName2, policyName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyAttachmentExists(ctx, t, resourceName),
					testAccCheckGroupPolicyAttachmentCount(ctx, t, groupName, 2),
				),
			},
		},
	})
}

func TestAccIAMGroupPolicyAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	groupName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	policyName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policy_attachment.test1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPolicyAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyAttachmentConfig_attach(groupName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyAttachmentExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourceGroupPolicyAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGroupPolicyAttachmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_group_policy_attachment" {
				continue
			}

			_, err := tfiam.FindAttachedGroupPolicyByTwoPartKey(ctx, conn, rs.Primary.Attributes["group"], rs.Primary.Attributes["policy_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM Group Policy Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckGroupPolicyAttachmentExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		_, err := tfiam.FindAttachedGroupPolicyByTwoPartKey(ctx, conn, rs.Primary.Attributes["group"], rs.Primary.Attributes["policy_arn"])

		return err
	}
}

func testAccCheckGroupPolicyAttachmentCount(ctx context.Context, t *testing.T, groupName string, want int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		input := &iam.ListAttachedGroupPoliciesInput{
			GroupName: aws.String(groupName),
		}
		output, err := tfiam.FindAttachedGroupPolicies(ctx, conn, input, tfslices.PredicateTrue[*awstypes.AttachedPolicy]())

		if err != nil {
			return err
		}

		if got := len(output); got != want {
			return fmt.Errorf("GroupPolicyAttachmentCount(%q) = %v, want %v", groupName, got, want)
		}

		return nil
	}
}

func testAccGroupPolicyAttachmentImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["group"], rs.Primary.Attributes["policy_arn"]), nil
	}
}

func testAccGroupPolicyAttachmentConfig_attach(groupName, policyName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "test" {
  name = %[1]q
}

resource "aws_iam_policy" "test1" {
  name        = %[2]q
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "iam:ChangePassword"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_group_policy_attachment" "test1" {
  group      = aws_iam_group.test.name
  policy_arn = aws_iam_policy.test1.arn
}
`, groupName, policyName)
}

func testAccGroupPolicyAttachmentConfig_attachUpdate(groupName, policyName1, policyName2, policyName3 string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "test" {
  name = %[1]q
}

resource "aws_iam_policy" "test1" {
  name        = %[2]q
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "iam:ChangePassword"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test2" {
  name        = %[3]q
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "iam:ChangePassword"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test3" {
  name        = %[4]q
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "iam:ChangePassword"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_group_policy_attachment" "test1" {
  group      = aws_iam_group.test.name
  policy_arn = aws_iam_policy.test2.arn
}

resource "aws_iam_group_policy_attachment" "test2" {
  group      = aws_iam_group.test.name
  policy_arn = aws_iam_policy.test3.arn
}
`, groupName, policyName1, policyName2, policyName3)
}
