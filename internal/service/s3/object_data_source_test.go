// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const rfc1123RegexPattern = `^[A-Za-z]{3}, [0-9]+ [A-Za-z]+ [0-9]{4} [0-9:]+ [A-Z]+$`

func TestAccS3ObjectDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object.test"
	dataSourceName := "data.aws_s3_object.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(dataSourceName, "body"),
					resource.TestCheckNoResourceAttr(dataSourceName, "checksum_mode"),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32c", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha1", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha256", ""),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "11"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrContentType, resourceName, names.AttrContentType),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexache.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttr(dataSourceName, "metadata.%", acctest.Ct0),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccS3ObjectDataSource_basicViaAccessPoint(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_object.test"
	resourceName := "aws_s3_object.test"
	accessPointResourceName := "aws_s3_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectDataSourceConfig_basicViaAccessPoint(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrBucket, accessPointResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrKey, resourceName, names.AttrKey),
				),
			},
		},
	})
}

func TestAccS3ObjectDataSource_readableBody(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object.test"
	dataSourceName := "data.aws_s3_object.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectDataSourceConfig_readableBody(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "body", "yes"),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", acctest.Ct3),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrContentType, resourceName, names.AttrContentType),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexache.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
				),
			},
		},
	})
}

func TestAccS3ObjectDataSource_kmsEncrypted(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object.test"
	dataSourceName := "data.aws_s3_object.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectDataSourceConfig_kmsEncrypted(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "body", "Keep Calm and Carry On"),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "22"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrContentType, resourceName, names.AttrContentType),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexache.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "server_side_encryption", resourceName, "server_side_encryption"),
					resource.TestCheckResourceAttrPair(dataSourceName, "sse_kms_key_id", resourceName, names.AttrKMSKeyID),
				),
			},
		},
	})
}

func TestAccS3ObjectDataSource_bucketKeyEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object.test"
	dataSourceName := "data.aws_s3_object.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectDataSourceConfig_bucketKeyEnabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "body", "Keep Calm and Carry On"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_key_enabled", resourceName, "bucket_key_enabled"),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "22"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrContentType, resourceName, names.AttrContentType),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexache.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "server_side_encryption", resourceName, "server_side_encryption"),
					resource.TestCheckResourceAttrPair(dataSourceName, "sse_kms_key_id", resourceName, names.AttrKMSKeyID),
				),
			},
		},
	})
}

func TestAccS3ObjectDataSource_allParams(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object.test"
	dataSourceName := "data.aws_s3_object.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectDataSourceConfig_allParams(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(dataSourceName, "body"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_key_enabled", resourceName, "bucket_key_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cache_control", resourceName, "cache_control"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_disposition", resourceName, "content_disposition"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_encoding", resourceName, "content_encoding"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_language", resourceName, "content_language"),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "25"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrContentType, resourceName, names.AttrContentType),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttr(dataSourceName, "expiration", ""),
					// Currently unsupported in aws_s3_object resource
					resource.TestCheckResourceAttr(dataSourceName, "expires", ""),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexache.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttr(dataSourceName, "metadata.%", acctest.Ct0),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
					// Encryption is off
					resource.TestCheckResourceAttrPair(dataSourceName, "server_side_encryption", resourceName, "server_side_encryption"),
					resource.TestCheckResourceAttr(dataSourceName, "sse_kms_key_id", ""),
					// Supported, but difficult to reproduce in short testing time
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStorageClass, resourceName, names.AttrStorageClass),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "version_id", resourceName, "version_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "website_redirect_location", resourceName, "website_redirect"),
				),
			},
		},
	})
}

func TestAccS3ObjectDataSource_objectLockLegalHoldOff(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object.test"
	dataSourceName := "data.aws_s3_object.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectDataSourceConfig_lockLegalHoldOff(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(dataSourceName, "body"),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "11"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrContentType, resourceName, names.AttrContentType),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexache.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
				),
			},
		},
	})
}

func TestAccS3ObjectDataSource_objectLockLegalHoldOn(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	retainUntilDate := time.Now().UTC().AddDate(0, 0, 10).Format(time.RFC3339)
	resourceName := "aws_s3_object.test"
	dataSourceName := "data.aws_s3_object.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectDataSourceConfig_lockLegalHoldOn(rName, retainUntilDate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(dataSourceName, "body"),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "11"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrContentType, resourceName, names.AttrContentType),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexache.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
				),
			},
		},
	})
}

