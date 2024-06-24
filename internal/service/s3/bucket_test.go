// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cloudformationtypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.S3ServiceID, testAccErrorCheckSkip)
}

// testAccErrorCheckSkip skips tests that have error messages indicating unsupported features
func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Number of distinct destination bucket ARNs cannot exceed",
		"destination is not allowed",
	)
}

func TestAccS3Bucket_Basic_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	region := acctest.Region()
	hostedZoneID, _ := tfs3.HostedZoneIDForRegion(region)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "acceleration_status", ""),
					resource.TestCheckNoResourceAttr(resourceName, "acl"),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, names.AttrARN, "s3", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					testAccCheckBucketDomainName(ctx, resourceName, "bucket_domain_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucketPrefix, ""),
					resource.TestCheckResourceAttr(resourceName, "bucket_regional_domain_name", testAccBucketRegionalDomainName(rName, region)),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "grant.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "grant.*", map[string]string{
						"permissions.#": acctest.Ct1,
						names.AttrType:  "CanonicalUser",
						names.AttrURI:   "",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrHostedZoneID, hostedZoneID),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "object_lock_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrRegion, region),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "request_payer", "BucketOwner"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", "AES256"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.bucket_key_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "versioning.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.mfa_delete", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "website.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "website_domain"),
					resource.TestCheckNoResourceAttr(resourceName, "website_endpoint"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

// Support for common Terraform 0.11 pattern
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7868
func TestAccS3Bucket_Basic_emptyString(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_emptyString,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucketPrefix, id.UniqueIdPrefix),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccS3Bucket_Basic_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_nameGenerated,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucketPrefix, id.UniqueIdPrefix),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccS3Bucket_Basic_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_namePrefix("tf-test-"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrBucket, "tf-test-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucketPrefix, "tf-test-"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccS3Bucket_Basic_forceDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket.test"
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_forceDestroy(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					testAccCheckBucketAddObjects(ctx, resourceName, "data.txt", "prefix/more_data.txt"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Basic_forceDestroyWithObjectVersions(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket.test"
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_forceDestroyObjectVersions(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					testAccCheckBucketAddObjects(ctx, resourceName, "data.txt", "prefix/more_data.txt"),
					testAccCheckBucketDeleteObjects(ctx, resourceName, "data.txt"), // Creates a delete marker.
					testAccCheckBucketAddObjects(ctx, resourceName, "data.txt"),
				),
			},
		},
	})
}

// By default, the AWS Go SDK cleans up URIs by removing extra slashes
// when the service API requests use the URI as part of making a request.
// While the aws_s3_object resource automatically cleans the key
// to not contain these extra slashes, out-of-band handling and other AWS
// services may create keys with extra slashes (empty "directory" prefixes).
func TestAccS3Bucket_Basic_forceDestroyWithEmptyPrefixes(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket.test"
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_forceDestroy(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					testAccCheckBucketAddObjects(ctx, resourceName, "data.txt", "/extraleadingslash.txt"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Basic_forceDestroyWithObjectLockEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket.test"
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_forceDestroyObjectLockEnabled(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					testAccCheckBucketAddObjectsWithLegalHold(ctx, resourceName, "data.txt", "prefix/more_data.txt"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Basic_acceleration(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, cloudfront.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_acceleration(bucketName, string(types.BucketAccelerateStatusEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "acceleration_status", string(types.BucketAccelerateStatusEnabled)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccBucketConfig_acceleration(bucketName, string(types.BucketAccelerateStatusSuspended)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "acceleration_status", string(types.BucketAccelerateStatusSuspended)),
				),
			},
		},
	})
}

func TestAccS3Bucket_Basic_keyEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_defaultEncryptionKeyEnabledKMSMasterKey(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", "aws:kms"),
					resource.TestMatchResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", regexache.MustCompile("^arn")),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.bucket_key_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Basic_requestPayer(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_requestPayer(bucketName, string(types.PayerBucketOwner)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "request_payer", string(types.PayerBucketOwner)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "acl"},
			},
			{
				Config: testAccBucketConfig_requestPayer(bucketName, string(types.PayerRequester)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "request_payer", string(types.PayerRequester)),
				),
			},
		},
	})
}

// Test TestAccS3Bucket_disappears is designed to fail with a "plan
// not empty" error in Terraform, to check against regressions.
// See https://github.com/hashicorp/terraform/pull/2925
func TestAccS3Bucket_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_basic(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucket(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3Bucket_Duplicate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	region := acctest.Region()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegionNot(t, names.USEast1RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketConfig_duplicate(region, bucketName),
				ExpectError: regexache.MustCompile(tfs3.ErrCodeBucketAlreadyOwnedByYou),
			},
		},
	})
}

func TestAccS3Bucket_Duplicate_UsEast1(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartition(t, names.StandardPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketConfig_duplicate(names.USEast1RegionID, bucketName),
				ExpectError: regexache.MustCompile(tfs3.ErrCodeBucketAlreadyExists),
			},
		},
	})
}

func TestAccS3Bucket_Duplicate_UsEast1AltAccount(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartition(t, names.StandardPartitionID)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketConfig_duplicateAltAccount(names.USEast1RegionID, bucketName),
				ExpectError: regexache.MustCompile(tfs3.ErrCodeBucketAlreadyExists),
			},
		},
	})
}

func TestAccS3Bucket_tags_withSystemTags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket.test"
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")

	var stackID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckBucketDestroy(ctx),
			func(s *terraform.State) error {
				// Tear down CF stack.
				conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationClient(ctx)

				requestToken := id.UniqueId()
				req := &cloudformation.DeleteStackInput{
					StackName:          aws.String(stackID),
					ClientRequestToken: aws.String(requestToken),
				}

				log.Printf("[DEBUG] Deleting CloudFormation stack: %s", stackID)
				if _, err := conn.DeleteStack(ctx, req); err != nil {
					return fmt.Errorf("error deleting CloudFormation stack: %w", err)
				}

				if _, err := tfcloudformation.WaitStackDeleted(ctx, conn, stackID, requestToken, 10*time.Minute); err != nil {
					return fmt.Errorf("Error waiting for CloudFormation stack deletion: %s", err)
				}

				return nil
			},
		),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_noTags(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucket(), resourceName),
					testAccCheckBucketCreateViaCloudFormation(ctx, bucketName, &stackID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccBucketConfig_tags(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
					testAccCheckBucketTagKeys(ctx, resourceName, "aws:cloudformation:stack-name", "aws:cloudformation:stack-id", "aws:cloudformation:logical-id"),
				),
			},
			{
				Config: testAccBucketConfig_updatedTags(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "XXX"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "EEE"),
					testAccCheckBucketTagKeys(ctx, resourceName, "aws:cloudformation:stack-name", "aws:cloudformation:stack-id", "aws:cloudformation:logical-id"),
				),
			},
			{
				Config: testAccBucketConfig_noTags(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					testAccCheckBucketTagKeys(ctx, resourceName, "aws:cloudformation:stack-name", "aws:cloudformation:stack-id", "aws:cloudformation:logical-id"),
				),
			},
		},
	})
}

