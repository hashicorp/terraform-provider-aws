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

func TestAccCloudFrontResponseHeadersPolicy_cors(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_response_headers_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResponseHeadersPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResponseHeadersPolicyConfig_cors(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponseHeadersPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "test comment"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_credentials", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_headers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_headers.0.items.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_config.0.access_control_allow_headers.0.items.*", "X-Header1"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_methods.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_methods.0.items.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_config.0.access_control_allow_methods.0.items.*", "GET"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_config.0.access_control_allow_methods.0.items.*", "POST"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_origins.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_origins.0.items.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_config.0.access_control_allow_origins.0.items.*", "test1.example.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_config.0.access_control_allow_origins.0.items.*", "test2.example.com"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_expose_headers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_max_age_sec", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.origin_override", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "custom_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, "remove_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "server_timing_headers_config.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccResponseHeadersPolicyConfig_corsUpdated(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponseHeadersPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "test comment updated"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_credentials", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_headers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_headers.0.items.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_config.0.access_control_allow_headers.0.items.*", "X-Header2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_config.0.access_control_allow_headers.0.items.*", "X-Header3"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_methods.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_config.0.access_control_allow_methods.0.items.*", "PUT"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_origins.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_origins.0.items.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_config.0.access_control_allow_origins.0.items.*", "test1.example.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_config.0.access_control_allow_origins.0.items.*", "test2.example.com"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_expose_headers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_expose_headers.0.items.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_config.0.access_control_expose_headers.0.items.*", "HEAD"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_max_age_sec", "3600"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.origin_override", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "custom_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, "remove_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "server_timing_headers_config.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccCloudFrontResponseHeadersPolicy_customHeaders(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_response_headers_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResponseHeadersPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResponseHeadersPolicyConfig_custom(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponseHeadersPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, ""),
					resource.TestCheckResourceAttr(resourceName, "cors_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "custom_headers_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_headers_config.0.items.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_headers_config.0.items.*", map[string]string{
						names.AttrHeader: "X-Header1",
						"override":       acctest.CtTrue,
						names.AttrValue:  acctest.CtValue1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_headers_config.0.items.*", map[string]string{
						names.AttrHeader: "X-Header2",
						"override":       acctest.CtFalse,
						names.AttrValue:  acctest.CtValue2,
					}),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "remove_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "server_timing_headers_config.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func TestAccCloudFrontResponseHeadersPolicy_RemoveHeadersConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_response_headers_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResponseHeadersPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResponseHeadersPolicyConfig_remove(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponseHeadersPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, ""),
					resource.TestCheckResourceAttr(resourceName, "cors_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "custom_headers_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_headers_config.0.items.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_headers_config.0.items.*", map[string]string{
						names.AttrHeader: "X-Header1",
						"override":       acctest.CtTrue,
						names.AttrValue:  acctest.CtValue1,
					}),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "remove_headers_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "remove_headers_config.0.items.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "remove_headers_config.0.items.*", map[string]string{
						names.AttrHeader: "X-Header3",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "remove_headers_config.0.items.*", map[string]string{
						names.AttrHeader: "X-Header4",
					}),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "server_timing_headers_config.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func TestAccCloudFrontResponseHeadersPolicy_securityHeaders(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_response_headers_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResponseHeadersPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResponseHeadersPolicyConfig_security(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponseHeadersPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, ""),
					resource.TestCheckResourceAttr(resourceName, "cors_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "custom_headers_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_headers_config.0.items.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_headers_config.0.items.*", map[string]string{
						names.AttrHeader: "X-Header1",
						"override":       acctest.CtTrue,
						names.AttrValue:  acctest.CtValue1,
					}),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "remove_headers_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "remove_headers_config.0.items.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "remove_headers_config.0.items.*", map[string]string{
						names.AttrHeader: "X-Header3",
					}),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.content_security_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.content_security_policy.0.content_security_policy", "policy1"),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.content_security_policy.0.override", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.content_type_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.frame_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.frame_options.0.frame_option", "DENY"),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.frame_options.0.override", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.referrer_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.strict_transport_security.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.strict_transport_security.0.access_control_max_age_sec", "90"),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.strict_transport_security.0.include_subdomains", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.strict_transport_security.0.override", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.strict_transport_security.0.preload", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.xss_protection.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "server_timing_headers_config.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccResponseHeadersPolicyConfig_securityUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponseHeadersPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, ""),
					resource.TestCheckResourceAttr(resourceName, "cors_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "custom_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "remove_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.content_security_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.content_type_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.content_type_options.0.override", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.frame_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.referrer_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.referrer_policy.0.override", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.referrer_policy.0.referrer_policy", "origin-when-cross-origin"),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.strict_transport_security.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.xss_protection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.xss_protection.0.mode_block", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.xss_protection.0.override", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.xss_protection.0.protection", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.0.xss_protection.0.report_uri", "https://example.com/"),
					resource.TestCheckResourceAttr(resourceName, "server_timing_headers_config.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccCloudFrontResponseHeadersPolicy_serverTimingHeaders(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_response_headers_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResponseHeadersPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResponseHeadersPolicyConfig_serverTiming(rName, true, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponseHeadersPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, ""),
					resource.TestCheckResourceAttr(resourceName, "cors_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "custom_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "remove_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "server_timing_headers_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_timing_headers_config.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "server_timing_headers_config.0.sampling_rate", acctest.Ct10),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccResponseHeadersPolicyConfig_serverTiming(rName, true, 90),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponseHeadersPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, ""),
					resource.TestCheckResourceAttr(resourceName, "cors_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "custom_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "remove_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "server_timing_headers_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_timing_headers_config.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "server_timing_headers_config.0.sampling_rate", "90"),
				),
			},
			{
				Config: testAccResponseHeadersPolicyConfig_serverTiming(rName, true, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponseHeadersPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, ""),
					resource.TestCheckResourceAttr(resourceName, "cors_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "custom_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "server_timing_headers_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_timing_headers_config.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "server_timing_headers_config.0.sampling_rate", acctest.Ct0),
				),
			},
			{
				Config: testAccResponseHeadersPolicyConfig_serverTiming(rName, false, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponseHeadersPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, ""),
					resource.TestCheckResourceAttr(resourceName, "cors_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "custom_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "remove_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "security_headers_config.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "server_timing_headers_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_timing_headers_config.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "server_timing_headers_config.0.sampling_rate", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccCloudFrontResponseHeadersPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_response_headers_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResponseHeadersPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResponseHeadersPolicyConfig_cors(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResponseHeadersPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceResponseHeadersPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckResponseHeadersPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_response_headers_policy" {
				continue
			}

			_, err := tfcloudfront.FindResponseHeadersPolicyByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Response Headers Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckResponseHeadersPolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		_, err := tfcloudfront.FindResponseHeadersPolicyByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccResponseHeadersPolicyConfig_cors(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_response_headers_policy" "test" {
  name    = %[1]q
  comment = "test comment"

  cors_config {
    access_control_allow_credentials = false

    access_control_allow_headers {
      items = ["X-Header1"]
    }

    access_control_allow_methods {
      items = ["GET", "POST"]
    }

    access_control_allow_origins {
      items = ["test1.example.com", "test2.example.com"]
    }

    origin_override = true
  }
}
`, rName)
}

func testAccResponseHeadersPolicyConfig_corsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_response_headers_policy" "test" {
  name    = %[1]q
  comment = "test comment updated"

  cors_config {
    access_control_allow_credentials = true

    access_control_allow_headers {
      items = ["X-Header2", "X-Header3"]
    }

    access_control_allow_methods {
      items = ["PUT"]
    }

    access_control_allow_origins {
      items = ["test1.example.com", "test2.example.com"]
    }

    access_control_expose_headers {
      items = ["HEAD"]
    }

    access_control_max_age_sec = 3600

    origin_override = false
  }
}
`, rName)
}