func TestAccS3ObjectDataSource_leadingSlash(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object.test"
	dataSourceName1 := "data.aws_s3_object.test1"
	dataSourceName2 := "data.aws_s3_object.test2"
	dataSourceName3 := "data.aws_s3_object.test3"

	resourceOnlyConf, conf := testAccObjectDataSourceConfig_leadingSlash(rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{ // nosemgrep:ci.test-config-funcs-correct-form
				Config: resourceOnlyConf,
			},
			{ // nosemgrep:ci.test-config-funcs-correct-form
				Config: conf,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName1, "body", "yes"),
					resource.TestCheckResourceAttr(dataSourceName1, "content_length", acctest.Ct3),
					resource.TestCheckResourceAttrPair(dataSourceName1, names.AttrContentType, resourceName, names.AttrContentType),
					resource.TestCheckResourceAttrPair(dataSourceName1, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName1, "last_modified", regexache.MustCompile(rfc1123RegexPattern)),

					resource.TestCheckResourceAttr(dataSourceName2, "body", "yes"),
					resource.TestCheckResourceAttr(dataSourceName2, "content_length", acctest.Ct3),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrContentType, resourceName, names.AttrContentType),
					resource.TestCheckResourceAttrPair(dataSourceName2, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName2, "last_modified", regexache.MustCompile(rfc1123RegexPattern)),

					resource.TestCheckResourceAttr(dataSourceName3, "body", "yes"),
					resource.TestCheckResourceAttr(dataSourceName3, "content_length", acctest.Ct3),
					resource.TestCheckResourceAttrPair(dataSourceName3, names.AttrContentType, resourceName, names.AttrContentType),
					resource.TestCheckResourceAttrPair(dataSourceName3, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName3, "last_modified", regexache.MustCompile(rfc1123RegexPattern)),
				),
			},
		},
	})
}

func TestAccS3ObjectDataSource_multipleSlashes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName1 := "aws_s3_object.test1"
	resourceName2 := "aws_s3_object.test2"
	dataSourceName1 := "data.aws_s3_object.test1"
	dataSourceName2 := "data.aws_s3_object.test2"
	dataSourceName3 := "data.aws_s3_object.test3"

	resourceOnlyConf, conf := testAccObjectDataSourceConfig_multipleSlashes(rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{ // nosemgrep:ci.test-config-funcs-correct-form
				Config: resourceOnlyConf,
			},
			{ // nosemgrep:ci.test-config-funcs-correct-form
				Config: conf,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName1, "body", "yes"),
					resource.TestCheckResourceAttr(dataSourceName1, "content_length", acctest.Ct3),
					resource.TestCheckResourceAttrPair(dataSourceName1, names.AttrContentType, resourceName1, names.AttrContentType),

					resource.TestCheckResourceAttr(dataSourceName2, "body", "yes"),
					resource.TestCheckResourceAttr(dataSourceName2, "content_length", acctest.Ct3),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrContentType, resourceName1, names.AttrContentType),

					resource.TestCheckResourceAttr(dataSourceName3, "body", "no"),
					resource.TestCheckResourceAttr(dataSourceName3, "content_length", acctest.Ct2),
					resource.TestCheckResourceAttrPair(dataSourceName3, names.AttrContentType, resourceName2, names.AttrContentType),
				),
			},
		},
	})
}

func TestAccS3ObjectDataSource_singleSlashAsKey(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config:      testAccObjectDataSourceConfig_singleSlashAsKey(rName),
				ExpectError: regexache.MustCompile(`input member Key must not be empty`),
			},
		},
	})
}