func TestAccS3Bucket_tags_ignoreTags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigIgnoreTagsKeyPrefixes1("ignorekey"),
					testAccBucketConfig_noTags(bucketName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					testAccCheckBucketUpdateTags(ctx, resourceName, nil, map[string]string{"ignorekey1": "ignorevalue1"}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					testAccCheckBucketCheckTags(ctx, resourceName, map[string]string{
						"ignorekey1": "ignorevalue1",
					}),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigIgnoreTagsKeyPrefixes1("ignorekey"),
					testAccBucketConfig_tags(bucketName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
					testAccCheckBucketCheckTags(ctx, resourceName, map[string]string{
						"ignorekey1": "ignorevalue1",
						"Key1":       "AAA",
						"Key2":       "BBB",
						"Key3":       "CCC",
					}),
				),
			},
		},
	})
}

func TestAccS3Bucket_Manage_lifecycleBasic(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_lifecycle(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.#", "6"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.days", "365"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.date", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.expired_object_delete_marker", acctest.CtFalse),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":                 "",
						"days":                 "30",
						names.AttrStorageClass: "STANDARD_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":                 "",
						"days":                 "60",
						names.AttrStorageClass: "INTELLIGENT_TIERING",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":                 "",
						"days":                 "90",
						names.AttrStorageClass: "ONEZONE_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":                 "",
						"days":                 "120",
						names.AttrStorageClass: "GLACIER",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":                 "",
						"days":                 "210",
						names.AttrStorageClass: "DEEP_ARCHIVE",
					}),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.id", "id2"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.prefix", "path2/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.expiration.0.date", "2016-01-12"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.expiration.0.days", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.expiration.0.expired_object_delete_marker", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.2.id", "id3"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.2.prefix", "path3/"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.2.transition.*", map[string]string{
						"days": acctest.Ct0,
					}),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.3.id", "id4"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.3.prefix", "path4/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.3.tags.tagKey", "tagValue"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.3.tags.terraform", "hashicorp"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.4.id", "id5"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.4.tags.tagKey", "tagValue"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.4.tags.terraform", "hashicorp"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.4.transition.*", map[string]string{
						"days":                 acctest.Ct0,
						names.AttrStorageClass: "GLACIER",
					}),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.5.id", "id6"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.5.tags.tagKey", "tagValue"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.5.transition.*", map[string]string{
						"days":                 acctest.Ct0,
						names.AttrStorageClass: "GLACIER",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "acl"},
			},
			{
				Config: testAccBucketConfig_basic(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccS3Bucket_Manage_lifecycleExpireMarkerOnly(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_lifecycleExpireMarker(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.days", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.date", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.expired_object_delete_marker", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "acl"},
			},
			{
				Config: testAccBucketConfig_basic(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11420
func TestAccS3Bucket_Manage_lifecycleRuleExpirationEmptyBlock(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_lifecycleRuleExpirationEmptyBlock(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/15138
func TestAccS3Bucket_Manage_lifecycleRuleAbortIncompleteMultipartUploadDaysNoExpiration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_lifecycleRuleAbortIncompleteMultipartUploadDays(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccS3Bucket_Manage_lifecycleRemove(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_lifecycle(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.#", "6"),
				),
			},
			{
				Config: testAccBucketConfig_basic(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					// As Lifecycle Rule is a Computed field, removing them from terraform will not
					// trigger an update to remove them from the S3 bucket.
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.#", "6"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Manage_objectLock(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_objectLockEnabledNoDefaultRetention(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.object_lock_enabled", string(types.ObjectLockEnabledEnabled)),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.rule.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccBucketConfig_objectLockEnabledDefaultRetention(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.object_lock_enabled", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.rule.0.default_retention.0.mode", "COMPLIANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.rule.0.default_retention.0.days", acctest.Ct3),
				),
			},
		},
	})
}

func TestAccS3Bucket_Manage_objectLock_deprecatedEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_objectLockEnabledNoDefaultRetentionDeprecatedEnabled(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.object_lock_enabled", string(types.ObjectLockEnabledEnabled)),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.rule.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccS3Bucket_Manage_objectLock_migrate(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_objectLockEnabledNoDefaultRetentionDeprecatedEnabled(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.object_lock_enabled", string(types.ObjectLockEnabledEnabled)),
				),
			},
			{
				Config:   testAccBucketConfig_objectLockEnabledNoDefaultRetention(bucketName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccS3Bucket_Manage_objectLockWithVersioning(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_objectLockEnabledVersioning(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.object_lock_enabled", string(types.ObjectLockEnabledEnabled)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccS3Bucket_Manage_objectLockWithVersioning_deprecatedEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_objectLockEnabledVersioningDeprecatedEnabled(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.object_lock_enabled", string(types.ObjectLockEnabledEnabled)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccS3Bucket_Manage_versioning(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_versioning(bucketName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.mfa_delete", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "acl"},
			},
			{
				Config: testAccBucketConfig_versioning(bucketName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.mfa_delete", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Manage_versioningDisabled(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_versioning(bucketName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.mfa_delete", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Manage_MFADeleteDisabled(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_versioningMFADelete(bucketName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.mfa_delete", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Manage_versioningAndMFADeleteDisabled(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_versioningDisabledAndMFADelete(bucketName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "versioning.0.mfa_delete", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Replication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	alternateRegion := acctest.AlternateRegion()
	region := acctest.Region()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket.source"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_replication(bucketName, string(types.StorageClassStandard)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct1),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
				),
			},
			{
				Config: testAccBucketConfig_replication(bucketName, string(types.StorageClassGlacier)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct1),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
				),
			},
			{
				Config: testAccBucketConfig_replicationSSEKMSEncryptedObjects(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccS3Bucket_Replication_multipleDestinationsEmptyFilter(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	alternateRegion := acctest.AlternateRegion()
	region := acctest.Region()
	resourceName := "aws_s3_bucket.source"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_replicationMultipleDestinationsEmptyFilter(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination2", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination3", acctest.RegionProviderFunc(alternateRegion, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						names.AttrID:                  "rule1",
						names.AttrPriority:            acctest.Ct1,
						names.AttrStatus:              "Enabled",
						"filter.#":                    acctest.Ct1,
						"filter.0.prefix":             "",
						"destination.#":               acctest.Ct1,
						"destination.0.storage_class": "STANDARD",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						names.AttrID:                  "rule2",
						names.AttrPriority:            acctest.Ct2,
						names.AttrStatus:              "Enabled",
						"filter.#":                    acctest.Ct1,
						"filter.0.prefix":             "",
						"destination.#":               acctest.Ct1,
						"destination.0.storage_class": "STANDARD_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						names.AttrID:                  "rule3",
						names.AttrPriority:            acctest.Ct3,
						names.AttrStatus:              "Disabled",
						"filter.#":                    acctest.Ct1,
						"filter.0.prefix":             "",
						"destination.#":               acctest.Ct1,
						"destination.0.storage_class": "ONEZONE_IA",
					}),
				),
			},
			{
				Config:            testAccBucketConfig_replicationMultipleDestinationsEmptyFilter(bucketName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDestroy,
					"acl",
				},
			},
		},
	})
}

func TestAccS3Bucket_Replication_multipleDestinationsNonEmptyFilter(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	alternateRegion := acctest.AlternateRegion()
	region := acctest.Region()
	resourceName := "aws_s3_bucket.source"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_replicationMultipleDestinationsNonEmptyFilter(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination2", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination3", acctest.RegionProviderFunc(alternateRegion, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						names.AttrID:                  "rule1",
						names.AttrPriority:            acctest.Ct1,
						names.AttrStatus:              "Enabled",
						"filter.#":                    acctest.Ct1,
						"filter.0.prefix":             "prefix1",
						"destination.#":               acctest.Ct1,
						"destination.0.storage_class": "STANDARD",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						names.AttrID:                  "rule2",
						names.AttrPriority:            acctest.Ct2,
						names.AttrStatus:              "Enabled",
						"filter.#":                    acctest.Ct1,
						"filter.0.tags.%":             acctest.Ct1,
						"filter.0.tags.Key2":          "Value2",
						"destination.#":               acctest.Ct1,
						"destination.0.storage_class": "STANDARD_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						names.AttrID:                  "rule3",
						names.AttrPriority:            acctest.Ct3,
						names.AttrStatus:              "Disabled",
						"filter.#":                    acctest.Ct1,
						"filter.0.prefix":             "prefix3",
						"filter.0.tags.%":             acctest.Ct1,
						"filter.0.tags.Key3":          "Value3",
						"destination.#":               acctest.Ct1,
						"destination.0.storage_class": "ONEZONE_IA",
					}),
				),
			},
			{
				Config:            testAccBucketConfig_replicationMultipleDestinationsNonEmptyFilter(bucketName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDestroy,
					"acl",
				},
			},
		},
	})
}

