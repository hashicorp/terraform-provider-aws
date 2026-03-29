// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestParseBucketACLResourceID(t *testing.T) {
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
			InputID:             tfs3.CreateBucketACLResourceID("example", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "example",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket that has hyphens",
			InputID:             tfs3.CreateBucketACLResourceID("my-example-bucket", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "my-example-bucket",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket that has dot and hyphens",
			InputID:             tfs3.CreateBucketACLResourceID("my-example.bucket", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "my-example.bucket",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket that has dots, hyphen, and numbers",
			InputID:             tfs3.CreateBucketACLResourceID("my-example.bucket.4000", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "my-example.bucket.4000",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket and acl",
			InputID:             tfs3.CreateBucketACLResourceID("example", "", string(types.BucketCannedACLPrivate)),
			ExpectedACL:         string(types.BucketCannedACLPrivate),
			ExpectedBucket:      "example",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket and acl that has hyphens",
			InputID:             tfs3.CreateBucketACLResourceID("example", "", string(types.BucketCannedACLPublicReadWrite)),
			ExpectedACL:         string(types.BucketCannedACLPublicReadWrite),
			ExpectedBucket:      "example",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket that has dot, hyphen, and number and acl that has hyphens",
			InputID:             tfs3.CreateBucketACLResourceID("my-example.bucket.4000", "", string(types.BucketCannedACLPublicReadWrite)),
			ExpectedACL:         string(types.BucketCannedACLPublicReadWrite),
			ExpectedBucket:      "my-example.bucket.4000",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket and bucket owner",
			InputID:             tfs3.CreateBucketACLResourceID("example", acctest.Ct12Digit, ""),
			ExpectedACL:         "",
			ExpectedBucket:      "example",
			ExpectedBucketOwner: acctest.Ct12Digit,
		},
		{
			TestName:            "valid ID with bucket that has dot, hyphen, and number and bucket owner",
			InputID:             tfs3.CreateBucketACLResourceID("my-example.bucket.4000", acctest.Ct12Digit, ""),
			ExpectedACL:         "",
			ExpectedBucket:      "my-example.bucket.4000",
			ExpectedBucketOwner: acctest.Ct12Digit,
		},
		{
			TestName:            "valid ID with bucket, bucket owner, and acl",
			InputID:             tfs3.CreateBucketACLResourceID("example", acctest.Ct12Digit, string(types.BucketCannedACLPrivate)),
			ExpectedACL:         string(types.BucketCannedACLPrivate),
			ExpectedBucket:      "example",
			ExpectedBucketOwner: acctest.Ct12Digit,
		},
		{
			TestName:            "valid ID with bucket, bucket owner, and acl that has hyphens",
			InputID:             tfs3.CreateBucketACLResourceID("example", acctest.Ct12Digit, string(types.BucketCannedACLPublicReadWrite)),
			ExpectedACL:         string(types.BucketCannedACLPublicReadWrite),
			ExpectedBucket:      "example",
			ExpectedBucketOwner: acctest.Ct12Digit,
		},
		{
			TestName:            "valid ID with bucket that has dot, hyphen, and numbers, bucket owner, and acl that has hyphens",
			InputID:             tfs3.CreateBucketACLResourceID("my-example.bucket.4000", acctest.Ct12Digit, string(types.BucketCannedACLPublicReadWrite)),
			ExpectedACL:         string(types.BucketCannedACLPublicReadWrite),
			ExpectedBucket:      "my-example.bucket.4000",
			ExpectedBucketOwner: acctest.Ct12Digit,
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1)", //lintignore:AWSAT003
			InputID:             tfs3.CreateBucketACLResourceID("Example", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "Example",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1) that has underscores", //lintignore:AWSAT003
			InputID:             tfs3.CreateBucketACLResourceID("My_Example_Bucket", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "My_Example_Bucket",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1) that has underscore, dot, and hyphens", //lintignore:AWSAT003
			InputID:             tfs3.CreateBucketACLResourceID("My_Example-Bucket.local", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "My_Example-Bucket.local",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1) that has underscore, dots, hyphen, and numbers", //lintignore:AWSAT003
			InputID:             tfs3.CreateBucketACLResourceID("My_Example-Bucket.4000", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "My_Example-Bucket.4000",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1) and acl", //lintignore:AWSAT003
			InputID:             tfs3.CreateBucketACLResourceID("Example", "", string(types.BucketCannedACLPrivate)),
			ExpectedACL:         string(types.BucketCannedACLPrivate),
			ExpectedBucket:      "Example",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1) and acl that has underscores", //lintignore:AWSAT003
			InputID:             tfs3.CreateBucketACLResourceID("My_Example_Bucket", "", string(types.BucketCannedACLPublicReadWrite)),
			ExpectedACL:         string(types.BucketCannedACLPublicReadWrite),
			ExpectedBucket:      "My_Example_Bucket",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1) that has underscore, dot, hyphen, and number and acl that has hyphens", //lintignore:AWSAT003
			InputID:             tfs3.CreateBucketACLResourceID("My_Example-Bucket.4000", "", string(types.BucketCannedACLPublicReadWrite)),
			ExpectedACL:         string(types.BucketCannedACLPublicReadWrite),
			ExpectedBucket:      "My_Example-Bucket.4000",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1) and bucket owner", //lintignore:AWSAT003
			InputID:             tfs3.CreateBucketACLResourceID("Example", acctest.Ct12Digit, ""),
			ExpectedACL:         "",
			ExpectedBucket:      "Example",
			ExpectedBucketOwner: acctest.Ct12Digit,
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1) that has underscore, dot, hyphen, and number and bucket owner", //lintignore:AWSAT003
			InputID:             tfs3.CreateBucketACLResourceID("My_Example-Bucket.4000", acctest.Ct12Digit, ""),
			ExpectedACL:         "",
			ExpectedBucket:      "My_Example-Bucket.4000",
			ExpectedBucketOwner: acctest.Ct12Digit,
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1), bucket owner, and acl", //lintignore:AWSAT003
			InputID:             tfs3.CreateBucketACLResourceID("Example", acctest.Ct12Digit, string(types.BucketCannedACLPrivate)),
			ExpectedACL:         string(types.BucketCannedACLPrivate),
			ExpectedBucket:      "Example",
			ExpectedBucketOwner: acctest.Ct12Digit,
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1), bucket owner, and acl that has hyphens", //lintignore:AWSAT003
			InputID:             tfs3.CreateBucketACLResourceID("Example", acctest.Ct12Digit, string(types.BucketCannedACLPublicReadWrite)),
			ExpectedACL:         string(types.BucketCannedACLPublicReadWrite),
			ExpectedBucket:      "Example",
			ExpectedBucketOwner: acctest.Ct12Digit,
		},
		{
			TestName:            "valid ID with bucket (pre-2018, us-east-1) that has underscore, dot, hyphen, and numbers, bucket owner, and acl that has hyphens", //lintignore:AWSAT003
			InputID:             tfs3.CreateBucketACLResourceID("My_Example-bucket.4000", acctest.Ct12Digit, string(types.BucketCannedACLPublicReadWrite)),
			ExpectedACL:         string(types.BucketCannedACLPublicReadWrite),
			ExpectedBucket:      "My_Example-bucket.4000",
			ExpectedBucketOwner: acctest.Ct12Digit,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			gotBucket, gotExpectedBucketOwner, gotAcl, err := tfs3.ParseBucketACLResourceID(testCase.InputID)

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
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketACLConfig_cannedACL(bucketName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPrivate)),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.grant.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      "1",
						"grantee.0.type": string(types.TypeCanonicalUser),
						"permission":     string(types.PermissionFullControl),
					}),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.owner.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, names.AttrExpectedBucketOwner, ""),
					acctest.CheckResourceAttrFormat(ctx, resourceName, names.AttrID, "{bucket},{acl}"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// DisplayName is deprecated and will be inconsistently returned between July and November 2025.
				ImportStateVerifyIgnore: []string{
					"access_control_policy.0.grant.0.grantee.0.display_name",
					"access_control_policy.0.owner.0.display_name",
				},
			},
		},
	})
}

