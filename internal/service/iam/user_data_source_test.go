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

func TestAccIAMUserDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_user.test"
	dataSourceName := "data.aws_iam_user.test"

	userName := fmt.Sprintf("test-datasource-user-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_basic(userName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "user_id", resourceName, "unique_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPath, resourceName, names.AttrPath),
					resource.TestCheckResourceAttr(dataSourceName, "permissions_boundary", ""),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrUserName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrTags, resourceName, names.AttrTags),
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