func TestAccS3ObjectDataSource_leadingDotSlash(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object.test"
	dataSourceName1 := "data.aws_s3_object.test1"
	dataSourceName2 := "data.aws_s3_object.test2"

	resourceOnlyConf, conf := testAccObjectDataSourceConfig_leadingDotSlash(rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{ // nosemgrep:ci.test-config-funcs-correct-form
				Config: resourceOnlyConf,
			},
			{ // nosemgrep:ci.test-config-funcs-correct-form
				Config: conf,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName1, "body", "yes"),
					resource.TestCheckResourceAttr(dataSourceName1, "content_length", acctest.Ct3),
					resource.TestCheckResourceAttrPair(dataSourceName1, names.AttrContentType, resourceName, names.AttrContentType),
					resource.TestCheckResourceAttrPair(dataSourceName1, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName1, "last_modified", regexache.MustCompile(rfc1123RegexPattern)),

					resource.TestCheckResourceAttr(dataSourceName2, "body", "yes"),
					resource.TestCheckResourceAttr(dataSourceName2, "content_length", acctest.Ct3),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrContentType, resourceName, names.AttrContentType),
					resource.TestCheckResourceAttrPair(dataSourceName2, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName2, "last_modified", regexache.MustCompile(rfc1123RegexPattern)),
				),
			},
		},
	})
}

func TestAccS3ObjectDataSource_leadingMultipleSlashes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object.test"
	dataSourceName1 := "data.aws_s3_object.test1"
	dataSourceName2 := "data.aws_s3_object.test2"
	dataSourceName3 := "data.aws_s3_object.test3"

	resourceOnlyConf, conf := testAccObjectDataSourceConfig_leadingMultipleSlashes(rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{ // nosemgrep:ci.test-config-funcs-correct-form
				Config: resourceOnlyConf,
			},
			{ // nosemgrep:ci.test-config-funcs-correct-form
				Config: conf,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName1, "body", "yes"),
					resource.TestCheckResourceAttr(dataSourceName1, "content_length", acctest.Ct3),
					resource.TestCheckResourceAttrPair(dataSourceName1, names.AttrContentType, resourceName, names.AttrContentType),
					resource.TestCheckResourceAttrPair(dataSourceName1, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName1, "last_modified", regexache.MustCompile(rfc1123RegexPattern)),

					resource.TestCheckResourceAttr(dataSourceName2, "body", "yes"),
					resource.TestCheckResourceAttr(dataSourceName2, "content_length", acctest.Ct3),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrContentType, resourceName, names.AttrContentType),
					resource.TestCheckResourceAttrPair(dataSourceName2, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName2, "last_modified", regexache.MustCompile(rfc1123RegexPattern)),

					resource.TestCheckResourceAttr(dataSourceName3, "body", "yes"),
					resource.TestCheckResourceAttr(dataSourceName3, "content_length", acctest.Ct3),
					resource.TestCheckResourceAttrPair(dataSourceName3, names.AttrContentType, resourceName, names.AttrContentType),
					resource.TestCheckResourceAttrPair(dataSourceName3, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName3, "last_modified", regexache.MustCompile(rfc1123RegexPattern)),
				),
			},
		},
	})
}

func TestAccS3ObjectDataSource_checksumMode(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object.test"
	dataSourceName := "data.aws_s3_object.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectDataSourceConfig_checksumMode(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "checksum_mode", "ENABLED"),
					resource.TestCheckResourceAttrPair(dataSourceName, "checksum_crc32", resourceName, "checksum_crc32"),
					resource.TestCheckResourceAttrPair(dataSourceName, "checksum_crc32c", resourceName, "checksum_crc32c"),
					resource.TestCheckResourceAttrPair(dataSourceName, "checksum_sha1", resourceName, "checksum_sha1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "checksum_sha256", resourceName, "checksum_sha256"),
					resource.TestCheckResourceAttrSet(dataSourceName, "checksum_sha256"),
				),
			},
		},
	})
}

func TestAccS3ObjectDataSource_metadata(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_object.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectDataSourceConfig_metadata(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "metadata.%", acctest.Ct2),
					resource.TestCheckResourceAttr(dataSourceName, "metadata.key1", acctest.CtValue1),
					resource.TestCheckResourceAttr(dataSourceName, "metadata.key2", "Value2"),
				),
			},
		},
	})
}

