// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"
	dataSourceName := "data.aws_s3_buckets.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketsDataSourceConfig_basic(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "buckets.#", 1),
					resource.TestCheckResourceAttrPair(dataSourceName, "buckets.0.bucket_arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "buckets.0.bucket_region", resourceName, "bucket_region"),
					resource.TestCheckResourceAttrPair(dataSourceName, "buckets.0.name", resourceName, names.AttrBucket),
				),
			},
		},
	})
}

func TestAccS3BucketsDataSource_maxBuckets(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"
	dataSourceName := "data.aws_s3_buckets.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketsDataSourceConfig_maxBuckets(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "buckets.#", 1),
					resource.TestCheckResourceAttrPair(dataSourceName, "buckets.0.bucket_arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "buckets.0.bucket_region", resourceName, "bucket_region"),
					resource.TestCheckResourceAttrPair(dataSourceName, "buckets.0.name", resourceName, names.AttrBucket),
				),
			},
		},
	})
}

func testAccBucketsDataSourceConfig_basic(rName, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket" "test2" {
  bucket = %[2]q
}

data "aws_s3_buckets" "test" {
  prefix = %[1]q

  depends_on = [aws_s3_bucket.test, aws_s3_bucket.test2]
}
`, rName, rName2)
}

func testAccBucketsDataSourceConfig_maxBuckets(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket" "test2" {
  bucket = "%[1]s-2"
}

data "aws_s3_buckets" "test" {
  prefix      = %[1]q
  max_buckets = 1

  depends_on = [aws_s3_bucket.test, aws_s3_bucket.test2]
}
`, rName)
}
