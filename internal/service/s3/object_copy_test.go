// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ObjectCopy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object_copy.test"
	sourceName := "aws_s3_object.source"
	sourceKey := "source"
	targetKey := "target"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectCopyConfig_basic(rName1, sourceKey, rName2, targetKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectCopyExists(ctx, resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "acl"),
					resource.TestCheckResourceAttr(resourceName, "bucket", rName2),
					resource.TestCheckResourceAttr(resourceName, "bucket_key_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "cache_control", ""),
					resource.TestCheckResourceAttr(resourceName, "content_disposition", ""),
					resource.TestCheckResourceAttr(resourceName, "content_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "content_language", ""),
					resource.TestCheckResourceAttr(resourceName, "content_type", "binary/octet-stream"),
					resource.TestCheckNoResourceAttr(resourceName, "copy_if_match"),
					resource.TestCheckNoResourceAttr(resourceName, "copy_if_modified_since"),
					resource.TestCheckNoResourceAttr(resourceName, "copy_if_none_match"),
					resource.TestCheckNoResourceAttr(resourceName, "copy_if_unmodified_since"),
					resource.TestCheckResourceAttr(resourceName, "customer_algorithm", ""),
					resource.TestCheckNoResourceAttr(resourceName, "customer_key"),
					resource.TestCheckResourceAttr(resourceName, "customer_key_md5", ""),
					resource.TestCheckResourceAttrPair(resourceName, "etag", sourceName, "etag"),
					resource.TestCheckNoResourceAttr(resourceName, "expected_bucket_owner"),
					resource.TestCheckNoResourceAttr(resourceName, "expected_source_bucket_owner"),
					resource.TestCheckResourceAttr(resourceName, "expiration", ""),
					resource.TestCheckNoResourceAttr(resourceName, "expires"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "false"),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "key", targetKey),
					resource.TestCheckResourceAttr(resourceName, "kms_encryption_context", ""),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, "metadata.%", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "metadata_directive"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
					resource.TestCheckResourceAttr(resourceName, "request_charged", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "request_payer"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption", "AES256"),
					resource.TestCheckResourceAttr(resourceName, "source", fmt.Sprintf("%s/%s", rName1, sourceKey)),
					resource.TestCheckNoResourceAttr(resourceName, "source_customer_algorithm"),
					resource.TestCheckNoResourceAttr(resourceName, "source_customer_key"),
					resource.TestCheckNoResourceAttr(resourceName, "source_customer_key_md5"),
					resource.TestCheckResourceAttr(resourceName, "source_version_id", ""),
					resource.TestCheckResourceAttr(resourceName, "storage_class", "STANDARD"),
					resource.TestCheckNoResourceAttr(resourceName, "tagging_directive"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", ""),
					resource.TestCheckResourceAttr(resourceName, "website_redirect", ""),
				),
			},
		},
	})
}

func TestAccS3ObjectCopy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object_copy.test"
	sourceKey := "source"
	targetKey := "target"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectCopyConfig_basic(rName1, sourceKey, rName2, targetKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectCopyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceObjectCopy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ObjectCopy_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object_copy.test"
	sourceKey := "source"
	targetKey := "target"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectCopyConfig_tags1(rName1, sourceKey, rName2, targetKey, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectCopyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tagging_directive", "REPLACE"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccObjectCopyConfig_tags2(rName1, sourceKey, rName2, targetKey, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectCopyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tagging_directive", "REPLACE"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccObjectCopyConfig_tags1(rName1, sourceKey, rName2, targetKey, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectCopyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tagging_directive", "REPLACE"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccS3ObjectCopy_metadata(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object_copy.test"
	sourceKey := "source"
	targetKey := "target"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectCopyConfig_metadata(rName1, sourceKey, rName2, targetKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectCopyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "metadata_directive", "REPLACE"),
					resource.TestCheckResourceAttr(resourceName, "metadata.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "metadata.mk1", "mv1"),
				),
			},
		},
	})
}

func TestAccS3ObjectCopy_grant(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object_copy.test"
	sourceKey := "source"
	targetKey := "target"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectCopyConfig_grant(rName1, sourceKey, rName2, targetKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectCopyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "grant.*", map[string]string{
						"permissions.#": "1",
						"type":          "Group",
						"uri":           "http://acs.amazonaws.com/groups/global/AllUsers",
					}),
				),
			},
		},
	})
}

func TestAccS3ObjectCopy_BucketKeyEnabled_bucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectCopyConfig_bucketKeyEnabledBucket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectCopyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket_key_enabled", "true"),
				),
			},
		},
	})
}

func TestAccS3ObjectCopy_BucketKeyEnabled_object(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectCopyConfig_bucketKeyEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectCopyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket_key_enabled", "true"),
				),
			},
		},
	})
}

func testAccCheckObjectCopyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_object_copy" {
				continue
			}

			_, err := tfs3.FindObjectByThreePartKey(ctx, conn, rs.Primary.Attributes["bucket"], rs.Primary.Attributes["key"], rs.Primary.Attributes["etag"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Object %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckObjectCopyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		_, err := tfs3.FindObjectByThreePartKey(ctx, conn, rs.Primary.Attributes["bucket"], rs.Primary.Attributes["key"], rs.Primary.Attributes["etag"])

		return err
	}
}