func TestAccS3BucketACL_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketACLConfig_cannedACL(bucketName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					// Bucket ACL cannot be destroyed, but we can verify Bucket deletion
					// will result in a missing Bucket ACL resource
					acctest.CheckSDKResourceDisappears(ctx, t, tfs3.ResourceBucket(), "aws_s3_bucket.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketACL_migrate_aclNoChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_acl(bucketName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, t, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "acl", string(types.BucketCannedACLPrivate)),
				),
			},
			{
				Config: testAccBucketACLConfig_cannedACL(bucketName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPrivate)),
				),
			},
		},
	})
}

func TestAccS3BucketACL_migrate_aclWithChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_acl(bucketName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, t, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "acl", string(types.BucketCannedACLPrivate)),
				),
			},
			{
				Config: testAccBucketACLConfig_basic_withDisabledPublicAccessBlock(bucketName, string(types.BucketCannedACLPublicRead)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPublicRead)),
				),
			},
		},
	})
}

func TestAccS3BucketACL_migrate_grantsWithChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_acl(bucketName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, t, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "acl", string(types.BucketCannedACLPrivate)),
				),
			},
			{
				Config: testAccBucketACLConfig_migrateGrantsChange(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.grant.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      "1",
						"grantee.0.type": string(types.TypeCanonicalUser),
						"permission":     string(types.PermissionRead),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "access_control_policy.0.grant.*.grantee.0.id", "data.aws_canonical_user_id.current", names.AttrID),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      "1",
						"grantee.0.type": string(types.TypeGroup),
						"permission":     string(types.PermissionReadAcp),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]*regexp.Regexp{
						"grantee.0.uri": regexache.MustCompile(`http://acs.*/groups/s3/LogDelivery`),
					}),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.owner.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_control_policy.0.owner.0.id", "data.aws_canonical_user_id.current", names.AttrID),
				),
			},
		},
	})
}

