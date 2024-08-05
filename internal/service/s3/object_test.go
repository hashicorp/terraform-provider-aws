// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestSDKv1CompatibleCleanKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		key  string
		want string
	}{
		{
			name: "empty string",
		},
		{
			name: "no slashes",
			key:  "test-key",
			want: "test-key",
		},
		{
			name: "just a slash",
			key:  "/",
			want: "",
		},
		{
			name: "simple slashes",
			key:  "dir1/dir2/test-key",
			want: "dir1/dir2/test-key",
		},
		{
			name: "trailing slash",
			key:  "a/b/c/",
			want: "a/b/c/",
		},
		{
			name: "leading slash",
			key:  "/a/b/c",
			want: "a/b/c",
		},
		{
			name: "leading and trailing slashes",
			key:  "/a/b/c/",
			want: "a/b/c/",
		},
		{
			name: "multiple leading slashes",
			key:  "/////a/b/c",
			want: "a/b/c",
		},
		{
			name: "multiple trailing slashes",
			key:  "a/b/c/////",
			want: "a/b/c/",
		},
		{
			name: "repeated inner slashes",
			key:  "a/b//c///d/////e",
			want: "a/b/c/d/e",
		},
		{
			name: "all the slashes",
			key:  "/a/b//c///d/////e/",
			want: "a/b/c/d/e/",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got, want := tfs3.SDKv1CompatibleCleanKey(testCase.key), testCase.want; got != want {
				t.Errorf("SDKv1CompatibleCleanKey(%q) = %v, want %v", testCase.key, got, want)
			}
		})
	}
}

func TestAccS3Object_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, ""),
					resource.TestCheckNoResourceAttr(resourceName, "acl"),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, names.AttrARN, "s3", fmt.Sprintf("%s/test-key", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					resource.TestCheckResourceAttr(resourceName, "bucket_key_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cache_control", ""),
					resource.TestCheckNoResourceAttr(resourceName, "checksum_algorithm"),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32c", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha1", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha256", ""),
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
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
				ImportStateIdFunc:       testAccObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3Object_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceObject(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3Object_Disappears_bucket(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucket(), "aws_s3_bucket.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3Object_upgradeFromV4(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "4.67.0",
					},
				},
				Config: testAccObjectConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccObjectConfig_basic(rName),
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccS3Object_source(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	source := testAccObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_source(rName, source),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "{anything will do }"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, names.AttrSource},
				ImportStateIdFunc:       testAccObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3Object_content(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_content(rName, "some_bucket_content"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "some_bucket_content"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrContent, "content_base64", names.AttrForceDestroy},
				ImportStateIdFunc:       testAccObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3Object_etagEncryption(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	source := testAccObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_etagEncryption(rName, source),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "{anything will do }"),
					resource.TestCheckResourceAttr(resourceName, "etag", "7b006ff4d70f68cc65061acf2f802e6f"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, names.AttrSource},
				ImportStateIdFunc:       testAccObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3Object_contentBase64(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_contentBase64(rName, base64.StdEncoding.EncodeToString([]byte("some_bucket_content"))),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "some_bucket_content"),
				),
			},
		},
	})
}

func TestAccS3Object_sourceHashTrigger(t *testing.T) {
	ctx := acctest.Context(t)
	var obj, updated_obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
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
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_sourceHashTrigger(rName, filename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "Ebben!"),
					resource.TestCheckResourceAttr(resourceName, "source_hash", "7c7e02a79f28968882bb1426c8f8bfc6"),
					rewriteFile,
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccObjectConfig_sourceHashTrigger(rName, filename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &updated_obj),
					testAccCheckObjectBody(&updated_obj, "Ne andrò lontana"),
					resource.TestCheckResourceAttr(resourceName, "source_hash", "cffc5e20de2d21764145b1124c9b337b"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrContent, "content_base64", names.AttrForceDestroy, names.AttrSource, "source_hash"},
				ImportStateIdFunc:       testAccObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3Object_withContentCharacteristics(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	source := testAccObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_contentCharacteristics(rName, source),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "{anything will do }"),
					resource.TestCheckResourceAttr(resourceName, names.AttrContentType, "binary/octet-stream"),
					resource.TestCheckResourceAttr(resourceName, "website_redirect", "http://google.com"),
				),
			},
		},
	})
}

func TestAccS3Object_nonVersioned(t *testing.T) {
	ctx := acctest.Context(t)
	sourceInitial := testAccObjectCreateTempFile(t, "initial object state")
	defer os.Remove(sourceInitial)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var originalObj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAssumeRoleARN(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_nonVersioned(rName, sourceInitial),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &originalObj),
					testAccCheckObjectBody(&originalObj, "initial object state"),
					resource.TestCheckResourceAttr(resourceName, "version_id", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, names.AttrSource},
				ImportStateId:           fmt.Sprintf("s3://%s/updateable-key", rName),
			},
		},
	})
}

func TestAccS3Object_updates(t *testing.T) {
	ctx := acctest.Context(t)
	var originalObj, modifiedObj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	sourceInitial := testAccObjectCreateTempFile(t, "initial object state")
	defer os.Remove(sourceInitial)
	sourceModified := testAccObjectCreateTempFile(t, "modified object")
	defer os.Remove(sourceInitial)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_updateable(rName, false, sourceInitial),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &originalObj),
					testAccCheckObjectBody(&originalObj, "initial object state"),
					resource.TestCheckResourceAttr(resourceName, "etag", "647d1d58e1011c743ec67d5e8af87b53"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
			{
				Config: testAccObjectConfig_updateable(rName, false, sourceModified),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &modifiedObj),
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
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, names.AttrSource},
				ImportStateId:           fmt.Sprintf("s3://%s/updateable-key", rName),
			},
		},
	})
}