func TestAccS3ObjectDataSource_metadataUppercaseKey(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := fmt.Sprintf("%[1]s-key", rName)
	bucketResourceName := "aws_s3_bucket.test"
	dataSourceName := "data.aws_s3_object.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectDataSourceConfig_metadataBucketOnly(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAddObjectWithMetadata(ctx, bucketResourceName, key, map[string]string{
						acctest.CtKey1: acctest.CtValue1,
						"Key2":         "Value2",
					}),
				),
			},
			{
				Config: testAccObjectDataSourceConfig_metadataBucketAndDS(rName, key),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "metadata.%", acctest.Ct2),
					resource.TestCheckResourceAttr(dataSourceName, "metadata.key1", acctest.CtValue1),
					// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3#HeadObjectOutput
					// Map keys will be normalized to lower-case.
					resource.TestCheckResourceAttr(dataSourceName, "metadata.key2", "Value2"),
				),
			},
		},
	})
}

func TestAccS3ObjectDataSource_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object.test"
	dataSourceName := "data.aws_s3_object.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectDataSourceConfig_directoryBucket(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(dataSourceName, "body"),
					resource.TestCheckNoResourceAttr(dataSourceName, "checksum_mode"),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32c", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha1", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha256", ""),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "11"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrContentType, resourceName, names.AttrContentType),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexache.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttr(dataSourceName, "metadata.%", acctest.Ct0),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func testAccObjectDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "%[1]s-key"
  content = "Hello World"
}

data "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = aws_s3_object.test.key
}
`, rName)
}

func testAccObjectDataSourceConfig_basicViaAccessPoint(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "%[1]s-key"
  content = "Hello World"
}

data "aws_s3_object" "test" {
  bucket = aws_s3_access_point.test.arn
  key    = aws_s3_object.test.key
}
`, rName)
}

func testAccObjectDataSourceConfig_readableBody(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket       = aws_s3_bucket.test.bucket
  key          = "%[1]s-key"
  content      = "yes"
  content_type = "text/plain"
}

data "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = aws_s3_object.test.key
}
`, rName)
}

func testAccObjectDataSourceConfig_kmsEncrypted(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_s3_object" "test" {
  bucket       = aws_s3_bucket.test.bucket
  key          = "%[1]s-key"
  content      = "Keep Calm and Carry On"
  content_type = "text/plain"
  kms_key_id   = aws_kms_key.test.arn
}

data "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = aws_s3_object.test.key
}
`, rName)
}

func testAccObjectDataSourceConfig_bucketKeyEnabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_s3_object" "test" {
  bucket             = aws_s3_bucket.test.bucket
  key                = "%[1]s-key"
  content            = "Keep Calm and Carry On"
  content_type       = "text/plain"
  kms_key_id         = aws_kms_key.test.arn
  bucket_key_enabled = true
}

data "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = aws_s3_object.test.key
}
`, rName)
}

func testAccObjectDataSourceConfig_allParams(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "test" {
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.test]

  bucket = aws_s3_bucket.test.bucket
  key    = "%[1]s-key"

  content             = <<CONTENT
{
  "msg": "Hi there!"
}
CONTENT
  content_type        = "application/unknown"
  cache_control       = "no-cache"
  content_disposition = "attachment"
  content_encoding    = "identity"
  content_language    = "en-GB"

  tags = {
    Name = %[1]q
  }
}

data "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = aws_s3_object.test.key
}
`, rName)
}

func testAccObjectDataSourceConfig_lockLegalHoldOff(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_enabled = true
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "test" {
  # Must have bucket versioning enabled first
  bucket                        = aws_s3_bucket_versioning.test.bucket
  key                           = "%[1]s-key"
  content                       = "Hello World"
  object_lock_legal_hold_status = "OFF"
}

data "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = aws_s3_object.test.key
}
`, rName)
}

func testAccObjectDataSourceConfig_lockLegalHoldOn(rName, retainUntilDate string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_enabled = true
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "test" {
  # Must have bucket versioning enabled first
  bucket                        = aws_s3_bucket_versioning.test.bucket
  key                           = "%[1]s-key"
  content                       = "Hello World"
  force_destroy                 = true
  object_lock_legal_hold_status = "ON"
  object_lock_mode              = "GOVERNANCE"
  object_lock_retain_until_date = %[2]q
}

data "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = aws_s3_object.test.key
}
`, rName, retainUntilDate)
}

