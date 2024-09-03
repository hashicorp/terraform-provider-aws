// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontDistribution_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_enabled(false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccCloudFrontDistribution_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_enabled(false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceDistribution(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontDistribution_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
			{
				Config: testAccDistributionConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDistributionConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

// TestAccCloudFrontDistribution_s3Origin runs an
// aws_cloudfront_distribution acceptance test with a single S3 origin.
//
// If you are testing manually and can't wait for deletion, set the
// TF_TEST_CLOUDFRONT_RETAIN environment variable.
func TestAccCloudFrontDistribution_s3Origin(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_s3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, "aws_cloudfront_distribution.s3_distribution", &distribution),
					resource.TestCheckResourceAttr("aws_cloudfront_distribution.s3_distribution", names.AttrHostedZoneID, "Z2FDTNDATAQYW2"),
				),
			},
			{
				ResourceName:      "aws_cloudfront_distribution.s3_distribution",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

// TestAccCloudFrontDistribution_customOrigin tests a single custom origin.
//
// If you are testing manually and can't wait for deletion, set the
// TF_TEST_CLOUDFRONT_RETAIN environment variable.
func TestAccCloudFrontDistribution_customOrigin(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_custom(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, "aws_cloudfront_distribution.custom_distribution", &distribution),
				),
			},
			{
				ResourceName:      "aws_cloudfront_distribution.custom_distribution",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccCloudFrontDistribution_originPolicyDefault(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_originRequestPolicyDefault(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_cloudfront_distribution.custom_distribution", "default_cache_behavior.0.origin_request_policy_id", regexache.MustCompile("[0-9A-z]+")),
				),
			},
			{
				ResourceName:      "aws_cloudfront_distribution.custom_distribution",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccCloudFrontDistribution_originPolicyOrdered(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_originRequestPolicyOrdered(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_cloudfront_distribution.custom_distribution", "ordered_cache_behavior.0.origin_request_policy_id", regexache.MustCompile("[0-9A-Za-z]+")),
				),
			},
			{
				ResourceName:      "aws_cloudfront_distribution.custom_distribution",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

// TestAccCloudFrontDistribution_multiOrigin runs an
// aws_cloudfront_distribution acceptance test with multiple origins.
//
// If you are testing manually and can't wait for deletion, set the
// TF_TEST_CLOUDFRONT_RETAIN environment variable.
func TestAccCloudFrontDistribution_multiOrigin(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.multi_origin_distribution"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_multiOrigin(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.default_ttl", "50"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.path_pattern", "images1/*.jpg"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/188
// TestAccCloudFrontDistribution_orderedCacheBehavior runs an
// aws_cloudfront_distribution acceptance test with multiple and ordered cache behaviors.
//
// If you are testing manually and can't wait for deletion, set the
// TF_TEST_CLOUDFRONT_RETAIN environment variable.
func TestAccCloudFrontDistribution_orderedCacheBehavior(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_orderedCacheBehavior(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.default_ttl", "50"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.path_pattern", "images1/*.jpg"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.1.default_ttl", "51"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.1.path_pattern", "images2/*.jpg"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccCloudFrontDistribution_orderedCacheBehaviorCachePolicy(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.main"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_orderedCacheBehaviorCachePolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.path_pattern", "images2/*.jpg"),
					resource.TestMatchResourceAttr(resourceName, "ordered_cache_behavior.0.cache_policy_id", regexache.MustCompile(`^[0-9a-z]+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccCloudFrontDistribution_orderedCacheBehaviorResponseHeadersPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.main"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_orderedCacheBehaviorResponseHeadersPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.path_pattern", "images2/*.jpg"),
					resource.TestMatchResourceAttr(resourceName, "ordered_cache_behavior.0.response_headers_policy_id", regexache.MustCompile(`^[0-9a-z]+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccCloudFrontDistribution_forwardedValuesToCachePolicy(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_distribution.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_orderedCacheBehavior(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
				),
			},
			{
				Config: testAccDistributionConfig_orderedCacheBehaviorCachePolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_Origin_emptyDomainName(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDistributionConfig_originEmptyDomainName(),
				ExpectError: regexache.MustCompile(`domain_name must not be empty`),
			},
		},
	})
}

func TestAccCloudFrontDistribution_Origin_emptyOriginID(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDistributionConfig_originEmptyOriginID(),
				ExpectError: regexache.MustCompile(`origin.0.origin_id must not be empty`),
			},
		},
	})
}

func TestAccCloudFrontDistribution_Origin_connectionAttempts(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, "cloudfront") },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDistributionConfig_originItem(rName, `connection_attempts = 0`),
				ExpectError: regexache.MustCompile(`expected origin.0.connection_attempts to be in the range`),
			},
			{
				Config:      testAccDistributionConfig_originItem(rName, `connection_attempts = 4`),
				ExpectError: regexache.MustCompile(`expected origin.0.connection_attempts to be in the range`),
			},
			{
				Config: testAccDistributionConfig_originItem(rName, `connection_attempts = 2`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "origin.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "origin.0.connection_attempts", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_Origin_connectionTimeout(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, "cloudfront") },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDistributionConfig_originItem(rName, `connection_timeout = 0`),
				ExpectError: regexache.MustCompile(`expected origin.0.connection_timeout to be in the range`),
			},
			{
				Config:      testAccDistributionConfig_originItem(rName, `connection_timeout = 11`),
				ExpectError: regexache.MustCompile(`expected origin.0.connection_timeout to be in the range`),
			},
			{
				Config: testAccDistributionConfig_originItem(rName, `connection_timeout = 6`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "origin.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "origin.0.connection_timeout", `6`),
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_Origin_originShield(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, "cloudfront") },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDistributionConfig_originItem(rName, originShieldItem(`null`, `data.aws_region.current.name`)),
				ExpectError: regexache.MustCompile(`Missing required argument`),
			},
			{
				Config:      testAccDistributionConfig_originItem(rName, originShieldItem(acctest.CtFalse, `""`)),
				ExpectError: regexache.MustCompile(`.*must be a valid AWS Region Code.*`),
			},
			{
				Config:      testAccDistributionConfig_originItem(rName, originShieldItem(acctest.CtTrue, `"US East (Ohio)"`)),
				ExpectError: regexache.MustCompile(`.*must be a valid AWS Region Code.*`),
			},
			{
				Config: testAccDistributionConfig_originItem(rName, originShieldItem(acctest.CtTrue, `"us-east-1"`)), //lintignore:AWSAT003
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "origin.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "origin.0.origin_shield.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "origin.0.origin_shield.0.origin_shield_region", "us-east-1"), //lintignore:AWSAT003
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_Origin_originAccessControl(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_originAccessControl(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, "aws_cloudfront_distribution.test", &distribution),
					resource.TestCheckResourceAttrPair("aws_cloudfront_distribution.test", "origin.0.origin_access_control_id", "aws_cloudfront_origin_access_control.test.0", names.AttrID),
				),
			},
			{
				ResourceName:      "aws_cloudfront_distribution.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
			{
				Config: testAccDistributionConfig_originAccessControl(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, "aws_cloudfront_distribution.test", &distribution),
					resource.TestCheckResourceAttrPair("aws_cloudfront_distribution.test", "origin.0.origin_access_control_id", "aws_cloudfront_origin_access_control.test.1", names.AttrID),
				),
			},
		},
	})
}

// TestAccCloudFrontDistribution_noOptionalItems runs an
// aws_cloudfront_distribution acceptance test with no optional items set.
//
// If you are testing manually and can't wait for deletion, set the
// TF_TEST_CLOUDFRONT_RETAIN environment variable.
func TestAccCloudFrontDistribution_noOptionalItems(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.no_optional_items"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_noOptionalItems(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", acctest.Ct0),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "cloudfront", regexache.MustCompile(`distribution/[0-9A-Z]+$`)),
					resource.TestCheckResourceAttr(resourceName, "custom_error_response.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.allowed_methods.#", "7"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.cached_methods.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.compress", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.0.forward", "all"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.0.whitelisted_names.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.headers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.query_string", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.query_string_cache_keys.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.lambda_function_association.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.function_association.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.min_ttl", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.smooth_streaming", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.target_origin_id", "myCustomOrigin"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.trusted_key_groups.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.trusted_signers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.viewer_protocol_policy", "allow-all"),
					resource.TestMatchResourceAttr(resourceName, names.AttrDomainName, regexache.MustCompile(`^[0-9a-z]+\.cloudfront\.net$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestMatchResourceAttr(resourceName, "etag", regexache.MustCompile(`^[0-9A-Z]+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrHostedZoneID, "Z2FDTNDATAQYW2"),
					resource.TestCheckResourceAttrSet(resourceName, "http_version"),
					resource.TestCheckResourceAttr(resourceName, "is_ipv6_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "logging_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "origin.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "origin.*", map[string]string{
						"custom_header.#":                                 acctest.Ct0,
						"custom_origin_config.#":                          acctest.Ct1,
						"custom_origin_config.0.http_port":                "80",
						"custom_origin_config.0.https_port":               "443",
						"custom_origin_config.0.origin_keepalive_timeout": "5",
						"custom_origin_config.0.origin_protocol_policy":   "http-only",
						"custom_origin_config.0.origin_read_timeout":      "30",
						"custom_origin_config.0.origin_ssl_protocols.#":   acctest.Ct2,
						names.AttrDomainName:                              "www.example.com",
					}),
					resource.TestCheckResourceAttr(resourceName, "price_class", "PriceClass_All"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.geo_restriction.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.geo_restriction.0.locations.#", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.geo_restriction.0.restriction_type", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "viewer_certificate.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "viewer_certificate.0.cloudfront_default_certificate", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "wait_for_deployment", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

// TestAccCloudFrontDistribution_http11 runs an
// aws_cloudfront_distribution acceptance test with the HTTP version set to
// 1.1.
//
// If you are testing manually and can't wait for deletion, set the
// TF_TEST_CLOUDFRONT_RETAIN environment variable.
func TestAccCloudFrontDistribution_http11(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_http11(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, "aws_cloudfront_distribution.http_1_1", &distribution),
				),
			},
			{
				ResourceName:      "aws_cloudfront_distribution.http_1_1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccCloudFrontDistribution_isIPV6Enabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_isIPV6Enabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, "aws_cloudfront_distribution.is_ipv6_enabled", &distribution),
					resource.TestCheckResourceAttr(
						"aws_cloudfront_distribution.is_ipv6_enabled", "is_ipv6_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      "aws_cloudfront_distribution.is_ipv6_enabled",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccCloudFrontDistribution_noCustomErrorResponse(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_noCustomErroResponseInfo(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, "aws_cloudfront_distribution.no_custom_error_responses", &distribution),
				),
			},
			{
				ResourceName:      "aws_cloudfront_distribution.no_custom_error_responses",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"custom_error_response.0.response_code",
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccCloudFrontDistribution_DefaultCacheBehaviorForwardedValuesCookies_whitelistedNames(t *testing.T) {
	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_defaultCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered3(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.0.whitelisted_names.#", acctest.Ct3),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
			{
				Config: testAccDistributionConfig_defaultCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered2(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.0.whitelisted_names.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_DefaultCacheBehaviorForwardedValues_headers(t *testing.T) {
	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_defaultCacheBehaviorForwardedValuesHeadersUnordered3(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.headers.#", acctest.Ct3),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
			{
				Config: testAccDistributionConfig_defaultCacheBehaviorForwardedValuesHeadersUnordered2(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.headers.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_DefaultCacheBehavior_trustedKeyGroups(t *testing.T) {
	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_defaultCacheBehaviorTrustedKeyGroups(retainOnDelete, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "trusted_key_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trusted_key_groups.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "trusted_key_groups.0.items.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "trusted_key_groups.0.items.0.key_group_id"),
					resource.TestCheckResourceAttr(resourceName, "trusted_key_groups.0.items.0.key_pair_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.trusted_key_groups.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccCloudFrontDistribution_DefaultCacheBehavior_trustedSigners(t *testing.T) {
	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_defaultCacheBehaviorTrustedSignersSelf(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "trusted_signers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trusted_signers.0.items.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "trusted_signers.0.items.0.aws_account_number", "self"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.trusted_signers.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccCloudFrontDistribution_DefaultCacheBehavior_realtimeLogARN(t *testing.T) {
	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_distribution.test"
	realtimeLogConfigResourceName := "aws_cloudfront_realtime_log_config.test"
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_defaultCacheBehaviorRealtimeLogARN(rName, retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "default_cache_behavior.0.realtime_log_config_arn", realtimeLogConfigResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccCloudFrontDistribution_OrderedCacheBehavior_realtimeLogARN(t *testing.T) {
	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_distribution.test"
	realtimeLogConfigResourceName := "aws_cloudfront_realtime_log_config.test"
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_orderedCacheBehaviorRealtimeLogARN(rName, retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "ordered_cache_behavior.0.realtime_log_config_arn", realtimeLogConfigResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccCloudFrontDistribution_enabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_enabled(false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
			{
				Config: testAccDistributionConfig_enabled(true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_OrderedCacheBehaviorForwardedValuesCookies_whitelistedNames(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_orderedCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered3(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.0.cookies.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.0.cookies.0.whitelisted_names.#", acctest.Ct3),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
			{
				Config: testAccDistributionConfig_orderedCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered2(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.0.cookies.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.0.cookies.0.whitelisted_names.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_OrderedCacheBehaviorForwardedValues_headers(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_orderedCacheBehaviorForwardedValuesHeadersUnordered3(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.0.headers.#", acctest.Ct3),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
			{
				Config: testAccDistributionConfig_orderedCacheBehaviorForwardedValuesHeadersUnordered2(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.0.headers.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_ViewerCertificate_acmCertificateARN(t *testing.T) {
	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_viewerCertificateACMCertificateARN(t, retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
				),
			},
			{
				Config:            testAccDistributionConfig_viewerCertificateACMCertificateARN(t, retainOnDelete),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7773
func TestAccCloudFrontDistribution_ViewerCertificateACMCertificateARN_conflictsWithCloudFrontDefaultCertificate(t *testing.T) {
	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_viewerCertificateACMCertificateARNConflictsDefaultCertificate(t, retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
				),
			},
			{
				Config:            testAccDistributionConfig_viewerCertificateACMCertificateARNConflictsDefaultCertificate(t, retainOnDelete),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccCloudFrontDistribution_waitForDeployment(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_waitForDeployment(false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					testAccCheckDistributionStatusInProgress(&distribution),
					testAccCheckDistributionWaitForDeployment(ctx, &distribution),
					resource.TestCheckResourceAttr(resourceName, "wait_for_deployment", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					names.AttrStatus,
					"wait_for_deployment",
				},
			},
			{
				Config: testAccDistributionConfig_waitForDeployment(true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					testAccCheckDistributionStatusInProgress(&distribution),
					resource.TestCheckResourceAttr(resourceName, "wait_for_deployment", acctest.CtFalse),
				),
			},
			{
				Config: testAccDistributionConfig_waitForDeployment(false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					testAccCheckDistributionStatusDeployed(&distribution),
					resource.TestCheckResourceAttr(resourceName, "wait_for_deployment", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_preconditionFailed(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_distribution.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_eTagInitial(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.#", acctest.Ct1),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.0.items.#", acctest.Ct1),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.0.items.0", "test"),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "Some comment"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
			{
				Config: testAccDistributionConfig_eTagUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.#", acctest.Ct1),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.0.items.#", acctest.Ct2),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.0.items.0", "test"),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.0.items.1", "updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "Some comment"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"etag",
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
			{
				Config: testAccDistributionConfig_eTagFinal(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.#", acctest.Ct1),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.0.items.#", acctest.Ct2),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.0.items.0", "test"),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.0.items.1", "updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "Updated comment"),
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_originGroups(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_distribution.failover_distribution"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_originGroups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "origin_group.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "origin_group.*", map[string]string{
						"origin_id":                          "groupS3",
						"failover_criteria.#":                acctest.Ct1,
						"failover_criteria.0.status_codes.#": acctest.Ct4,
						"member.#":                           acctest.Ct2,
						"member.0.origin_id":                 "primaryS3",
						"member.1.origin_id":                 "failoverS3",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "origin_group.*.failover_criteria.0.status_codes.*", "403"),
					resource.TestCheckTypeSetElemAttr(resourceName, "origin_group.*.failover_criteria.0.status_codes.*", "404"),
					resource.TestCheckTypeSetElemAttr(resourceName, "origin_group.*.failover_criteria.0.status_codes.*", "500"),
					resource.TestCheckTypeSetElemAttr(resourceName, "origin_group.*.failover_criteria.0.status_codes.*", "502"),
				),
			},
		},
	})
}

func testAccCheckDistributionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_distribution" {
				continue
			}

			output, err := tfcloudfront.FindDistributionByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if !testAccDistributionRetainOnDeleteFromEnv() {
				return fmt.Errorf("CloudFront Distribution (%s) still exists", rs.Primary.ID)
			}

			if aws.ToBool(output.Distribution.DistributionConfig.Enabled) {
				return fmt.Errorf("CloudFront Distribution (%s) not disabled", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckDistributionExists(ctx context.Context, n string, v *awstypes.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		output, err := tfcloudfront.FindDistributionByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output.Distribution

		return nil
	}
}

func testAccCheckDistributionStatusDeployed(distribution *awstypes.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if distribution == nil {
			return fmt.Errorf("CloudFront Distribution empty")
		}

		if aws.ToString(distribution.Status) != "Deployed" {
			return fmt.Errorf("CloudFront Distribution (%s) status not Deployed: %s", aws.ToString(distribution.Id), aws.ToString(distribution.Status))
		}

		return nil
	}
}

func testAccCheckDistributionStatusInProgress(distribution *awstypes.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if distribution == nil {
			return fmt.Errorf("CloudFront Distribution empty")
		}

		if aws.ToString(distribution.Status) != "InProgress" {
			return fmt.Errorf("CloudFront Distribution (%s) status not InProgress: %s", aws.ToString(distribution.Id), aws.ToString(distribution.Status))
		}

		return nil
	}
}

func testAccCheckDistributionWaitForDeployment(ctx context.Context, distribution *awstypes.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := tfcloudfront.WaitDistributionDeployed(ctx, acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx), aws.ToString(distribution.Id))
		return err
	}
}

func testAccDistributionRetainOnDeleteFromEnv() bool {
	_, ok := os.LookupEnv("TF_TEST_CLOUDFRONT_RETAIN")
	return ok
}

func testAccDistributionRetainConfig() string {
	if testAccDistributionRetainOnDeleteFromEnv() {
		return "retain_on_delete = true"
	}
	return ""
}

// testAccRegionProviderConfig is the Terraform provider configuration for CloudFront region testing
//
// Testing CloudFront assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccRegionProviderConfig() string {
	switch acctest.Partition() {
	case names.StandardPartitionID:
		return acctest.ConfigRegionalProvider(names.USEast1RegionID)
	case names.ChinaPartitionID:
		return acctest.ConfigRegionalProvider(names.CNNorthwest1RegionID)
	default:
		return acctest.ConfigRegionalProvider(acctest.Region())
	}
}

func originBucket(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "s3_bucket_origin" {
  bucket = "%[1]s.origin-bucket"
}

resource "aws_s3_bucket_public_access_block" "s3_bucket_origin" {
  bucket = aws_s3_bucket.s3_bucket_origin.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_ownership_controls" "s3_bucket_origin" {
  bucket = aws_s3_bucket.s3_bucket_origin.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "s3_bucket_origin_acl" {
  depends_on = [
    aws_s3_bucket_public_access_block.s3_bucket_origin,
    aws_s3_bucket_ownership_controls.s3_bucket_origin,
  ]

  bucket = aws_s3_bucket.s3_bucket_origin.id
  acl    = "public-read"
}
`, rName)
}

func backupBucket(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "s3_backup_bucket_origin" {
  bucket = "%[1]s.backup-bucket"
}

resource "aws_s3_bucket_public_access_block" "s3_backup_bucket_origin" {
  bucket = aws_s3_bucket.s3_backup_bucket_origin.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_ownership_controls" "s3_backup_bucket_origin" {
  bucket = aws_s3_bucket.s3_backup_bucket_origin.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "s3_backup_bucket_origin_acl" {
  depends_on = [
    aws_s3_bucket_public_access_block.s3_backup_bucket_origin,
    aws_s3_bucket_ownership_controls.s3_backup_bucket_origin,
  ]

  bucket = aws_s3_bucket.s3_backup_bucket_origin.id
  acl    = "public-read"
}
`, rName)
}

func logBucket(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "s3_bucket_logs" {
  bucket        = "%[1]s.log-bucket"
  force_destroy = true
}

resource "aws_s3_bucket_public_access_block" "s3_bucket_logs" {
  bucket = aws_s3_bucket.s3_bucket_logs.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_ownership_controls" "s3_bucket_logs" {
  bucket = aws_s3_bucket.s3_bucket_logs.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}


resource "aws_s3_bucket_acl" "s3_bucket_logs_acl" {
  depends_on = [
    aws_s3_bucket_public_access_block.s3_bucket_logs,
    aws_s3_bucket_ownership_controls.s3_bucket_logs,
  ]

  bucket = aws_s3_bucket.s3_bucket_logs.id
  acl    = "public-read"
}
`, rName)
}

func testAccDistributionConfig_s3(rName string) string {
	return acctest.ConfigCompose(
		originBucket(rName),
		logBucket(rName),
		fmt.Sprintf(`
resource "aws_cloudfront_distribution" "s3_distribution" {
  depends_on = [
    aws_s3_bucket_acl.s3_bucket_origin_acl,
    aws_s3_bucket_acl.s3_bucket_logs_acl,
  ]

  origin {
    domain_name = aws_s3_bucket.s3_bucket_origin.bucket_regional_domain_name
    origin_id   = "myS3Origin"
  }

  enabled             = true
  default_root_object = "index.html"

  logging_config {
    include_cookies = false
    bucket          = aws_s3_bucket.s3_bucket_logs.bucket_regional_domain_name
    prefix          = "myprefix"
  }

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myS3Origin"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "allow-all"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  price_class = "PriceClass_200"

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[1]s
}
`, testAccDistributionRetainConfig()))
}

func testAccDistributionConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  enabled          = false
  retain_on_delete = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccDistributionConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  enabled          = false
  retain_on_delete = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDistributionConfig_custom(rName string) string {
	return acctest.ConfigCompose(
		logBucket(rName),
		fmt.Sprintf(`
resource "aws_cloudfront_distribution" "custom_distribution" {
  depends_on = [aws_s3_bucket_acl.s3_bucket_logs_acl]

  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port                = 80
      https_port               = 443
      origin_protocol_policy   = "http-only"
      origin_ssl_protocols     = ["SSLv3", "TLSv1"]
      origin_read_timeout      = 30
      origin_keepalive_timeout = 5
    }
  }

  enabled             = true
  comment             = "Some comment"
  default_root_object = "index.html"

  logging_config {
    include_cookies = false
    bucket          = aws_s3_bucket.s3_bucket_logs.bucket_regional_domain_name
    prefix          = "myprefix"
  }

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "allow-all"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  price_class = "PriceClass_200"

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[1]s
}
`, testAccDistributionRetainConfig()))
}

func testAccDistributionConfig_originRequestPolicyDefault(rName string) string {
	return acctest.ConfigCompose(
		logBucket(rName),
		fmt.Sprintf(`
resource "aws_cloudfront_cache_policy" "example" {
  name        = "test-policy-%[1]s"
  comment     = "test comment"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1
  parameters_in_cache_key_and_forwarded_to_origin {
    cookies_config {
      cookie_behavior = "whitelist"
      cookies {
        items = ["test"]
      }
    }
    headers_config {
      header_behavior = "whitelist"
      headers {
        items = ["test"]
      }
    }
    query_strings_config {
      query_string_behavior = "whitelist"
      query_strings {
        items = ["test"]
      }
    }
  }
}

resource "aws_cloudfront_response_headers_policy" "example" {
  name    = "test-policy-%[1]s"
  comment = "test comment"

  cors_config {
    access_control_allow_credentials = true

    access_control_allow_headers {
      items = ["test"]
    }

    access_control_allow_methods {
      items = ["GET"]
    }

    access_control_allow_origins {
      items = ["test.example.comtest"]
    }

    origin_override = true
  }
}

resource "aws_cloudfront_origin_request_policy" "test_policy" {
  name    = "test-policy-%[1]s"
  comment = "test comment"
  cookies_config {
    cookie_behavior = "whitelist"
    cookies {
      items = ["test"]
    }
  }
  headers_config {
    header_behavior = "whitelist"
    headers {
      items = ["test"]
    }
  }
  query_strings_config {
    query_string_behavior = "whitelist"
    query_strings {
      items = ["test"]
    }
  }
}

resource "aws_cloudfront_distribution" "custom_distribution" {
  depends_on = [aws_s3_bucket_acl.s3_bucket_logs_acl]

  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port                = 80
      https_port               = 443
      origin_protocol_policy   = "http-only"
      origin_ssl_protocols     = ["SSLv3", "TLSv1"]
      origin_read_timeout      = 30
      origin_keepalive_timeout = 5
    }
  }

  enabled             = true
  comment             = "Some comment"
  default_root_object = "index.html"

  logging_config {
    include_cookies = false
    bucket          = aws_s3_bucket.s3_bucket_logs.bucket_regional_domain_name
    prefix          = "myprefix"
  }

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    origin_request_policy_id   = aws_cloudfront_origin_request_policy.test_policy.id
    cache_policy_id            = aws_cloudfront_cache_policy.example.id
    response_headers_policy_id = aws_cloudfront_response_headers_policy.example.id

    viewer_protocol_policy = "allow-all"
  }

  price_class = "PriceClass_200"

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[2]s
}
`, rName, testAccDistributionRetainConfig()))
}

func testAccDistributionConfig_originRequestPolicyOrdered(rName string) string {
	return acctest.ConfigCompose(
		logBucket(rName),
		fmt.Sprintf(`
resource "aws_cloudfront_cache_policy" "example" {
  name        = "test-policy-%[1]s"
  comment     = "test comment"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1
  parameters_in_cache_key_and_forwarded_to_origin {
    cookies_config {
      cookie_behavior = "whitelist"
      cookies {
        items = ["test"]
      }
    }
    headers_config {
      header_behavior = "whitelist"
      headers {
        items = ["test"]
      }
    }
    query_strings_config {
      query_string_behavior = "whitelist"
      query_strings {
        items = ["test"]
      }
    }
  }
}

resource "aws_cloudfront_response_headers_policy" "example" {
  name    = "test-policy-%[1]s"
  comment = "test comment"

  cors_config {
    access_control_allow_credentials = true

    access_control_allow_headers {
      items = ["test"]
    }

    access_control_allow_methods {
      items = ["GET"]
    }

    access_control_allow_origins {
      items = ["test.example.comtest"]
    }

    origin_override = true
  }
}

resource "aws_cloudfront_origin_request_policy" "test_policy" {
  name    = "test-policy-%[1]s"
  comment = "test comment"
  cookies_config {
    cookie_behavior = "whitelist"
    cookies {
      items = ["test"]
    }
  }
  headers_config {
    header_behavior = "whitelist"
    headers {
      items = ["test"]
    }
  }
  query_strings_config {
    query_string_behavior = "whitelist"
    query_strings {
      items = ["test"]
    }
  }
}

resource "aws_cloudfront_distribution" "custom_distribution" {
  depends_on = [aws_s3_bucket_acl.s3_bucket_logs_acl]

  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port                = 80
      https_port               = 443
      origin_protocol_policy   = "http-only"
      origin_ssl_protocols     = ["SSLv3", "TLSv1"]
      origin_read_timeout      = 30
      origin_keepalive_timeout = 5
    }
  }

  enabled             = true
  comment             = "Some comment"
  default_root_object = "index.html"

  logging_config {
    include_cookies = false
    bucket          = aws_s3_bucket.s3_bucket_logs.bucket_regional_domain_name
    prefix          = "myprefix"
  }

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    origin_request_policy_id   = aws_cloudfront_origin_request_policy.test_policy.id
    cache_policy_id            = aws_cloudfront_cache_policy.example.id
    response_headers_policy_id = aws_cloudfront_response_headers_policy.example.id

    viewer_protocol_policy = "allow-all"
  }

  ordered_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false
    path_pattern     = "/*"

    origin_request_policy_id   = aws_cloudfront_origin_request_policy.test_policy.id
    cache_policy_id            = aws_cloudfront_cache_policy.example.id
    response_headers_policy_id = aws_cloudfront_response_headers_policy.example.id

    viewer_protocol_policy = "allow-all"
  }

  price_class = "PriceClass_200"

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[2]s
}
`, rName, testAccDistributionRetainConfig()))
}

func testAccDistributionConfig_multiOrigin(rName string) string {
	return acctest.ConfigCompose(
		originBucket(rName),
		logBucket(rName),
		fmt.Sprintf(`
resource "aws_cloudfront_distribution" "multi_origin_distribution" {
  depends_on = [
    aws_s3_bucket_acl.s3_bucket_origin_acl,
    aws_s3_bucket_acl.s3_bucket_logs_acl,
  ]

  origin {
    domain_name = aws_s3_bucket.s3_bucket_origin.bucket_regional_domain_name
    origin_id   = "myS3Origin"
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port                = 80
      https_port               = 443
      origin_protocol_policy   = "http-only"
      origin_ssl_protocols     = ["SSLv3", "TLSv1"]
      origin_keepalive_timeout = 45
    }
  }

  enabled             = true
  comment             = "Some comment"
  default_root_object = "index.html"

  logging_config {
    include_cookies = false
    bucket          = aws_s3_bucket.s3_bucket_logs.bucket_regional_domain_name
    prefix          = "myprefix"
  }

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myS3Origin"
    smooth_streaming = true

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    min_ttl                = 100
    default_ttl            = 100
    max_ttl                = 100
    viewer_protocol_policy = "allow-all"
  }

  ordered_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myS3Origin"

    forwarded_values {
      query_string = true

      cookies {
        forward = "none"
      }
    }

    min_ttl                = 50
    default_ttl            = 50
    max_ttl                = 50
    viewer_protocol_policy = "allow-all"
    path_pattern           = "images1/*.jpg"
  }

  ordered_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"

    forwarded_values {
      query_string = true

      cookies {
        forward = "none"
      }
    }

    min_ttl                = 50
    default_ttl            = 50
    max_ttl                = 50
    viewer_protocol_policy = "allow-all"
    path_pattern           = "images2/*.jpg"
  }

  price_class = "PriceClass_All"

  custom_error_response {
    error_code            = 404
    response_page_path    = "/error-pages/404.html"
    response_code         = 200
    error_caching_min_ttl = 30
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[1]s
}
`, testAccDistributionRetainConfig()))
}

func testAccDistributionConfig_noCustomErroResponseInfo() string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "no_custom_error_responses" {
  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["SSLv3", "TLSv1"]
    }
  }

  enabled = true
  comment = "Some comment"

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "allow-all"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  custom_error_response {
    error_code            = 404
    error_caching_min_ttl = 30
  }

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[1]s
}
`, testAccDistributionRetainConfig())
}

func testAccDistributionConfig_noOptionalItems() string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "no_optional_items" {
  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["SSLv3", "TLSv1"]
    }
  }

  enabled = true

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "allow-all"
  }

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[1]s
}
`, testAccDistributionRetainConfig())
}

func testAccDistributionConfig_originEmptyDomainName() string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "Origin_EmptyDomainName" {
  origin {
    domain_name = ""
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["SSLv3", "TLSv1"]
    }
  }

  enabled = true

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "allow-all"
  }

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[1]s
}
`, testAccDistributionRetainConfig())
}

func testAccDistributionConfig_originEmptyOriginID() string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "Origin_EmptyOriginID" {
  origin {
    domain_name = "www.example.com"
    origin_id   = ""

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["SSLv3", "TLSv1"]
    }
  }

  enabled = true

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "allow-all"
  }

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[1]s
}
`, testAccDistributionRetainConfig())
}

func testAccDistributionConfig_http11() string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "http_1_1" {
  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["SSLv3", "TLSv1"]
    }
  }

  enabled = true
  comment = "Some comment"

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "allow-all"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  http_version = "http1.1"

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[1]s
}
`, testAccDistributionRetainConfig())
}

func testAccDistributionConfig_isIPV6Enabled() string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "is_ipv6_enabled" {
  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["SSLv3", "TLSv1"]
    }
  }

  enabled         = true
  is_ipv6_enabled = true
  comment         = "Some comment"

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "allow-all"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  http_version = "http1.1"

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[1]s
}
`, testAccDistributionRetainConfig())
}

func testAccDistributionConfig_orderedCacheBehavior() string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "main" {
  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["SSLv3", "TLSv1"]
    }
  }

  enabled = true
  comment = "Some comment"

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = true

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    min_ttl                = 100
    default_ttl            = 100
    max_ttl                = 100
    viewer_protocol_policy = "allow-all"
  }

  ordered_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"

    forwarded_values {
      query_string = true

      cookies {
        forward = "none"
      }
    }

    min_ttl                = 50
    default_ttl            = 50
    max_ttl                = 50
    viewer_protocol_policy = "allow-all"
    path_pattern           = "images1/*.jpg"
  }

  ordered_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"

    forwarded_values {
      query_string = true

      cookies {
        forward = "none"
      }
    }

    min_ttl                = 51
    default_ttl            = 51
    max_ttl                = 51
    viewer_protocol_policy = "allow-all"
    path_pattern           = "images2/*.jpg"
  }

  price_class = "PriceClass_All"

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[1]s
}
`, testAccDistributionRetainConfig())
}

func testAccDistributionConfig_orderedCacheBehaviorCachePolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "main" {
  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["SSLv3", "TLSv1"]
    }
  }

  enabled = true
  comment = "Some comment"

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = true

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "allow-all"
  }

  ordered_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"

    cache_policy_id = aws_cloudfront_cache_policy.cache_policy.id

    viewer_protocol_policy = "allow-all"
    path_pattern           = "images2/*.jpg"
  }

  price_class = "PriceClass_All"

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[2]s
}

resource "aws_cloudfront_cache_policy" "cache_policy" {
  name        = "test-policy-%[1]s"
  comment     = "test comment"
  default_ttl = 50
  max_ttl     = 100
  parameters_in_cache_key_and_forwarded_to_origin {
    cookies_config {
      cookie_behavior = "none"
    }
    headers_config {
      header_behavior = "none"
    }
    query_strings_config {
      query_string_behavior = "none"
    }
  }
}
`, rName, testAccDistributionRetainConfig())
}

func testAccDistributionConfig_orderedCacheBehaviorResponseHeadersPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "main" {
  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["SSLv3", "TLSv1"]
    }
  }

  enabled = true
  comment = "Some comment"

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = true

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "allow-all"
  }

  ordered_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    cache_policy_id  = aws_cloudfront_cache_policy.cache_policy.id

    response_headers_policy_id = aws_cloudfront_response_headers_policy.response_headers_policy.id

    viewer_protocol_policy = "allow-all"
    path_pattern           = "images2/*.jpg"
  }

  price_class = "PriceClass_All"

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[2]s
}

resource "aws_cloudfront_cache_policy" "cache_policy" {
  name        = "test-policy-%[1]s"
  comment     = "test comment"
  default_ttl = 50
  max_ttl     = 100
  parameters_in_cache_key_and_forwarded_to_origin {
    cookies_config {
      cookie_behavior = "none"
    }
    headers_config {
      header_behavior = "none"
    }
    query_strings_config {
      query_string_behavior = "none"
    }
  }
}

resource "aws_cloudfront_response_headers_policy" "response_headers_policy" {
  name    = "test-policy-%[1]s"
  comment = "test comment"

  cors_config {
    access_control_allow_credentials = true

    access_control_allow_headers {
      items = ["test"]
    }

    access_control_allow_methods {
      items = ["GET"]
    }

    access_control_allow_origins {
      items = ["test.example.comtest"]
    }

    origin_override = true
  }
}
`, rName, testAccDistributionRetainConfig())
}

func testAccDistributionConfig_originGroups(rName string) string {
	return acctest.ConfigCompose(
		originBucket(rName),
		backupBucket(rName),
		fmt.Sprintf(`
resource "aws_cloudfront_distribution" "failover_distribution" {
  depends_on = [
    aws_s3_bucket_acl.s3_bucket_origin_acl,
    aws_s3_bucket_acl.s3_backup_bucket_origin_acl,
  ]

  origin {
    domain_name = aws_s3_bucket.s3_bucket_origin.bucket_regional_domain_name
    origin_id   = "primaryS3"
  }

  origin {
    domain_name = aws_s3_bucket.s3_backup_bucket_origin.bucket_regional_domain_name
    origin_id   = "failoverS3"
  }

  origin_group {
    origin_id = "groupS3"

    failover_criteria {
      status_codes = [403, 404, 500, 502]
    }

    member {
      origin_id = "primaryS3"
    }

    member {
      origin_id = "failoverS3"
    }
  }

  enabled = true

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  default_cache_behavior {
    allowed_methods  = ["GET", "HEAD"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "groupS3"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "allow-all"
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
  %[1]s
}
`, testAccDistributionRetainConfig()))
}

func testAccDistributionConfig_defaultCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered2(retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward           = "whitelist"
        whitelisted_names = ["test2", "test1"]
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete)
}

func testAccDistributionConfig_defaultCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered3(retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward           = "whitelist"
        whitelisted_names = ["test2", "test3", "test1"]
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete)
}

func testAccDistributionConfig_defaultCacheBehaviorForwardedValuesHeadersUnordered2(retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      headers      = ["Origin", "Access-Control-Request-Headers"]
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete)
}

func testAccDistributionConfig_defaultCacheBehaviorForwardedValuesHeadersUnordered3(retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      headers      = ["Origin", "Access-Control-Request-Headers", "Access-Control-Request-Method"]
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete)
}

func testAccDistributionConfig_enabled(enabled, retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  enabled          = %[1]t
  retain_on_delete = %[2]t

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, enabled, retainOnDelete)
}

func testAccDistributionConfig_orderedCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered2(retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  ordered_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    path_pattern           = "/test/*"
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward           = "whitelist"
        whitelisted_names = ["test2", "test1"]
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete)
}

func testAccDistributionConfig_orderedCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered3(retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  ordered_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    path_pattern           = "/test/*"
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward           = "whitelist"
        whitelisted_names = ["test2", "test3", "test1"]
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete)
}

func testAccDistributionConfig_orderedCacheBehaviorForwardedValuesHeadersUnordered2(retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  ordered_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    path_pattern           = "/test/*"
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      headers      = ["Origin", "Access-Control-Request-Headers"]
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete)
}

func testAccDistributionConfig_orderedCacheBehaviorForwardedValuesHeadersUnordered3(retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  ordered_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    path_pattern           = "/test/*"
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      headers      = ["Origin", "Access-Control-Request-Headers", "Access-Control-Request-Method"]
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete)
}

func testAccDistributionConfig_defaultCacheBehaviorTrustedKeyGroups(retainOnDelete bool, rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    trusted_key_groups     = [aws_cloudfront_key_group.test.id]
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}

resource "aws_cloudfront_public_key" "test" {
  comment     = "test key"
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
  name        = %[2]q
}

resource "aws_cloudfront_key_group" "test" {
  comment = "test key group"
  items   = [aws_cloudfront_public_key.test.id]
  name    = %[2]q
}
`, retainOnDelete, rName)
}

func testAccDistributionConfig_defaultCacheBehaviorTrustedSignersSelf(retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    trusted_signers        = ["self"]
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete)
}

// CloudFront Distribution ACM Certificates must be created in us-east-1
func testAccDistributionViewerCertificateACMCertificateARNBaseConfig(t *testing.T, commonName string) string {
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, commonName)

	return acctest.ConfigCompose(testAccRegionProviderConfig(), fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}
`, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccDistributionConfig_viewerCertificateACMCertificateARN(t *testing.T, retainOnDelete bool) string {
	return acctest.ConfigCompose(testAccDistributionViewerCertificateACMCertificateARNBaseConfig(t, "example.com"), fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  enabled          = false
  retain_on_delete = %[1]t

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn = aws_acm_certificate.test.arn
    ssl_support_method  = "sni-only"
  }
}
`, retainOnDelete))
}

func testAccDistributionConfig_viewerCertificateACMCertificateARNConflictsDefaultCertificate(t *testing.T, retainOnDelete bool) string {
	return acctest.ConfigCompose(testAccDistributionViewerCertificateACMCertificateARNBaseConfig(t, "example.com"), fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  enabled          = false
  retain_on_delete = %[1]t

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn            = aws_acm_certificate.test.arn
    cloudfront_default_certificate = false
    ssl_support_method             = "sni-only"
  }
}
`, retainOnDelete))
}

func testAccDistributionConfig_waitForDeployment(enabled, waitForDeployment bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  enabled             = %[1]t
  wait_for_deployment = %[2]t

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, enabled, waitForDeployment)
}

func testAccDistributionCacheBehaviorRealtimeLogBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 2
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "cloudfront.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "kinesis:DescribeStreamSummary",
      "kinesis:DescribeStream",
      "kinesis:PutRecord",
      "kinesis:PutRecords"
    ],
    "Resource": "${aws_kinesis_stream.test.arn}"
  }]
}
EOF
}

resource "aws_cloudfront_realtime_log_config" "test" {
  name          = %[1]q
  sampling_rate = 50
  fields        = ["timestamp", "c-ip"]

  endpoint {
    stream_type = "Kinesis"

    kinesis_stream_config {
      role_arn   = aws_iam_role.test.arn
      stream_arn = aws_kinesis_stream.test.arn
    }
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName)
}

func testAccDistributionConfig_defaultCacheBehaviorRealtimeLogARN(rName string, retainOnDelete bool) string {
	return acctest.ConfigCompose(
		testAccDistributionCacheBehaviorRealtimeLogBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods         = ["GET", "HEAD"]
    cached_methods          = ["GET", "HEAD"]
    target_origin_id        = "test"
    viewer_protocol_policy  = "allow-all"
    realtime_log_config_arn = aws_cloudfront_realtime_log_config.test.arn

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete))
}

func testAccDistributionConfig_orderedCacheBehaviorRealtimeLogARN(rName string, retainOnDelete bool) string {
	return acctest.ConfigCompose(
		testAccDistributionCacheBehaviorRealtimeLogBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  ordered_cache_behavior {
    allowed_methods         = ["GET", "HEAD"]
    cached_methods          = ["GET", "HEAD"]
    path_pattern            = "/test/*"
    target_origin_id        = "test"
    viewer_protocol_policy  = "allow-all"
    realtime_log_config_arn = aws_cloudfront_realtime_log_config.test.arn

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete))
}

func originShieldItem(enabled, region string) string {
	return fmt.Sprintf(`
origin_shield {
  enabled              = %[1]s
  origin_shield_region = %[2]s
}
`, enabled, region)
}

func testAccDistributionConfig_originItem(rName string, item string) string {
	return acctest.ConfigCompose(
		originBucket(rName),
		testAccDistributionCacheBehaviorRealtimeLogBaseConfig(rName),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_cloudfront_distribution" "test" {
  depends_on = [aws_s3_bucket_acl.s3_bucket_origin_acl]

  origin {
    domain_name = aws_s3_bucket.s3_bucket_origin.bucket_regional_domain_name
    origin_id   = "myOrigin"
    %[1]s
  }
  enabled = true
  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myOrigin"
    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }
    viewer_protocol_policy = "allow-all"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }
  price_class = "PriceClass_200"
  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }
  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, item))
}

func testAccDistributionConfig_eTagInitial(rName string) string {
	return acctest.ConfigCompose(
		logBucket(rName),
		fmt.Sprintf(`
resource "aws_cloudfront_cache_policy" "example" {
  name        = "test-policy-%[1]s"
  comment     = "test comment"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1
  parameters_in_cache_key_and_forwarded_to_origin {
    cookies_config {
      cookie_behavior = "whitelist"
      cookies {
        items = ["test"]
      }
    }
    headers_config {
      header_behavior = "whitelist"
      headers {
        items = ["test"]
      }
    }
    query_strings_config {
      query_string_behavior = "whitelist"
      query_strings {
        items = ["test"]
      }
    }
  }
}

resource "aws_cloudfront_response_headers_policy" "example" {
  name    = "test-policy-%[1]s"
  comment = "test comment"

  cors_config {
    access_control_allow_credentials = true

    access_control_allow_headers {
      items = ["test"]
    }

    access_control_allow_methods {
      items = ["GET"]
    }

    access_control_allow_origins {
      items = ["test.example.comtest"]
    }

    origin_override = true
  }
}

resource "aws_cloudfront_origin_request_policy" "test_policy" {
  name    = "test-policy-%[1]s"
  comment = "test comment"
  cookies_config {
    cookie_behavior = "whitelist"
    cookies {
      items = ["test"]
    }
  }
  headers_config {
    header_behavior = "whitelist"
    headers {
      items = ["test"]
    }
  }
  query_strings_config {
    query_string_behavior = "whitelist"
    query_strings {
      items = ["test"]
    }
  }
}

resource "aws_cloudfront_distribution" "main" {
  depends_on = [aws_s3_bucket_acl.s3_bucket_logs_acl]

  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port                = 80
      https_port               = 443
      origin_protocol_policy   = "http-only"
      origin_ssl_protocols     = ["SSLv3", "TLSv1"]
      origin_read_timeout      = 30
      origin_keepalive_timeout = 5
    }
  }

  enabled             = true
  comment             = "Some comment"
  default_root_object = "index.html"

  logging_config {
    include_cookies = false
    bucket          = aws_s3_bucket.s3_bucket_logs.bucket_regional_domain_name
    prefix          = "myprefix"
  }

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    origin_request_policy_id   = aws_cloudfront_origin_request_policy.test_policy.id
    cache_policy_id            = aws_cloudfront_cache_policy.example.id
    response_headers_policy_id = aws_cloudfront_response_headers_policy.example.id

    viewer_protocol_policy = "allow-all"
  }

  price_class = "PriceClass_200"

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[2]s
}
`, rName, testAccDistributionRetainConfig()))
}

func testAccDistributionConfig_eTagUpdated(rName string) string {
	return acctest.ConfigCompose(
		logBucket(rName),
		fmt.Sprintf(`
resource "aws_cloudfront_cache_policy" "example" {
  name        = "test-policy-%[1]s"
  comment     = "test comment"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1
  parameters_in_cache_key_and_forwarded_to_origin {
    cookies_config {
      cookie_behavior = "whitelist"
      cookies {
        items = ["test"]
      }
    }
    headers_config {
      header_behavior = "whitelist"
      headers {
        items = ["test"]
      }
    }
    query_strings_config {
      query_string_behavior = "whitelist"
      query_strings {
        items = ["test"]
      }
    }
  }
}

resource "aws_cloudfront_response_headers_policy" "example" {
  name    = "test-policy-%[1]s"
  comment = "test comment"

  cors_config {
    access_control_allow_credentials = true

    access_control_allow_headers {
      items = ["test", "updated"]
    }

    access_control_allow_methods {
      items = ["GET"]
    }

    access_control_allow_origins {
      items = ["test.example.comtest"]
    }

    origin_override = true
  }
}

resource "aws_cloudfront_origin_request_policy" "test_policy" {
  name    = "test-policy-%[1]s"
  comment = "test comment"
  cookies_config {
    cookie_behavior = "whitelist"
    cookies {
      items = ["test"]
    }
  }
  headers_config {
    header_behavior = "whitelist"
    headers {
      items = ["test"]
    }
  }
  query_strings_config {
    query_string_behavior = "whitelist"
    query_strings {
      items = ["test"]
    }
  }
}

resource "aws_cloudfront_distribution" "main" {
  depends_on = [aws_s3_bucket_acl.s3_bucket_logs_acl]

  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port                = 80
      https_port               = 443
      origin_protocol_policy   = "http-only"
      origin_ssl_protocols     = ["SSLv3", "TLSv1"]
      origin_read_timeout      = 30
      origin_keepalive_timeout = 5
    }
  }

  enabled             = true
  comment             = "Some comment"
  default_root_object = "index.html"

  logging_config {
    include_cookies = false
    bucket          = aws_s3_bucket.s3_bucket_logs.bucket_regional_domain_name
    prefix          = "myprefix"
  }

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    origin_request_policy_id   = aws_cloudfront_origin_request_policy.test_policy.id
    cache_policy_id            = aws_cloudfront_cache_policy.example.id
    response_headers_policy_id = aws_cloudfront_response_headers_policy.example.id

    viewer_protocol_policy = "allow-all"
  }

  price_class = "PriceClass_200"

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[2]s
}
`, rName, testAccDistributionRetainConfig()))
}

func testAccDistributionConfig_eTagFinal(rName string) string {
	return acctest.ConfigCompose(
		logBucket(rName),
		fmt.Sprintf(`
resource "aws_cloudfront_cache_policy" "example" {
  name        = "test-policy-%[1]s"
  comment     = "test comment"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1
  parameters_in_cache_key_and_forwarded_to_origin {
    cookies_config {
      cookie_behavior = "whitelist"
      cookies {
        items = ["test"]
      }
    }
    headers_config {
      header_behavior = "whitelist"
      headers {
        items = ["test"]
      }
    }
    query_strings_config {
      query_string_behavior = "whitelist"
      query_strings {
        items = ["test"]
      }
    }
  }
}

resource "aws_cloudfront_response_headers_policy" "example" {
  name    = "test-policy-%[1]s"
  comment = "test comment"

  cors_config {
    access_control_allow_credentials = true

    access_control_allow_headers {
      items = ["test", "updated"]
    }

    access_control_allow_methods {
      items = ["GET"]
    }

    access_control_allow_origins {
      items = ["test.example.comtest"]
    }

    origin_override = true
  }
}

resource "aws_cloudfront_origin_request_policy" "test_policy" {
  name    = "test-policy-%[1]s"
  comment = "test comment"
  cookies_config {
    cookie_behavior = "whitelist"
    cookies {
      items = ["test"]
    }
  }
  headers_config {
    header_behavior = "whitelist"
    headers {
      items = ["test"]
    }
  }
  query_strings_config {
    query_string_behavior = "whitelist"
    query_strings {
      items = ["test"]
    }
  }
}

resource "aws_cloudfront_distribution" "main" {
  depends_on = [aws_s3_bucket_acl.s3_bucket_logs_acl]

  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port                = 80
      https_port               = 443
      origin_protocol_policy   = "http-only"
      origin_ssl_protocols     = ["SSLv3", "TLSv1"]
      origin_read_timeout      = 30
      origin_keepalive_timeout = 5
    }
  }

  enabled             = true
  comment             = "Updated comment"
  default_root_object = "index.html"

  logging_config {
    include_cookies = false
    bucket          = aws_s3_bucket.s3_bucket_logs.bucket_regional_domain_name
    prefix          = "myprefix"
  }

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    origin_request_policy_id   = aws_cloudfront_origin_request_policy.test_policy.id
    cache_policy_id            = aws_cloudfront_cache_policy.example.id
    response_headers_policy_id = aws_cloudfront_response_headers_policy.example.id

    viewer_protocol_policy = "allow-all"
  }

  price_class = "PriceClass_200"

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[2]s
}
`, rName, testAccDistributionRetainConfig()))
}

func testAccDistributionConfig_originAccessControl(rName string, which int) string {
	return acctest.ConfigCompose(
		originBucket(rName),
		logBucket(rName),
		fmt.Sprintf(`
locals {
  rName = %[1]q
}

resource "aws_cloudfront_origin_access_control" "test" {
  count = 2

  name                              = "${local.rName}-${count.index}"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

resource "aws_cloudfront_distribution" "test" {
  depends_on = [
    aws_s3_bucket_acl.s3_bucket_origin_acl,
    aws_s3_bucket_acl.s3_bucket_logs_acl,
  ]

  origin {
    domain_name = aws_s3_bucket.s3_bucket_origin.bucket_regional_domain_name
    origin_id   = "myS3Origin"

    origin_access_control_id = aws_cloudfront_origin_access_control.test[%[3]d].id
  }

  enabled             = true
  default_root_object = "index.html"

  logging_config {
    include_cookies = false
    bucket          = aws_s3_bucket.s3_bucket_logs.bucket_regional_domain_name
    prefix          = "myprefix"
  }

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myS3Origin"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "allow-all"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  price_class = "PriceClass_200"

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[2]s
}
`, rName, testAccDistributionRetainConfig(), which))
}