func TestAccS3Bucket_Replication_twoDestination(t *testing.T) {
	ctx := acctest.Context(t)

	// This tests 2 destinations since GovCloud and possibly other non-standard partitions allow a max of 2
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	alternateRegion := acctest.AlternateRegion()
	region := acctest.Region()
	resourceName := "aws_s3_bucket.source"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_replicationMultipleDestinationsTwoDestination(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination2", acctest.RegionProviderFunc(alternateRegion, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						names.AttrID:                  "rule1",
						names.AttrPriority:            acctest.Ct1,
						names.AttrStatus:              "Enabled",
						"filter.#":                    acctest.Ct1,
						"filter.0.prefix":             "prefix1",
						"destination.#":               acctest.Ct1,
						"destination.0.storage_class": "STANDARD",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						names.AttrID:                  "rule2",
						names.AttrPriority:            acctest.Ct2,
						names.AttrStatus:              "Enabled",
						"filter.#":                    acctest.Ct1,
						"filter.0.tags.%":             acctest.Ct1,
						"filter.0.tags.Key2":          "Value2",
						"destination.#":               acctest.Ct1,
						"destination.0.storage_class": "STANDARD_IA",
					}),
				),
			},
			{
				Config:            testAccBucketConfig_replicationMultipleDestinationsTwoDestination(bucketName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDestroy,
					"acl",
				},
			},
		},
	})
}

func TestAccS3Bucket_Replication_ruleDestinationAccessControlTranslation(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	region := acctest.Region()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket.source"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_replicationAccessControlTranslation(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct1),
				),
			},
			{
				Config:            testAccBucketConfig_replicationAccessControlTranslation(bucketName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDestroy,
					"acl",
					"versioning",
					"replication_configuration.0.rules.0.priority",
				},
			},
			{
				Config: testAccBucketConfig_replicationSSEKMSEncryptedObjectsAndAccessControlTranslation(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct1),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12480
func TestAccS3Bucket_Replication_ruleDestinationAddAccessControlTranslation(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	region := acctest.Region()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket.source"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_replicationRulesDestination(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct1),
				),
			},
			{
				Config:            testAccBucketConfig_replicationAccessControlTranslation(bucketName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDestroy,
					"acl",
					"versioning",
					"replication_configuration.0.rules.0.priority",
				},
			},
			{
				Config: testAccBucketConfig_replicationAccessControlTranslation(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct1),
				),
			},
		},
	})
}

// StorageClass issue: https://github.com/hashicorp/terraform/issues/10909
func TestAccS3Bucket_Replication_withoutStorageClass(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	alternateRegion := acctest.AlternateRegion()
	region := acctest.Region()
	resourceName := "aws_s3_bucket.source"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_replicationNoStorageClass(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
				),
			},
			{
				Config:            testAccBucketConfig_replicationNoStorageClass(bucketName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDestroy,
					"acl",
					"replication_configuration.0.rules.0.priority",
				},
			},
		},
	})
}

func TestAccS3Bucket_Replication_expectVersioningValidationError(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketConfig_replicationNoVersioning(bucketName),
				ExpectError: regexache.MustCompile(`versioning must be enabled on S3 Bucket \(.*\) to allow replication`),
			},
		},
	})
}

// Prefix issue: https://github.com/hashicorp/terraform-provider-aws/issues/6340
func TestAccS3Bucket_Replication_withoutPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	alternateRegion := acctest.AlternateRegion()
	region := acctest.Region()
	resourceName := "aws_s3_bucket.source"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_replicationNoPrefix(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
				),
			},
			{
				Config:            testAccBucketConfig_replicationNoPrefix(bucketName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDestroy,
					"acl",
					"replication_configuration.0.rules.0.priority",
				},
			},
		},
	})
}

func TestAccS3Bucket_Replication_schemaV2(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	alternateRegion := acctest.AlternateRegion()
	region := acctest.Region()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket.source"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_replicationV2DeleteMarkerReplicationDisabled(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct1),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
				),
			},
			{
				Config: testAccBucketConfig_replicationV2NoTags(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct1),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
				),
			},
			{
				Config:            testAccBucketConfig_replicationV2NoTags(bucketName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDestroy,
					"acl",
					"replication_configuration.0.rules.0.priority",
				},
			},
			{
				Config: testAccBucketConfig_replicationV2OnlyOneTag(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct1),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
				),
			},
			{
				Config: testAccBucketConfig_replicationV2PrefixAndTags(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct1),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
				),
			},
			{
				Config: testAccBucketConfig_replicationV2MultipleTags(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct1),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
				),
			},
		},
	})
}

func TestAccS3Bucket_Replication_schemaV2SameRegion(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket.source"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationResourceName := "aws_s3_bucket.destination"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_replicationV2SameRegionNoTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					acctest.CheckResourceAttrGlobalARN(resourceName, "replication_configuration.0.role", "iam", fmt.Sprintf("role/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct1),
					testAccCheckBucketExists(ctx, destinationResourceName),
				),
			},
			{
				Config:            testAccBucketConfig_replicationV2SameRegionNoTags(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDestroy,
					"acl",
					"replication_configuration.0.rules.0.priority",
				},
			},
		},
	})
}

