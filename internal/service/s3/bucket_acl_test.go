// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestBucketACLParseResourceID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName            string
		InputID             string
		ExpectError         bool
		ExpectedACL         string
		ExpectedBucket      string
		ExpectedBucketOwner string
	}{
		{
			TestName:    "empty ID",
			InputID:     "",
			ExpectError: true,
		},
		{
			TestName:    "incorrect bucket and account ID format with slash separator",
			InputID:     "test/123456789012",
			ExpectError: true,
		},
		{
			TestName:    "incorrect bucket, account ID, and ACL format with slash separators",
			InputID:     "test/123456789012/private",
			ExpectError: true,
		},
		{
			TestName:    "incorrect bucket, account ID, and ACL format with mixed separators",
			InputID:     "test/123456789012,private",
			ExpectError: true,
		},
		{
			TestName:    "incorrect bucket, ACL, and account ID format",
			InputID:     "test,private,123456789012",
			ExpectError: true,
		},
		{
			TestName:            "valid ID with bucket",
			InputID:             tfs3.BucketACLCreateResourceID("example", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "example",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket that has hyphens",
			InputID:             tfs3.BucketACLCreateResourceID("my-example-bucket", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "my-example-bucket",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket that has dot and hyphens",
			InputID:             tfs3.BucketACLCreateResourceID("my-example.bucket", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "my-example.bucket",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket that has dots, hyphen, and numbers",
			InputID:             tfs3.BucketACLCreateResourceID("my-example.bucket.4000", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "my-example.bucket.4000",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket and acl",
			InputID:             tfs3.BucketACLCreateResourceID("example", "", string(types.BucketCannedACLPrivate)),
			ExpectedACL:         string(types.BucketCannedACLPrivate),
			ExpectedBucket:      "example",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket and acl that has hyphens",
			InputID:             tfs3.BucketACLCreateResourceID("example", "", string(types.BucketCannedACLPublicReadWrite)),
			ExpectedACL:         string(types.BucketCannedACLPublicReadWrite),
			ExpectedBucket:      "example",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket that has dot, hyphen, and number and acl that has hyphens",
			InputID:             tfs3.BucketACLCreateResourceID("my-example.bucket.4000", "", string(types.BucketCannedACLPublicReadWrite)),
			ExpectedACL:         string(types.BucketCannedACLPublicReadWrite),
			ExpectedBucket:      "my-example.bucket.4000",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket and bucket owner",
			InputID:             tfs3.BucketACLCreateResourceID("example", "123456789012", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "example",
			ExpectedBucketOwner: "123456789012",
		},
		{
			TestName:            "valid ID with bucket that has dot, hyphen, and number and bucket owner",
			InputID:             tfs3.BucketACLCreateResourceID("my-example.bucket.4000", "123456789012", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "my-example.bucket.4000",
			ExpectedBucketOwner: "123456789012",
		},
		{
			TestName:            "valid ID with bucket, bucket owner, and acl",
			InputID:             tfs3.BucketACLCreateResourceID("example", "123456789012", string(types.BucketCannedACLPrivate)),
			ExpectedACL:         string(types.BucketCannedACLPrivate),
			ExpectedBucket:      "example",
			ExpectedBucketOwner: "123456789012",
		},
		{
			TestName:            "valid ID with bucket, bucket owner, and acl that has hyphens",
			InputID:             tfs3.BucketACLCreateResourceID("example", "123456789012", string(types.BucketCannedACLPublicReadWrite)),
			ExpectedACL:         string(types.BucketCannedACLPublicReadWrite),
			ExpectedBucket:      "example",
			ExpectedBucketOwner: "123456789012",
		},
		{
			TestName:            "valid ID with bucket that has dot, hyphen, and numbers, bucket owner, and acl that has hyphens",
			InputID:             tfs3.BucketACLCreateResourceID("my-example.bucket.4000", "123456789012", string(types.BucketCannedACLPublicReadWrite)),
			ExpectedACL:         string(types.BucketCannedACLPublicReadWrite),
			ExpectedBucket:      "my-example.bucket.4000",
			ExpectedBucketOwner: "123456789012",
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1)", //lintignore:AWSAT003
			InputID:             tfs3.BucketACLCreateResourceID("Example", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "Example",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1) that has underscores", //lintignore:AWSAT003
			InputID:             tfs3.BucketACLCreateResourceID("My_Example_Bucket", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "My_Example_Bucket",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1) that has underscore, dot, and hyphens", //lintignore:AWSAT003
			InputID:             tfs3.BucketACLCreateResourceID("My_Example-Bucket.local", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "My_Example-Bucket.local",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1) that has underscore, dots, hyphen, and numbers", //lintignore:AWSAT003
			InputID:             tfs3.BucketACLCreateResourceID("My_Example-Bucket.4000", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "My_Example-Bucket.4000",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1) and acl", //lintignore:AWSAT003
			InputID:             tfs3.BucketACLCreateResourceID("Example", "", string(types.BucketCannedACLPrivate)),
			ExpectedACL:         string(types.BucketCannedACLPrivate),
			ExpectedBucket:      "Example",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1) and acl that has underscores", //lintignore:AWSAT003
			InputID:             tfs3.BucketACLCreateResourceID("My_Example_Bucket", "", string(types.BucketCannedACLPublicReadWrite)),
			ExpectedACL:         string(types.BucketCannedACLPublicReadWrite),
			ExpectedBucket:      "My_Example_Bucket",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1) that has underscore, dot, hyphen, and number and acl that has hyphens", //lintignore:AWSAT003
			InputID:             tfs3.BucketACLCreateResourceID("My_Example-Bucket.4000", "", string(types.BucketCannedACLPublicReadWrite)),
			ExpectedACL:         string(types.BucketCannedACLPublicReadWrite),
			ExpectedBucket:      "My_Example-Bucket.4000",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1) and bucket owner", //lintignore:AWSAT003
			InputID:             tfs3.BucketACLCreateResourceID("Example", "123456789012", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "Example",
			ExpectedBucketOwner: "123456789012",
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1) that has underscore, dot, hyphen, and number and bucket owner", //lintignore:AWSAT003
			InputID:             tfs3.BucketACLCreateResourceID("My_Example-Bucket.4000", "123456789012", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "My_Example-Bucket.4000",
			ExpectedBucketOwner: "123456789012",
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1), bucket owner, and acl", //lintignore:AWSAT003
			InputID:             tfs3.BucketACLCreateResourceID("Example", "123456789012", string(types.BucketCannedACLPrivate)),
			ExpectedACL:         string(types.BucketCannedACLPrivate),
			ExpectedBucket:      "Example",
			ExpectedBucketOwner: "123456789012",
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1), bucket owner, and acl that has hyphens", //lintignore:AWSAT003
			InputID:             tfs3.BucketACLCreateResourceID("Example", "123456789012", string(types.BucketCannedACLPublicReadWrite)),
			ExpectedACL:         string(types.BucketCannedACLPublicReadWrite),
			ExpectedBucket:      "Example",
			ExpectedBucketOwner: "123456789012",
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1) that has underscore, dot, hyphen, and numbers, bucket owner, and acl that has hyphens", //lintignore:AWSAT003
			InputID:             tfs3.BucketACLCreateResourceID("My_Example-bucket.4000", "123456789012", string(types.BucketCannedACLPublicReadWrite)),
			ExpectedACL:         string(types.BucketCannedACLPublicReadWrite),
			ExpectedBucket:      "My_Example-bucket.4000",
			ExpectedBucketOwner: "123456789012",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			gotBucket, gotExpectedBucketOwner, gotAcl, err := tfs3.BucketACLParseResourceID(testCase.InputID)

			if err == nil && testCase.ExpectError {
				t.Fatalf("expected error")
			}

			if err != nil && !testCase.ExpectError {
				t.Fatalf("unexpected error: %s", err)
			}

			if gotAcl != testCase.ExpectedACL {
				t.Errorf("got ACL %s, expected %s", gotAcl, testCase.ExpectedACL)
			}

			if gotBucket != testCase.ExpectedBucket {
				t.Errorf("got bucket %s, expected %s", gotBucket, testCase.ExpectedBucket)
			}

			if gotExpectedBucketOwner != testCase.ExpectedBucketOwner {
				t.Errorf("got ExpectedBucketOwner %s, expected %s", gotExpectedBucketOwner, testCase.ExpectedBucketOwner)
			}
		})
	}
}