func TestAccS3Object_updateSameFile(t *testing.T) {
	ctx := acctest.Context(t)
	var originalObj, modifiedObj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
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
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_updateable(rName, false, filename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &originalObj),
					testAccCheckObjectBody(&originalObj, startingData),
					resource.TestCheckResourceAttr(resourceName, "etag", "aa48b42f36a2652cbee40c30a5df7d25"),
					rewriteFile,
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccObjectConfig_updateable(rName, false, filename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &modifiedObj),
					testAccCheckObjectBody(&modifiedObj, changingData),
					resource.TestCheckResourceAttr(resourceName, "etag", "fafc05f8c4da0266a99154681ab86e8c"),
				),
			},
		},
	})
}

func TestAccS3Object_updatesWithVersioning(t *testing.T) {
	ctx := acctest.Context(t)
	var originalObj, modifiedObj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	sourceInitial := testAccObjectCreateTempFile(t, "initial versioned object state")
	defer os.Remove(sourceInitial)
	sourceModified := testAccObjectCreateTempFile(t, "modified versioned object")
	defer os.Remove(sourceInitial)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_updateable(rName, true, sourceInitial),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &originalObj),
					testAccCheckObjectBody(&originalObj, "initial versioned object state"),
					resource.TestCheckResourceAttr(resourceName, "etag", "cee4407fa91906284e2a5e5e03e86b1b"),
				),
			},
			{
				Config: testAccObjectConfig_updateable(rName, true, sourceModified),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &modifiedObj),
					testAccCheckObjectBody(&modifiedObj, "modified versioned object"),
					resource.TestCheckResourceAttr(resourceName, "etag", "00b8c73b1b50e7cc932362c7225b8e29"),
					testAccCheckObjectVersionIDDiffers(&modifiedObj, &originalObj),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, names.AttrSource},
				ImportStateId:           fmt.Sprintf("s3://%s/updateable-key", rName),
			},
		},
	})
}

func TestAccS3Object_updatesWithVersioningViaAccessPoint(t *testing.T) {
	ctx := acctest.Context(t)
	var originalObj, modifiedObj s3.GetObjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object.test"
	accessPointResourceName := "aws_s3_access_point.test"

	sourceInitial := testAccObjectCreateTempFile(t, "initial versioned object state")
	defer os.Remove(sourceInitial)
	sourceModified := testAccObjectCreateTempFile(t, "modified versioned object")
	defer os.Remove(sourceInitial)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_updateableViaAccessPoint(rName, sourceInitial),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &originalObj),
					testAccCheckObjectBody(&originalObj, "initial versioned object state"),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "s3", fmt.Sprintf("accesspoint/%s/updateable-key", rName)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, accessPointResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "etag", "cee4407fa91906284e2a5e5e03e86b1b"),
				),
			},
			{
				Config: testAccObjectConfig_updateableViaAccessPoint(rName, sourceModified),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &modifiedObj),
					testAccCheckObjectBody(&modifiedObj, "modified versioned object"),
					resource.TestCheckResourceAttr(resourceName, "etag", "00b8c73b1b50e7cc932362c7225b8e29"),
					testAccCheckObjectVersionIDDiffers(&modifiedObj, &originalObj),
				),
			},
		},
	})
}

func TestAccS3Object_kms(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	source := testAccObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_kmsID(rName, source),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectSSE(ctx, resourceName, "aws:kms"),
					testAccCheckObjectBody(&obj, "{anything will do }"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, names.AttrSource},
				ImportStateIdFunc:       testAccObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3Object_sse(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	source := testAccObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_sse(rName, source),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectSSE(ctx, resourceName, "AES256"),
					testAccCheckObjectBody(&obj, "{anything will do }"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, names.AttrSource},
				ImportStateIdFunc:       testAccObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3Object_acl(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2, obj3 s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_acl(rName, "some_bucket_content", string(types.BucketCannedACLPrivate), true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "some_bucket_content"),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPrivate)),
					testAccCheckObjectACL(ctx, resourceName, []string{"FULL_CONTROL"}),
				),
			},
			{
				Config: testAccObjectConfig_acl(rName, "some_bucket_content", string(types.BucketCannedACLPublicRead), false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj2),
					testAccCheckObjectVersionIDEquals(&obj2, &obj1),
					testAccCheckObjectBody(&obj2, "some_bucket_content"),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPublicRead)),
					testAccCheckObjectACL(ctx, resourceName, []string{"FULL_CONTROL", "READ"}),
				),
			},
			{
				Config: testAccObjectConfig_acl(rName, "changed_some_bucket_content", string(types.BucketCannedACLPrivate), true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj3),
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
				ImportStateIdFunc:       testAccObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3Object_metadata(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_metadata(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, "metadata.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "metadata.key1", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "metadata.key2", acctest.CtValue2),
				),
			},
			{
				Config: testAccObjectConfig_metadata(rName, acctest.CtKey1, acctest.CtValue1Updated, "key3", "value3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, "metadata.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "metadata.key1", acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, "metadata.key3", "value3"),
				),
			},
			{
				Config: testAccObjectConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, "metadata.%", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
				ImportStateIdFunc:       testAccObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3Object_storageClass(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_content(rName, "some_bucket_content"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageClass, "STANDARD"),
					testAccCheckObjectStorageClass(ctx, resourceName, "STANDARD"),
				),
			},
			{
				Config: testAccObjectConfig_storageClass(rName, "REDUCED_REDUNDANCY"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageClass, "REDUCED_REDUNDANCY"),
					testAccCheckObjectStorageClass(ctx, resourceName, "REDUCED_REDUNDANCY"),
				),
			},
			{
				Config: testAccObjectConfig_storageClass(rName, "GLACIER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Can't GetObject on an object in Glacier without restoring it.
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageClass, "GLACIER"),
					testAccCheckObjectStorageClass(ctx, resourceName, "GLACIER"),
				),
			},
			{
				Config: testAccObjectConfig_storageClass(rName, "INTELLIGENT_TIERING"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageClass, "INTELLIGENT_TIERING"),
					testAccCheckObjectStorageClass(ctx, resourceName, "INTELLIGENT_TIERING"),
				),
			},
			{
				Config: testAccObjectConfig_storageClass(rName, "DEEP_ARCHIVE"),
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
				ImportStateVerifyIgnore: []string{names.AttrContent, names.AttrForceDestroy},
				ImportStateIdFunc:       testAccObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3Object_tagsLeadingSingleSlash(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2, obj3, obj4 s3.GetObjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object.object"
	key := "/test-key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_tags(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "A@AA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				Config: testAccObjectConfig_updatedTags(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj2),
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
				Config: testAccObjectConfig_noTags(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj3),
					testAccCheckObjectVersionIDEquals(&obj3, &obj2),
					testAccCheckObjectBody(&obj3, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccObjectConfig_tags(rName, key, "changed stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj4),
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
				ImportStateVerifyIgnore: []string{names.AttrContent, names.AttrForceDestroy},
				ImportStateId:           fmt.Sprintf("s3://%s/%s", rName, key),
			},
		},
	})
}

