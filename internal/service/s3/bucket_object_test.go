// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

// WARNING: This code is DEPRECATED and will be removed in a future release!!
// DO NOT apply fixes or enhancements to this file.
// INSTEAD, apply fixes and enhancements to "object_test.go".

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketObject_noNameNoKey(t *testing.T) {
	ctx := acctest.Context(t)
	bucketError := regexache.MustCompile(`bucket must not be empty`)
	keyError := regexache.MustCompile(`key must not be empty`)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig:   func() {},
				Config:      testAccBucketObjectConfig_invalid("", "a key"),
				ExpectError: bucketError,
			},
			{
				PreConfig:   func() {},
				Config:      testAccBucketObjectConfig_invalid("a name", ""),
				ExpectError: keyError,
			},
		},
	})
}

func TestAccS3BucketObject_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, ""),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPrivate)),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, names.AttrARN, "s3", fmt.Sprintf("%s/test-key", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					resource.TestCheckResourceAttr(resourceName, "bucket_key_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cache_control", ""),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrContent),
					resource.TestCheckNoResourceAttr(resourceName, "content_base64"),
					resource.TestCheckResourceAttr(resourceName, "content_disposition", ""),
					resource.TestCheckResourceAttr(resourceName, "content_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "content_language", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrContentType, "application/octet-stream"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, "test-key"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttr(resourceName, "metadata.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption", "AES256"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrSource),
					resource.TestCheckNoResourceAttr(resourceName, "source_hash"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageClass, "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", ""),
					resource.TestCheckResourceAttr(resourceName, "website_redirect", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"acl", names.AttrForceDestroy},
				ImportStateIdFunc:       testAccBucketObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3BucketObject_source(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	source := testAccObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectConfig_source(rName, source),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "{anything will do }"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"acl", names.AttrSource, names.AttrForceDestroy},
				ImportStateIdFunc:       testAccBucketObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3BucketObject_content(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_content(rName, "some_bucket_content"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "some_bucket_content"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"acl", names.AttrContent, "content_base64", names.AttrForceDestroy},
				ImportStateIdFunc:       testAccBucketObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3BucketObject_etagEncryption(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	source := testAccObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_etagEncryption(rName, source),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "{anything will do }"),
					resource.TestCheckResourceAttr(resourceName, "etag", "7b006ff4d70f68cc65061acf2f802e6f"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"acl", names.AttrSource, names.AttrForceDestroy},
				ImportStateIdFunc:       testAccBucketObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3BucketObject_contentBase64(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_contentBase64(rName, base64.StdEncoding.EncodeToString([]byte("some_bucket_content"))),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "some_bucket_content"),
				),
			},
		},
	})
}

func TestAccS3BucketObject_sourceHashTrigger(t *testing.T) {
	ctx := acctest.Context(t)
	var obj, updated_obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	startingData := "Ebben!"
	changingData := "Ne andrò lontana"

	filename := testAccObjectCreateTempFile(t, startingData)
	defer os.Remove(filename)

	rewriteFile := func(*terraform.State) error {
		if err := os.WriteFile(filename, []byte(changingData), 0644); err != nil {
			os.Remove(filename)
			t.Fatal(err)
		}
		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_sourceHashTrigger(rName, filename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "Ebben!"),
					resource.TestCheckResourceAttr(resourceName, "source_hash", "7c7e02a79f28968882bb1426c8f8bfc6"),
					rewriteFile,
				),
				ExpectNonEmptyPlan: true,
			},
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_sourceHashTrigger(rName, filename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &updated_obj),
					testAccCheckObjectBody(&updated_obj, "Ne andrò lontana"),
					resource.TestCheckResourceAttr(resourceName, "source_hash", "cffc5e20de2d21764145b1124c9b337b"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"acl", names.AttrContent, "content_base64", names.AttrForceDestroy, names.AttrSource, "source_hash"},
				ImportStateIdFunc:       testAccBucketObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3BucketObject_withContentCharacteristics(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	source := testAccObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectConfig_contentCharacteristics(rName, source),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "{anything will do }"),
					resource.TestCheckResourceAttr(resourceName, names.AttrContentType, "binary/octet-stream"),
					resource.TestCheckResourceAttr(resourceName, "website_redirect", "http://google.com"),
				),
			},
		},
	})
}