func TestAccS3BucketACL_basic(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketACLConfig_basic(bucketName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPrivate)),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.grant.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      acctest.Ct1,
						"grantee.0.type": string(types.TypeCanonicalUser),
						"permission":     string(types.PermissionFullControl),
					}),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.owner.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketACL_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketACLConfig_basic(bucketName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, resourceName),
					// Bucket ACL cannot be destroyed, but we can verify Bucket deletion
					// will result in a missing Bucket ACL resource
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucket(), "aws_s3_bucket.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketACL_migrate_aclNoChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_acl(bucketName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "acl", string(types.BucketCannedACLPrivate)),
				),
			},
			{
				Config: testAccBucketACLConfig_basic(bucketName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPrivate)),
				),
			},
		},
	})
}

func TestAccS3BucketACL_migrate_aclWithChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_acl(bucketName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "acl", string(types.BucketCannedACLPrivate)),
				),
			},
			{
				Config: testAccBucketACLConfig_basic_withDisabledPublicAccessBlock(bucketName, string(types.BucketCannedACLPublicRead)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPublicRead)),
				),
			},
		},
	})
}

func TestAccS3BucketACL_migrate_grantsWithChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_acl(bucketName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "acl", string(types.BucketCannedACLPrivate)),
				),
			},
			{
				Config: testAccBucketACLConfig_migrateGrantsChange(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.grant.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      acctest.Ct1,
						"grantee.0.type": string(types.TypeCanonicalUser),
						"permission":     string(types.PermissionRead),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "access_control_policy.0.grant.*.grantee.0.id", "data.aws_canonical_user_id.current", names.AttrID),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      acctest.Ct1,
						"grantee.0.type": string(types.TypeGroup),
						"permission":     string(types.PermissionReadAcp),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]*regexp.Regexp{
						"grantee.0.uri": regexache.MustCompile(`http://acs.*/groups/s3/LogDelivery`),
					}),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.owner.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "access_control_policy.0.owner.0.id", "data.aws_canonical_user_id.current", names.AttrID),
				),
			},
		},
	})
}

