// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// serializing tests so that we do not hit the lightsail rate limit for distributions
func TestAccLightsailDistribution_serial(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	testCases := map[string]map[string]func(t *testing.T){
		"distribution": {
			"basic":                   testAccDistribution_basic,
			"disappears":              testAccDistribution_disappears,
			"is_enabled":              testAccDistribution_isEnabled,
			"cache_behavior":          testAccDistribution_cacheBehavior,
			"cache_behavior_settings": testAccDistribution_cacheBehaviorSettings,
			"default_cache_behavior":  testAccDistribution_defaultCacheBehavior,
			"ip_address_type":         testAccDistribution_ipAddressType,
			"tags":                    testAccDistribution_tags,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccDistribution_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_distribution.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
			acctest.PreCheckRegion(t, string(types.RegionNameUsEast1))
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_basic(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "alternative_domain_names.#", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "lightsail", regexp.MustCompile(`Distribution/*`)),
					resource.TestCheckResourceAttr(resourceName, "bundle_id", "small_1_0"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.allowed_http_methods", "GET,HEAD,OPTIONS,PUT,PATCH,POST,DELETE"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.cached_http_methods", "GET,HEAD"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.default_ttl", "86400"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_cookies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_cookies.0.cookies_allow_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_cookies.0.option", "none"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_headers.0.headers_allow_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_headers.0.option", "default"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_query_strings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_query_strings.0.query_strings_allowed_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_query_strings.0.option", "false"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.maximum_ttl", "31536000"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.minimum_ttl", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.behavior", "cache"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "dualstack"),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "location.0.region_name"),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "origin.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.name", bucketName),
					resource.TestCheckResourceAttrSet(resourceName, "origin.0.region_name"),
					resource.TestCheckResourceAttrSet(resourceName, "origin.0.resource_type"),
					resource.TestCheckResourceAttrSet(resourceName, "origin_public_dns"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_type"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "support_code"),
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

func testAccDistribution_isEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_distribution.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	isEnabled := "true"
	isDisabled := "false"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
			acctest.PreCheckRegion(t, string(types.RegionNameUsEast1))
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_isEnabled(rName, bucketName, isDisabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", isDisabled),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfig_isEnabled(rName, bucketName, isEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", isEnabled),
				),
			},
		},
	})
}

func testAccDistribution_cacheBehavior(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_distribution.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	path1 := "/path1"
	behaviorCache := "cache"
	path2 := "/path2"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
			acctest.PreCheckRegion(t, string(types.RegionNameUsEast1))
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_cacheBehavior1(rName, bucketName, path1, behaviorCache),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cache_behavior.*", map[string]string{
						"path": path1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cache_behavior.*", map[string]string{
						"behavior": behaviorCache,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfig_cacheBehavior2(rName, bucketName, path1, behaviorCache, path2, behaviorCache),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cache_behavior.*", map[string]string{
						"path": path1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cache_behavior.*", map[string]string{
						"behavior": behaviorCache,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cache_behavior.*", map[string]string{
						"path": path2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cache_behavior.*", map[string]string{
						"behavior": behaviorCache,
					}),
				),
			},
		},
	})
}

func testAccDistribution_defaultCacheBehavior(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_distribution.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	instanceName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ipName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
			acctest.PreCheckRegion(t, string(types.RegionNameUsEast1))
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_defaultCacheBehaviorDontCache(rName, instanceName, ipName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "default_cache_behavior.*", map[string]string{
						"behavior": "dont-cache",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfig_defaultCacheBehaviorCache(rName, instanceName, ipName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "default_cache_behavior.*", map[string]string{
						"behavior": "cache",
					}),
				),
			},
		},
	})
}

func testAccDistribution_ipAddressType(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_distribution.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
			acctest.PreCheckRegion(t, string(types.RegionNameUsEast1))
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_ipAddressType(rName, bucketName, string(types.IpAddressTypeIpv4)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", string(types.IpAddressTypeIpv4)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfig_ipAddressType(rName, bucketName, string(types.IpAddressTypeDualstack)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", string(types.IpAddressTypeDualstack)),
				),
			},
		},
	})
}

