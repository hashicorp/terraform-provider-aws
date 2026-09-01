// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketObjectLockConfigurationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_bucket_object_lock_configuration.test"
	resourceName := "aws_s3_bucket_object_lock_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectLockConfigurationDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrBucket, resourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(dataSourceName, "object_lock_enabled", string(types.ObjectLockEnabledEnabled)),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtRulePound, "1"),
					resource.TestCheckResourceAttr(dataSourceName, "rule.0.default_retention.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "rule.0.default_retention.0.days", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "rule.0.default_retention.0.mode", string(types.ObjectLockRetentionModeCompliance)),
				),
			},
		},
	})
}

func TestAccS3BucketObjectLockConfigurationDataSource_noRule(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_bucket_object_lock_configuration.test"
	resourceName := "aws_s3_bucket_object_lock_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectLockConfigurationDataSourceConfig_noRule(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrBucket, resourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(dataSourceName, "object_lock_enabled", string(types.ObjectLockEnabledEnabled)),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtRulePound, "0"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectLockConfigurationDataSource_notConfigured(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketObjectLockConfigurationDataSourceConfig_notConfigured(rName),
				ExpectError: regexache.MustCompile(`couldn't find resource`),
			},
		},
	})
}

func testAccBucketObjectLockConfigurationDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_enabled = true
}

resource "aws_s3_bucket_object_lock_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    default_retention {
      mode = %[2]q
      days = 3
    }
  }
}

data "aws_s3_bucket_object_lock_configuration" "test" {
  bucket = aws_s3_bucket_object_lock_configuration.test.bucket
}
`, rName, types.ObjectLockRetentionModeCompliance)
}

func testAccBucketObjectLockConfigurationDataSourceConfig_noRule(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_enabled = true
}

resource "aws_s3_bucket_object_lock_configuration" "test" {
  bucket = aws_s3_bucket.test.id
}

data "aws_s3_bucket_object_lock_configuration" "test" {
  bucket = aws_s3_bucket_object_lock_configuration.test.bucket
}
`, rName)
}

func testAccBucketObjectLockConfigurationDataSourceConfig_notConfigured(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

data "aws_s3_bucket_object_lock_configuration" "test" {
  bucket = aws_s3_bucket.test.id
}
`, rName)
}