func TestAccS3BucketACL_updateACL(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketACLConfig_basic(bucketName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPrivate)),
				),
			},
			{
				Config: testAccBucketACLConfig_basic_withDisabledPublicAccessBlock(bucketName, string(types.BucketCannedACLPublicRead)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPublicRead)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketACL_updateGrant(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketACLConfig_grants(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.grant.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      acctest.Ct1,
						"grantee.0.type": string(types.TypeCanonicalUser),
						"permission":     string(types.PermissionFullControl),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      acctest.Ct1,
						"grantee.0.type": string(types.TypeCanonicalUser),
						"permission":     string(types.PermissionWrite),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "access_control_policy.0.grant.*.grantee.0.id", "data.aws_canonical_user_id.current", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.owner.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "access_control_policy.0.owner.0.id", "data.aws_canonical_user_id.current", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "acl", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketACLConfig_grantsUpdate(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.grant.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      acctest.Ct1,
						"grantee.0.type": string(types.TypeCanonicalUser),
						"permission":     string(types.PermissionRead),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "access_control_policy.0.grant.*.grantee.0.id", "data.aws_canonical_user_id.current", names.AttrID),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      acctest.Ct1,
						"grantee.0.type": string(types.TypeGroup),
						"permission":     string(types.PermissionReadAcp),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]*regexp.Regexp{
						"grantee.0.uri": regexache.MustCompile(`http://acs.*/groups/s3/LogDelivery`),
					}),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.owner.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "access_control_policy.0.owner.0.id", "data.aws_canonical_user_id.current", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "acl", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketACL_ACLToGrant(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketACLConfig_basic(bucketName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPrivate)),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", acctest.Ct1),
				),
			},
			{
				Config: testAccBucketACLConfig_grants(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.grant.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      acctest.Ct1,
						"grantee.0.type": string(types.TypeCanonicalUser),
						"permission":     string(types.PermissionFullControl),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      acctest.Ct1,
						"grantee.0.type": string(types.TypeCanonicalUser),
						"permission":     string(types.PermissionWrite),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "access_control_policy.0.grant.*.grantee.0.id", "data.aws_canonical_user_id.current", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.owner.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "access_control_policy.0.owner.0.id", "data.aws_canonical_user_id.current", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "acl", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketACL_grantToACL(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketACLConfig_grants(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.grant.#", acctest.Ct2),
				),
			},
			{
				Config: testAccBucketACLConfig_basic(bucketName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPrivate)),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.grant.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      acctest.Ct1,
						"grantee.0.type": string(types.TypeCanonicalUser),
						"permission":     string(types.PermissionFullControl),
					}),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.owner.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketACL_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketACLConfig_directoryBucket(bucketName, string(types.BucketCannedACLPrivate)),
				ExpectError: regexache.MustCompile(`directory buckets are not supported`),
			},
		},
	})
}

func testAccCheckBucketACLExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		bucket, expectedBucketOwner, _, err := tfs3.BucketACLParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		_, err = tfs3.FindBucketACL(ctx, conn, bucket, expectedBucketOwner)

		return err
	}
}

func testAccBucketACLConfig_basic(rName, acl string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "test" {
  depends_on = [aws_s3_bucket_ownership_controls.test]

  bucket = aws_s3_bucket.test.bucket
  acl    = %[2]q
}
`, rName, acl)
}

func testAccBucketACLConfig_basic_withDisabledPublicAccessBlock(rName, acl string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.bucket

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "test" {
  depends_on = [
    aws_s3_bucket_public_access_block.test,
    aws_s3_bucket_ownership_controls.test,
  ]

  bucket = aws_s3_bucket.test.bucket
  acl    = %[2]q
}
`, rName, acl)
}

func testAccBucketACLConfig_grants(bucketName string) string {
	return fmt.Sprintf(`
data "aws_canonical_user_id" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "test" {
  depends_on = [aws_s3_bucket_ownership_controls.test]

  bucket = aws_s3_bucket.test.bucket
  access_control_policy {
    grant {
      grantee {
        id   = data.aws_canonical_user_id.current.id
        type = "CanonicalUser"
      }
      permission = "FULL_CONTROL"
    }

    grant {
      grantee {
        id   = data.aws_canonical_user_id.current.id
        type = "CanonicalUser"
      }
      permission = "WRITE"
    }

    owner {
      id = data.aws_canonical_user_id.current.id
    }
  }
}
`, bucketName)
}

func testAccBucketACLConfig_grantsUpdate(bucketName string) string {
	return fmt.Sprintf(`
data "aws_canonical_user_id" "current" {}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "test" {
  depends_on = [aws_s3_bucket_ownership_controls.test]

  bucket = aws_s3_bucket.test.bucket
  access_control_policy {
    grant {
      grantee {
        id   = data.aws_canonical_user_id.current.id
        type = "CanonicalUser"
      }
      permission = "READ"
    }

    grant {
      grantee {
        type = "Group"
        uri  = "http://acs.${data.aws_partition.current.dns_suffix}/groups/s3/LogDelivery"
      }
      permission = "READ_ACP"
    }

    owner {
      id = data.aws_canonical_user_id.current.id
    }
  }
}
`, bucketName)
}

func testAccBucketACLConfig_migrateGrantsChange(rName string) string {
	return fmt.Sprintf(`
data "aws_canonical_user_id" "current" {}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "test" {
  depends_on = [aws_s3_bucket_ownership_controls.test]

  bucket = aws_s3_bucket.test.bucket
  access_control_policy {
    grant {
      grantee {
        id   = data.aws_canonical_user_id.current.id
        type = "CanonicalUser"
      }
      permission = "READ"
    }

    grant {
      grantee {
        type = "Group"
        uri  = "http://acs.${data.aws_partition.current.dns_suffix}/groups/s3/LogDelivery"
      }
      permission = "READ_ACP"
    }

    owner {
      id = data.aws_canonical_user_id.current.id
    }
  }
}
`, rName)
}

func testAccBucketACLConfig_directoryBucket(rName, acl string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_directory_bucket.test.id
  acl    = %[1]q
}
`, acl))
}