func TestAccS3Bucket_Replication_RTC_valid(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	alternateRegion := acctest.AlternateRegion()
	region := acctest.Region()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket.source"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_replicationV2RTC(bucketName, 15),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct1),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
				),
			},
			{
				Config: testAccBucketConfig_replicationV2RTCNoMinutes(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct1),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
				),
			},
			{
				Config: testAccBucketConfig_replicationV2RTCNoStatus(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct1),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
				),
			},
			{
				Config: testAccBucketConfig_replicationV2RTCNotConfigured(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExistsWithProvider(ctx, resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.0.destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.0.destination.0.replication_time.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.0.destination.0.metrics.#", acctest.Ct1),
					testAccCheckBucketExistsWithProvider(ctx, "aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
				),
			},
		},
	})
}

func TestAccS3Bucket_Security_corsUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.test"

	modifyBucketCors := func(n string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			rs, ok := s.RootModule().Resources[n]
			if !ok {
				return fmt.Errorf("Not found: %s", n)
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)
			_, err := conn.PutBucketCors(ctx, &s3.PutBucketCorsInput{
				Bucket: aws.String(rs.Primary.ID),
				CORSConfiguration: &types.CORSConfiguration{
					CORSRules: []types.CORSRule{
						{
							AllowedHeaders: []string{"*"},
							AllowedMethods: []string{"GET"},
							AllowedOrigins: []string{"https://www.example.com"},
						},
					},
				},
			})
			if err != nil && !tfawserr.ErrCodeEquals(err, tfs3.ErrCodeNoSuchCORSConfiguration) {
				return err
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_cors(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_headers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_headers.0", "*"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.0", "PUT"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.1", "POST"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_origins.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_origins.0", "https://www.example.com"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.expose_headers.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.expose_headers.0", "x-amz-server-side-encryption"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.expose_headers.1", "ETag"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.max_age_seconds", "3000"),

					modifyBucketCors(resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccBucketConfig_cors(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_headers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_headers.0", "*"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.0", "PUT"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.1", "POST"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_origins.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_origins.0", "https://www.example.com"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.expose_headers.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.expose_headers.0", "x-amz-server-side-encryption"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.expose_headers.1", "ETag"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.max_age_seconds", "3000"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Security_corsDelete(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.test"

	deleteBucketCors := func(n string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			rs, ok := s.RootModule().Resources[n]
			if !ok {
				return fmt.Errorf("Not found: %s", n)
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)
			_, err := conn.DeleteBucketCors(ctx, &s3.DeleteBucketCorsInput{
				Bucket: aws.String(rs.Primary.ID),
			})
			if err != nil && !tfawserr.ErrCodeEquals(err, tfs3.ErrCodeNoSuchCORSConfiguration) {
				return err
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_cors(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					deleteBucketCors(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3Bucket_Security_corsEmptyOrigin(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_corsEmptyOrigin(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_headers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_headers.0", "*"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.0", "PUT"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.1", "POST"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_origins.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_origins.0", ""),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.expose_headers.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.expose_headers.0", "x-amz-server-side-encryption"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.expose_headers.1", "ETag"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.max_age_seconds", "3000"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccS3Bucket_Security_corsSingleMethodAndEmptyOrigin(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_corsSingleMethodAndEmptyOrigin(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccS3Bucket_Security_logging(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_logging(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "logging.0.target_bucket", "aws_s3_bucket.log_bucket", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "logging.0.target_prefix", "log/"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Security_enableDefaultEncryptionWhenTypical(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_defaultEncryptionKMSMasterKey(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", string(types.ServerSideEncryptionAwsKms)),
					resource.TestMatchResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", regexache.MustCompile("^arn")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Security_enableDefaultEncryptionWhenAES256IsUsed(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_defaultEncryptionDefaultKey(bucketName, string(types.ServerSideEncryptionAes256)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", string(types.ServerSideEncryptionAes256)),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Security_disableDefaultEncryptionWhenDefaultEncryptionIsEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_defaultEncryptionDefaultKey(bucketName, string(types.ServerSideEncryptionAwsKms)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "acl"},
			},
			{
				// As ServerSide Encryption Configuration is a Computed field, removing them from terraform will not
				// trigger an update to remove it from the S3 bucket.
				Config: testAccBucketConfig_basic(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccS3Bucket_Web_simple(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	region := acctest.Region()
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_website(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "website.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "website.0.index_document", "index.html"),
					testAccCheckBucketWebsiteEndpoint(resourceName, "website_endpoint", bucketName, region),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "acl", "grant"},
			},
			{
				Config: testAccBucketConfig_websiteAndError(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "website.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "website.0.index_document", "index.html"),
					resource.TestCheckResourceAttr(resourceName, "website.0.error_document", "error.html"),
					testAccCheckBucketWebsiteEndpoint(resourceName, "website_endpoint", bucketName, region),
				),
			},
			{
				// As Website is a Computed field, removing them from terraform will not
				// trigger an update to remove them from the S3 bucket.
				Config: testAccBucketConfig_basic(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "website.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "website.0.index_document", "index.html"),
					resource.TestCheckResourceAttr(resourceName, "website.0.error_document", "error.html"),
					testAccCheckBucketWebsiteEndpoint(resourceName, "website_endpoint", bucketName, region),
				),
			},
		},
	})
}

func TestAccS3Bucket_Web_redirect(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	region := acctest.Region()
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_websiteAndRedirect(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "website.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "website.0.redirect_all_requests_to", "hashicorp.com?my=query"),
					testAccCheckBucketWebsiteEndpoint(resourceName, "website_endpoint", bucketName, region),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "acl", "grant"},
			},
			{
				Config: testAccBucketConfig_websiteAndHTTPSRedirect(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "website.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "website.0.redirect_all_requests_to", "https://hashicorp.com?my=query"),
					testAccCheckBucketWebsiteEndpoint(resourceName, "website_endpoint", bucketName, region),
				),
			},
			{
				// As Website is a Computed field, removing them from terraform will not
				// trigger an update to remove them from the S3 bucket.
				Config: testAccBucketConfig_basic(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "website.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "website.0.redirect_all_requests_to", "https://hashicorp.com?my=query"),
					testAccCheckBucketWebsiteEndpoint(resourceName, "website_endpoint", bucketName, region),
				),
			},
		},
	})
}

func TestAccS3Bucket_Web_routingRules(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	region := acctest.Region()
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_websiteAndRoutingRules(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "website.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "website.0.error_document", "error.html"),
					resource.TestCheckResourceAttr(resourceName, "website.0.index_document", "index.html"),
					resource.TestCheckResourceAttrSet(resourceName, "website.0.routing_rules"),
					testAccCheckBucketWebsiteEndpoint(resourceName, "website_endpoint", bucketName, region),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy, "acl", "grant"},
			},
			{
				// As Website is a Computed field, removing them from terraform will not
				// trigger an update to remove them from the S3 bucket.
				Config: testAccBucketConfig_basic(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "website.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "website.0.error_document", "error.html"),
					resource.TestCheckResourceAttr(resourceName, "website.0.index_document", "index.html"),
					resource.TestCheckResourceAttrSet(resourceName, "website.0.routing_rules"),
					testAccCheckBucketWebsiteEndpoint(resourceName, "website_endpoint", bucketName, region),
				),
			},
		},
	})
}