func testAccObjectCopyConfig_base(sourceBucket, sourceKey, targetBucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "source" {
  bucket = %[1]q
}

resource "aws_s3_object" "source" {
  bucket  = aws_s3_bucket.source.bucket
  key     = %[2]q
  content = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
}

resource "aws_s3_bucket" "target" {
  bucket = %[3]q
}
`, sourceBucket, sourceKey, targetBucket)
}

func testAccObjectCopyConfig_basic(sourceBucket, sourceKey, targetBucket, targetKey string) string {
	return acctest.ConfigCompose(testAccObjectCopyConfig_base(sourceBucket, sourceKey, targetBucket), fmt.Sprintf(`
resource "aws_s3_object_copy" "test" {
  bucket = aws_s3_bucket.target.bucket
  key    = %[1]q
  source = "${aws_s3_bucket.source.bucket}/${aws_s3_object.source.key}"
}
`, targetKey))
}

func testAccObjectCopyConfig_tags1(sourceBucket, sourceKey, targetBucket, targetKey, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccObjectCopyConfig_base(sourceBucket, sourceKey, targetBucket), fmt.Sprintf(`
resource "aws_s3_object_copy" "test" {
  bucket = aws_s3_bucket.target.bucket
  key    = %[1]q
  source = "${aws_s3_bucket.source.bucket}/${aws_s3_object.source.key}"

  tagging_directive = "REPLACE"

  tags = {
    %[2]q = %[3]q
  }
}
`, targetKey, tagKey1, tagValue1))
}

func testAccObjectCopyConfig_tags2(sourceBucket, sourceKey, targetBucket, targetKey, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccObjectCopyConfig_base(sourceBucket, sourceKey, targetBucket), fmt.Sprintf(`
resource "aws_s3_object_copy" "test" {
  bucket = aws_s3_bucket.target.bucket
  key    = %[1]q
  source = "${aws_s3_bucket.source.bucket}/${aws_s3_object.source.key}"

  tagging_directive = "REPLACE"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, targetKey, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccObjectCopyConfig_metadata(sourceBucket, sourceKey, targetBucket, targetKey string) string {
	return acctest.ConfigCompose(testAccObjectCopyConfig_base(sourceBucket, sourceKey, targetBucket), fmt.Sprintf(`
resource "aws_s3_object_copy" "test" {
  bucket = aws_s3_bucket.target.bucket
  key    = %[1]q
  source = "${aws_s3_bucket.source.bucket}/${aws_s3_object.source.key}"

  metadata_directive = "REPLACE"

  metadata = {
    "mk1" = "mv1"
  }
}
`, targetKey))
}

func testAccObjectCopyConfig_grant(sourceBucket, sourceKey, targetBucket, targetKey string) string {
	return acctest.ConfigCompose(testAccObjectCopyConfig_base(sourceBucket, sourceKey, targetBucket), fmt.Sprintf(`
resource "aws_s3_bucket_public_access_block" "target" {
  bucket = aws_s3_bucket.target.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_ownership_controls" "target" {
  bucket = aws_s3_bucket.target.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_object_copy" "test" {
  depends_on = [
    aws_s3_bucket_public_access_block.target,
    aws_s3_bucket_ownership_controls.target,
  ]

  bucket = aws_s3_bucket.target.bucket
  key    = %[1]q
  source = "${aws_s3_bucket.source.bucket}/${aws_s3_object.source.key}"

  grant {
    uri         = "http://acs.amazonaws.com/groups/global/AllUsers"
    type        = "Group"
    permissions = ["READ"]
  }
}
`, targetKey))
}

func testAccObjectCopyConfig_bucketKeyEnabledBucket(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Encrypts test objects"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"
}

resource "aws_s3_object" "source" {
  bucket  = aws_s3_bucket.source.bucket
  content = "Ingen ko på isen"
  key     = "test"
}

resource "aws_s3_bucket" "target" {
  bucket = "%[1]s-target"
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.target.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.test.arn
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_object_copy" "test" {
  # Must have bucket SSE enabled first
  depends_on = [aws_s3_bucket_server_side_encryption_configuration.test]

  bucket = aws_s3_bucket.target.bucket
  key    = "test"
  source = "${aws_s3_bucket.source.bucket}/${aws_s3_object.source.key}"
}
`, rName)
}

func testAccObjectCopyConfig_bucketKeyEnabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Encrypts test objects"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"
}

resource "aws_s3_object" "source" {
  bucket  = aws_s3_bucket.source.bucket
  content = "Ingen ko på isen"
  key     = "test"
}

resource "aws_s3_bucket" "target" {
  bucket = "%[1]s-target"
}

resource "aws_s3_object_copy" "test" {
  bucket             = aws_s3_bucket.target.bucket
  bucket_key_enabled = true
  key                = "test"
  kms_key_id         = aws_kms_key.test.arn
  source             = "${aws_s3_bucket.source.bucket}/${aws_s3_object.source.key}"
}
`, rName)
}