func testAccResponseHeadersPolicyConfig_custom(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_response_headers_policy" "test" {
  name = %[1]q

  custom_headers_config {
    items {
      header   = "X-Header2"
      override = false
      value    = "value2"
    }

    items {
      header   = "X-Header1"
      override = true
      value    = "value1"
    }
  }
}
`, rName)
}

func testAccResponseHeadersPolicyConfig_remove(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_response_headers_policy" "test" {
  name = %[1]q

  custom_headers_config {
    items {
      header   = "X-Header1"
      override = true
      value    = "value1"
    }
  }

  remove_headers_config {
    items {
      header = "X-Header3"
    }

    items {
      header = "X-Header4"
    }
  }
}
`, rName)
}

func testAccResponseHeadersPolicyConfig_security(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_response_headers_policy" "test" {
  name = %[1]q

  custom_headers_config {
    items {
      header   = "X-Header1"
      override = true
      value    = "value1"
    }
  }

  remove_headers_config {
    items {
      header = "X-Header3"
    }
  }

  security_headers_config {
    content_security_policy {
      content_security_policy = "policy1"
      override                = true
    }

    frame_options {
      frame_option = "DENY"
      override     = false
    }

    strict_transport_security {
      access_control_max_age_sec = 90
      override                   = true
      preload                    = true
    }
  }
}
`, rName)
}

func testAccResponseHeadersPolicyConfig_securityUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_response_headers_policy" "test" {
  name = %[1]q

  security_headers_config {
    content_type_options {
      override = true
    }

    referrer_policy {
      referrer_policy = "origin-when-cross-origin"
      override        = false
    }

    xss_protection {
      override   = true
      protection = true
      report_uri = "https://example.com/"
    }
  }
}
`, rName)
}

func testAccResponseHeadersPolicyConfig_serverTiming(rName string, enabled bool, rate float64) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_response_headers_policy" "test" {
  name = %[1]q

  server_timing_headers_config {
    enabled       = %[2]t
    sampling_rate = %[3]f
  }
}
`, rName, enabled, rate)
}