func TestAccS3BucketObject_nonVersioned(t *testing.T) {
	ctx := acctest.Context(t)
	sourceInitial := testAccObjectCreateTempFile(t, "initial object state")
	defer os.Remove(sourceInitial)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var originalObj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAssumeRoleARN(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectConfig_nonVersioned(rName, sourceInitial),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &originalObj),
					testAccCheckObjectBody(&originalObj, "initial object state"),
					resource.TestCheckResourceAttr(resourceName, "version_id", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"acl", names.AttrSource, names.AttrForceDestroy},
				ImportStateIdFunc:       testAccBucketObjectImportStateIdFunc(resourceName)},
		},
	})
}

func TestAccS3BucketObject_updates(t *testing.T) {
	ctx := acctest.Context(t)
	var originalObj, modifiedObj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	sourceInitial := testAccObjectCreateTempFile(t, "initial object state")
	defer os.Remove(sourceInitial)
	sourceModified := testAccObjectCreateTempFile(t, "modified object")
	defer os.Remove(sourceInitial)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectConfig_updateable(rName, false, sourceInitial),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &originalObj),
					testAccCheckObjectBody(&originalObj, "initial object state"),
					resource.TestCheckResourceAttr(resourceName, "etag", "647d1d58e1011c743ec67d5e8af87b53"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
			{
				Config: testAccBucketObjectConfig_updateable(rName, false, sourceModified),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &modifiedObj),
					testAccCheckObjectBody(&modifiedObj, "modified object"),
					resource.TestCheckResourceAttr(resourceName, "etag", "1c7fd13df1515c2a13ad9eb068931f09"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"acl", names.AttrSource, names.AttrForceDestroy},
				ImportStateIdFunc:       testAccBucketObjectImportStateIdFunc(resourceName)},
		},
	})
}

func TestAccS3BucketObject_updateSameFile(t *testing.T) {
	ctx := acctest.Context(t)
	var originalObj, modifiedObj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	startingData := "lane 8"
	changingData := "chicane"

	filename := testAccObjectCreateTempFile(t, startingData)
	defer os.Remove(filename)

	rewriteFile := func(*terraform.State) error {
		if err := os.WriteFile(filename, []byte(changingData), 0644); err != nil {
			os.Remove(filename)
			t.Fatal(err)
		}
		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectConfig_updateable(rName, false, filename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &originalObj),
					testAccCheckObjectBody(&originalObj, startingData),
					resource.TestCheckResourceAttr(resourceName, "etag", "aa48b42f36a2652cbee40c30a5df7d25"),
					rewriteFile,
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccBucketObjectConfig_updateable(rName, false, filename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &modifiedObj),
					testAccCheckObjectBody(&modifiedObj, changingData),
					resource.TestCheckResourceAttr(resourceName, "etag", "fafc05f8c4da0266a99154681ab86e8c"),
				),
			},
		},
	})
}

func TestAccS3BucketObject_updatesWithVersioning(t *testing.T) {
	ctx := acctest.Context(t)
	var originalObj, modifiedObj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	sourceInitial := testAccObjectCreateTempFile(t, "initial versioned object state")
	defer os.Remove(sourceInitial)
	sourceModified := testAccObjectCreateTempFile(t, "modified versioned object")
	defer os.Remove(sourceInitial)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectConfig_updateable(rName, true, sourceInitial),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &originalObj),
					testAccCheckObjectBody(&originalObj, "initial versioned object state"),
					resource.TestCheckResourceAttr(resourceName, "etag", "cee4407fa91906284e2a5e5e03e86b1b"),
				),
			},
			{
				Config: testAccBucketObjectConfig_updateable(rName, true, sourceModified),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &modifiedObj),
					testAccCheckObjectBody(&modifiedObj, "modified versioned object"),
					resource.TestCheckResourceAttr(resourceName, "etag", "00b8c73b1b50e7cc932362c7225b8e29"),
					testAccCheckObjectVersionIDDiffers(&modifiedObj, &originalObj),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"acl", names.AttrSource, names.AttrForceDestroy},
				ImportStateIdFunc:       testAccBucketObjectImportStateIdFunc(resourceName)},
		},
	})
}