func TestAccS3BucketACL_updateACL(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, "tf-test-bucket")
	resourceName := "aws_s3_bucket_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketACLConfig_cannedACL(bucketName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPrivate)),
				),
			},
			{
				Config: testAccBucketACLConfig_basic_withDisabledPublicAccessBlock(bucketName, string(types.BucketCannedACLPublicRead)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPublicRead)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// DisplayName is deprecated and will be inconsistently returned between July and November 2025.
				ImportStateVerifyIgnore: []string{
					"access_control_policy.0.grant.0.grantee.0.display_name",
					"access_control_policy.0.grant.1.grantee.0.display_name",
					"access_control_policy.0.owner.0.display_name",
				},
			},
		},
	})
}

func TestAccS3BucketACL_updateGrant(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketACLConfig_grants(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.grant.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      "1",
						"grantee.0.type": string(types.TypeCanonicalUser),
						"permission":     string(types.PermissionFullControl),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      "1",
						"grantee.0.type": string(types.TypeCanonicalUser),
						"permission":     string(types.PermissionWrite),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "access_control_policy.0.grant.*.grantee.0.id", "data.aws_canonical_user_id.current", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.owner.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_control_policy.0.owner.0.id", "data.aws_canonical_user_id.current", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "acl", ""),
					acctest.CheckResourceAttrFormat(ctx, resourceName, names.AttrID, "{bucket}"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// DisplayName is deprecated and will be inconsistently returned between July and November 2025.
				ImportStateVerifyIgnore: []string{
					"access_control_policy.0.grant.0.grantee.0.display_name",
					"access_control_policy.0.grant.1.grantee.0.display_name",
					"access_control_policy.0.owner.0.display_name",
					// Set order is not guaranteed on import. Permissions may be swapped.
					"access_control_policy.0.grant.0.permission",
					"access_control_policy.0.grant.1.permission",
				},
			},
			{
				Config: testAccBucketACLConfig_grantsUpdate(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.grant.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      "1",
						"grantee.0.type": string(types.TypeCanonicalUser),
						"permission":     string(types.PermissionRead),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "access_control_policy.0.grant.*.grantee.0.id", "data.aws_canonical_user_id.current", names.AttrID),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      "1",
						"grantee.0.type": string(types.TypeGroup),
						"permission":     string(types.PermissionReadAcp),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]*regexp.Regexp{
						"grantee.0.uri": regexache.MustCompile(`http://acs.*/groups/s3/LogDelivery`),
					}),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.owner.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_control_policy.0.owner.0.id", "data.aws_canonical_user_id.current", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "acl", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// DisplayName is deprecated and will be inconsistently returned between July and November 2025.
				ImportStateVerifyIgnore: []string{
					"access_control_policy.0.grant.0.grantee.0.display_name",
					"access_control_policy.0.grant.1.grantee.0.display_name",
					"access_control_policy.0.owner.0.display_name",
					// Set order is not guaranteed on import. Permissions may be swapped.
					"access_control_policy.0.grant.0.permission",
					"access_control_policy.0.grant.1.permission",
				},
			},
		},
	})
}