func testAccDistribution_cacheBehaviorSettings(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_distribution.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	allow1 := "test"
	allow2 := "special"
	header1 := "Host"
	header2 := "Origin"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
			acctest.PreCheckRegion(t, string(types.RegionNameUsEast1))
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_cacheBehaviorSettings(rName, bucketName, allow1, allow2, header1, header2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_cookies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_query_strings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.allowed_http_methods", "GET,HEAD,OPTIONS"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.cached_http_methods", "GET,HEAD,OPTIONS"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.default_ttl", "50000"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.maximum_ttl", "100000"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.minimum_ttl", "10000"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_cookies.0.option", "allow-list"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_behavior_settings.0.forwarded_cookies.0.cookies_allow_list.*", allow1),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_behavior_settings.0.forwarded_cookies.0.cookies_allow_list.*", allow2),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_headers.0.option", "allow-list"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_behavior_settings.0.forwarded_headers.0.headers_allow_list.*", header1),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_behavior_settings.0.forwarded_headers.0.headers_allow_list.*", header2),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_query_strings.0.option", "true"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_behavior_settings.0.forwarded_query_strings.0.query_strings_allowed_list.*", allow1),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_behavior_settings.0.forwarded_query_strings.0.query_strings_allowed_list.*", allow2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfig_basic(rName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_cookies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_query_strings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.allowed_http_methods", "GET,HEAD,OPTIONS,PUT,PATCH,POST,DELETE"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.cached_http_methods", "GET,HEAD"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.default_ttl", "86400"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.maximum_ttl", "31536000"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.minimum_ttl", "0"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_cookies.0.option", "none"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_cookies.0.cookies_allow_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_cookies.0.cookies_allow_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_headers.0.option", "default"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_headers.0.headers_allow_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_headers.0.headers_allow_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_query_strings.0.option", "false"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_query_strings.0.query_strings_allowed_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior_settings.0.forwarded_query_strings.0.query_strings_allowed_list.#", "0"),
				),
			},
		},
	})
}