func TestAccS3BucketObject_updatesWithVersioningViaAccessPoint(t *testing.T) {
	ctx := acctest.Context(t)
	var originalObj, modifiedObj s3.GetObjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_object.test"
	accessPointResourceName := "aws_s3_access_point.test"

	sourceInitial := testAccObjectCreateTempFile(t, "initial versioned object state")
	defer os.Remove(sourceInitial)
	sourceModified := testAccObjectCreateTempFile(t, "modified versioned object")
	defer os.Remove(sourceInitial)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectConfig_updateableViaAccessPoint(rName, string(types.BucketVersioningStatusEnabled), sourceInitial),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &originalObj),
					testAccCheckObjectBody(&originalObj, "initial versioned object state"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, accessPointResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "etag", "cee4407fa91906284e2a5e5e03e86b1b"),
				),
			},
			{
				Config: testAccBucketObjectConfig_updateableViaAccessPoint(rName, string(types.BucketVersioningStatusEnabled), sourceModified),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &modifiedObj),
					testAccCheckObjectBody(&modifiedObj, "modified versioned object"),
					resource.TestCheckResourceAttr(resourceName, "etag", "00b8c73b1b50e7cc932362c7225b8e29"),
					testAccCheckObjectVersionIDDiffers(&modifiedObj, &originalObj),
				),
			},
		},
	})
}

func TestAccS3BucketObject_kms(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	source := testAccObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_kmsID(rName, source),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectSSE(ctx, resourceName, "aws:kms"),
					testAccCheckObjectBody(&obj, "{anything will do }"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"acl", names.AttrSource, names.AttrForceDestroy},
				ImportStateIdFunc:       testAccBucketObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3BucketObject_sse(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	source := testAccObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_sse(rName, source),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectSSE(ctx, resourceName, "AES256"),
					testAccCheckObjectBody(&obj, "{anything will do }"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"acl", names.AttrSource, names.AttrForceDestroy},
				ImportStateIdFunc:       testAccBucketObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3BucketObject_acl(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2, obj3 s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectConfig_acl(rName, "some_bucket_content", string(types.BucketCannedACLPrivate), true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "some_bucket_content"),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPrivate)),
					testAccCheckObjectACL(ctx, resourceName, []string{"FULL_CONTROL"}),
				),
			},
			{
				Config: testAccBucketObjectConfig_acl(rName, "some_bucket_content", string(types.BucketCannedACLPublicRead), false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj2),
					testAccCheckObjectVersionIDEquals(&obj2, &obj1),
					testAccCheckObjectBody(&obj2, "some_bucket_content"),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPublicRead)),
					testAccCheckObjectACL(ctx, resourceName, []string{"FULL_CONTROL", "READ"}),
				),
			},
			{
				Config: testAccBucketObjectConfig_acl(rName, "changed_some_bucket_content", string(types.BucketCannedACLPrivate), true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj3),
					testAccCheckObjectVersionIDDiffers(&obj3, &obj2),
					testAccCheckObjectBody(&obj3, "changed_some_bucket_content"),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPrivate)),
					testAccCheckObjectACL(ctx, resourceName, []string{"FULL_CONTROL"}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"acl", names.AttrContent, names.AttrForceDestroy},
				ImportStateIdFunc:       testAccBucketObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3BucketObject_metadata(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectConfig_metadata(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, "metadata.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "metadata.key1", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "metadata.key2", acctest.CtValue2),
				),
			},
			{
				Config: testAccBucketObjectConfig_metadata(rName, acctest.CtKey1, acctest.CtValue1Updated, "key3", "value3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, "metadata.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "metadata.key1", acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, "metadata.key3", "value3"),
				),
			},
			{
				Config: testAccBucketObjectConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, "metadata.%", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"acl", names.AttrForceDestroy},
				ImportStateIdFunc:       testAccBucketObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3BucketObject_storageClass(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_content(rName, "some_bucket_content"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageClass, "STANDARD"),
					testAccCheckObjectStorageClass(ctx, resourceName, "STANDARD"),
				),
			},
			{
				Config: testAccBucketObjectConfig_storageClass(rName, "REDUCED_REDUNDANCY"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageClass, "REDUCED_REDUNDANCY"),
					testAccCheckObjectStorageClass(ctx, resourceName, "REDUCED_REDUNDANCY"),
				),
			},
			{
				Config: testAccBucketObjectConfig_storageClass(rName, "GLACIER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Can't GetObject on an object in Glacier without restoring it.
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageClass, "GLACIER"),
					testAccCheckObjectStorageClass(ctx, resourceName, "GLACIER"),
				),
			},
			{
				Config: testAccBucketObjectConfig_storageClass(rName, "INTELLIGENT_TIERING"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageClass, "INTELLIGENT_TIERING"),
					testAccCheckObjectStorageClass(ctx, resourceName, "INTELLIGENT_TIERING"),
				),
			},
			{
				Config: testAccBucketObjectConfig_storageClass(rName, "DEEP_ARCHIVE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// 	Can't GetObject on an object in DEEP_ARCHIVE without restoring it.
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageClass, "DEEP_ARCHIVE"),
					testAccCheckObjectStorageClass(ctx, resourceName, "DEEP_ARCHIVE"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrContent, "acl", names.AttrForceDestroy},
				ImportStateIdFunc:       testAccBucketObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3BucketObject_tagsLeadingSingleSlash(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2, obj3, obj4 s3.GetObjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_object.object"
	key := "/test-key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_tags(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "A@AA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_updatedTags(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj2),
					testAccCheckObjectVersionIDEquals(&obj2, &obj1),
					testAccCheckObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "B@BB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "X X"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "E:/"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_noTags(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj3),
					testAccCheckObjectVersionIDEquals(&obj3, &obj2),
					testAccCheckObjectBody(&obj3, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_tags(rName, key, "changed stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj4),
					testAccCheckObjectVersionIDDiffers(&obj4, &obj3),
					testAccCheckObjectBody(&obj4, "changed stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "A@AA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrContent, "acl", names.AttrForceDestroy},
				ImportStateIdFunc:       testAccBucketObjectImportStateIdFunc(resourceName)},
		},
	})
}