func TestAccS3BucketACL_ACLToGrant(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, "tf-test-bucket")
	resourceName := "aws_s3_bucket_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketACLConfig_cannedACL(bucketName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPrivate)),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", "1"),
				),
			},
			{
				Config: testAccBucketACLConfig_grants(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.grant.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      "1",
						"grantee.0.type": string(types.TypeCanonicalUser),
						"permission":     string(types.PermissionFullControl),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      "1",
						"grantee.0.type": string(types.TypeCanonicalUser),
						"permission":     string(types.PermissionWrite),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "access_control_policy.0.grant.*.grantee.0.id", "data.aws_canonical_user_id.current", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.owner.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_control_policy.0.owner.0.id", "data.aws_canonical_user_id.current", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "acl", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// DisplayName is deprecated and will be inconsistently returned between July and November 2025.
				ImportStateVerifyIgnore: []string{
					"access_control_policy.0.grant.0.grantee.0.display_name",
					"access_control_policy.0.grant.1.grantee.0.display_name",
					"access_control_policy.0.owner.0.display_name",
					// Set order is not guaranteed on import. Permissions may be swapped.
					"access_control_policy.0.grant.0.permission",
					"access_control_policy.0.grant.1.permission",
				},
			},
		},
	})
}

func TestAccS3BucketACL_grantToACL(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, "tf-test-bucket")
	resourceName := "aws_s3_bucket_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketACLConfig_grants(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.grant.#", "2"),
				),
			},
			{
				Config: testAccBucketACLConfig_cannedACL(bucketName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", string(types.BucketCannedACLPrivate)),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.grant.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      "1",
						"grantee.0.type": string(types.TypeCanonicalUser),
						"permission":     string(types.PermissionFullControl),
					}),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.owner.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// DisplayName is deprecated and will be inconsistently returned between July and November 2025.
				ImportStateVerifyIgnore: []string{
					"access_control_policy.0.grant.0.grantee.0.display_name",
					"access_control_policy.0.owner.0.display_name",
				},
			},
		},
	})
}

func TestAccS3BucketACL_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccDirectoryBucketPreCheck(ctx, t)
		},
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

func TestAccS3BucketACL_Identity_CannedACL(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_s3_bucket_acl.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/BucketACL/canned_acl/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrBucket:    knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrBucket)),
				},
			},

			// Step 2: Import command
			{
				ConfigDirectory: config.StaticDirectory("testdata/BucketACL/canned_acl/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateKind:   resource.ImportCommandWithID,
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"access_control_policy.0.grant.0.grantee.0.display_name", "access_control_policy.0.owner.0.display_name",
				},
			},

			// Step 3: Import block with Import ID
			{
				ConfigDirectory: config.StaticDirectory("testdata/BucketACL/canned_acl/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithID,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrBucket), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("acl"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
			},

			// Step 4: Import block with Resource Identity
			{
				ConfigDirectory: config.StaticDirectory("testdata/BucketACL/canned_acl/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrBucket), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("acl"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Resource Identity was added after v6.10.0
func TestAccS3BucketACL_Identity_ExistingResource_CannedACL(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_s3_bucket_acl.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			// Step 1: Create pre-Identity
			{
				ConfigDirectory: config.StaticDirectory("testdata/BucketACL/canned_acl_v6.10.0/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoIdentity(resourceName),
				},
			},

			// Step 2: Current version
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/BucketACL/canned_acl/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrBucket:    knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrBucket)),
				},
			},
		},
	})
}

