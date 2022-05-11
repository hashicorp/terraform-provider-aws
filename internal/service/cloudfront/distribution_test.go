package cloudfront_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCloudFrontDistribution_disappears(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionEnabledConfig(false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					testAccCheckDistributionDisappears(&distribution),
				),
				ExpectNonEmptyPlan: true,
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionS3Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists("aws_cloudfront_distribution.s3_distribution", &distribution),
					resource.TestCheckResourceAttr("aws_cloudfront_distribution.s3_distribution", "hosted_zone_id", "Z2FDTNDATAQYW2"),
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

func TestAccCloudFrontDistribution_s3OriginWithTags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionS3WithTagsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists("aws_cloudfront_distribution.s3_distribution", &distribution),
					resource.TestCheckResourceAttr("aws_cloudfront_distribution.s3_distribution", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_cloudfront_distribution.s3_distribution", "tags.environment", "production"),
					resource.TestCheckResourceAttr("aws_cloudfront_distribution.s3_distribution", "tags.account", "main"),
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
			{
				Config: testAccDistributionS3WithTagsUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists("aws_cloudfront_distribution.s3_distribution", &distribution),
					resource.TestCheckResourceAttr("aws_cloudfront_distribution.s3_distribution", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_cloudfront_distribution.s3_distribution", "tags.environment", "dev"),
				),
			},
		},
	})
}

