// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3DirectoryBucketsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_directory_buckets.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "arns.#", 1),
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "buckets.#", 1),
				),
			},
		},
	})
}

func testAccDirectoryBucketsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_baseAZ(rName), `
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

data "aws_s3_directory_buckets" "test" {
  depends_on = [aws_s3_directory_bucket.test]
}
`)
}