// Resource Identity version 1 was added in version 6.31.0
func TestAccS3BucketACL_Identity_Upgrade_CannedACL(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_s3_bucket_acl.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			// Step 1: Create with Identity version 0
			{
				ConfigDirectory: config.StaticDirectory("testdata/BucketACL/canned_acl_v6.30.0/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectHasIdentity(resourceName),
				},
			},

			// Step 2: Current version
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/BucketACL/canned_acl/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrBucket:    knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrBucket)),
				},
			},
		},
	})
}

func TestAccS3BucketACL_cannedACL_expectedBucketOwner(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketACLConfig_cannedACL_expectedBucketOwner(bucketName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrExpectedBucketOwner),
					acctest.CheckResourceAttrFormat(ctx, resourceName, names.AttrID, "{bucket},{expected_bucket_owner},{acl}"),
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

func TestAccS3BucketACL_grant_expectedBucketOwner(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketACLConfig_grant_expectedBucketOwner(bucketName, string(types.BucketAccelerateStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrExpectedBucketOwner),
					acctest.CheckResourceAttrFormat(ctx, resourceName, names.AttrID, "{bucket},{expected_bucket_owner}"),
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

func TestAccS3BucketACL_Identity_grant_expectedBucketOwner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				Config: testAccBucketACLConfig_grant_expectedBucketOwner(rName, string(types.BucketAccelerateStatusEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrExpectedBucketOwner),
					acctest.CheckResourceAttrFormat(ctx, resourceName, names.AttrID, "{bucket},{expected_bucket_owner}"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrBucket:    knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrBucket)),
				},
			},

			// Step 2: Import command
			{
				Config:            testAccBucketACLConfig_grant_expectedBucketOwner(rName, string(types.BucketAccelerateStatusEnabled)),
				ImportStateKind:   resource.ImportCommandWithID,
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},

			// Step 3: Import block with Import ID
			{
				Config:          testAccBucketACLConfig_grant_expectedBucketOwner(rName, string(types.BucketAccelerateStatusEnabled)),
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithID,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrBucket), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
			},

			// Step 4: Import block with Resource Identity
			{
				Config:          testAccBucketACLConfig_grant_expectedBucketOwner(rName, string(types.BucketAccelerateStatusEnabled)),
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrBucket), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Resource Identity was added after v6.10.0
func TestAccS3BucketACL_Identity_ExistingResource_grant_expectedBucketOwner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			// Step 1: Create pre-Identity
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.10.0",
					},
				},
				Config: testAccBucketACLConfig_grant_expectedBucketOwner(rName, string(types.BucketAccelerateStatusEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoIdentity(resourceName),
				},
			},

			// Step 2: Current version
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketACLConfig_grant_expectedBucketOwner(rName, string(types.BucketAccelerateStatusEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrExpectedBucketOwner),
					acctest.CheckResourceAttrFormat(ctx, resourceName, names.AttrID, "{bucket},{expected_bucket_owner}"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrBucket:    knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrBucket)),
				},
			},
		},
	})
}

// Resource Identity version 1 was added in version 6.31.0
func TestAccS3BucketACL_Identity_Upgrade_grant_expectedBucketOwner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			// Step 1: Create pre-Identity
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.30.0",
					},
				},
				Config: testAccBucketACLConfig_grant_expectedBucketOwner(rName, string(types.BucketAccelerateStatusEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectHasIdentity(resourceName),
				},
			},

			// Step 2: Current version
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketACLConfig_grant_expectedBucketOwner(rName, string(types.BucketAccelerateStatusEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrExpectedBucketOwner),
					acctest.CheckResourceAttrFormat(ctx, resourceName, names.AttrID, "{bucket},{expected_bucket_owner}"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrBucket:    knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrBucket)),
				},
			},
		},
	})
}