func TestAccS3Object_tagsLeadingMultipleSlashes(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2, obj3, obj4 s3.GetObjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object.object"
	key := "/////test-key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_tags(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "A@AA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				Config: testAccObjectConfig_updatedTags(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj2),
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
				Config: testAccObjectConfig_noTags(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj3),
					testAccCheckObjectVersionIDEquals(&obj3, &obj2),
					testAccCheckObjectBody(&obj3, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccObjectConfig_tags(rName, key, "changed stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj4),
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

func TestAccS3Object_tagsMultipleSlashes(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2, obj3, obj4 s3.GetObjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object.object"
	key := "first//second///third//"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_tags(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "A@AA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				Config: testAccObjectConfig_updatedTags(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj2),
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
				Config: testAccObjectConfig_noTags(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj3),
					testAccCheckObjectVersionIDEquals(&obj3, &obj2),
					testAccCheckObjectBody(&obj3, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccObjectConfig_tags(rName, key, "changed stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj4),
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

func TestAccS3Object_tagsViaAccessPointARN(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2 s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	key := "test-key"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_tagsViaAccessPointARN(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "A@AA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				Config: testAccObjectConfig_updatedTagsViaAccessPointARN(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj2),
					testAccCheckObjectVersionIDEquals(&obj2, &obj1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "B@BB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "X X"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "E:/"),
				),
			},
		},
	})
}

func TestAccS3Object_tagsViaAccessPointAlias(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2 s3.GetObjectOutput
	const resourceName = "aws_s3_object.object"
	const key = "test-key"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop, // Cannot access the object via the access point alias after the access point is destroyed
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_tagsViaAccessPointAlias(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj1),
					acctest.MatchResourceAttrGlobalARNNoAccount(resourceName, names.AttrARN, "s3", regexache.MustCompile(fmt.Sprintf(`%s-\w+-s3alias/%s`, rName[:20], key))),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "A@AA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				Config: testAccObjectConfig_updatedTagsViaAccessPointAlias(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj2),
					testAccCheckObjectVersionIDEquals(&obj2, &obj1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "B@BB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "X X"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "E:/"),
				),
			},
		},
	})
}

func TestAccS3Object_tagsViaMultiRegionAccessPoint(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2 s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	key := "test-key"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop, // Cannot access the object via the access point alias after the access point is destroyed
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_tagsViaMultiRegionAccessPoint(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "A@AA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				Config: testAccObjectConfig_updatedTagsViaMultiRegionAccessPoint(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj2),
					testAccCheckObjectVersionIDEquals(&obj2, &obj1),
					testAccCheckObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "B@BB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "X X"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "E:/"),
				),
			},
		},
	})
}

func TestAccS3Object_tagsViaObjectLambdaAccessPointARN(t *testing.T) {
	t.Skip("Accessing Objects via Lambda Access Points is not yet supported")
	ctx := acctest.Context(t)
	var obj1, obj2 s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	key := "test-key"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_tagsViaObjectLambdaAccessPointARN(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "A@AA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				Config: testAccObjectConfig_updatedTagsViaObjectLambdaAccessPointARN(rName, key, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj2),
					testAccCheckObjectVersionIDEquals(&obj2, &obj1),
					testAccCheckObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "B@BB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "X X"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "E:/"),
				),
			},
		},
	})
}

