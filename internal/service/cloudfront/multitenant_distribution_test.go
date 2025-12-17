// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontMultiTenantDistribution_basic(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_multitenant_distribution.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiTenantDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiTenantDistributionConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiTenantDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "tenant_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tenant_config.0.parameter_definition.0.definition.0.string_schema.0.required", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccCloudFrontMultiTenantDistribution_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_multitenant_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiTenantDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiTenantDistributionConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiTenantDistributionExists(ctx, resourceName, &distribution),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceMultiTenantDistribution, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMultiTenantDistributionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_multitenant_distribution" {
				continue
			}

			_, err := tfcloudfront.FindDistributionByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Multi-tenant Distribution %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckMultiTenantDistributionExists(ctx context.Context, n string, v *awstypes.Distribution) resource.TestCheckFunc {
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

func TestAccCloudFrontMultiTenantDistribution_comprehensive(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_multitenant_distribution.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiTenantDistributionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiTenantDistributionConfig_comprehensive(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiTenantDistributionExists(ctx, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "Comprehensive multi-tenant distribution test"),
					resource.TestCheckResourceAttr(resourceName, "default_root_object", "index.html"),
					resource.TestCheckResourceAttr(resourceName, "http_version", "http2"),

					// Check connection_mode is computed
					resource.TestCheckResourceAttrSet(resourceName, "connection_mode"),

					// Check multiple origins
					resource.TestCheckResourceAttr(resourceName, "origin.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.id", "custom-origin"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.origin_path", "/api"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.custom_header.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.custom_header.0.header_name", "X-Custom-Header"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.custom_header.0.header_value", "test-value"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.origin_shield.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.origin_shield.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "origin.0.origin_shield.0.origin_shield_region", "us-east-1"),

					resource.TestCheckResourceAttr(resourceName, "origin.1.id", "s3-origin"),
					resource.TestCheckResourceAttr(resourceName, "origin.1.s3_origin_config.#", "1"),

					// Check cache behaviors
					resource.TestCheckResourceAttr(resourceName, "cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior.0.path_pattern", "/api/*"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior.0.target_origin_id", "custom-origin"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior.0.compress", acctest.CtTrue),

					// Check custom error responses
					resource.TestCheckResourceAttr(resourceName, "custom_error_response.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "custom_error_response.0.error_code", "404"),
					resource.TestCheckResourceAttr(resourceName, "custom_error_response.0.response_code", "200"),
					resource.TestCheckResourceAttr(resourceName, "custom_error_response.0.response_page_path", "/404.html"),

					// Check tags
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),

					// Check tenant config with multiple parameters
					resource.TestCheckResourceAttr(resourceName, "tenant_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tenant_config.0.parameter_definition.#", "2"),
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

func testAccMultiTenantDistributionConfig_comprehensive(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_cloudfront_origin_access_control" "test" {
  name                              = %[1]q
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

resource "aws_cloudfront_multitenant_distribution" "test" {
  enabled             = true
  comment             = "Comprehensive multi-tenant distribution test"
  default_root_object = "index.html"
  http_version        = "http2"

  # Custom origin with headers and shield
  origin {
    domain_name = "example.com"
    id          = "custom-origin"
    origin_path = "/api"

    custom_header {
      header_name  = "X-Custom-Header"
      header_value = "test-value"
    }

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }

    origin_shield {
      enabled              = true
      origin_shield_region = "us-east-1"
    }
  }

  # S3 origin
  origin {
    domain_name              = aws_s3_bucket.test.bucket_regional_domain_name
    id                       = "s3-origin"
    origin_access_control_id = aws_cloudfront_origin_access_control.test.id

    s3_origin_config {
      origin_access_identity = ""
    }
  }

  default_cache_behavior {
    target_origin_id       = "s3-origin"
    viewer_protocol_policy = "redirect-to-https"
    cache_policy_id        = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad" # AWS Managed CachingDisabled policy

    allowed_methods {
      items          = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
      cached_methods = ["GET", "HEAD"]
    }
  }

  # Additional cache behavior
  cache_behavior {
    path_pattern           = "/api/*"
    target_origin_id       = "custom-origin"
    viewer_protocol_policy = "https-only"
    cache_policy_id        = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad"
    compress               = true

    allowed_methods {
      items          = ["GET", "HEAD", "OPTIONS"]
      cached_methods = ["GET", "HEAD"]
    }
  }

  # Custom error response
  custom_error_response {
    error_code         = 404
    response_code      = "200"
    response_page_path = "/404.html"
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  # Multi-parameter tenant config
  tenant_config {
    parameter_definition {
      name = "origin_domain"
      definition {
        string_schema {
          required = true
          comment  = "Origin domain parameter for tenants"
        }
      }
    }

    parameter_definition {
      name = "cache_prefix"
      definition {
        string_schema {
          required = false
          comment  = "Cache prefix parameter for tenants"
        }
      }
    }
  }

  tags = {
    Environment = "test"
    Name        = %[1]q
  }
}
`, rName)
}

func testAccMultiTenantDistributionConfig_basic() string {
	return `
resource "aws_cloudfront_multitenant_distribution" "test" {
  enabled = false
  comment = "Test multi-tenant distribution"

  origin {
    domain_name = "example.com"
    id          = "example"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  default_cache_behavior {
    target_origin_id       = "example"
    viewer_protocol_policy = "redirect-to-https"
    cache_policy_id        = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad" # AWS Managed CachingDisabled policy

    allowed_methods {
      items          = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
      cached_methods = ["GET", "HEAD", "OPTIONS"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  tenant_config {
    parameter_definition {
      name = "origin_domain"
      definition {
        string_schema {
          required = true
          comment  = "Origin domain parameter for tenants"
        }
      }
    }
  }
}
`
}
