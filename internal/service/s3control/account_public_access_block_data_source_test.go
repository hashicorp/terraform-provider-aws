// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccountPublicAccessBlockDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3_account_public_access_block.test"
	dataSourceName := "data.aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "block_public_acls", dataSourceName, "block_public_acls"),
					resource.TestCheckResourceAttrPair(resourceName, "block_public_policy", dataSourceName, "block_public_policy"),
					resource.TestCheckResourceAttrPair(resourceName, "ignore_public_acls", dataSourceName, "ignore_public_acls"),
					resource.TestCheckResourceAttrPair(resourceName, "restrict_public_buckets", dataSourceName, "restrict_public_buckets"),
				),
			},
		},
	})
}

func testAccAccountPublicAccessBlockDataSourceConfig_base() string {
	return `
resource "aws_s3_account_public_access_block" "test" {
  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}
`
}

func testAccAccountPublicAccessBlockDataSourceConfig_basic() string {
	return acctest.ConfigCompose(testAccAccountPublicAccessBlockDataSourceConfig_base(), `
data "aws_s3_account_public_access_block" "test" {
  depends_on = [aws_s3_account_public_access_block.test]
}
`)
}