// TestAccCloudFrontDistribution_customOrigin tests a single custom origin.
//
// If you are testing manually and can't wait for deletion, set the
// TF_TEST_CLOUDFRONT_RETAIN environment variable.
func TestAccCloudFrontDistribution_customOrigin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionCustomConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists("aws_cloudfront_distribution.custom_distribution", &distribution),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionOriginRequestPolicyDefaultConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_cloudfront_distribution.custom_distribution", "default_cache_behavior.0.origin_request_policy_id", regexp.MustCompile("[A-z0-9]+")),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionOriginRequestPolicyOrderedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_cloudfront_distribution.custom_distribution", "ordered_cache_behavior.0.origin_request_policy_id", regexp.MustCompile("[A-z0-9]+")),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.multi_origin_distribution"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionMultiOriginConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", "2"),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionOrderedCacheBehavior(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", "2"),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.main"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionOrderedCacheBehaviorCachePolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.path_pattern", "images2/*.jpg"),
					resource.TestMatchResourceAttr(resourceName, "ordered_cache_behavior.0.cache_policy_id", regexp.MustCompile(`^[a-z0-9]+`)),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.main"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderedCacheBehaviorResponseHeadersPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.path_pattern", "images2/*.jpg"),
					resource.TestMatchResourceAttr(resourceName, "ordered_cache_behavior.0.response_headers_policy_id", regexp.MustCompile(`^[a-z0-9]+`)),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_distribution.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionOrderedCacheBehavior(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
				),
			},
			{
				Config: testAccDistributionOrderedCacheBehaviorCachePolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_Origin_emptyDomainName(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccDistributionConfig_Origin_EmptyDomainName(),
				ExpectError: regexp.MustCompile(`domain_name must not be empty`),
			},
		},
	})
}

func TestAccCloudFrontDistribution_Origin_emptyOriginID(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccDistributionConfig_Origin_EmptyOriginID(),
				ExpectError: regexp.MustCompile(`origin.0.origin_id must not be empty`),
			},
		},
	})
}

func TestAccCloudFrontDistribution_Origin_connectionAttempts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService("cloudfront", t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccDistributionOriginItem(rName, `connection_attempts = 0`),
				ExpectError: regexp.MustCompile(`expected origin.0.connection_attempts to be in the range`),
			},
			{
				Config:      testAccDistributionOriginItem(rName, `connection_attempts = 4`),
				ExpectError: regexp.MustCompile(`expected origin.0.connection_attempts to be in the range`),
			},
			{
				Config: testAccDistributionOriginItem(rName, `connection_attempts = 2`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "origin.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.connection_attempts", `2`),
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_Origin_connectionTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService("cloudfront", t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccDistributionOriginItem(rName, `connection_timeout = 0`),
				ExpectError: regexp.MustCompile(`expected origin.0.connection_timeout to be in the range`),
			},
			{
				Config:      testAccDistributionOriginItem(rName, `connection_timeout = 11`),
				ExpectError: regexp.MustCompile(`expected origin.0.connection_timeout to be in the range`),
			},
			{
				Config: testAccDistributionOriginItem(rName, `connection_timeout = 6`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "origin.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.connection_timeout", `6`),
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_Origin_originShield(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService("cloudfront", t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccDistributionOriginItem(rName, originShieldItem(`null`, `data.aws_region.current.name`)),
				ExpectError: regexp.MustCompile(`Missing required argument`),
			},
			{
				Config:      testAccDistributionOriginItem(rName, originShieldItem(`false`, `null`)),
				ExpectError: regexp.MustCompile(`Missing required argument`),
			},
			{
				Config:      testAccDistributionOriginItem(rName, originShieldItem(`true`, `null`)),
				ExpectError: regexp.MustCompile(`Missing required argument`),
			},
			{
				Config:      testAccDistributionOriginItem(rName, originShieldItem(`false`, `""`)),
				ExpectError: regexp.MustCompile(`.*must be a valid AWS Region Code.*`),
			},
			{
				Config:      testAccDistributionOriginItem(rName, originShieldItem(`true`, `"US East (Ohio)"`)),
				ExpectError: regexp.MustCompile(`.*must be a valid AWS Region Code.*`),
			},
			{
				Config: testAccDistributionOriginItem(rName, originShieldItem(`true`, `"us-east-1"`)), //lintignore:AWSAT003
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "origin.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.origin_shield.0.enabled", `true`),
					resource.TestCheckResourceAttr(resourceName, "origin.0.origin_shield.0.origin_shield_region", "us-east-1"), //lintignore:AWSAT003
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.no_optional_items"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionNoOptionalItemsConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", "0"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "cloudfront", regexp.MustCompile(`distribution/[A-Z0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, "custom_error_response.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.allowed_methods.#", "7"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.cached_methods.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.compress", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.0.forward", "all"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.0.whitelisted_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.headers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.query_string", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.query_string_cache_keys.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.lambda_function_association.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.function_association.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.min_ttl", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.smooth_streaming", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.target_origin_id", "myCustomOrigin"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.trusted_key_groups.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.trusted_signers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.viewer_protocol_policy", "allow-all"),
					resource.TestMatchResourceAttr(resourceName, "domain_name", regexp.MustCompile(`^[a-z0-9]+\.cloudfront\.net$`)),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestMatchResourceAttr(resourceName, "etag", regexp.MustCompile(`^[A-Z0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, "hosted_zone_id", "Z2FDTNDATAQYW2"),
					resource.TestCheckResourceAttrSet(resourceName, "http_version"),
					resource.TestCheckResourceAttr(resourceName, "is_ipv6_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "origin.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "origin.*", map[string]string{
						"custom_header.#":                                 "0",
						"custom_origin_config.#":                          "1",
						"custom_origin_config.0.http_port":                "80",
						"custom_origin_config.0.https_port":               "443",
						"custom_origin_config.0.origin_keepalive_timeout": "5",
						"custom_origin_config.0.origin_protocol_policy":   "http-only",
						"custom_origin_config.0.origin_read_timeout":      "30",
						"custom_origin_config.0.origin_ssl_protocols.#":   "2",
						"domain_name": "www.example.com",
					}),
					resource.TestCheckResourceAttr(resourceName, "price_class", "PriceClass_All"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.geo_restriction.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.geo_restriction.0.locations.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.geo_restriction.0.restriction_type", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "viewer_certificate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "viewer_certificate.0.cloudfront_default_certificate", "true"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_deployment", "true"),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionHTTP11Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists("aws_cloudfront_distribution.http_1_1", &distribution),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionIsIPV6EnabledConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists("aws_cloudfront_distribution.is_ipv6_enabled", &distribution),
					resource.TestCheckResourceAttr(
						"aws_cloudfront_distribution.is_ipv6_enabled", "is_ipv6_enabled", "true"),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionNoCustomErroResponseInfo(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists("aws_cloudfront_distribution.no_custom_error_responses", &distribution),
				),
			},
			{
				ResourceName:      "aws_cloudfront_distribution.no_custom_error_responses",
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

func TestAccCloudFrontDistribution_DefaultCacheBehaviorForwardedValuesCookies_whitelistedNames(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionDefaultCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered3Config(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.0.whitelisted_names.#", "3"),
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
				Config: testAccDistributionDefaultCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered2Config(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.0.whitelisted_names.#", "2"),
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_DefaultCacheBehaviorForwardedValues_headers(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionDefaultCacheBehaviorForwardedValuesHeadersUnordered3Config(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.headers.#", "3"),
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
				Config: testAccDistributionDefaultCacheBehaviorForwardedValuesHeadersUnordered2Config(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.headers.#", "2"),
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_DefaultCacheBehavior_trustedKeyGroups(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionDefaultCacheBehaviorTrustedKeyGroupsConfig(retainOnDelete, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "trusted_key_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "trusted_key_groups.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "trusted_key_groups.0.items.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "trusted_key_groups.0.items.0.key_group_id"),
					resource.TestCheckResourceAttr(resourceName, "trusted_key_groups.0.items.0.key_pair_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.trusted_key_groups.#", "1"),
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
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionDefaultCacheBehaviorTrustedSignersSelfConfig(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "trusted_signers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "trusted_signers.0.items.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "trusted_signers.0.items.0.aws_account_number", "self"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.trusted_signers.#", "1"),
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
	var distribution cloudfront.Distribution
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_distribution.test"
	realtimeLogConfigResourceName := "aws_cloudfront_realtime_log_config.test"
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionDefaultCacheBehaviorRealtimeLogARNConfig(rName, retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_cache_behavior.0.realtime_log_config_arn", realtimeLogConfigResourceName, "arn"),
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
	var distribution cloudfront.Distribution
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_distribution.test"
	realtimeLogConfigResourceName := "aws_cloudfront_realtime_log_config.test"
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionOrderedCacheBehaviorRealtimeLogARNConfig(rName, retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "ordered_cache_behavior.0.realtime_log_config_arn", realtimeLogConfigResourceName, "arn"),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionEnabledConfig(false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
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
				Config: testAccDistributionEnabledConfig(true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
		},
	})
}

// TestAccCloudFrontDistribution_retainOnDelete verifies retain_on_delete = true
// This acceptance test performs the following steps:
//  * Trigger a Terraform destroy of the resource, which should only disable the distribution
//  * Check it still exists and is disabled outside Terraform
//  * Destroy for real outside Terraform
func TestAccCloudFrontDistribution_retainOnDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionEnabledConfig(true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
				),
			},
			{
				Config:  testAccDistributionEnabledConfig(true, true),
				Destroy: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExistsAPIOnly(&distribution),
					testAccCheckDistributionWaitForDeployment(&distribution),
					testAccCheckDistributionDisabled(&distribution),
					testAccCheckDistributionDisappears(&distribution),
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_OrderedCacheBehaviorForwardedValuesCookies_whitelistedNames(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionOrderedCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered3Config(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.0.cookies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.0.cookies.0.whitelisted_names.#", "3"),
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
				Config: testAccDistributionOrderedCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered2Config(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.0.cookies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.0.cookies.0.whitelisted_names.#", "2"),
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_OrderedCacheBehaviorForwardedValues_headers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionOrderedCacheBehaviorForwardedValuesHeadersUnordered3Config(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.0.headers.#", "3"),
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
				Config: testAccDistributionOrderedCacheBehaviorForwardedValuesHeadersUnordered2Config(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.0.headers.#", "2"),
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_ViewerCertificate_acmCertificateARN(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionViewerCertificateACMCertificateARNConfig(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
				),
			},
			{
				Config:            testAccDistributionViewerCertificateACMCertificateARNConfig(retainOnDelete),
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
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionViewerCertificateACMCertificateARNConflictsWithDefaultCertificateConfig(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
				),
			},
			{
				Config:            testAccDistributionViewerCertificateACMCertificateARNConflictsWithDefaultCertificateConfig(retainOnDelete),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionWaitForDeploymentConfig(false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					testAccCheckDistributionStatusInProgress(&distribution),
					testAccCheckDistributionWaitForDeployment(&distribution),
					resource.TestCheckResourceAttr(resourceName, "wait_for_deployment", "false"),
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
				Config: testAccDistributionWaitForDeploymentConfig(true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					testAccCheckDistributionStatusInProgress(&distribution),
					resource.TestCheckResourceAttr(resourceName, "wait_for_deployment", "false"),
				),
			},
			{
				Config: testAccDistributionWaitForDeploymentConfig(false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					testAccCheckDistributionStatusDeployed(&distribution),
					resource.TestCheckResourceAttr(resourceName, "wait_for_deployment", "true"),
				),
			},
		},
	})
}

func TestAccCloudFrontDistribution_preconditionFailed(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_distribution.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionETagInitialConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.#", "1"),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.#", "1"),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.0.items.#", "1"),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.0.items.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "comment", "Some comment"),
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
				Config: testAccDistributionETagUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.#", "1"),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.#", "1"),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.0.items.#", "2"),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.0.items.0", "test"),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.0.items.1", "updated"),
					resource.TestCheckResourceAttr(resourceName, "comment", "Some comment"),
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
				Config: testAccDistributionETagFinalConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.#", "1"),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.#", "1"),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.0.items.#", "2"),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.0.items.0", "test"),
					resource.TestCheckResourceAttr("aws_cloudfront_response_headers_policy.example", "cors_config.0.access_control_allow_headers.0.items.1", "updated"),
					resource.TestCheckResourceAttr(resourceName, "comment", "Updated comment"),
				),
			},
		},
	})
}

func testAccCheckDistributionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_distribution" {
			continue
		}

		input := &cloudfront.GetDistributionInput{
			Id: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetDistribution(input)

		if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchDistribution) {
			continue
		}

		if err != nil {
			return err
		}

		if !testAccDistributionRetainOnDeleteFromEnv() {
			return fmt.Errorf("CloudFront Distribution (%s) still exists", rs.Primary.ID)
		}

		if output != nil && output.Distribution != nil && output.Distribution.DistributionConfig != nil && aws.BoolValue(output.Distribution.DistributionConfig.Enabled) {
			return fmt.Errorf("CloudFront Distribution (%s) not disabled", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckDistributionExists(resourceName string, distribution *cloudfront.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource ID not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

		input := &cloudfront.GetDistributionInput{
			Id: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetDistribution(input)

		if err != nil {
			return fmt.Errorf("Error retrieving CloudFront distribution: %s", err)
		}

		*distribution = *output.Distribution

		return nil
	}
}

func testAccCheckDistributionExistsAPIOnly(distribution *cloudfront.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

		input := &cloudfront.GetDistributionInput{
			Id: distribution.Id,
		}

		output, err := conn.GetDistribution(input)

		if err != nil {
			return err
		}

		*distribution = *output.Distribution

		return nil
	}
}

func testAccCheckDistributionStatusDeployed(distribution *cloudfront.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if distribution == nil {
			return fmt.Errorf("CloudFront Distribution empty")
		}

		if aws.StringValue(distribution.Status) != "Deployed" {
			return fmt.Errorf("CloudFront Distribution (%s) status not Deployed: %s", aws.StringValue(distribution.Id), aws.StringValue(distribution.Status))
		}

		return nil
	}
}

func testAccCheckDistributionStatusInProgress(distribution *cloudfront.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if distribution == nil {
			return fmt.Errorf("CloudFront Distribution empty")
		}

		if aws.StringValue(distribution.Status) != "InProgress" {
			return fmt.Errorf("CloudFront Distribution (%s) status not InProgress: %s", aws.StringValue(distribution.Id), aws.StringValue(distribution.Status))
		}

		return nil
	}
}

func testAccCheckDistributionDisabled(distribution *cloudfront.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if distribution == nil || distribution.DistributionConfig == nil {
			return fmt.Errorf("CloudFront Distribution configuration empty")
		}

		if aws.BoolValue(distribution.DistributionConfig.Enabled) {
			return fmt.Errorf("CloudFront Distribution (%s) enabled", aws.StringValue(distribution.Id))
		}

		return nil
	}
}

// testAccCheckDistributionDisappears deletes a CloudFront Distribution outside Terraform
// This requires the CloudFront Distribution to previously be disabled and fetches latest ETag automatically.
func testAccCheckDistributionDisappears(distribution *cloudfront.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

		getDistributionInput := &cloudfront.GetDistributionInput{
			Id: distribution.Id,
		}

		getDistributionOutput, err := conn.GetDistribution(getDistributionInput)

		if err != nil {
			return err
		}

		deleteDistributionInput := &cloudfront.DeleteDistributionInput{
			Id:      distribution.Id,
			IfMatch: getDistributionOutput.ETag,
		}

		err = resource.Retry(2*time.Minute, func() *resource.RetryError {
			_, err = conn.DeleteDistribution(deleteDistributionInput)

			if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeDistributionNotDisabled) {
				return resource.RetryableError(err)
			}

			if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchDistribution) {
				return nil
			}

			if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodePreconditionFailed) {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.DeleteDistribution(deleteDistributionInput)
		}

		return err
	}
}

func testAccCheckDistributionWaitForDeployment(distribution *cloudfront.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return tfcloudfront.DistributionWaitUntilDeployed(aws.StringValue(distribution.Id), acctest.Provider.Meta())
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

func TestAccCloudFrontDistribution_originGroups(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.failover_distribution"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionOriginGroupsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "origin_group.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "origin_group.*", map[string]string{
						"origin_id":                          "groupS3",
						"failover_criteria.#":                "1",
						"failover_criteria.0.status_codes.#": "4",
						"member.#":                           "2",
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

func originBucket(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "s3_bucket_origin" {
  bucket = "%[1]s.origin-bucket"
}

resource "aws_s3_bucket_acl" "s3_bucket_origin_acl" {
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

resource "aws_s3_bucket_acl" "s3_backup_bucket_origin_acl" {
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

resource "aws_s3_bucket_acl" "s3_bucket_logs_acl" {
  bucket = aws_s3_bucket.s3_bucket_logs.id
  acl    = "public-read"
}
`, rName)
}

func testAccDistributionS3Config(rName string) string {
	return acctest.ConfigCompose(
		originBucket(rName),
		logBucket(rName),
		fmt.Sprintf(`
resource "aws_cloudfront_distribution" "s3_distribution" {
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

func testAccDistributionS3WithTagsConfig(rName string) string {
	return acctest.ConfigCompose(
		originBucket(rName),
		logBucket(rName),
		fmt.Sprintf(`
resource "aws_cloudfront_distribution" "s3_distribution" {
  origin {
    domain_name = aws_s3_bucket.s3_bucket_origin.bucket_regional_domain_name
    origin_id   = "myS3Origin"
  }

  enabled             = true
  default_root_object = "index.html"

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

  tags = {
    environment = "production"
    account     = "main"
  }

  %[1]s
}
`, testAccDistributionRetainConfig()))
}

func testAccDistributionS3WithTagsUpdatedConfig(rName string) string {
	return acctest.ConfigCompose(
		originBucket(rName),
		logBucket(rName),
		fmt.Sprintf(`
resource "aws_cloudfront_distribution" "s3_distribution" {
  origin {
    domain_name = aws_s3_bucket.s3_bucket_origin.bucket_regional_domain_name
    origin_id   = "myS3Origin"
  }

  enabled             = true
  default_root_object = "index.html"

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

  tags = {
    environment = "dev"
  }

  %[1]s
}
`, testAccDistributionRetainConfig()))
}

func testAccDistributionCustomConfig(rName string) string {
	return acctest.ConfigCompose(
		logBucket(rName),
		fmt.Sprintf(`
resource "aws_cloudfront_distribution" "custom_distribution" {
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

func testAccDistributionOriginRequestPolicyDefaultConfig(rName string) string {
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

func testAccDistributionOriginRequestPolicyOrderedConfig(rName string) string {
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

func testAccDistributionMultiOriginConfig(rName string) string {
	return acctest.ConfigCompose(
		originBucket(rName),
		logBucket(rName),
		fmt.Sprintf(`
resource "aws_cloudfront_distribution" "multi_origin_distribution" {
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

func testAccDistributionNoCustomErroResponseInfo() string {
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

func testAccDistributionNoOptionalItemsConfig() string {
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

func testAccDistributionConfig_Origin_EmptyDomainName() string {
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

func testAccDistributionConfig_Origin_EmptyOriginID() string {
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

func testAccDistributionHTTP11Config() string {
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

func testAccDistributionIsIPV6EnabledConfig() string {
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

func testAccDistributionOrderedCacheBehavior() string {
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

func testAccDistributionOrderedCacheBehaviorCachePolicy(rName string) string {
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

func testAccOrderedCacheBehaviorResponseHeadersPolicy(rName string) string {
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

func testAccDistributionOriginGroupsConfig(rName string) string {
	return acctest.ConfigCompose(
		originBucket(rName),
		backupBucket(rName),
		fmt.Sprintf(`
resource "aws_cloudfront_distribution" "failover_distribution" {
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

func testAccDistributionDefaultCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered2Config(retainOnDelete bool) string {
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

func testAccDistributionDefaultCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered3Config(retainOnDelete bool) string {
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

func testAccDistributionDefaultCacheBehaviorForwardedValuesHeadersUnordered2Config(retainOnDelete bool) string {
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

func testAccDistributionDefaultCacheBehaviorForwardedValuesHeadersUnordered3Config(retainOnDelete bool) string {
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

func testAccDistributionEnabledConfig(enabled, retainOnDelete bool) string {
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

func testAccDistributionOrderedCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered2Config(retainOnDelete bool) string {
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

func testAccDistributionOrderedCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered3Config(retainOnDelete bool) string {
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

func testAccDistributionOrderedCacheBehaviorForwardedValuesHeadersUnordered2Config(retainOnDelete bool) string {
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

func testAccDistributionOrderedCacheBehaviorForwardedValuesHeadersUnordered3Config(retainOnDelete bool) string {
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

func testAccDistributionDefaultCacheBehaviorTrustedKeyGroupsConfig(retainOnDelete bool, rName string) string {
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

func testAccDistributionDefaultCacheBehaviorTrustedSignersSelfConfig(retainOnDelete bool) string {
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
func testAccDistributionViewerCertificateACMCertificateARNBaseConfig(commonName string) string {
	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, commonName)

	return testAccCloudfrontRegionProviderConfig() + fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}
`, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccDistributionViewerCertificateACMCertificateARNConfig(retainOnDelete bool) string {
	return testAccDistributionViewerCertificateACMCertificateARNBaseConfig("example.com") + fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  enabled          = false
  retain_on_delete = %t

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
`, retainOnDelete)
}

func testAccDistributionViewerCertificateACMCertificateARNConflictsWithDefaultCertificateConfig(retainOnDelete bool) string {
	return testAccDistributionViewerCertificateACMCertificateARNBaseConfig("example.com") + fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  enabled          = false
  retain_on_delete = %t

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
`, retainOnDelete)
}

func testAccDistributionWaitForDeploymentConfig(enabled, waitForDeployment bool) string {
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

func testAccDistributionDefaultCacheBehaviorRealtimeLogARNConfig(rName string, retainOnDelete bool) string {
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

func testAccDistributionOrderedCacheBehaviorRealtimeLogARNConfig(rName string, retainOnDelete bool) string {
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

func testAccDistributionOriginItem(rName string, item string) string {
	return acctest.ConfigCompose(
		originBucket(rName),
		testAccDistributionCacheBehaviorRealtimeLogBaseConfig(rName),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_cloudfront_distribution" "test" {
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

func testAccDistributionETagInitialConfig(rName string) string {
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

func testAccDistributionETagUpdatedConfig(rName string) string {
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

func testAccDistributionETagFinalConfig(rName string) string {
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