func TestAccS3BucketObject_tagsLeadingMultipleSlashes(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2, obj3, obj4 s3.GetObjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_object.object"
	key := "/////test-key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_tags(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "A@AA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_updatedTags(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj2),
					testAccCheckObjectVersionIDEquals(&obj2, &obj1),
					testAccCheckObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "B@BB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "X X"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "E:/"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_noTags(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj3),
					testAccCheckObjectVersionIDEquals(&obj3, &obj2),
					testAccCheckObjectBody(&obj3, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_tags(rName, key, "changed stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj4),
					testAccCheckObjectVersionIDDiffers(&obj4, &obj3),
					testAccCheckObjectBody(&obj4, "changed stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "A@AA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
		},
	})
}

func TestAccS3BucketObject_tagsMultipleSlashes(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2, obj3, obj4 s3.GetObjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_object.object"
	key := "first//second///third//"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_tags(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "A@AA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_updatedTags(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj2),
					testAccCheckObjectVersionIDEquals(&obj2, &obj1),
					testAccCheckObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "B@BB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "X X"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "E:/"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_noTags(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj3),
					testAccCheckObjectVersionIDEquals(&obj3, &obj2),
					testAccCheckObjectBody(&obj3, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccBucketObjectConfig_tags(rName, key, "changed stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj4),
					testAccCheckObjectVersionIDDiffers(&obj4, &obj3),
					testAccCheckObjectBody(&obj4, "changed stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "A@AA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
		},
	})
}

func TestAccS3BucketObject_objectLockLegalHoldStartWithNone(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2, obj3 s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectConfig_noLockLegalHold(rName, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
			{
				Config: testAccBucketObjectConfig_lockLegalHold(rName, "stuff", "ON"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj2),
					testAccCheckObjectVersionIDEquals(&obj2, &obj1),
					testAccCheckObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", "ON"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
			// Remove legal hold but create a new object version to test force_destroy
			{
				Config: testAccBucketObjectConfig_lockLegalHold(rName, "changed stuff", "OFF"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj3),
					testAccCheckObjectVersionIDDiffers(&obj3, &obj2),
					testAccCheckObjectBody(&obj3, "changed stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
		},
	})
}

func TestAccS3BucketObject_objectLockLegalHoldStartWithOn(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2 s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectConfig_lockLegalHold(rName, "stuff", "ON"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", "ON"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
			{
				Config: testAccBucketObjectConfig_lockLegalHold(rName, "stuff", "OFF"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj2),
					testAccCheckObjectVersionIDEquals(&obj2, &obj1),
					testAccCheckObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
		},
	})
}

func TestAccS3BucketObject_objectLockRetentionStartWithNone(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2, obj3 s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	retainUntilDate := time.Now().UTC().AddDate(0, 0, 10).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectConfig_noLockRetention(rName, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
			{
				Config: testAccBucketObjectConfig_lockRetention(rName, "stuff", retainUntilDate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj2),
					testAccCheckObjectVersionIDEquals(&obj2, &obj1),
					testAccCheckObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", retainUntilDate),
				),
			},
			// Remove retention period but create a new object version to test force_destroy
			{
				Config: testAccBucketObjectConfig_noLockRetention(rName, "changed stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj3),
					testAccCheckObjectVersionIDDiffers(&obj3, &obj2),
					testAccCheckObjectBody(&obj3, "changed stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
		},
	})
}

func TestAccS3BucketObject_objectLockRetentionStartWithSet(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2, obj3, obj4 s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	retainUntilDate1 := time.Now().UTC().AddDate(0, 0, 20).Format(time.RFC3339)
	retainUntilDate2 := time.Now().UTC().AddDate(0, 0, 30).Format(time.RFC3339)
	retainUntilDate3 := time.Now().UTC().AddDate(0, 0, 10).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectConfig_lockRetention(rName, "stuff", retainUntilDate1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", retainUntilDate1),
				),
			},
			{
				Config: testAccBucketObjectConfig_lockRetention(rName, "stuff", retainUntilDate2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj2),
					testAccCheckObjectVersionIDEquals(&obj2, &obj1),
					testAccCheckObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", retainUntilDate2),
				),
			},
			{
				Config: testAccBucketObjectConfig_lockRetention(rName, "stuff", retainUntilDate3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj3),
					testAccCheckObjectVersionIDEquals(&obj3, &obj2),
					testAccCheckObjectBody(&obj3, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", retainUntilDate3),
				),
			},
			{
				Config: testAccBucketObjectConfig_noLockRetention(rName, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj4),
					testAccCheckObjectVersionIDEquals(&obj4, &obj3),
					testAccCheckObjectBody(&obj4, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
		},
	})
}

