// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3EmptyBucketAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.S3ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmptyBucketActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketIsEmpty(ctx, rName),
				),
			},
		},
	})
}

func TestAccS3EmptyBucketAction_withPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.S3ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmptyBucketActionConfig_withPrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketHasObjectsWithPrefix(ctx, rName, "keep/", 1),
					testAccCheckBucketHasObjectsWithPrefix(ctx, rName, "delete/", 0),
				),
			},
		},
	})
}

func testAccCheckBucketIsEmpty(ctx context.Context, bucketName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		input := &s3.ListObjectsV2Input{
			Bucket: &bucketName,
		}

		output, err := conn.ListObjectsV2(ctx, input)
		if err != nil {
			return fmt.Errorf("error listing objects in bucket %s: %w", bucketName, err)
		}

		if len(output.Contents) > 0 {
			return fmt.Errorf("expected bucket %s to be empty, but found %d objects", bucketName, len(output.Contents))
		}

		return nil
	}
}

func testAccCheckBucketHasObjectsWithPrefix(ctx context.Context, bucketName, prefix string, expectedCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		input := &s3.ListObjectsV2Input{
			Bucket: &bucketName,
			Prefix: &prefix,
		}

		output, err := conn.ListObjectsV2(ctx, input)
		if err != nil {
			return fmt.Errorf("error listing objects in bucket %s with prefix %s: %w", bucketName, prefix, err)
		}

		actualCount := len(output.Contents)
		if actualCount != expectedCount {
			return fmt.Errorf("expected %d objects with prefix %s in bucket %s, but found %d", expectedCount, prefix, bucketName, actualCount)
		}

		return nil
	}
}

func testAccEmptyBucketActionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "test1" {
  bucket = aws_s3_bucket.test.bucket
  key    = "test1.txt"
  source = "/dev/null"
}

resource "aws_s3_object" "test2" {
  bucket = aws_s3_bucket.test.bucket
  key    = "test2.txt"
  source = "/dev/null"
}

action "aws_s3_empty_bucket" "test" {
  config {
    bucket_name = aws_s3_bucket.test.bucket
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_s3_empty_bucket.test]
    }
  }
  depends_on = [aws_s3_object.test1, aws_s3_object.test2]
}
`, rName)
}

func testAccEmptyBucketActionConfig_withPrefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "keep" {
  bucket = aws_s3_bucket.test.bucket
  key    = "keep/file.txt"
  source = "/dev/null"
}

resource "aws_s3_object" "delete1" {
  bucket = aws_s3_bucket.test.bucket
  key    = "delete/file1.txt"
  source = "/dev/null"
}

resource "aws_s3_object" "delete2" {
  bucket = aws_s3_bucket.test.bucket
  key    = "delete/file2.txt"
  source = "/dev/null"
}

action "aws_s3_empty_bucket" "test" {
  config {
    bucket_name = aws_s3_bucket.test.bucket
    prefix      = "delete/"
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_s3_empty_bucket.test]
    }
  }
  depends_on = [aws_s3_object.keep, aws_s3_object.delete1, aws_s3_object.delete2]
}
`, rName)
}
