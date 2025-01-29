// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMPolicyDataSource_arn(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_iam_policy.test"
	resourceName := "aws_iam_policy.test"
	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDataSourceConfig_arn(policyName, "/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrPath, resourceName, names.AttrPath),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrPolicy, resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttrPair(datasourceName, "policy_id", resourceName, "policy_id"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "attachment_count", resourceName, "attachment_count"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccIAMPolicyDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_iam_policy.test"
	resourceName := "aws_iam_policy.test"
	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDataSourceConfig_name(policyName, "/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrPath, resourceName, names.AttrPath),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrPolicy, resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttrPair(datasourceName, "policy_id", resourceName, "policy_id"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "attachment_count", resourceName, "attachment_count"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccIAMPolicyDataSource_nameAndPathPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_iam_policy.test"
	resourceName := "aws_iam_policy.test"

	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyPath := "/test-path/"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDataSourceConfig_pathPrefix(policyName, policyPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrPath, resourceName, names.AttrPath),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrPolicy, resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttrPair(datasourceName, "policy_id", resourceName, "policy_id"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "attachment_count", resourceName, "attachment_count"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccIAMPolicyDataSource_nonExistent(t *testing.T) {
	ctx := acctest.Context(t)
	policyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyPath := "/test-path/"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPolicyDataSourceConfig_nonExistent(policyName, policyPath),
				ExpectError: regexache.MustCompile(`no matching IAM Policy found`),
			},
		},
	})
}

func testAccPolicyBaseDataSourceConfig(policyName, policyPath string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name        = %q
  path        = %q
  description = "My test policy"

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
}`, policyName, policyPath)
}

func testAccPolicyDataSourceConfig_arn(policyName, policyPath string) string {
	return acctest.ConfigCompose(
		testAccPolicyBaseDataSourceConfig(policyName, policyPath), `
data "aws_iam_policy" "test" {
  arn = aws_iam_policy.test.arn
}
`)
}

func testAccPolicyDataSourceConfig_name(policyName, policyPath string) string {
	return acctest.ConfigCompose(
		testAccPolicyBaseDataSourceConfig(policyName, policyPath), `
data "aws_iam_policy" "test" {
  name = aws_iam_policy.test.name
}
`)
}

func testAccPolicyDataSourceConfig_pathPrefix(policyName, policyPath string) string {
	return acctest.ConfigCompose(
		testAccPolicyBaseDataSourceConfig(policyName, policyPath),
		fmt.Sprintf(`
data "aws_iam_policy" "test" {
  name        = aws_iam_policy.test.name
  path_prefix = %q
}
`, policyPath))
}

func testAccPolicyDataSourceConfig_nonExistent(policyName, policyPath string) string {
	return acctest.ConfigCompose(
		testAccPolicyBaseDataSourceConfig(policyName, policyPath),
		fmt.Sprintf(`
data "aws_iam_policy" "test" {
  name        = "non-existent"
  path_prefix = %q
}
`, policyPath))
}
