// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMRoleDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_iam_role.test"
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleDataSourceConfig_basic(roleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					acctest.CheckResourceAttrEquivalentJSON(dataSourceName, "assume_role_policy", testAccRoleDataSourceConfig_AssumeRolePolicy_ExpectedJSON),
					resource.TestCheckResourceAttrPair(dataSourceName, "create_date", resourceName, "create_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "max_session_duration", resourceName, "max_session_duration"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPath, resourceName, names.AttrPath),
					resource.TestCheckResourceAttrPair(dataSourceName, "unique_id", resourceName, "unique_id"),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

const testAccRoleDataSourceConfig_AssumeRolePolicy_ExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      }
    }
  ]
}`

func testAccRoleDataSourceConfigBase() string {
	return `
data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}
`
}

func testAccRoleDataSourceConfig_basic(roleName string) string {
	return acctest.ConfigCompose(
		testAccRoleDataSourceConfigBase(),
		fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/testpath/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_role" "test" {
  name = aws_iam_role.test.name
}
`, roleName))
}