func TestAccS3Object_objectLockLegalHoldStartWithNone(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2, obj3 s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_noLockLegalHold(rName, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
			{
				Config: testAccObjectConfig_lockLegalHold(rName, "stuff", "ON"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj2),
					testAccCheckObjectVersionIDEquals(&obj2, &obj1),
					testAccCheckObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", "ON"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
			// Remove legal hold but create a new object version to test force_destroy
			{
				Config: testAccObjectConfig_lockLegalHold(rName, "changed stuff", "OFF"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj3),
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

func TestAccS3Object_objectLockLegalHoldStartWithOn(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2 s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_lockLegalHold(rName, "stuff", "ON"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", "ON"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
			{
				Config: testAccObjectConfig_lockLegalHold(rName, "stuff", "OFF"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj2),
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

func TestAccS3Object_objectLockRetentionStartWithNone(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2, obj3 s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	retainUntilDate := time.Now().UTC().AddDate(0, 0, 10).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_noLockRetention(rName, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
			{
				Config: testAccObjectConfig_lockRetention(rName, "stuff", retainUntilDate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj2),
					testAccCheckObjectVersionIDEquals(&obj2, &obj1),
					testAccCheckObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", retainUntilDate),
				),
			},
			// Remove retention period but create a new object version to test force_destroy
			{
				Config: testAccObjectConfig_noLockRetention(rName, "changed stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj3),
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

func TestAccS3Object_objectLockRetentionStartWithSet(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2, obj3, obj4 s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	retainUntilDate1 := time.Now().UTC().AddDate(0, 0, 20).Format(time.RFC3339)
	retainUntilDate2 := time.Now().UTC().AddDate(0, 0, 30).Format(time.RFC3339)
	retainUntilDate3 := time.Now().UTC().AddDate(0, 0, 10).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_lockRetention(rName, "stuff", retainUntilDate1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", retainUntilDate1),
				),
			},
			{
				Config: testAccObjectConfig_lockRetention(rName, "stuff", retainUntilDate2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj2),
					testAccCheckObjectVersionIDEquals(&obj2, &obj1),
					testAccCheckObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", retainUntilDate2),
				),
			},
			{
				Config: testAccObjectConfig_lockRetention(rName, "stuff", retainUntilDate3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj3),
					testAccCheckObjectVersionIDEquals(&obj3, &obj2),
					testAccCheckObjectBody(&obj3, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", retainUntilDate3),
				),
			},
			{
				Config: testAccObjectConfig_noLockRetention(rName, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj4),
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

func TestAccS3Object_objectBucketKeyEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_bucketKeyEnabled(rName, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "bucket_key_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccS3Object_bucketBucketKeyEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_bucketBucketKeyEnabled(rName, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "bucket_key_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccS3Object_defaultBucketSSE(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1 s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_defaultBucketSSE(rName, "stuff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
				),
			},
		},
	})
}

func TestAccS3Object_ignoreTags(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object.object"
	key := "test-key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigIgnoreTagsKeyPrefixes1("ignorekey"),
					testAccObjectConfig_noTags(rName, key, "stuff")),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "stuff"),
					testAccCheckObjectUpdateTags(ctx, resourceName, nil, map[string]string{"ignorekey1": "ignorevalue1"}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					testAccCheckAllObjectTags(ctx, resourceName, map[string]string{
						"ignorekey1": "ignorevalue1",
					}),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigIgnoreTagsKeyPrefixes1("ignorekey"),
					testAccObjectConfig_tags(rName, key, "stuff")),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
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

func TestAccS3Object_checksumAlgorithm(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_checksumAlgorithm(rName, "CRC32"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "ABCDEFGHIJKLMNOPQRSTUVWXYZ"),
					resource.TestCheckResourceAttr(resourceName, "checksum_algorithm", "CRC32"),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32", "q/d4Ig=="),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32c", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha1", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha256", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"checksum_algorithm", "checksum_crc32", names.AttrContent, names.AttrForceDestroy},
				ImportStateIdFunc:       testAccObjectImportStateIdFunc(resourceName),
			},
			{
				Config: testAccObjectConfig_checksumAlgorithm(rName, "SHA256"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, "ABCDEFGHIJKLMNOPQRSTUVWXYZ"),
					resource.TestCheckResourceAttr(resourceName, "checksum_algorithm", "SHA256"),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32c", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha1", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha256", "1uxomN6H3axuWzYRcIp6ocLSmCkzScwabCmaHbcUnTg="),
				),
			},
		},
	})
}

func TestAccS3Object_keyWithSlashesMigrated(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					// Final version for aws_s3_object using AWS SDK for Go v1.
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.16.0",
					},
				},
				Config: testAccObjectConfig_keyWithSlashes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, "/a/b//c///d/////e/"),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccObjectConfig_keyWithSlashes(rName),
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccS3Object_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_directoryBucket(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					testAccCheckObjectBody(&obj, ""),
					resource.TestCheckNoResourceAttr(resourceName, "acl"),
					acctest.MatchResourceAttrGlobalARNNoAccount(resourceName, names.AttrARN, "s3", regexache.MustCompile(fmt.Sprintf(`%s--[-a-z0-9]+--x-s3/%s$`, rName, "test-key"))),
					resource.TestMatchResourceAttr(resourceName, names.AttrBucket, regexache.MustCompile(fmt.Sprintf(`^%s--[-a-z0-9]+--x-s3$`, rName))),
					resource.TestCheckResourceAttr(resourceName, "bucket_key_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cache_control", ""),
					resource.TestCheckNoResourceAttr(resourceName, "checksum_algorithm"),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_crc32c", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha1", ""),
					resource.TestCheckResourceAttr(resourceName, "checksum_sha256", ""),
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
					resource.TestCheckResourceAttr(resourceName, "override_provider.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "override_provider.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "override_provider.0.default_tags.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "override_provider.0.default_tags.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption", "AES256"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrSource),
					resource.TestCheckNoResourceAttr(resourceName, "source_hash"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageClass, "EXPRESS_ONEZONE"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", ""),
					resource.TestCheckResourceAttr(resourceName, "website_redirect", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "override_provider"},
				ImportStateIdFunc:       testAccObjectImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccS3Object_DirectoryBucket_disappears(t *testing.T) { // nosemgrep:ci.acceptance-test-naming-parent-disappears
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_directoryBucket(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceObject(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3Object_DirectoryBucket_DefaultTags_providerOnly(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(acctest.CtProviderKey1, acctest.CtProviderValue1),
					testAccObjectConfig_directoryBucket(rName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/32385.
func TestAccS3Object_prefix(t *testing.T) {
	ctx := acctest.Context(t)
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_prefix(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, names.AttrContentType, "application/x-directory"),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, "pfx/"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
				ImportStateId:           fmt.Sprintf("s3://%s/pfx/", rName),
			},
		},
	})
}

// S3 bucket is created in the alternate Region.
func TestAccS3Object_crossRegion(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccObjectConfig_crossRegion(rName),
				ExpectError: regexache.MustCompile(`PermanentRedirect`),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/35325.
func TestAccS3Object_optInRegion(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	optInRegion := names.APEast1RegionID // Hong Kong.
	providers := make(map[string]*schema.Provider)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegionNot(t, optInRegion)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckRegionOptIn(ctx, t, optInRegion)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccObjectConfig_optInRegion(rName, optInRegion),
				ExpectError: regexache.MustCompile(`IllegalLocationConstraintException`),
			},
		},
	})
}