func TestBucketName(t *testing.T) {
	t.Parallel()

	validDnsNames := []string{
		"foobar",
		"foo.bar",
		"foo.bar.baz",
		"1234",
		"foo-bar",
		strings.Repeat("x", 63),
	}

	for _, v := range validDnsNames {
		if err := tfs3.ValidBucketName(v, names.USWest2RegionID); err != nil {
			t.Fatalf("%q should be a valid S3 bucket name", v)
		}
	}

	invalidDnsNames := []string{
		"foo..bar",
		"Foo.Bar",
		"192.168.0.1",
		"127.0.0.1",
		".foo",
		"bar.",
		"foo_bar",
		strings.Repeat("x", 64),
	}

	for _, v := range invalidDnsNames {
		if err := tfs3.ValidBucketName(v, names.USWest2RegionID); err == nil {
			t.Fatalf("%q should not be a valid S3 bucket name", v)
		}
	}

	validEastNames := []string{
		"foobar",
		"foo_bar",
		"127.0.0.1",
		"foo..bar",
		"foo_bar_baz",
		"foo.bar.baz",
		"Foo.Bar",
		strings.Repeat("x", 255),
	}

	for _, v := range validEastNames {
		if err := tfs3.ValidBucketName(v, names.USEast1RegionID); err != nil {
			t.Fatalf("%q should be a valid S3 bucket name", v)
		}
	}

	invalidEastNames := []string{
		"foo;bar",
		strings.Repeat("x", 256),
	}

	for _, v := range invalidEastNames {
		if err := tfs3.ValidBucketName(v, names.USEast1RegionID); err == nil {
			t.Fatalf("%q should not be a valid S3 bucket name", v)
		}
	}
}

func TestBucketRegionalDomainName(t *testing.T) {
	t.Parallel()

	const bucket = "bucket-name"

	var testCases = []struct {
		ExpectedErrCount int
		ExpectedOutput   string
		Region           string
	}{
		{
			Region:           "",
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + ".s3.amazonaws.com",
		},
		{
			Region:           "custom",
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + ".s3.custom.amazonaws.com",
		},
		{
			Region:           names.USEast1RegionID,
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + fmt.Sprintf(".s3.%s.%s", names.USEast1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			Region:           names.USWest2RegionID,
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + fmt.Sprintf(".s3.%s.%s", names.USWest2RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			Region:           names.USGovWest1RegionID,
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + fmt.Sprintf(".s3.%s.%s", names.USGovWest1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			Region:           names.CNNorth1RegionID,
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + fmt.Sprintf(".s3.%s.amazonaws.com.cn", names.CNNorth1RegionID),
		},
	}

	for _, tc := range testCases {
		output := tfs3.BucketRegionalDomainName(bucket, tc.Region)
		if output != tc.ExpectedOutput {
			t.Fatalf("expected %q, received %q", tc.ExpectedOutput, output)
		}
	}
}

func TestWebsiteEndpoint(t *testing.T) {
	t.Parallel()

	// https://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteEndpoints.html
	testCases := []struct {
		TestingClient      *conns.AWSClient
		LocationConstraint string
		Expected           string
	}{
		{
			LocationConstraint: "",
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", names.USEast1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			LocationConstraint: names.USEast2RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.%s", names.USEast2RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			LocationConstraint: names.USGovEast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.%s", names.USGovEast1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			LocationConstraint: names.USISOEast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.c2s.ic.gov", names.USISOEast1RegionID),
		},
		{
			LocationConstraint: names.USISOBEast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.sc2s.sgov.gov", names.USISOBEast1RegionID),
		},
		{
			LocationConstraint: names.CNNorth1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.amazonaws.com.cn", names.CNNorth1RegionID),
		},
	}

	for _, testCase := range testCases {
		got, _ := tfs3.BucketWebsiteEndpointAndDomain("bucket-name", testCase.LocationConstraint)
		if got != testCase.Expected {
			t.Errorf("BucketWebsiteEndpointAndDomain(\"bucket-name\", %q) => %q, want %q", testCase.LocationConstraint, got, testCase.Expected)
		}
	}
}

func testAccCheckBucketDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error { return testAccCheckBucketDestroyWithProvider(ctx)(s, acctest.Provider) }
}

func testAccCheckBucketDestroyWithProvider(ctx context.Context) acctest.TestCheckWithProviderFunc {
	return func(s *terraform.State, provider *schema.Provider) error {
		conn := provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket" {
				continue
			}

			// S3 seems to be highly eventually consistent. Even if one connection reports that the queue is gone,
			// another connection may still report it as present.
			_, err := tfresource.RetryUntilNotFound(ctx, tfs3.BucketPropagationTimeout, func() (interface{}, error) {
				return nil, tfs3.FindBucket(ctx, conn, rs.Primary.ID)
			})

			if errors.Is(err, tfresource.ErrFoundResource) {
				return fmt.Errorf("S3 Bucket %s still exists", rs.Primary.ID)
			}

			if err != nil {
				return err
			}

			continue
		}

		return nil
	}
}

func testAccCheckBucketExists(ctx context.Context, n string) resource.TestCheckFunc {
	return testAccCheckBucketExistsWithProvider(ctx, n, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckBucketExistsWithProvider(ctx context.Context, n string, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := providerF().Meta().(*conns.AWSClient).S3Client(ctx)

		return tfs3.FindBucket(ctx, conn, rs.Primary.ID)
	}
}

func testAccCheckBucketAddObjects(ctx context.Context, n string, keys ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)
		if tfs3.IsDirectoryBucket(rs.Primary.ID) {
			conn = acctest.Provider.Meta().(*conns.AWSClient).S3ExpressClient(ctx)
		}

		for _, key := range keys {
			_, err := conn.PutObject(ctx, &s3.PutObjectInput{
				Bucket: aws.String(rs.Primary.ID),
				Key:    aws.String(key),
			})

			if err != nil {
				return fmt.Errorf("PutObject error: %w", err)
			}
		}

		return nil
	}
}

func testAccCheckBucketAddObjectsWithLegalHold(ctx context.Context, n string, keys ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, key := range keys {
			_, err := conn.PutObject(ctx, &s3.PutObjectInput{
				Bucket:                    aws.String(rs.Primary.ID),
				ChecksumAlgorithm:         types.ChecksumAlgorithmCrc32,
				Key:                       aws.String(key),
				ObjectLockLegalHoldStatus: types.ObjectLockLegalHoldStatusOn,
			})

			if err != nil {
				return fmt.Errorf("PutObject error: %w", err)
			}
		}

		return nil
	}
}

func testAccCheckBucketAddObjectWithMetadata(ctx context.Context, n string, key string, metadata map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		_, err := conn.PutObject(ctx, &s3.PutObjectInput{
			Bucket:   aws.String(rs.Primary.ID),
			Key:      aws.String(key),
			Metadata: metadata,
		})

		if err != nil {
			return fmt.Errorf("PutObject error: %w", err)
		}

		return nil
	}
}