func TestAccS3BucketObject_objectBucketKeyEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectConfig_objectKeyEnabled(rName, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "bucket_key_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccS3BucketObject_bucketBucketKeyEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectConfig_bucketKeyEnabled(rName, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "bucket_key_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccS3BucketObject_defaultBucketSSE(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1 s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectConfig_defaultSSE(rName, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
				),
			},
		},
	})
}

func TestAccS3BucketObject_ignoreTags(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_object.object"
	key := "test-key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config: acctest.ConfigCompose(
					acctest.ConfigIgnoreTagsKeyPrefixes1("ignorekey"),
					testAccBucketObjectConfig_noTags(rName, key, "stuff")),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "stuff"),
					testAccCheckObjectUpdateTags(ctx, resourceName, nil, map[string]string{"ignorekey1": "ignorevalue1"}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					testAccCheckAllObjectTags(ctx, resourceName, map[string]string{
						"ignorekey1": "ignorevalue1",
					}),
				),
			},
			{
				PreConfig: func() {},
				Config: acctest.ConfigCompose(
					acctest.ConfigIgnoreTagsKeyPrefixes1("ignorekey"),
					testAccBucketObjectConfig_tags(rName, key, "stuff")),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "A@AA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
					testAccCheckAllObjectTags(ctx, resourceName, map[string]string{
						"ignorekey1": "ignorevalue1",
						"Key1":       "A@AA",
						"Key2":       "BBB",
						"Key3":       "CCC",
					}),
				),
			},
		},
	})
}

func testAccCheckBucketObjectDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_object" {
				continue
			}

			_, err := tfs3.FindObjectByBucketAndKey(ctx, conn, rs.Primary.Attributes[names.AttrBucket], tfs3.SDKv1CompatibleCleanKey(rs.Primary.Attributes[names.AttrKey]), rs.Primary.Attributes["etag"], rs.Primary.Attributes["checksum_algorithm"])

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

func testAccCheckBucketObjectExists(ctx context.Context, n string, v *s3.GetObjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		input := &s3.GetObjectInput{
			Bucket:  aws.String(rs.Primary.Attributes[names.AttrBucket]),
			Key:     aws.String(tfs3.SDKv1CompatibleCleanKey(rs.Primary.Attributes[names.AttrKey])),
			IfMatch: aws.String(rs.Primary.Attributes["etag"]),
		}

		output, err := conn.GetObject(ctx, input)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccBucketObjectImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("s3://%s/%s", rs.Primary.Attributes[names.AttrBucket], rs.Primary.Attributes[names.AttrKey]), nil
	}
}

