// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIAMUserDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_user.test"
	dataSourceName := "data.aws_iam_user.test"

	userName := fmt.Sprintf("test-datasource-user-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_basic(userName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "user_id", resourceName, "unique_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "path", resourceName, "path"),
					resource.TestCheckResourceAttr(dataSourceName, "permissions_boundary", ""),
					resource.TestCheckResourceAttrPair(dataSourceName, "user_name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags", resourceName, "tags"),
				),
			},
		},
	})
}

func TestAccIAMUserDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_user.test"
	dataSourceName := "data.aws_iam_user.test"

	userName := fmt.Sprintf("test-datasource-user-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_tags(userName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "tags", resourceName, "tags"),
				),
			},
		},
	})
}

func testAccUserDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = "%s"
  path = "/"
}

data "aws_iam_user" "test" {
  user_name = aws_iam_user.test.name
}
`, name)
}

func testAccUserDataSourceConfig_tags(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = "%s"
  path = "/"

  tags = {
    tag1 = "test-value1"
    tag2 = "test-value2"
  }
}

data "aws_iam_user" "test" {
  user_name = aws_iam_user.test.name
}
`, name)
}
