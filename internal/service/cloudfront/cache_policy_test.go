// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontCachePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_cache_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCachePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCachePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, ""),
					resource.TestCheckResourceAttr(resourceName, "default_ttl", "86400"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, "min_ttl", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "max_ttl", "31536000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookie_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_brotli", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_gzip", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.header_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_string_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.#", acctest.Ct0),
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

func TestAccCloudFrontCachePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_cache_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCachePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCachePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachePolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceCachePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontCachePolicy_Items(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_cache_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCachePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCachePolicyConfig_items(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "test comment"),
					resource.TestCheckResourceAttr(resourceName, "default_ttl", "50"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, "min_ttl", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "max_ttl", "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookie_behavior", "allExcept"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.0.items.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.0.items.*", "test1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_brotli", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_gzip", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.header_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.0.items.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.0.items.*", "test1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.0.items.*", "test2"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_string_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.0.items.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.0.items.*", "test1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.0.items.*", "test2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.0.items.*", "test3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCachePolicyConfig_itemsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "test comment updated"),
					resource.TestCheckResourceAttr(resourceName, "default_ttl", "51"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, "min_ttl", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "max_ttl", "99"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookie_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.0.items.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.0.items.*", "test1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.0.items.*", "test2"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_brotli", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_gzip", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.header_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.0.items.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.0.items.*", "test1"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_string_behavior", "all"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccCloudFrontCachePolicy_ZeroTTLs(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_cache_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCachePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCachePolicyConfig_zeroTTLs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, ""),
					resource.TestCheckResourceAttr(resourceName, "default_ttl", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, "min_ttl", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "max_ttl", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookie_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.cookies_config.0.cookies.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_brotli", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.enable_accept_encoding_gzip", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.header_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.headers_config.0.headers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_string_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "parameters_in_cache_key_and_forwarded_to_origin.0.query_strings_config.0.query_strings.#", acctest.Ct0),
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

func testAccCheckCachePolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_cache_policy" {
				continue
			}

			_, err := tfcloudfront.FindCachePolicyByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Cache Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCachePolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		_, err := tfcloudfront.FindCachePolicyByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCachePolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_cache_policy" "test" {
  name = %[1]q

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
`, rName)
}

func testAccCachePolicyConfig_items(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_cache_policy" "test" {
  name        = %[1]q
  comment     = "test comment"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1

  parameters_in_cache_key_and_forwarded_to_origin {
    enable_accept_encoding_brotli = true
    enable_accept_encoding_gzip   = true

    cookies_config {
      cookie_behavior = "allExcept"

      cookies {
        items = ["test1"]
      }
    }

    headers_config {
      header_behavior = "whitelist"

      headers {
        items = ["test1", "test2"]
      }
    }

    query_strings_config {
      query_string_behavior = "whitelist"

      query_strings {
        items = ["test1", "test2", "test3"]
      }
    }
  }
}
`, rName)
}

func testAccCachePolicyConfig_itemsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_cache_policy" "test" {
  name        = %[1]q
  comment     = "test comment updated"
  default_ttl = 51
  max_ttl     = 99
  min_ttl     = 10

  parameters_in_cache_key_and_forwarded_to_origin {
    enable_accept_encoding_brotli = false
    enable_accept_encoding_gzip   = false

    cookies_config {
      cookie_behavior = "whitelist"

      cookies {
        items = ["test2", "test1"]
      }
    }

    headers_config {
      header_behavior = "whitelist"

      headers {
        items = ["test1"]
      }
    }

    query_strings_config {
      query_string_behavior = "all"
    }
  }
}
`, rName)
}

func testAccCachePolicyConfig_zeroTTLs(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_cache_policy" "test" {
  name = %[1]q

  default_ttl = 0
  max_ttl     = 0
  min_ttl     = 0

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
`, rName)
}