func TestAccS3BucketACL_Identity_cannedACL_expectedBucketOwner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				Config: testAccBucketACLConfig_cannedACL_expectedBucketOwner(rName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrExpectedBucketOwner),
					acctest.CheckResourceAttrFormat(ctx, resourceName, names.AttrID, "{bucket},{expected_bucket_owner},{acl}"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrBucket:    knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrBucket)),
				},
			},

			// Step 2: Import command
			{
				Config:            testAccBucketACLConfig_cannedACL_expectedBucketOwner(rName, string(types.BucketCannedACLPrivate)),
				ImportStateKind:   resource.ImportCommandWithID,
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},

			// Step 3: Import block with Import ID
			{
				Config:          testAccBucketACLConfig_cannedACL_expectedBucketOwner(rName, string(types.BucketCannedACLPrivate)),
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithID,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrBucket), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
			},

			// Step 4: Import block with Resource Identity
			{
				Config:          testAccBucketACLConfig_cannedACL_expectedBucketOwner(rName, string(types.BucketCannedACLPrivate)),
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrBucket), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrExpectedBucketOwner), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Resource Identity was added after v6.10.0
func TestAccS3BucketACL_Identity_ExistingResource_cannedACL_expectedBucketOwner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			// Step 1: Create pre-Identity
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.10.0",
					},
				},
				Config: testAccBucketACLConfig_cannedACL_expectedBucketOwner(rName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoIdentity(resourceName),
				},
			},

			// Step 2: Current version
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketACLConfig_cannedACL_expectedBucketOwner(rName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrExpectedBucketOwner),
					acctest.CheckResourceAttrFormat(ctx, resourceName, names.AttrID, "{bucket},{expected_bucket_owner},{acl}"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrBucket:    knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrBucket)),
				},
			},
		},
	})
}

// Resource Identity version 1 was added in version 6.31.0
func TestAccS3BucketACL_Identity_Upgrade_cannedACL_expectedBucketOwner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_acl.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			// Step 1: Create pre-Identity
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.30.0",
					},
				},
				Config: testAccBucketACLConfig_cannedACL_expectedBucketOwner(rName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectHasIdentity(resourceName),
				},
			},

			// Step 2: Current version
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccBucketACLConfig_cannedACL_expectedBucketOwner(rName, string(types.BucketCannedACLPrivate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketACLExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrExpectedBucketOwner),
					acctest.CheckResourceAttrFormat(ctx, resourceName, names.AttrID, "{bucket},{expected_bucket_owner},{acl}"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrBucket:    knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrBucket)),
				},
			},
		},
	})
}

func testAccCheckBucketACLExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		bucket, expectedBucketOwner, _, err := tfs3.ParseBucketACLResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.ProviderMeta(ctx, t).S3Client(ctx)
		if tfs3.IsDirectoryBucket(bucket) {
			conn = acctest.ProviderMeta(ctx, t).S3ExpressClient(ctx)
		}

		_, err = tfs3.FindBucketACL(ctx, conn, bucket, expectedBucketOwner)

		return err
	}
}

func testAccBucketACLConfig_cannedACL(rName, acl string) string {
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
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_baseAZ(rName), fmt.Sprintf(`
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  acl    = %[1]q
}
`, acl))
}

func testAccBucketACLConfig_cannedACL_expectedBucketOwner(rName, acl string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_acl" "test" {
  bucket                = aws_s3_bucket.test.bucket
  expected_bucket_owner = data.aws_caller_identity.current.account_id

  acl = %[2]q

  depends_on = [aws_s3_bucket_ownership_controls.test]
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

data "aws_caller_identity" "current" {}
`, rName, acl)
}

func testAccBucketACLConfig_grant_expectedBucketOwner(rName, acl string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_acl" "test" {
  bucket                = aws_s3_bucket.test.bucket
  expected_bucket_owner = data.aws_caller_identity.current.account_id

  access_control_policy {
    grant {
      grantee {
        id   = data.aws_canonical_user_id.current.id
        type = "CanonicalUser"
      }
      permission = "FULL_CONTROL"
    }

    owner {
      id = data.aws_canonical_user_id.current.id
    }
  }

  depends_on = [aws_s3_bucket_ownership_controls.test]
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

data "aws_canonical_user_id" "current" {}

data "aws_caller_identity" "current" {}
`, rName, acl)
}
