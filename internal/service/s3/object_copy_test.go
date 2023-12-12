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
					resource.TestCheckNoResourceAttr(resourceName, "checksum_algorithm"),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32c", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha1", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha256", ""),
					resource.TestCheckResourceAttr(resourceName, "content_disposition", ""),
					resource.TestCheckResourceAttr(resourceName, "content_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "content_language", ""),
					resource.TestCheckResourceAttr(resourceName, "content_type", "application/octet-stream"),
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
				Config: testAccObjectCopyConfig_bucketKeyEnabledBucket(rName1, sourceKey, rName2, targetKey),
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
				Config: testAccObjectCopyConfig_bucketKeyEnabledObject(rName1, sourceKey, rName2, targetKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectCopyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket_key_enabled", "true"),
				),
			},
		},
	})
}

func TestAccS3ObjectCopy_sourceWithSlashes(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object_copy.test"
	sourceKey := "dir1/dir2/source"
	targetKey := "target"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectCopyConfig_baseSourceAndTargetBuckets(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAddObjects(ctx, "aws_s3_bucket.source", sourceKey),
				),
			},
			{
				Config: testAccObjectCopyConfig_externalSourceObject(rName1, sourceKey, rName2, targetKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectCopyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "source", fmt.Sprintf("%s/%s", rName1, sourceKey)),
				),
			},
		},
	})
}

func TestAccS3ObjectCopy_checksumAlgorithm(t *testing.T) {
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
				Config: testAccObjectCopyConfig_checksumAlgorithm(rName1, sourceKey, rName2, targetKey, "CRC32C"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectCopyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "checksum_algorithm", "CRC32C"),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32c", "7y1BJA=="),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha1", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha256", ""),
				),
			},
			{
				Config: testAccObjectCopyConfig_checksumAlgorithm(rName1, sourceKey, rName2, targetKey, "SHA1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectCopyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "checksum_algorithm", "SHA1"),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32c", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha1", "7MuLDoLjuZB9Uv63Krr4E7U5x30="),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha256", ""),
				),
			},
		},
	})
}

func TestAccS3ObjectCopy_objectLockLegalHold(t *testing.T) {
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
				Config: testAccObjectCopyConfig_lockLegalHold(rName1, sourceKey, rName2, targetKey, "ON"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectCopyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", "ON"),
				),
			},
			{
				Config: testAccObjectCopyConfig_lockLegalHold(rName1, sourceKey, rName2, targetKey, "OFF"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectCopyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", "OFF"),
				),
			},
		},
	})
}

func TestAccS3ObjectCopy_targetWithMultipleSlashes(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object_copy.test"
	sourceKey := "source"
	targetKey := "/dir//target/"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectCopyConfig_basic(rName1, sourceKey, rName2, targetKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", targetKey),
					resource.TestCheckResourceAttr(resourceName, "source", fmt.Sprintf("%s/%s", rName1, sourceKey)),
				),
			},
		},
	})
}

