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

func TestAccCloudFrontOriginRequestPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_origin_request_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginRequestPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginRequestPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOriginRequestPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, ""),
					resource.TestCheckResourceAttr(resourceName, "cookies_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cookies_config.0.cookie_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "cookies_config.0.cookies.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, "headers_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "headers_config.0.header_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "headers_config.0.headers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "query_strings_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "query_strings_config.0.query_string_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "query_strings_config.0.query_strings.#", acctest.Ct0),
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

func TestAccCloudFrontOriginRequestPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_origin_request_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginRequestPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginRequestPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOriginRequestPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceOriginRequestPolicy(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceOriginRequestPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontOriginRequestPolicy_Items(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_origin_request_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginRequestPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginRequestPolicyConfig_items(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOriginRequestPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "test comment"),
					resource.TestCheckResourceAttr(resourceName, "cookies_config.0.cookie_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "cookies_config.0.cookies.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cookies_config.0.cookies.0.items.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "cookies_config.0.cookies.0.items.*", "test1"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, "headers_config.0.header_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "headers_config.0.headers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "headers_config.0.headers.0.items.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "headers_config.0.headers.0.items.*", "test1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "headers_config.0.headers.0.items.*", "test2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "query_strings_config.0.query_string_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "query_strings_config.0.query_strings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "query_strings_config.0.query_strings.0.items.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttr(resourceName, "query_strings_config.0.query_strings.0.items.*", "test1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "query_strings_config.0.query_strings.0.items.*", "test2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "query_strings_config.0.query_strings.0.items.*", "test3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOriginRequestPolicyConfig_itemsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOriginRequestPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "test comment updated"),
					resource.TestCheckResourceAttr(resourceName, "cookies_config.0.cookie_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "cookies_config.0.cookies.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cookies_config.0.cookies.0.items.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "cookies_config.0.cookies.0.items.*", "test1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cookies_config.0.cookies.0.items.*", "test2"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, "headers_config.0.header_behavior", "allViewerAndWhitelistCloudFront"),
					resource.TestCheckResourceAttr(resourceName, "headers_config.0.headers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "headers_config.0.headers.0.items.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "headers_config.0.headers.0.items.*", "CloudFront-Viewer-City"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "query_strings_config.0.query_string_behavior", "all"),
					resource.TestCheckResourceAttr(resourceName, "query_strings_config.0.query_strings.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccCheckOriginRequestPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_origin_request_policy" {
				continue
			}

			_, err := tfcloudfront.FindOriginRequestPolicyByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Origin Request Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckOriginRequestPolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		_, err := tfcloudfront.FindOriginRequestPolicyByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccOriginRequestPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_request_policy" "test" {
  name = %[1]q

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
`, rName)
}

func testAccOriginRequestPolicyConfig_items(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_request_policy" "test" {
  name    = %[1]q
  comment = "test comment"

  cookies_config {
    cookie_behavior = "whitelist"

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
`, rName)
}

func testAccOriginRequestPolicyConfig_itemsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_request_policy" "test" {
  name    = %[1]q
  comment = "test comment updated"

  cookies_config {
    cookie_behavior = "whitelist"

    cookies {
      items = ["test2", "test1"]
    }
  }

  headers_config {
    header_behavior = "allViewerAndWhitelistCloudFront"

    headers {
      items = ["CloudFront-Viewer-City"]
    }
  }

  query_strings_config {
    query_string_behavior = "all"
  }
}
`, rName)
}