func testAccCheckObjectVersionIDDiffers(first, second *s3.GetObjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(first.VersionId) == aws.ToString(second.VersionId) {
			return errors.New("S3 Object version IDs are equal")
		}

		return nil
	}
}

func testAccCheckObjectVersionIDEquals(first, second *s3.GetObjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(first.VersionId) != aws.ToString(second.VersionId) {
			return errors.New("S3 Object version IDs differ")
		}

		return nil
	}
}

func testAccCheckObjectDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_object" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)
			if tfs3.IsDirectoryBucket(rs.Primary.Attributes[names.AttrBucket]) {
				conn = acctest.Provider.Meta().(*conns.AWSClient).S3ExpressClient(ctx)
			}

			var optFns []func(*s3.Options)
			if arn.IsARN(rs.Primary.Attributes[names.AttrBucket]) && conn.Options().Region == names.GlobalRegionID {
				optFns = append(optFns, func(o *s3.Options) { o.UseARNRegion = true })
			}

			_, err := tfs3.FindObjectByBucketAndKey(ctx, conn, rs.Primary.Attributes[names.AttrBucket], tfs3.SDKv1CompatibleCleanKey(rs.Primary.Attributes[names.AttrKey]), rs.Primary.Attributes["etag"], rs.Primary.Attributes["checksum_algorithm"], optFns...)

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

func testAccCheckObjectExists(ctx context.Context, n string, v *s3.GetObjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)
		if tfs3.IsDirectoryBucket(rs.Primary.Attributes[names.AttrBucket]) {
			conn = acctest.Provider.Meta().(*conns.AWSClient).S3ExpressClient(ctx)
		}

		var optFns []func(*s3.Options)
		if arn.IsARN(rs.Primary.Attributes[names.AttrBucket]) && conn.Options().Region == names.GlobalRegionID {
			optFns = append(optFns, func(o *s3.Options) { o.UseARNRegion = true })
		}

		input := &s3.GetObjectInput{
			Bucket:  aws.String(rs.Primary.Attributes[names.AttrBucket]),
			Key:     aws.String(tfs3.SDKv1CompatibleCleanKey(rs.Primary.Attributes[names.AttrKey])),
			IfMatch: aws.String(rs.Primary.Attributes["etag"]),
		}

		output, err := conn.GetObject(ctx, input, optFns...)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccObjectImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("s3://%s/%s", rs.Primary.Attributes[names.AttrBucket], rs.Primary.Attributes[names.AttrKey]), nil
	}
}

func testAccCheckObjectBody(obj *s3.GetObjectOutput, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		body, err := io.ReadAll(obj.Body)
		if err != nil {
			return err
		}
		obj.Body.Close()

		if got := string(body); got != want {
			return fmt.Errorf("S3 Object body = %v, want %v", got, want)
		}

		return nil
	}
}

func testAccCheckObjectACL(ctx context.Context, n string, want []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)
		if tfs3.IsDirectoryBucket(rs.Primary.Attributes[names.AttrBucket]) {
			conn = acctest.Provider.Meta().(*conns.AWSClient).S3ExpressClient(ctx)
		}

		var optFns []func(*s3.Options)
		if arn.IsARN(rs.Primary.Attributes[names.AttrBucket]) && conn.Options().Region == names.GlobalRegionID {
			optFns = append(optFns, func(o *s3.Options) { o.UseARNRegion = true })
		}

		input := &s3.GetObjectAclInput{
			Bucket: aws.String(rs.Primary.Attributes[names.AttrBucket]),
			Key:    aws.String(tfs3.SDKv1CompatibleCleanKey(rs.Primary.Attributes[names.AttrKey])),
		}

		output, err := conn.GetObjectAcl(ctx, input, optFns...)

		if err != nil {
			return err
		}

		var got []string
		for _, v := range output.Grants {
			got = append(got, string(v.Permission))
		}
		sort.Strings(got)

		if diff := cmp.Diff(got, want); diff != "" {
			return fmt.Errorf("unexpected S3 Object ACL diff (+wanted, -got): %s", diff)
		}

		return nil
	}
}

func testAccCheckObjectStorageClass(ctx context.Context, n, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)
		if tfs3.IsDirectoryBucket(rs.Primary.Attributes[names.AttrBucket]) {
			conn = acctest.Provider.Meta().(*conns.AWSClient).S3ExpressClient(ctx)
		}

		var optFns []func(*s3.Options)
		if arn.IsARN(rs.Primary.Attributes[names.AttrBucket]) && conn.Options().Region == names.GlobalRegionID {
			optFns = append(optFns, func(o *s3.Options) { o.UseARNRegion = true })
		}

		output, err := tfs3.FindObjectByBucketAndKey(ctx, conn, rs.Primary.Attributes[names.AttrBucket], tfs3.SDKv1CompatibleCleanKey(rs.Primary.Attributes[names.AttrKey]), "", "", optFns...)

		if err != nil {
			return err
		}

		// The "STANDARD" (which is also the default) storage
		// class when set would not be included in the results.
		storageClass := types.StorageClassStandard
		if output.StorageClass != "" {
			storageClass = output.StorageClass
		}

		if got := string(storageClass); got != want {
			return fmt.Errorf("S3 Object storage class = %v, want %v", got, want)
		}

		return nil
	}
}