func testAccCheckBucketDeleteObjects(ctx context.Context, n string, keys ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, key := range keys {
			_, err := conn.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(rs.Primary.ID),
				Key:    aws.String(key),
			})

			if err != nil {
				return fmt.Errorf("DeleteObject error: %w", err)
			}
		}

		return nil
	}
}

// Create an S3 bucket via a CF stack so that it has system tags.
func testAccCheckBucketCreateViaCloudFormation(ctx context.Context, n string, v *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationClient(ctx)
		stackName := sdkacctest.RandomWithPrefix("tf-acc-test-s3tags")
		templateBody := fmt.Sprintf(`{
  "Resources": {
    "TfTestBucket": {
      "Type": "AWS::S3::Bucket",
      "Properties": {
        "BucketName": "%s"
      }
    }
  }
}`, n)

		requestToken := id.UniqueId()
		input := &cloudformation.CreateStackInput{
			ClientRequestToken: aws.String(requestToken),
			StackName:          aws.String(stackName),
			TemplateBody:       aws.String(templateBody),
		}

		output, err := conn.CreateStack(ctx, input)

		if err != nil {
			return fmt.Errorf("creating CloudFormation Stack: %w", err)
		}

		stackID := aws.ToString(output.StackId)
		stack, err := tfcloudformation.WaitStackCreated(ctx, conn, stackID, requestToken, 10*time.Minute)

		if err != nil {
			return fmt.Errorf("waiting for CloudFormation Stack (%s) create: %w", stackID, err)
		}

		if stack.StackStatus != cloudformationtypes.StackStatusCreateComplete {
			return fmt.Errorf("invalid CloudFormation Stack (%s) status: %s", stackID, stack.StackStatus)
		}

		*v = stackID

		return nil
	}
}

func testAccCheckBucketTagKeys(ctx context.Context, n string, keys ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		bucket := rs.Primary.Attributes[names.AttrBucket]
		got, err := tfs3.BucketListTags(ctx, conn, bucket)

		if err != nil {
			return err
		}

		for _, want := range keys {
			ok := false
			for _, key := range got.Keys() {
				if want == key {
					ok = true
					break
				}
			}
			if !ok {
				return fmt.Errorf("key %s not found in S3 Bucket (%s) tag set", bucket, want)
			}
		}

		return nil
	}
}

func testAccCheckBucketDomainName(ctx context.Context, resourceName string, attributeName string, bucketName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		expectedValue := acctest.Provider.Meta().(*conns.AWSClient).PartitionHostname(ctx, fmt.Sprintf("%s.s3", bucketName))

		return resource.TestCheckResourceAttr(resourceName, attributeName, expectedValue)(s)
	}
}

func testAccBucketRegionalDomainName(bucket, region string) string {
	return tfs3.BucketRegionalDomainName(bucket, region)
}

func testAccCheckBucketWebsiteEndpoint(resourceName string, attributeName string, bucketName string, region string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		expectedValue, _ := tfs3.BucketWebsiteEndpointAndDomain(bucketName, region)

		return resource.TestCheckResourceAttr(resourceName, attributeName, expectedValue)(s)
	}
}

func testAccCheckBucketUpdateTags(ctx context.Context, n string, oldTags, newTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		return tfs3.BucketUpdateTags(ctx, conn, rs.Primary.Attributes[names.AttrBucket], oldTags, newTags)
	}
}

func testAccCheckBucketCheckTags(ctx context.Context, n string, expectedTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		got, err := tfs3.BucketListTags(ctx, conn, rs.Primary.Attributes[names.AttrBucket])
		if err != nil {
			return err
		}

		want := tftags.New(ctx, expectedTags)
		if !reflect.DeepEqual(want, got) {
			return fmt.Errorf("Incorrect tags, want: %v got: %v", want, got)
		}

		return nil
	}
}

func testAccBucketConfig_basic(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, bucketName)
}

func testAccBucketConfig_acceleration(bucketName, acceleration string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket              = %[1]q
  acceleration_status = %[2]q
}
`, bucketName, acceleration)
}

func testAccBucketConfig_acl(bucketName, acl string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  acl    = %[2]q
}
`, bucketName, acl)
}

func testAccBucketConfig_cors(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST"]
    allowed_origins = ["https://www.example.com"]
    expose_headers  = ["x-amz-server-side-encryption", "ETag"]
    max_age_seconds = 3000
  }
}
`, bucketName)
}

func testAccBucketConfig_corsSingleMethodAndEmptyOrigin(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  cors_rule {
    allowed_methods = ["GET"]
    allowed_origins = [""]
  }
}
`, bucketName)
}

func testAccBucketConfig_corsEmptyOrigin(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST"]
    allowed_origins = [""]
    expose_headers  = ["x-amz-server-side-encryption", "ETag"]
    max_age_seconds = 3000
  }
}
`, bucketName)
}

func testAccBucketConfig_defaultEncryptionDefaultKey(bucketName, sseAlgorithm string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = %[2]q
      }
    }
  }
}
`, bucketName, sseAlgorithm)
}

func testAccBucketConfig_defaultEncryptionKMSMasterKey(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "KMS Key for Bucket %[1]s"
  deletion_window_in_days = 10
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = aws_kms_key.test.arn
        sse_algorithm     = "aws:kms"
      }
    }
  }
}
`, bucketName)
}

func testAccBucketConfig_defaultEncryptionKeyEnabledKMSMasterKey(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "KMS Key for Bucket %[1]s"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = aws_kms_key.test.arn
        sse_algorithm     = "aws:kms"
      }
      bucket_key_enabled = true
    }
  }
}
`, bucketName)
}

func testAccBucketConfig_lifecycle(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  lifecycle_rule {
    id      = "id1"
    prefix  = "path1/"
    enabled = true

    expiration {
      days = 365
    }

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 60
      storage_class = "INTELLIGENT_TIERING"
    }

    transition {
      days          = 90
      storage_class = "ONEZONE_IA"
    }

    transition {
      days          = 120
      storage_class = "GLACIER"
    }

    transition {
      days          = 210
      storage_class = "DEEP_ARCHIVE"
    }
  }

  lifecycle_rule {
    id      = "id2"
    prefix  = "path2/"
    enabled = true

    expiration {
      date = "2016-01-12"
    }
  }

  lifecycle_rule {
    id      = "id3"
    prefix  = "path3/"
    enabled = true

    transition {
      days          = 0
      storage_class = "GLACIER"
    }
  }

  lifecycle_rule {
    id      = "id4"
    prefix  = "path4/"
    enabled = true

    tags = {
      "tagKey"    = "tagValue"
      "terraform" = "hashicorp"
    }

    expiration {
      date = "2016-01-12"
    }
  }

  lifecycle_rule {
    id      = "id5"
    enabled = true

    tags = {
      "tagKey"    = "tagValue"
      "terraform" = "hashicorp"
    }

    transition {
      days          = 0
      storage_class = "GLACIER"
    }
  }

  lifecycle_rule {
    id      = "id6"
    enabled = true

    tags = {
      "tagKey" = "tagValue"
    }

    transition {
      days          = 0
      storage_class = "GLACIER"
    }
  }
}
`, bucketName)
}

func testAccBucketConfig_lifecycleExpireMarker(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  lifecycle_rule {
    id      = "id1"
    prefix  = "path1/"
    enabled = true

    expiration {
      expired_object_delete_marker = "true"
    }
  }
}
`, bucketName)
}