func testAccBucketObjectConfig_invalid(bucket, key string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_object" "object" {
  bucket = %[1]q
  key    = %[2]q
}
`, bucket, key)
}

func testAccBucketObjectConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "object" {
  bucket = aws_s3_bucket.test.bucket
  key    = "test-key"
}
`, rName)
}

func testAccBucketObjectConfig_source(rName string, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "object" {
  bucket       = aws_s3_bucket.test.bucket
  key          = "test-key"
  source       = %[2]q
  content_type = "binary/octet-stream"
}
`, rName, source)
}

func testAccBucketObjectConfig_contentCharacteristics(rName string, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "object" {
  bucket           = aws_s3_bucket.test.bucket
  key              = "test-key"
  source           = %[2]q
  content_language = "en"
  content_type     = "binary/octet-stream"
  website_redirect = "http://google.com"
}
`, rName, source)
}

func testAccBucketObjectConfig_content(rName string, content string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "object" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "test-key"
  content = %[2]q
}
`, rName, content)
}

func testAccBucketObjectConfig_etagEncryption(rName string, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "object" {
  bucket                 = aws_s3_bucket.test.bucket
  key                    = "test-key"
  server_side_encryption = "AES256"
  source                 = %[2]q
  etag                   = filemd5(%[2]q)
}
`, rName, source)
}

func testAccBucketObjectConfig_contentBase64(rName string, contentBase64 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "object" {
  bucket         = aws_s3_bucket.test.bucket
  key            = "test-key"
  content_base64 = %[2]q
}
`, rName, contentBase64)
}

func testAccBucketObjectConfig_sourceHashTrigger(rName string, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "object" {
  bucket      = aws_s3_bucket.test.bucket
  key         = "test-key"
  source      = %[2]q
  source_hash = filemd5(%[2]q)
}
`, rName, source)
}

func testAccBucketObjectConfig_updateable(rName string, bucketVersioning bool, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket_3" {
  bucket = %[1]q
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.object_bucket_3.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_object" "object" {
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.test]

  bucket = aws_s3_bucket.object_bucket_3.bucket
  key    = "updateable-key"
  source = %[3]q
  etag   = filemd5(%[3]q)
}
`, rName, bucketVersioning, source)
}

func testAccBucketObjectConfig_updateableViaAccessPoint(rName, bucketVersioning, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = %[2]q
  }
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q
}

resource "aws_s3_bucket_object" "test" {
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.test]

  bucket = aws_s3_access_point.test.arn
  key    = "updateable-key"
  source = %[3]q
  etag   = filemd5(%[3]q)
}
`, rName, bucketVersioning, source)
}

func testAccBucketObjectConfig_kmsID(rName string, source string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "kms_key_1" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "object" {
  bucket     = aws_s3_bucket.test.bucket
  key        = "test-key"
  source     = %[2]q
  kms_key_id = aws_kms_key.kms_key_1.arn
}
`, rName, source)
}

func testAccBucketObjectConfig_sse(rName string, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "object" {
  bucket                 = aws_s3_bucket.test.bucket
  key                    = "test-key"
  source                 = %[2]q
  server_side_encryption = "AES256"
}
`, rName, source)
}

func testAccBucketObjectConfig_acl(rName, content, acl string, blockPublicAccess bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.id

  block_public_acls       = %[4]t
  block_public_policy     = %[4]t
  ignore_public_acls      = %[4]t
  restrict_public_buckets = %[4]t
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_object" "object" {
  depends_on = [
    aws_s3_bucket_public_access_block.test,
    aws_s3_bucket_ownership_controls.test,
    aws_s3_bucket_versioning.test,
  ]

  bucket  = aws_s3_bucket.test.bucket
  key     = "test-key"
  content = %[2]q
  acl     = %[3]q
}
`, rName, content, acl, blockPublicAccess)
}

func testAccBucketObjectConfig_storageClass(rName string, storage_class string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "object" {
  bucket        = aws_s3_bucket.test.bucket
  key           = "test-key"
  content       = "some_bucket_content"
  storage_class = %[2]q
}
`, rName, storage_class)
}

func testAccBucketObjectConfig_tags(rName, key, content string) string {
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

resource "aws_s3_bucket_object" "object" {
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.test]

  bucket  = aws_s3_bucket.test.bucket
  key     = %[2]q
  content = %[3]q

  tags = {
    Key1 = "A@AA"
    Key2 = "BBB"
    Key3 = "CCC"
  }
}
`, rName, key, content)
}

func testAccBucketObjectConfig_updatedTags(rName, key, content string) string {
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

resource "aws_s3_bucket_object" "object" {
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.test]

  bucket  = aws_s3_bucket.test.bucket
  key     = %[2]q
  content = %[3]q

  tags = {
    Key2 = "B@BB"
    Key3 = "X X"
    Key4 = "DDD"
    Key5 = "E:/"
  }
}
`, rName, key, content)
}