func testAccCheckObjectSSE(ctx context.Context, n, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)
		if tfs3.IsDirectoryBucket(rs.Primary.Attributes[names.AttrBucket]) {
			conn = acctest.Provider.Meta().(*conns.AWSClient).S3ExpressClient(ctx)
		}

		var optFns []func(*s3.Options)
		if arn.IsARN(rs.Primary.Attributes[names.AttrBucket]) && conn.Options().Region == names.GlobalRegionID {
			optFns = append(optFns, func(o *s3.Options) { o.UseARNRegion = true })
		}

		output, err := tfs3.FindObjectByBucketAndKey(ctx, conn, rs.Primary.Attributes[names.AttrBucket], tfs3.SDKv1CompatibleCleanKey(rs.Primary.Attributes[names.AttrKey]), "", "", optFns...)

		if err != nil {
			return err
		}

		if got := string(output.ServerSideEncryption); got != want {
			return fmt.Errorf("S3 Object server-side encryption = %v, want %v", got, want)
		}

		return nil
	}
}

func testAccObjectCreateTempFile(t *testing.T, data string) string {
	tmpFile, err := os.CreateTemp("", "tf-acc-s3-obj")
	if err != nil {
		t.Fatal(err)
	}
	filename := tmpFile.Name()

	err = os.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		os.Remove(filename)
		t.Fatal(err)
	}

	return filename
}

func testAccCheckObjectUpdateTags(ctx context.Context, n string, oldTags, newTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)
		if tfs3.IsDirectoryBucket(rs.Primary.Attributes[names.AttrBucket]) {
			conn = acctest.Provider.Meta().(*conns.AWSClient).S3ExpressClient(ctx)
		}

		var optFns []func(*s3.Options)
		if arn.IsARN(rs.Primary.Attributes[names.AttrBucket]) && conn.Options().Region == names.GlobalRegionID {
			optFns = append(optFns, func(o *s3.Options) { o.UseARNRegion = true })
		}

		return tfs3.ObjectUpdateTags(ctx, conn, rs.Primary.Attributes[names.AttrBucket], tfs3.SDKv1CompatibleCleanKey(rs.Primary.Attributes[names.AttrKey]), oldTags, newTags, optFns...)
	}
}

func testAccCheckAllObjectTags(ctx context.Context, n string, expectedTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)
		if tfs3.IsDirectoryBucket(rs.Primary.Attributes[names.AttrBucket]) {
			conn = acctest.Provider.Meta().(*conns.AWSClient).S3ExpressClient(ctx)
		}

		var optFns []func(*s3.Options)
		if arn.IsARN(rs.Primary.Attributes[names.AttrBucket]) && conn.Options().Region == names.GlobalRegionID {
			optFns = append(optFns, func(o *s3.Options) { o.UseARNRegion = true })
		}

		got, err := tfs3.ObjectListTags(ctx, conn, rs.Primary.Attributes[names.AttrBucket], tfs3.SDKv1CompatibleCleanKey(rs.Primary.Attributes[names.AttrKey]), optFns...)
		if err != nil {
			return err
		}

		want := tftags.New(ctx, expectedTags)
		if diff := cmp.Diff(got, want); diff != "" {
			return fmt.Errorf("unexpected S3 Object tags diff (+wanted, -got): %s", diff)
		}

		return nil
	}
}

func testAccObjectConfig_baseAccessPoint(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.bucket
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_access_point" "test" {
  # Must have bucket versioning enabled first
  bucket = aws_s3_bucket_versioning.test.bucket
  name   = %[1]q
}
`, rName)
}

func testAccObjectConfig_baseMultiRegionAccessPoint(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.bucket
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3control_multi_region_access_point" "test" {
  details {
    name = %[1]q

    region {
      bucket = aws_s3_bucket_versioning.test.bucket
    }
  }
}
`, rName)
}

func testAccObjectConfig_baseObjectLambdaAccessPoint(rName string) string {
	return acctest.ConfigCompose(testAccObjectConfig_baseAccessPoint(rName), acctest.ConfigLambdaBase(rName, rName, rName), fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs20.x"
}

resource "aws_s3control_object_lambda_access_point" "test" {
  name = %[1]q

  configuration {
    supporting_access_point = aws_s3_access_point.test.arn

    transformation_configuration {
      actions = ["GetObject"]

      content_transformation {
        aws_lambda {
          function_arn = aws_lambda_function.test.arn
        }
      }
    }
  }
}
`, rName))
}

func testAccObjectConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "object" {
  bucket = aws_s3_bucket.test.bucket
  key    = "test-key"
}
`, rName)
}

func testAccObjectConfig_source(rName string, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "object" {
  bucket       = aws_s3_bucket.test.bucket
  key          = "test-key"
  source       = %[2]q
  content_type = "binary/octet-stream"
}
`, rName, source)
}

func testAccObjectConfig_contentCharacteristics(rName string, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "object" {
  bucket           = aws_s3_bucket.test.bucket
  key              = "test-key"
  source           = %[2]q
  content_language = "en"
  content_type     = "binary/octet-stream"
  website_redirect = "http://google.com"
}
`, rName, source)
}

func testAccObjectConfig_content(rName string, content string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "object" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "test-key"
  content = %[2]q
}
`, rName, content)
}

func testAccObjectConfig_etagEncryption(rName string, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "object" {
  bucket                 = aws_s3_bucket.test.bucket
  key                    = "test-key"
  server_side_encryption = "AES256"
  source                 = %[2]q
  etag                   = filemd5(%[2]q)
}
`, rName, source)
}

func testAccObjectConfig_contentBase64(rName string, contentBase64 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "object" {
  bucket         = aws_s3_bucket.test.bucket
  key            = "test-key"
  content_base64 = %[2]q
}
`, rName, contentBase64)
}

func testAccObjectConfig_sourceHashTrigger(rName string, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "object" {
  bucket      = aws_s3_bucket.test.bucket
  key         = "test-key"
  source      = %[2]q
  source_hash = filemd5(%[2]q)
}
`, rName, source)
}

func testAccObjectConfig_updateable(rName string, bucketVersioning bool, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket_3" {
  bucket = %[1]q
}