func TestAccS3ObjectCopy_targetWithMultipleSlashesMigrated(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object_copy.test"
	sourceKey := "source"
	targetKey := "/dir//target/"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3EndpointID),
		CheckDestroy: testAccCheckObjectCopyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					// Final version for aws_s3_object_copy using AWS SDK for Go v1.
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.15.0",
					},
				},
				Config: testAccObjectCopyConfig_basic(rName1, sourceKey, rName2, targetKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", targetKey),
					resource.TestCheckResourceAttr(resourceName, "source", fmt.Sprintf("%s/%s", rName1, sourceKey)),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccObjectCopyConfig_basic(rName1, sourceKey, rName2, targetKey),
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccS3ObjectCopy_directoryBucket(t *testing.T) {
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
		// FIXME "Error running post-test destroy, there may be dangling resources: operation error S3: HeadObject, https response error StatusCode: 403, RequestID: 0033eada6b00018c1826f0b80509eee5684ca4b6, HostID: T7lA2Yxglq, api error Forbidden: Forbidden"
		// CheckDestroy:             testAccCheckObjectCopyDestroy(ctx),
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectCopyConfig_directoryBucket(rName1, sourceKey, rName2, targetKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectCopyExists(ctx, resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "acl"),
					resource.TestCheckResourceAttrSet(resourceName, "bucket"),
					resource.TestCheckResourceAttr(resourceName, "bucket_key_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "cache_control", ""),
					resource.TestCheckNoResourceAttr(resourceName, "checksum_algorithm"),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32c", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha1", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha256", ""),
					resource.TestCheckResourceAttr(resourceName, "content_disposition", ""),
					resource.TestCheckResourceAttr(resourceName, "content_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "content_language", ""),
					resource.TestCheckResourceAttr(resourceName, "content_type", "application/octet-stream"),
					resource.TestCheckNoResourceAttr(resourceName, "copy_if_match"),
					resource.TestCheckNoResourceAttr(resourceName, "copy_if_modified_since"),
					resource.TestCheckNoResourceAttr(resourceName, "copy_if_none_match"),
					resource.TestCheckNoResourceAttr(resourceName, "copy_if_unmodified_since"),
					resource.TestCheckResourceAttr(resourceName, "customer_algorithm", ""),
					resource.TestCheckNoResourceAttr(resourceName, "customer_key"),
					resource.TestCheckResourceAttr(resourceName, "customer_key_md5", ""),
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
					resource.TestCheckResourceAttrSet(resourceName, "source"),
					resource.TestCheckNoResourceAttr(resourceName, "source_customer_algorithm"),
					resource.TestCheckNoResourceAttr(resourceName, "source_customer_key"),
					resource.TestCheckNoResourceAttr(resourceName, "source_customer_key_md5"),
					resource.TestCheckResourceAttr(resourceName, "source_version_id", ""),
					resource.TestCheckResourceAttr(resourceName, "storage_class", "EXPRESS_ONEZONE"),
					resource.TestCheckNoResourceAttr(resourceName, "tagging_directive"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", ""),
					resource.TestCheckResourceAttr(resourceName, "website_redirect", ""),
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

			_, err := tfs3.FindObjectByBucketAndKey(ctx, conn, rs.Primary.Attributes["bucket"], tfs3.SDKv1CompatibleCleanKey(rs.Primary.Attributes["key"]), rs.Primary.Attributes["etag"], "")

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

		_, err := tfs3.FindObjectByBucketAndKey(ctx, conn, rs.Primary.Attributes["bucket"], tfs3.SDKv1CompatibleCleanKey(rs.Primary.Attributes["key"]), rs.Primary.Attributes["etag"], "")

		return err
	}
}

func testAccObjectCopyConfig_baseSourceAndTargetBuckets(sourceBucket, targetBucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "source" {
  bucket = %[1]q

  force_destroy = true
}

resource "aws_s3_bucket" "target" {
  bucket = %[2]q
}
`, sourceBucket, targetBucket)
}

func testAccObjectCopyConfig_baseSourceObject(sourceBucket, sourceKey, targetBucket string) string {
	return acctest.ConfigCompose(testAccObjectCopyConfig_baseSourceAndTargetBuckets(sourceBucket, targetBucket), fmt.Sprintf(`
resource "aws_s3_object" "source" {
  bucket  = aws_s3_bucket.source.bucket
  key     = %[1]q
  content = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
}
`, sourceKey))
}

func testAccObjectCopyConfig_basic(sourceBucket, sourceKey, targetBucket, targetKey string) string {
	return acctest.ConfigCompose(testAccObjectCopyConfig_baseSourceObject(sourceBucket, sourceKey, targetBucket), fmt.Sprintf(`
resource "aws_s3_object_copy" "test" {
  bucket = aws_s3_bucket.target.bucket
  key    = %[1]q
  source = "${aws_s3_bucket.source.bucket}/${aws_s3_object.source.key}"
}
`, targetKey))
}

func testAccObjectCopyConfig_tags1(sourceBucket, sourceKey, targetBucket, targetKey, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccObjectCopyConfig_baseSourceObject(sourceBucket, sourceKey, targetBucket), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccObjectCopyConfig_baseSourceObject(sourceBucket, sourceKey, targetBucket), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccObjectCopyConfig_baseSourceObject(sourceBucket, sourceKey, targetBucket), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccObjectCopyConfig_baseSourceObject(sourceBucket, sourceKey, targetBucket), fmt.Sprintf(`
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

func testAccObjectCopyConfig_baseBucketKeyEnabled(sourceBucket, sourceKey, targetBucket string) string {
	return acctest.ConfigCompose(testAccObjectCopyConfig_baseSourceObject(sourceBucket, sourceKey, targetBucket), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}
`, targetBucket))
}

func testAccObjectCopyConfig_bucketKeyEnabledBucket(sourceBucket, sourceKey, targetBucket, targetKey string) string {
	return acctest.ConfigCompose(testAccObjectCopyConfig_baseBucketKeyEnabled(sourceBucket, sourceKey, targetBucket), fmt.Sprintf(`
resource "aws_s3_bucket_server_side_encryption_configuration" "target" {
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
  depends_on = [aws_s3_bucket_server_side_encryption_configuration.target]

  bucket = aws_s3_bucket.target.bucket
  key    = %[1]q
  source = "${aws_s3_bucket.source.bucket}/${aws_s3_object.source.key}"
}
`, targetKey))
}

func testAccObjectCopyConfig_bucketKeyEnabledObject(sourceBucket, sourceKey, targetBucket, targetKey string) string {
	return acctest.ConfigCompose(testAccObjectCopyConfig_baseBucketKeyEnabled(sourceBucket, sourceKey, targetBucket), fmt.Sprintf(`
resource "aws_s3_object_copy" "test" {
  bucket             = aws_s3_bucket.target.bucket
  bucket_key_enabled = true
  key                = %[1]q
  kms_key_id         = aws_kms_key.test.arn
  source             = "${aws_s3_bucket.source.bucket}/${aws_s3_object.source.key}"
}
`, targetKey))
}

func testAccObjectCopyConfig_externalSourceObject(sourceBucket, sourceKey, targetBucket, targetKey string) string {
	return acctest.ConfigCompose(testAccObjectCopyConfig_baseSourceAndTargetBuckets(sourceBucket, targetBucket), fmt.Sprintf(`
resource "aws_s3_object_copy" "test" {
  bucket = aws_s3_bucket.target.bucket
  key    = %[2]q
  source = "${aws_s3_bucket.source.bucket}/%[1]s"
}
`, sourceKey, targetKey))
}

func testAccObjectCopyConfig_checksumAlgorithm(sourceBucket, sourceKey, targetBucket, targetKey, checksumAlgorithm string) string {
	return acctest.ConfigCompose(testAccObjectCopyConfig_baseSourceObject(sourceBucket, sourceKey, targetBucket), fmt.Sprintf(`
resource "aws_s3_object_copy" "test" {
  bucket = aws_s3_bucket.target.bucket
  key    = %[1]q
  source = "${aws_s3_bucket.source.bucket}/${aws_s3_object.source.key}"

  checksum_algorithm = %[2]q
}
`, targetKey, checksumAlgorithm))
}

func testAccObjectCopyConfig_lockLegalHold(sourceBucket, sourceKey, targetBucket, targetKey, legalHoldStatus string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "source" {
  bucket = %[1]q

  force_destroy = true
}

resource "aws_s3_bucket" "target" {
  bucket = %[3]q

  object_lock_enabled = true

  force_destroy = true
}

resource "aws_s3_bucket_versioning" "target" {
  bucket = aws_s3_bucket.target.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "source" {
  bucket  = aws_s3_bucket.source.bucket
  key     = %[2]q
  content = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
}

resource "aws_s3_object_copy" "test" {
  # Must have bucket versioning enabled first
  bucket = aws_s3_bucket_versioning.target.bucket
  key    = %[4]q
  source = "${aws_s3_bucket.source.bucket}/${aws_s3_object.source.key}"

  object_lock_legal_hold_status = %[5]q
  force_destroy                 = true
}
`, sourceBucket, sourceKey, targetBucket, targetKey, legalHoldStatus)
}

func testAccObjectCopyConfig_directoryBucket(sourceBucket, sourceKey, targetBucket, targetKey string) string {
	return acctest.ConfigCompose(testAccConfigAvailableAZsDirectoryBucket(), fmt.Sprintf(`
locals {
  location_name = data.aws_availability_zones.available.zone_ids[0]
  source_bucket = "%[1]s--${local.location_name}--x-s3"
  target_bucket = "%[3]s--${local.location_name}--x-s3"
}

resource "aws_s3_directory_bucket" "source" {
  bucket = local.source_bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_directory_bucket" "test" {
  bucket = local.target_bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_object" "source" {
  bucket  = aws_s3_directory_bucket.source.bucket
  key     = %[2]q
  content = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
}

resource "aws_s3_object_copy" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  key    = %[4]q
  source = "${aws_s3_object.source.bucket}/${aws_s3_object.source.key}"
}
`, sourceBucket, sourceKey, targetBucket, targetKey))
}
