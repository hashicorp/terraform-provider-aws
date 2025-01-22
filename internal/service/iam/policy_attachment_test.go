// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMPolicyAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	userName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	userName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	groupName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	groupName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	groupName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	attachmentName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyAttachmentConfig_attach(userName1, roleName1, roleName2, policyName, attachmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyAttachmentExists(ctx, resourceName),
					testAccCheckPolicyAttachmentCounts(ctx, resourceName, 0, 2, 1),
				),
			},
			{
				Config: testAccPolicyAttachmentConfig_attachUpdate(userName1, userName2, roleName1, groupName1, groupName2, groupName3, policyName, attachmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyAttachmentExists(ctx, resourceName),
					testAccCheckPolicyAttachmentCounts(ctx, resourceName, 3, 1, 2),
				),
			},
		},
	})
}

func TestAccIAMPolicyAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	userName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	attachmentName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyAttachmentConfig_attach(userName1, roleName1, roleName2, policyName, attachmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyAttachmentExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourcePolicyAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMPolicyAttachment_paginatedEntities(t *testing.T) {
	ctx := acctest.Context(t)
	userNamePrefix := fmt.Sprintf("%s-%s-", acctest.ResourcePrefix, sdkacctest.RandString(3))
	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	attachmentName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyAttachmentConfig_paginatedAttach(userNamePrefix, policyName, attachmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyAttachmentExists(ctx, resourceName),
					testAccCheckPolicyAttachmentCounts(ctx, resourceName, 0, 0, 101),
				),
			},
		},
	})
}

func testAccCheckPolicyAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_policy_attachment" {
				continue
			}

			_, _, _, err := tfiam.FindEntitiesForPolicyByARN(ctx, conn, rs.Primary.Attributes["policy_arn"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM Policy Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPolicyAttachmentExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		_, _, _, err := tfiam.FindEntitiesForPolicyByARN(ctx, conn, rs.Primary.Attributes["policy_arn"])

		return err
	}
}

func testAccCheckPolicyAttachmentCounts(ctx context.Context, n string, wantGroups, wantRoles, wantUsers int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		groups, roles, users, err := tfiam.FindEntitiesForPolicyByARN(ctx, conn, rs.Primary.Attributes["policy_arn"])

		if err != nil {
			return err
		}

		if got, want := len(groups), wantGroups; got != want {
			return fmt.Errorf("GroupPolicyAttachmentCount = %v, want %v", got, want)
		}
		if got, want := len(roles), wantRoles; got != want {
			return fmt.Errorf("RolePolicyAttachmentCount = %v, want %v", got, want)
		}
		if got, want := len(users), wantUsers; got != want {
			return fmt.Errorf("GroupPolicyAttachmentCount = %v, want %v", got, want)
		}

		return nil
	}
}

func testAccPolicyAttachmentConfig_attach(userName1, roleName1, roleName2, policyName, attachmentName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test1" {
  name = %[1]q
}

resource "aws_iam_role" "test1" {
  name = %[2]q

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

resource "aws_iam_role" "test2" {
  name = %[3]q

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

resource "aws_iam_policy" "test" {
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

resource "aws_iam_policy_attachment" "test" {
  name       = %[5]q
  users      = [aws_iam_user.test1.name]
  roles      = [aws_iam_role.test1.name, aws_iam_role.test2.name]
  policy_arn = aws_iam_policy.test.arn
}
`, userName1, roleName1, roleName2, policyName, attachmentName)
}

func testAccPolicyAttachmentConfig_attachUpdate(userName1, userName2, roleName1, groupName1, groupName2, groupName3, policyName, attachmentName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test1" {
  name = %[1]q
}

resource "aws_iam_user" "test2" {
  name = %[2]q
}

resource "aws_iam_role" "test1" {
  name = %[3]q

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

resource "aws_iam_group" "test1" {
  name = %[4]q
}

resource "aws_iam_group" "test2" {
  name = %[5]q
}

resource "aws_iam_group" "test3" {
  name = %[6]q
}

resource "aws_iam_policy" "test" {
  name        = %[7]q
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

resource "aws_iam_policy_attachment" "test" {
  name = %[8]q

  users = [
    aws_iam_user.test1.name,
    aws_iam_user.test2.name,
  ]

  roles = [
    aws_iam_role.test1.name,
  ]

  groups = [
    aws_iam_group.test1.name,
    aws_iam_group.test2.name,
    aws_iam_group.test3.name,
  ]

  policy_arn = aws_iam_policy.test.arn
}
`, userName1, userName2, roleName1, groupName1, groupName2, groupName3, policyName, attachmentName)
}

func testAccPolicyAttachmentConfig_paginatedAttach(userNamePrefix, policyName, attachmentName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  count = 101
  name  = format("%[1]s%%d", count.index + 1)
}

resource "aws_iam_policy" "test" {
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

resource "aws_iam_policy_attachment" "test" {
  name       = %[3]q
  policy_arn = aws_iam_policy.test.arn

  users = aws_iam_user.test[*].name
}
`, userNamePrefix, policyName, attachmentName)
}
