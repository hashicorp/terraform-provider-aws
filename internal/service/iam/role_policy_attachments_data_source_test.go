// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMRolePolicyAttachmentsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_iam_role_policy_attachments.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRolePolicyAttachmentsDataSourceConfig_basic(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("role_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("attached_policies"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"policy_arn":  tfknownvalue.GlobalARNExact("iam", "policy/"+rName),
							"policy_name": knownvalue.StringExact(rName),
						}),
					})),
				},
			},
		},
	})
}

func TestAccIAMRolePolicyAttachmentsDataSource_empty(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_iam_role_policy_attachments.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRolePolicyAttachmentsDataSourceConfig_empty(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("role_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("attached_policies"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccIAMRolePolicyAttachmentsDataSource_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_iam_role_policy_attachments.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRolePolicyAttachmentsDataSourceConfig_multiple(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("role_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("attached_policies"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"policy_arn":  tfknownvalue.GlobalARNExact("iam", "policy/"+rName+"-1"),
							"policy_name": knownvalue.StringExact(rName + "-1"),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"policy_arn":  tfknownvalue.GlobalARNExact("iam", "policy/"+rName+"-2"),
							"policy_name": knownvalue.StringExact(rName + "-2"),
						}),
					})),
				},
			},
		},
	})
}

func testAccRolePolicyAttachmentsDataSourceConfig_Base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

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
`, rName)
}

func testAccRolePolicyAttachmentsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccRolePolicyAttachmentsDataSourceConfig_Base(rName),
		fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name = %[1]q

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

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

data "aws_iam_role_policy_attachments" "test" {
  role_name = aws_iam_role.test.name

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccRolePolicyAttachmentsDataSourceConfig_empty(rName string) string {
	return acctest.ConfigCompose(testAccRolePolicyAttachmentsDataSourceConfig_Base(rName),
		`
data "aws_iam_role_policy_attachments" "test" {
  role_name = aws_iam_role.test.name
}
`)
}

func testAccRolePolicyAttachmentsDataSourceConfig_multiple(rName string) string {
	return acctest.ConfigCompose(testAccRolePolicyAttachmentsDataSourceConfig_Base(rName),
		fmt.Sprintf(`
resource "aws_iam_policy" "test1" {
  name = "%[1]s-1"

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

resource "aws_iam_policy" "test2" {
  name = "%[1]s-2"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:Get*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test1" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test1.arn
}

resource "aws_iam_role_policy_attachment" "test2" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test2.arn
}

data "aws_iam_role_policy_attachments" "test" {
  role_name = aws_iam_role.test.name

  depends_on = [
    aws_iam_role_policy_attachment.test1,
    aws_iam_role_policy_attachment.test2
  ]
}
`, rName))
}