func testAccDistribution_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_distribution.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
			acctest.PreCheckRegion(t, string(types.RegionNameUsEast1))
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_tags1(rName, bucketName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfig_tags2(rName, bucketName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDistributionConfig_tags1(rName, bucketName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccDistribution_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_distribution.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
			acctest.PreCheckRegion(t, string(types.RegionNameUsEast1))
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfig_basic(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflightsail.ResourceDistribution(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDistributionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_distribution" {
				continue
			}

			_, err := tflightsail.FindDistributionByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Lightsail, create.ErrActionCheckingDestroyed, tflightsail.ResNameDistribution, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDistributionExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Lightsail, create.ErrActionCheckingExistence, tflightsail.ResNameDistribution, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Lightsail, create.ErrActionCheckingExistence, tflightsail.ResNameDistribution, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)
		resp, err := tflightsail.FindDistributionByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("Distribution %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDistributionConfig_base(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_bucket" "test" {
  name      = %[1]q
  bundle_id = "small_1_0"
}`, bucketName)
}

func testAccDistributionConfig_basic(rName, bucketName string) string {
	return acctest.ConfigCompose(
		testAccDistributionConfig_base(bucketName),
		fmt.Sprintf(`
resource "aws_lightsail_distribution" "test" {
  name      = %[1]q
  bundle_id = "small_1_0"
  origin {
    name        = aws_lightsail_bucket.test.name
    region_name = aws_lightsail_bucket.test.region
  }
  default_cache_behavior {
    behavior = "cache"
  }
  cache_behavior_settings {
    allowed_http_methods = "GET,HEAD,OPTIONS,PUT,PATCH,POST,DELETE"
    cached_http_methods  = "GET,HEAD"
    default_ttl          = 86400
    maximum_ttl          = 31536000
    minimum_ttl          = 0
    forwarded_cookies {
      option = "none"
    }
    forwarded_headers {
      option = "default"
    }
    forwarded_query_strings {
      option = false
    }
  }
}
`, rName))
}

func testAccDistributionConfig_isEnabled(rName, bucketName, isEnabled string) string {
	return acctest.ConfigCompose(
		testAccDistributionConfig_base(bucketName),
		fmt.Sprintf(`
resource "aws_lightsail_distribution" "test" {
  name       = %[1]q
  bundle_id  = "small_1_0"
  is_enabled = %[2]s
  origin {
    name        = aws_lightsail_bucket.test.name
    region_name = aws_lightsail_bucket.test.region
  }
  default_cache_behavior {
    behavior = "cache"
  }
  cache_behavior_settings {
    allowed_http_methods = "GET,HEAD,OPTIONS,PUT,PATCH,POST,DELETE"
    cached_http_methods  = "GET,HEAD"
    default_ttl          = 86400
    maximum_ttl          = 31536000
    minimum_ttl          = 0
    forwarded_cookies {
      option = "none"
    }
    forwarded_headers {
      option = "default"
    }
    forwarded_query_strings {
      option = false
    }
  }
}
`, rName, isEnabled))
}

func testAccDistributionConfig_tags1(rName, bucketName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccDistributionConfig_base(bucketName),
		fmt.Sprintf(`	
resource "aws_lightsail_distribution" "test" {
  name      = %[1]q
  bundle_id = "small_1_0"
  origin {
    name        = aws_lightsail_bucket.test.name
    region_name = aws_lightsail_bucket.test.region
  }
  default_cache_behavior {
    behavior = "cache"
  }
  cache_behavior_settings {
    allowed_http_methods = "GET,HEAD,OPTIONS,PUT,PATCH,POST,DELETE"
    cached_http_methods  = "GET,HEAD"
    default_ttl          = 86400
    maximum_ttl          = 31536000
    minimum_ttl          = 0
    forwarded_cookies {
      option = "none"
    }
    forwarded_headers {
      option = "default"
    }
    forwarded_query_strings {
      option = false
    }
  }
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccDistributionConfig_tags2(rName, bucketName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccDistributionConfig_base(bucketName),
		fmt.Sprintf(`	
resource "aws_lightsail_distribution" "test" {
  name      = %[1]q
  bundle_id = "small_1_0"
  origin {
    name        = aws_lightsail_bucket.test.name
    region_name = aws_lightsail_bucket.test.region
  }
  default_cache_behavior {
    behavior = "cache"
  }
  cache_behavior_settings {
    allowed_http_methods = "GET,HEAD,OPTIONS,PUT,PATCH,POST,DELETE"
    cached_http_methods  = "GET,HEAD"
    default_ttl          = 86400
    maximum_ttl          = 31536000
    minimum_ttl          = 0
    forwarded_cookies {
      option = "none"
    }
    forwarded_headers {
      option = "default"
    }
    forwarded_query_strings {
      option = false
    }
  }
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccDistributionConfig_cacheBehavior1(rName, bucketName, path1, behavior1 string) string {
	return acctest.ConfigCompose(
		testAccDistributionConfig_base(bucketName),
		fmt.Sprintf(`	
resource "aws_lightsail_distribution" "test" {
  name      = %[1]q
  bundle_id = "small_1_0"
  origin {
    name        = aws_lightsail_bucket.test.name
    region_name = aws_lightsail_bucket.test.region
  }
  default_cache_behavior {
    behavior = "cache"
  }
  cache_behavior_settings {
    allowed_http_methods = "GET,HEAD,OPTIONS,PUT,PATCH,POST,DELETE"
    cached_http_methods  = "GET,HEAD"
    default_ttl          = 86400
    maximum_ttl          = 31536000
    minimum_ttl          = 0
    forwarded_cookies {
      option = "none"
    }
    forwarded_headers {
      option = "default"
    }
    forwarded_query_strings {
      option = false
    }
  }
  cache_behavior {
    path     = %[2]q
    behavior = %[3]q
  }
}
`, rName, path1, behavior1))
}

func testAccDistributionConfig_cacheBehavior2(rName, bucketName, path1, behavior1, path2, behavior2 string) string {
	return acctest.ConfigCompose(
		testAccDistributionConfig_base(bucketName),
		fmt.Sprintf(`	
resource "aws_lightsail_distribution" "test" {
  name      = %[1]q
  bundle_id = "small_1_0"
  origin {
    name        = aws_lightsail_bucket.test.name
    region_name = aws_lightsail_bucket.test.region
  }
  default_cache_behavior {
    behavior = "cache"
  }
  cache_behavior_settings {
    allowed_http_methods = "GET,HEAD,OPTIONS,PUT,PATCH,POST,DELETE"
    cached_http_methods  = "GET,HEAD"
    default_ttl          = 86400
    maximum_ttl          = 31536000
    minimum_ttl          = 0
    forwarded_cookies {
      option = "none"
    }
    forwarded_headers {
      option = "default"
    }
    forwarded_query_strings {
      option = false
    }
  }
  cache_behavior {
    path     = %[2]q
    behavior = %[3]q
  }

  cache_behavior {
    path     = %[4]q
    behavior = %[5]q
  }
}
`, rName, path1, behavior1, path2, behavior2))
}

func testAccDistributionConfig_cacheBehaviorSettings(rName, bucketName, allow1, allow2, header1, header2 string) string {
	return acctest.ConfigCompose(
		testAccDistributionConfig_base(bucketName),
		fmt.Sprintf(`	
resource "aws_lightsail_distribution" "test" {
  name      = %[1]q
  bundle_id = "small_1_0"
  origin {
    name        = aws_lightsail_bucket.test.name
    region_name = aws_lightsail_bucket.test.region
  }
  default_cache_behavior {
    behavior = "cache"
  }
  cache_behavior_settings {
    allowed_http_methods = "GET,HEAD,OPTIONS"
    cached_http_methods  = "GET,HEAD,OPTIONS"
    default_ttl          = 50000
    forwarded_cookies {
      cookies_allow_list = [%[2]q, %[3]q]
      option             = "allow-list"
    }
    forwarded_headers {
      headers_allow_list = [%[4]q, %[5]q]
      option             = "allow-list"
    }
    forwarded_query_strings {
      query_strings_allowed_list = [%[2]q, %[3]q]
      option                     = true
    }
    maximum_ttl = 100000
    minimum_ttl = 10000
  }
}
`, rName, allow1, allow2, header1, header2))
}

func testAccDistributionConfig_defaultCacheBehaviorDontCache(rName, instanceName, ipName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_static_ip_attachment" "test" {
  static_ip_name = aws_lightsail_static_ip.test.name
  instance_name  = aws_lightsail_instance.test.name
}

resource "aws_lightsail_static_ip" "test" {
  name = %[3]q
}

resource "aws_lightsail_instance" "test" {
  name              = %[2]q
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "micro_1_0"
}

resource "aws_lightsail_distribution" "test" {
  name       = %[1]q
  depends_on = [aws_lightsail_static_ip_attachment.test]
  bundle_id  = "small_1_0"
  origin {
    name            = aws_lightsail_instance.test.name
    region_name     = data.aws_availability_zones.available.id
    protocol_policy = "http-only"
  }
  default_cache_behavior {
    behavior = "dont-cache"
  }
  cache_behavior_settings {}

}
`, rName, instanceName, ipName)
}

func testAccDistributionConfig_defaultCacheBehaviorCache(rName, instanceName, ipName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_static_ip_attachment" "test" {
  static_ip_name = aws_lightsail_static_ip.test.name
  instance_name  = aws_lightsail_instance.test.name
}

resource "aws_lightsail_static_ip" "test" {
  name = %[3]q
}

resource "aws_lightsail_instance" "test" {
  name              = %[2]q
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "micro_1_0"
}

resource "aws_lightsail_distribution" "test" {
  name       = %[1]q
  depends_on = [aws_lightsail_static_ip_attachment.test]
  bundle_id  = "small_1_0"
  origin {
    name            = aws_lightsail_instance.test.name
    region_name     = data.aws_availability_zones.available.id
    protocol_policy = "http-only"
  }
  default_cache_behavior {
    behavior = "cache"
  }
  cache_behavior_settings {
    allowed_http_methods = "GET,HEAD,OPTIONS,PUT,PATCH,POST,DELETE"
    cached_http_methods  = "GET,HEAD"
    default_ttl          = 86400
    maximum_ttl          = 31536000

    forwarded_cookies {
      option = "none"
    }

    forwarded_headers {
      headers_allow_list = ["Host"]
      option             = "allow-list"
    }

    forwarded_query_strings {
      option = false
    }

  }
}
`, rName, instanceName, ipName)
}

func testAccDistributionConfig_ipAddressType(rName, bucketName, IpAddressType string) string {
	return acctest.ConfigCompose(
		testAccDistributionConfig_base(bucketName),
		fmt.Sprintf(`
resource "aws_lightsail_distribution" "test" {
  name            = %[1]q
  bundle_id       = "small_1_0"
  ip_address_type = %[2]q
  origin {
    name        = aws_lightsail_bucket.test.name
    region_name = aws_lightsail_bucket.test.region
  }
  default_cache_behavior {
    behavior = "cache"
  }
  cache_behavior_settings {
    allowed_http_methods = "GET,HEAD,OPTIONS,PUT,PATCH,POST,DELETE"
    cached_http_methods  = "GET,HEAD"
    default_ttl          = 86400
    maximum_ttl          = 31536000
    minimum_ttl          = 0
    forwarded_cookies {
      option = "none"
    }
    forwarded_headers {
      option = "default"
    }
    forwarded_query_strings {
      option = false
    }
  }
}
`, rName, IpAddressType))
}