resource "aws_s3_bucket_versioning" "object_bucket_3" {
  bucket = aws_s3_bucket.object_bucket_3.bucket
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "object" {
  # Must have bucket versioning enabled first
  bucket = aws_s3_bucket_versioning.object_bucket_3.bucket
  key    = "updateable-key"
  source = %[3]q
  etag   = filemd5(%[3]q)
}
`, rName, bucketVersioning, source)
}

func testAccObjectConfig_updateableViaAccessPoint(rName string, source string) string {
	return acctest.ConfigCompose(testAccObjectConfig_baseAccessPoint(rName), fmt.Sprintf(`
resource "aws_s3_object" "test" {
  bucket = aws_s3_access_point.test.arn
  key    = "updateable-key"
  source = %[1]q
  etag   = filemd5(%[1]q)
}
`, source))
}

func testAccObjectConfig_kmsID(rName string, source string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "kms_key_1" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "object" {
  bucket     = aws_s3_bucket.test.bucket
  key        = "test-key"
  source     = %[2]q
  kms_key_id = aws_kms_key.kms_key_1.arn
}
`, rName, source)
}

func testAccObjectConfig_sse(rName string, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "object" {
  bucket                 = aws_s3_bucket.test.bucket
  key                    = "test-key"
  source                 = %[2]q
  server_side_encryption = "AES256"
}
`, rName, source)
}

func testAccObjectConfig_acl(rName, content, acl string, blockPublicAccess bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.bucket

  block_public_acls       = %[4]t
  block_public_policy     = %[4]t
  ignore_public_acls      = %[4]t
  restrict_public_buckets = %[4]t
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.bucket
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "object" {
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

func testAccObjectConfig_storageClass(rName string, storage_class string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "object" {
  bucket        = aws_s3_bucket.test.bucket
  key           = "test-key"
  content       = "some_bucket_content"
  storage_class = %[2]q
}
`, rName, storage_class)
}

func testAccObjectConfig_tags(rName, key, content string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.bucket
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "object" {
  # Must have bucket versioning enabled first
  bucket  = aws_s3_bucket_versioning.test.bucket
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

func testAccObjectConfig_updatedTags(rName, key, content string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.bucket
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "object" {
  # Must have bucket versioning enabled first
  bucket  = aws_s3_bucket_versioning.test.bucket
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

func testAccObjectConfig_noTags(rName, key, content string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.bucket
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "object" {
  # Must have bucket versioning enabled first
  bucket  = aws_s3_bucket_versioning.test.bucket
  key     = %[2]q
  content = %[3]q
}
`, rName, key, content)
}

func testAccObjectConfig_tagsViaAccessPointARN(rName, key, content string) string {
	return acctest.ConfigCompose(testAccObjectConfig_baseAccessPoint(rName), fmt.Sprintf(`
resource "aws_s3_object" "object" {
  bucket  = aws_s3_access_point.test.arn
  key     = %[1]q
  content = %[2]q

  tags = {
    Key1 = "A@AA"
    Key2 = "BBB"
    Key3 = "CCC"
  }
}
`, key, content))
}

func testAccObjectConfig_updatedTagsViaAccessPointARN(rName, key, content string) string {
	return acctest.ConfigCompose(testAccObjectConfig_baseAccessPoint(rName), fmt.Sprintf(`
resource "aws_s3_object" "object" {
  bucket  = aws_s3_access_point.test.arn
  key     = %[1]q
  content = %[2]q

  tags = {
    Key2 = "B@BB"
    Key3 = "X X"
    Key4 = "DDD"
    Key5 = "E:/"
  }
}
`, key, content))
}

func testAccObjectConfig_tagsViaAccessPointAlias(rName, key, content string) string {
	return acctest.ConfigCompose(testAccObjectConfig_baseAccessPoint(rName), fmt.Sprintf(`
resource "aws_s3_object" "object" {
  bucket  = aws_s3_access_point.test.alias
  key     = %[1]q
  content = %[2]q

  tags = {
    Key1 = "A@AA"
    Key2 = "BBB"
    Key3 = "CCC"
  }
}
`, key, content))
}

func testAccObjectConfig_updatedTagsViaAccessPointAlias(rName, key, content string) string {
	return acctest.ConfigCompose(testAccObjectConfig_baseAccessPoint(rName), fmt.Sprintf(`
resource "aws_s3_object" "object" {
  bucket  = aws_s3_access_point.test.alias
  key     = %[1]q
  content = %[2]q

  tags = {
    Key2 = "B@BB"
    Key3 = "X X"
    Key4 = "DDD"
    Key5 = "E:/"
  }
}
`, key, content))
}

func testAccObjectConfig_tagsViaMultiRegionAccessPoint(rName, key, content string) string {
	return acctest.ConfigCompose(testAccObjectConfig_baseMultiRegionAccessPoint(rName), fmt.Sprintf(`
resource "aws_s3_object" "object" {
  bucket  = aws_s3control_multi_region_access_point.test.arn
  key     = %[1]q
  content = %[2]q

  tags = {
    Key1 = "A@AA"
    Key2 = "BBB"
    Key3 = "CCC"
  }
}
`, key, content))
}

func testAccObjectConfig_updatedTagsViaMultiRegionAccessPoint(rName, key, content string) string {
	return acctest.ConfigCompose(testAccObjectConfig_baseMultiRegionAccessPoint(rName), fmt.Sprintf(`
resource "aws_s3_object" "object" {
  bucket  = aws_s3control_multi_region_access_point.test.arn
  key     = %[1]q
  content = %[2]q

  tags = {
    Key2 = "B@BB"
    Key3 = "X X"
    Key4 = "DDD"
    Key5 = "E:/"
  }
}
`, key, content))
}

func testAccObjectConfig_tagsViaObjectLambdaAccessPointARN(rName, key, content string) string {
	return acctest.ConfigCompose(testAccObjectConfig_baseObjectLambdaAccessPoint(rName), fmt.Sprintf(`
resource "aws_s3_object" "object" {
  bucket  = aws_s3control_object_lambda_access_point.test.arn
  key     = %[1]q
  content = %[2]q

  tags = {
    Key1 = "A@AA"
    Key2 = "BBB"
    Key3 = "CCC"
  }
}
`, key, content))
}

func testAccObjectConfig_updatedTagsViaObjectLambdaAccessPointARN(rName, key, content string) string {
	return acctest.ConfigCompose(testAccObjectConfig_baseObjectLambdaAccessPoint(rName), fmt.Sprintf(`
resource "aws_s3_object" "object" {
  bucket  = aws_s3control_object_lambda_access_point.test.arn
  key     = %[1]q
  content = %[2]q

  tags = {
    Key2 = "B@BB"
    Key3 = "X X"
    Key4 = "DDD"
    Key5 = "E:/"
  }
}
`, key, content))
}

func testAccObjectConfig_metadata(rName string, metadataKey1, metadataValue1, metadataKey2, metadataValue2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "object" {
  bucket = aws_s3_bucket.test.bucket
  key    = "test-key"

  metadata = {
    %[2]s = %[3]q
    %[4]s = %[5]q
  }
}
`, rName, metadataKey1, metadataValue1, metadataKey2, metadataValue2)
}

func testAccObjectConfig_noLockLegalHold(rName string, content string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_enabled = true
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.bucket
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "object" {
  # Must have bucket versioning enabled first
  bucket        = aws_s3_bucket_versioning.test.bucket
  key           = "test-key"
  content       = %[2]q
  force_destroy = true
}
`, rName, content)
}

func testAccObjectConfig_lockLegalHold(rName string, content, legalHoldStatus string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_enabled = true
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.bucket
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "object" {
  # Must have bucket versioning enabled first
  bucket                        = aws_s3_bucket_versioning.test.bucket
  key                           = "test-key"
  content                       = %[2]q
  object_lock_legal_hold_status = %[3]q
  force_destroy                 = true
}
`, rName, content, legalHoldStatus)
}

func testAccObjectConfig_noLockRetention(rName string, content string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_enabled = true
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.bucket
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "object" {
  # Must have bucket versioning enabled first
  bucket        = aws_s3_bucket_versioning.test.bucket
  key           = "test-key"
  content       = %[2]q
  force_destroy = true
}
`, rName, content)
}

func testAccObjectConfig_lockRetention(rName string, content, retainUntilDate string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_enabled = true
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.bucket
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "object" {
  # Must have bucket versioning enabled first
  bucket                        = aws_s3_bucket_versioning.test.bucket
  key                           = "test-key"
  content                       = %[2]q
  force_destroy                 = true
  object_lock_mode              = "GOVERNANCE"
  object_lock_retain_until_date = %[3]q
}
`, rName, content, retainUntilDate)
}

func testAccObjectConfig_nonVersioned(rName string, source string) string {
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

resource "aws_s3_object" "object" {
  bucket = aws_s3_bucket.object_bucket_3.bucket
  key    = "updateable-key"
  source = %[2]q
  etag   = filemd5(%[2]q)
}
`, rName, source)
}

func testAccObjectConfig_bucketKeyEnabled(rName string, content string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Encrypts test objects"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "object" {
  bucket             = aws_s3_bucket.test.bucket
  key                = "test-key"
  content            = %[2]q
  kms_key_id         = aws_kms_key.test.arn
  bucket_key_enabled = true
}
`, rName, content)
}