func testAccBucketObjectConfig_noTags(rName, key, content string) string {
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

resource "aws_s3_bucket_object" "object" {
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.test]

  bucket  = aws_s3_bucket.test.bucket
  key     = %[2]q
  content = %[3]q
}
`, rName, key, content)
}

func testAccBucketObjectConfig_metadata(rName string, metadataKey1, metadataValue1, metadataKey2, metadataValue2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "object" {
  bucket = aws_s3_bucket.test.bucket
  key    = "test-key"

  metadata = {
    %[2]s = %[3]q
    %[4]s = %[5]q
  }
}
`, rName, metadataKey1, metadataValue1, metadataKey2, metadataValue2)
}

func testAccBucketObjectConfig_noLockLegalHold(rName string, content string) string {
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

resource "aws_s3_bucket_object" "object" {
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.test]

  bucket        = aws_s3_bucket.test.bucket
  key           = "test-key"
  content       = %[2]q
  force_destroy = true
}
`, rName, content)
}

func testAccBucketObjectConfig_lockLegalHold(rName string, content, legalHoldStatus string) string {
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

resource "aws_s3_bucket_object" "object" {
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.test]

  bucket                        = aws_s3_bucket.test.bucket
  key                           = "test-key"
  content                       = %[2]q
  object_lock_legal_hold_status = %[3]q
  force_destroy                 = true
}
`, rName, content, legalHoldStatus)
}

func testAccBucketObjectConfig_noLockRetention(rName string, content string) string {
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

resource "aws_s3_bucket_object" "object" {
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.test]

  bucket        = aws_s3_bucket.test.bucket
  key           = "test-key"
  content       = %[2]q
  force_destroy = true
}
`, rName, content)
}

func testAccBucketObjectConfig_lockRetention(rName string, content, retainUntilDate string) string {
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

resource "aws_s3_bucket_object" "object" {
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.test]

  bucket                        = aws_s3_bucket.test.bucket
  key                           = "test-key"
  content                       = %[2]q
  force_destroy                 = true
  object_lock_mode              = "GOVERNANCE"
  object_lock_retain_until_date = %[3]q
}
`, rName, content, retainUntilDate)
}

func testAccBucketObjectConfig_nonVersioned(rName string, source string) string {
	policy := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowYeah",
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": "*"
    },
    {
      "Sid": "DenyStm1",
      "Effect": "Deny",
      "Action": [
        "s3:GetObjectVersion*",
        "s3:ListBucketVersions"
      ],
      "Resource": "*"
    }
  ]
}`

	return acctest.ConfigAssumeRolePolicy(policy) + fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket_3" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "object" {
  bucket = aws_s3_bucket.object_bucket_3.bucket
  key    = "updateable-key"
  source = %[2]q
  etag   = filemd5(%[2]q)
}
`, rName, source)
}

func testAccBucketObjectConfig_objectKeyEnabled(rName string, content string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Encrypts test objects"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "object" {
  bucket             = aws_s3_bucket.test.bucket
  key                = "test-key"
  content            = %q
  kms_key_id         = aws_kms_key.test.arn
  bucket_key_enabled = true
}
`, rName, content)
}

func testAccBucketObjectConfig_bucketKeyEnabled(rName string, content string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Encrypts test objects"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.test.arn
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_object" "object" {
  # Must have bucket SSE enabled first
  depends_on = [aws_s3_bucket_server_side_encryption_configuration.test]

  bucket  = aws_s3_bucket.test.bucket
  key     = "test-key"
  content = %q
}
`, rName, content)
}

func testAccBucketObjectConfig_defaultSSE(rName string, content string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Encrypts test objects"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.test.arn
      sse_algorithm     = "aws:kms"
    }
  }
}

resource "aws_s3_bucket_object" "object" {
  # Must have bucket SSE enabled first
  depends_on = [aws_s3_bucket_server_side_encryption_configuration.test]

  bucket  = aws_s3_bucket.test.bucket
  key     = "test-key"
  content = %[2]q
}
`, rName, content)
}