func testAccBucketConfig_lifecycleRuleExpirationEmptyBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  lifecycle_rule {
    enabled = true
    id      = "id1"

    expiration {}
  }
}
`, rName)
}

func testAccBucketConfig_lifecycleRuleAbortIncompleteMultipartUploadDays(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  lifecycle_rule {
    abort_incomplete_multipart_upload_days = 7
    enabled                                = true
    id                                     = "id1"
  }
}
`, rName)
}

func testAccBucketConfig_logging(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "log_bucket" {
  bucket = "%[1]s-log"
}

resource "aws_s3_bucket_ownership_controls" "log_bucket_ownership" {
  bucket = aws_s3_bucket.log_bucket.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "log_bucket_acl" {
  depends_on = [aws_s3_bucket_ownership_controls.log_bucket_ownership]

  bucket = aws_s3_bucket.log_bucket.id
  acl    = "log-delivery-write"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  logging {
    target_bucket = aws_s3_bucket.log_bucket.id
    target_prefix = "log/"
  }
}
`, bucketName)
}

func testAccBucketConfig_policy(bucketName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "policy" {
  statement {
    effect = "Allow"

    actions = [
      "s3:*",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::%[1]s",
      "arn:${data.aws_partition.current.partition}:s3:::%[1]s/*",
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  policy = data.aws_iam_policy_document.policy.json
}
`, bucketName)
}

func testAccBucketConfig_ReplicationBase(bucketName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "role" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "destination" {
  provider = "awsalternate"
  bucket   = "%[1]s-destination"

  versioning {
    enabled = true
  }
}
`, bucketName))
}

func testAccBucketConfig_replication(bucketName, storageClass string) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = %[2]q
      }
    }
  }
}
`, bucketName, storageClass))
}

func testAccBucketConfig_replicationAccessControlTranslation(bucketName string) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        account_id    = data.aws_caller_identity.current.account_id
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"

        access_control_translation {
          owner = "Destination"
        }
      }
    }
  }
}
`, bucketName))
}

func testAccBucketConfig_replicationMultipleDestinationsEmptyFilter(bucketName string) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "destination2" {
  provider = "awsalternate"
  bucket   = "%[1]s-destination2"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket" "destination3" {
  provider = "awsalternate"
  bucket   = "%[1]s-destination3"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id       = "rule1"
      priority = 1
      status   = "Enabled"

      filter {}

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }

    rules {
      id       = "rule2"
      priority = 2
      status   = "Enabled"

      filter {}

      destination {
        bucket        = aws_s3_bucket.destination2.arn
        storage_class = "STANDARD_IA"
      }
    }

    rules {
      id       = "rule3"
      priority = 3
      status   = "Disabled"

      filter {}

      destination {
        bucket        = aws_s3_bucket.destination3.arn
        storage_class = "ONEZONE_IA"
      }
    }
  }
}
`, bucketName))
}

func testAccBucketConfig_replicationMultipleDestinationsNonEmptyFilter(bucketName string) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "destination2" {
  provider = "awsalternate"
  bucket   = "%[1]s-destination2"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket" "destination3" {
  provider = "awsalternate"
  bucket   = "%[1]s-destination3"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id       = "rule1"
      priority = 1
      status   = "Enabled"

      filter {
        prefix = "prefix1"
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }

    rules {
      id       = "rule2"
      priority = 2
      status   = "Enabled"

      filter {
        tags = {
          Key2 = "Value2"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination2.arn
        storage_class = "STANDARD_IA"
      }
    }

    rules {
      id       = "rule3"
      priority = 3
      status   = "Disabled"

      filter {
        prefix = "prefix3"

        tags = {
          Key3 = "Value3"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination3.arn
        storage_class = "ONEZONE_IA"
      }
    }
  }
}
`, bucketName))
}

func testAccBucketConfig_replicationMultipleDestinationsTwoDestination(bucketName string) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "destination2" {
  provider = "awsalternate"
  bucket   = "%[1]s-destination2"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id       = "rule1"
      priority = 1
      status   = "Enabled"

      filter {
        prefix = "prefix1"
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }

    rules {
      id       = "rule2"
      priority = 2
      status   = "Enabled"

      filter {
        tags = {
          Key2 = "Value2"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination2.arn
        storage_class = "STANDARD_IA"
      }
    }
  }
}
`, bucketName))
}

func testAccBucketConfig_replicationNoVersioning(bucketName string) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "source" {
  bucket = "%[1]s"

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }
}
`, bucketName))
}

func testAccBucketConfig_replicationRulesDestination(bucketName string) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        account_id    = data.aws_caller_identity.current.account_id
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }

  versioning {
    enabled = true
  }
}
`, bucketName))
}

func testAccBucketConfig_replicationSSEKMSEncryptedObjects(bucketName string) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
resource "aws_kms_key" "replica" {
  provider                = "awsalternate"
  description             = "TF Acceptance Test S3 repl KMS key"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        bucket             = aws_s3_bucket.destination.arn
        storage_class      = "STANDARD"
        replica_kms_key_id = aws_kms_key.replica.arn
      }

      source_selection_criteria {
        sse_kms_encrypted_objects {
          enabled = true
        }
      }
    }
  }
}
`, bucketName))
}

func testAccBucketConfig_replicationSSEKMSEncryptedObjectsAndAccessControlTranslation(bucketName string) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_kms_key" "replica" {
  provider                = "awsalternate"
  description             = "TF Acceptance Test S3 repl KMS key"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        account_id         = data.aws_caller_identity.current.account_id
        bucket             = aws_s3_bucket.destination.arn
        storage_class      = "STANDARD"
        replica_kms_key_id = aws_kms_key.replica.arn

        access_control_translation {
          owner = "Destination"
        }
      }

      source_selection_criteria {
        sse_kms_encrypted_objects {
          enabled = true
        }
      }
    }
  }
}
`, bucketName))
}

func testAccBucketConfig_replicationNoPrefix(bucketName string) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }
}
`, bucketName))
}

func testAccBucketConfig_replicationNoStorageClass(bucketName string) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        bucket = aws_s3_bucket.destination.arn
      }
    }
  }
}
`, bucketName))
}

func testAccBucketConfig_replicationV2SameRegionNoTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.test.arn

    rules {
      id     = "testid"
      status = "Enabled"

      filter {
        prefix = "testprefix"
      }

      delete_marker_replication_status = "Enabled"

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }
}

resource "aws_s3_bucket" "destination" {
  bucket = "%[1]s-destination"

  versioning {
    enabled = true
  }
}
`, rName)
}

func testAccBucketConfig_replicationV2DeleteMarkerReplicationDisabled(bucketName string) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      filter {
        prefix = "foo"
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }
}
`, bucketName))
}