func testAccObjectConfig_bucketBucketKeyEnabled(rName string, content string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Encrypts test objects"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.test.arn
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_object" "object" {
  # Must have bucket SSE enabled first
  depends_on = [aws_s3_bucket_server_side_encryption_configuration.test]

  bucket  = aws_s3_bucket.test.bucket
  key     = "test-key"
  content = %[2]q
}
`, rName, content)
}

func testAccObjectConfig_defaultBucketSSE(rName string, content string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Encrypts test objects"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.test.arn
      sse_algorithm     = "aws:kms"
    }
  }
}

resource "aws_s3_object" "object" {
  # Must have bucket SSE enabled first
  depends_on = [aws_s3_bucket_server_side_encryption_configuration.test]

  bucket  = aws_s3_bucket.test.bucket
  key     = "test-key"
  content = %[2]q
}
`, rName, content)
}

func testAccObjectConfig_checksumAlgorithm(rName, checksumAlgorithm string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "object" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "test-key"
  content = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

  checksum_algorithm = %[2]q
}
`, rName, checksumAlgorithm)
}

func testAccObjectConfig_keyWithSlashes(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "object" {
  bucket = aws_s3_bucket.test.bucket
  key    = "/a/b//c///d/////e/"
}
`, rName)
}

func testAccObjectConfig_directoryBucket(rName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(rName), `
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }

  force_destroy = true
}

resource "aws_s3_object" "object" {
  bucket = aws_s3_directory_bucket.test.bucket
  key    = "test-key"

  override_provider {
    default_tags {
      tags = {}
    }
  }
}
`)
}

func testAccObjectConfig_prefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "object" {
  bucket       = aws_s3_bucket.test.bucket
  key          = "pfx/"
  content_type = "application/x-directory"
}
`, rName)
}

func testAccObjectConfig_crossRegion(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  provider = awsalternate

  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "object" {
  bucket = aws_s3_bucket.test.bucket
  key    = "test-key"
}
`, rName))
}

func testAccObjectConfig_optInRegion(rName, region string) string {
	return acctest.ConfigCompose(acctest.ConfigNamedRegionalProvider(acctest.ProviderNameAlternate, region), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  provider = awsalternate

  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "object" {
  bucket = aws_s3_bucket.test.bucket
  key    = "test-key"
}
`, rName))
}