func testAccObjectDataSourceConfig_leadingSlash(rName string) (string, string) {
	resources := fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket       = aws_s3_bucket.test.bucket
  key          = "/%[1]s-key"
  content      = "yes"
  content_type = "text/plain"
}
`, rName)

	both := acctest.ConfigCompose(resources, fmt.Sprintf(`
data "aws_s3_object" "test1" {
  bucket = aws_s3_bucket.test.bucket
  key    = "%[1]s-key"
}

data "aws_s3_object" "test2" {
  bucket = aws_s3_bucket.test.bucket
  key    = "/%[1]s-key"
}

data "aws_s3_object" "test3" {
  bucket = aws_s3_bucket.test.bucket
  key    = "//%[1]s-key"
}
`, rName))

	return resources, both
}

func testAccObjectDataSourceConfig_multipleSlashes(rName string) (string, string) {
	resources := fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test1" {
  bucket       = aws_s3_bucket.test.bucket
  key          = "first//second///third//"
  content      = "yes"
  content_type = "text/plain"
}

# Without a trailing slash.
resource "aws_s3_object" "test2" {
  bucket       = aws_s3_bucket.test.bucket
  key          = "/first////second/third"
  content      = "no"
  content_type = "text/plain"
}
`, rName)

	both := acctest.ConfigCompose(resources, `
data "aws_s3_object" "test1" {
  bucket = aws_s3_bucket.test.bucket
  key    = "first/second/third/"
}

data "aws_s3_object" "test2" {
  bucket = aws_s3_bucket.test.bucket
  key    = "first//second///third//"
}

data "aws_s3_object" "test3" {
  bucket = aws_s3_bucket.test.bucket
  key    = "first/second/third"
}
`)

	return resources, both
}

func testAccObjectDataSourceConfig_singleSlashAsKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

data "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "/"
}
`, rName)
}

func testAccObjectDataSourceConfig_leadingDotSlash(rName string) (string, string) {
	resources := fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket       = aws_s3_bucket.test.bucket
  key          = "./%[1]s-key"
  content      = "yes"
  content_type = "text/plain"
}
`, rName)

	both := acctest.ConfigCompose(resources, fmt.Sprintf(`
data "aws_s3_object" "test1" {
  bucket = aws_s3_bucket.test.bucket
  key    = "%[1]s-key"
}

data "aws_s3_object" "test2" {
  bucket = aws_s3_bucket.test.bucket
  key    = "/%[1]s-key"
}
`, rName))

	return resources, both
}

func testAccObjectDataSourceConfig_leadingMultipleSlashes(rName string) (string, string) {
	resources := fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket       = aws_s3_bucket.test.bucket
  key          = "///%[1]s-key"
  content      = "yes"
  content_type = "text/plain"
}
`, rName)

	both := acctest.ConfigCompose(resources, fmt.Sprintf(`
data "aws_s3_object" "test1" {
  bucket = aws_s3_bucket.test.bucket
  key    = "%[1]s-key"
}

data "aws_s3_object" "test2" {
  bucket = aws_s3_bucket.test.bucket
  key    = "/%[1]s-key"
}

data "aws_s3_object" "test3" {
  bucket = aws_s3_bucket.test.bucket
  key    = "//%[1]s-key"
}
`, rName))

	return resources, both
}

func testAccObjectDataSourceConfig_checksumMode(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "%[1]s-key"
  content = "Keep Calm and Carry On"

  checksum_algorithm = "SHA256"
}

data "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = aws_s3_object.test.key

  checksum_mode = "ENABLED"
}
`, rName)
}

func testAccObjectDataSourceConfig_metadata(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "%[1]s-key"
  content = "Hello World"

  metadata = {
    key1 = "value1"
    key2 = "Value2"
  }
}

data "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = aws_s3_object.test.key
}
`, rName)
}

func testAccObjectDataSourceConfig_metadataBucketOnly(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}
`, rName)
}

func testAccObjectDataSourceConfig_metadataBucketAndDS(rName, key string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = %[2]q
}
`, rName, key)
}

func testAccObjectDataSourceConfig_directoryBucket(rName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_directory_bucket.test.bucket
  key     = "%[1]s-key"
  content = "Hello World"
}

data "aws_s3_object" "test" {
  bucket = aws_s3_object.test.bucket
  key    = aws_s3_object.test.key
}
`, rName))
}