func testAccBucketConfig_replicationV2NoTags(bucketName string) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      filter {
        prefix = "foo"
      }

      delete_marker_replication_status = "Enabled"

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }
}
`, bucketName))
}

func testAccBucketConfig_replicationV2OnlyOneTag(bucketName string) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      priority = 42

      filter {
        tags = {
          ReplicateMe = "Yes"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }
}
`, bucketName))
}

func testAccBucketConfig_replicationV2PrefixAndTags(bucketName string) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      priority = 41

      filter {
        prefix = "foo"

        tags = {
          AnotherTag  = "OK"
          ReplicateMe = "Yes"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }
}
`, bucketName))
}

func testAccBucketConfig_replicationV2MultipleTags(bucketName string) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      filter {
        tags = {
          AnotherTag  = "OK"
          Foo         = "Bar"
          ReplicateMe = "Yes"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }
}
`, bucketName))
}

func testAccBucketConfig_replicationV2RTC(bucketName string, minutes int) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn
    rules {
      id     = "rtc"
      status = "Enabled"
      filter {
        tags = {}
      }
      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
        metrics {
          status  = "Enabled"
          minutes = %[2]d
        }
        replication_time {
          status  = "Enabled"
          minutes = %[2]d
        }
      }
    }
  }
}
`, bucketName, minutes))
}

func testAccBucketConfig_replicationV2RTCNoMinutes(bucketName string) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn
    rules {
      id     = "rtc-no-minutes"
      status = "Enabled"
      filter {
        tags = {}
      }
      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
        metrics {}
        replication_time {
          status = "Enabled"
        }
      }
    }
  }
}
`, bucketName))
}

func testAccBucketConfig_replicationV2RTCNoStatus(bucketName string) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn
    rules {
      id     = "rtc-no-status"
      status = "Enabled"
      filter {
        prefix = "foo"
      }
      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
        metrics {}
        replication_time {
          minutes = 15
        }
      }
    }
  }
}
`, bucketName))
}

func testAccBucketConfig_replicationV2RTCNotConfigured(bucketName string) string {
	return acctest.ConfigCompose(
		testAccBucketConfig_ReplicationBase(bucketName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"
  versioning {
    enabled = true
  }
  replication_configuration {
    role = aws_iam_role.role.arn
    rules {
      id     = "rtc-no-config"
      status = "Enabled"
      filter {
        prefix = "foo"
      }
      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
        metrics {}
        replication_time {}
      }
    }
  }
}
`, bucketName))
}

func testAccBucketConfig_requestPayer(bucketName, requestPayer string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  request_payer = %[2]q
}
`, bucketName, requestPayer)
}

func testAccBucketConfig_versioning(bucketName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  versioning {
    enabled = %[2]t
  }
}
`, bucketName, enabled)
}

func testAccBucketConfig_versioningMFADelete(bucketName string, mfaDelete bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  versioning {
    mfa_delete = %[2]t
  }
}
`, bucketName, mfaDelete)
}

func testAccBucketConfig_versioningDisabledAndMFADelete(bucketName string, mfaDelete bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  versioning {
    enabled    = false
    mfa_delete = %[2]t
  }
}
`, bucketName, mfaDelete)
}

func testAccBucketConfig_website(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  website {
    index_document = "index.html"
  }
}
`, bucketName)
}

func testAccBucketConfig_websiteAndError(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  website {
    index_document = "index.html"
    error_document = "error.html"
  }
}
`, bucketName)
}

func testAccBucketConfig_websiteAndRedirect(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  website {
    redirect_all_requests_to = "hashicorp.com?my=query"
  }
}
`, bucketName)
}

func testAccBucketConfig_websiteAndHTTPSRedirect(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  website {
    redirect_all_requests_to = "https://hashicorp.com?my=query"
  }
}
`, bucketName)
}

func testAccBucketConfig_websiteAndRoutingRules(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  website {
    index_document = "index.html"
    error_document = "error.html"

    routing_rules = <<EOF
[
  {
    "Condition": {
      "KeyPrefixEquals": "docs/"
    },
    "Redirect": {
      "ReplaceKeyPrefixWith": "documents/"
    }
  }
]
EOF

  }
}
`, bucketName)
}

func testAccBucketConfig_noTags(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = false
}
`, bucketName)
}

func testAccBucketConfig_tags(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = false

  tags = {
    Key1 = "AAA"
    Key2 = "BBB"
    Key3 = "CCC"
  }
}
`, bucketName)
}

func testAccBucketConfig_updatedTags(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = false

  tags = {
    Key2 = "BBB"
    Key3 = "XXX"
    Key4 = "DDD"
    Key5 = "EEE"
  }
}
`, bucketName)
}

func testAccBucketConfig_objectLockEnabledNoDefaultRetention(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_enabled = true
}
`, bucketName)
}

func testAccBucketConfig_objectLockEnabledDefaultRetention(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_configuration {
    object_lock_enabled = "Enabled"

    rule {
      default_retention {
        mode = "COMPLIANCE"
        days = 3
      }
    }
  }
}
`, bucketName)
}

func testAccBucketConfig_objectLockEnabledNoDefaultRetentionDeprecatedEnabled(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_configuration {
    object_lock_enabled = "Enabled"
  }
}
`, bucketName)
}

func testAccBucketConfig_objectLockEnabledVersioning(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true

  object_lock_enabled = true
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = "Enabled"
  }
}
`, bucketName)
}

func testAccBucketConfig_objectLockEnabledVersioningDeprecatedEnabled(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true

  object_lock_configuration {
    object_lock_enabled = "Enabled"
  }
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = "Enabled"
  }
}
`, bucketName)
}

func testAccBucketConfig_forceDestroy(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}
`, bucketName)
}

func testAccBucketConfig_forceDestroyObjectVersions(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_versioning" "bucket" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = "Enabled"
  }
}
`, bucketName)
}

func testAccBucketConfig_forceDestroyObjectLockEnabled(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true

  object_lock_enabled = true
}

resource "aws_s3_bucket_versioning" "bucket" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = "Enabled"
  }
}
`, bucketName)
}

const testAccBucketConfig_emptyString = `
resource "aws_s3_bucket" "test" {
  bucket = ""
}
`

func testAccBucketConfig_namePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket_prefix = %[1]q
}
`, namePrefix)
}

const testAccBucketConfig_nameGenerated = `
resource "aws_s3_bucket" "test" {}
`

func testAccBucketConfig_duplicate(region, bucketName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigRegionalProvider(region),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  depends_on = [aws_s3_bucket.duplicate]
}

resource "aws_s3_bucket" "duplicate" {
  bucket = %[1]q
}
  `, bucketName),
	)
}

func testAccBucketConfig_duplicateAltAccount(region, bucketName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigRegionalProvider(region),
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  depends_on = [aws_s3_bucket.duplicate]
}

resource "aws_s3_bucket" "duplicate" {
  provider = "awsalternate"
  bucket   = %[1]q
}
  `, bucketName),
	)
}
