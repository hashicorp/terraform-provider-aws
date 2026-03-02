// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontMultiTenantDistribution_basic(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_multitenant_distribution.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiTenantDistributionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiTenantDistributionConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiTenantDistributionExists(ctx, t, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "tenant_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tenant_config.0.parameter_definition.0.definition.0.string_schema.0.required", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
		},
	})
}

func TestAccCloudFrontMultiTenantDistribution_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_multitenant_distribution.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiTenantDistributionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiTenantDistributionConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiTenantDistributionExists(ctx, t, resourceName, &distribution),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudfront.ResourceMultiTenantDistribution, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontMultiTenantDistribution_comprehensive(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_multitenant_distribution.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiTenantDistributionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiTenantDistributionConfig_comprehensive(rName, "Comprehensive multi-tenant distribution test", "index.html", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiTenantDistributionExists(ctx, t, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "Comprehensive multi-tenant distribution test"),
					resource.TestCheckResourceAttr(resourceName, "default_root_object", "index.html"),
					resource.TestCheckResourceAttr(resourceName, "http_version", "http2"),

					// Check connection_mode is computed
					resource.TestCheckResourceAttrSet(resourceName, "connection_mode"),

					// Check multiple origins
					resource.TestCheckResourceAttr(resourceName, "origin.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.id", "custom-origin"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.origin_path", "/api"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.custom_header.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.custom_header.0.header_name", "X-Custom-Header"),
					resource.TestCheckResourceAttr(resourceName, "origin.0.custom_header.0.header_value", "test-value"),

					// Check cache behaviors
					resource.TestCheckResourceAttr(resourceName, "cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior.0.path_pattern", "/api/*"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior.0.target_origin_id", "custom-origin"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior.0.compress", acctest.CtFalse),

					// Check custom error responses
					resource.TestCheckResourceAttr(resourceName, "custom_error_response.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "custom_error_response.0.error_code", "404"),
					resource.TestCheckResourceAttr(resourceName, "custom_error_response.0.response_code", "200"),
					resource.TestCheckResourceAttr(resourceName, "custom_error_response.0.response_page_path", "/404.html"),

					// Check tags
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),

					// Check tenant config with single parameter
					resource.TestCheckResourceAttr(resourceName, "tenant_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tenant_config.0.parameter_definition.#", "1"),
				),
			},
			{
				Config: testAccMultiTenantDistributionConfig_comprehensive(rName, "Updated comprehensive test", "updated.html", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiTenantDistributionExists(ctx, t, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "Updated comprehensive test"),
					resource.TestCheckResourceAttr(resourceName, "default_root_object", "updated.html"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior.0.compress", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
		},
	})
}

func TestAccCloudFrontMultiTenantDistribution_s3OriginWithOAC(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_multitenant_distribution.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiTenantDistributionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiTenantDistributionConfig_s3OriginWithOAC(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiTenantDistributionExists(ctx, t, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "origin.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "origin.0.origin_access_control_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
		},
	})
}

// Ref: https://github.com/hashicorp/terraform-provider-aws/issues/46302
func TestAccCloudFrontMultiTenantDistribution_customErrorResponseWithoutResponsePagePath(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_multitenant_distribution.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiTenantDistributionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiTenantDistributionConfig_customErrorResponseWithoutResponsePagePath(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiTenantDistributionExists(ctx, t, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "tenant_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tenant_config.0.parameter_definition.0.definition.0.string_schema.0.required", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "custom_error_response.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "custom_error_response.0.error_code", "404"),
					resource.TestCheckNoResourceAttr(resourceName, "custom_error_response.0.response_code"),
					resource.TestCheckNoResourceAttr(resourceName, "custom_error_response.0.response_page_path"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
		},
	})
}

func TestAccCloudFrontMultiTenantDistribution_tags(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_multitenant_distribution.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiTenantDistributionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiTenantDistributionConfig_tags(map[string]string{acctest.CtKey1: acctest.CtValue1}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiTenantDistributionExists(ctx, t, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
			{
				Config: testAccMultiTenantDistributionConfig_tags(map[string]string{acctest.CtKey1: acctest.CtValue1Updated, acctest.CtKey2: acctest.CtValue2}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiTenantDistributionExists(ctx, t, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccMultiTenantDistributionConfig_tags(map[string]string{acctest.CtKey2: acctest.CtValue2}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiTenantDistributionExists(ctx, t, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccCloudFrontMultiTenantDistribution_update(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)
	var distribution awstypes.Distribution
	resourceName := "aws_cloudfront_multitenant_distribution.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiTenantDistributionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiTenantDistributionConfig_update("Initial comment", "http1.1", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiTenantDistributionExists(ctx, t, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "Initial comment"),
					resource.TestCheckResourceAttr(resourceName, "http_version", "http1.1"),
				),
			},
			{
				Config: testAccMultiTenantDistributionConfig_update("Updated comment", "http2", "updated.html"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiTenantDistributionExists(ctx, t, resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "Updated comment"),
					resource.TestCheckResourceAttr(resourceName, "http_version", "http2"),
					resource.TestCheckResourceAttr(resourceName, "default_root_object", "updated.html"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
		},
	})
}

func testAccCheckMultiTenantDistributionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

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

func testAccCheckMultiTenantDistributionExists(ctx context.Context, t *testing.T, n string, v *awstypes.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

		output, err := tfcloudfront.FindDistributionByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output.Distribution

		return nil
	}
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

func testAccMultiTenantDistributionConfig_comprehensive(rName, comment, defaultRootObject string, compress bool) string {
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
  enabled             = false
  comment             = %[2]q
  default_root_object = %[3]q

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
  }

  default_cache_behavior {
    target_origin_id       = "custom-origin"
    viewer_protocol_policy = "redirect-to-https"
    cache_policy_id        = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad" # AWS Managed CachingDisabled policy

    allowed_methods {
      items          = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
      cached_methods = ["GET", "HEAD"]
    }
  }

  # Single cache behavior
  cache_behavior {
    path_pattern           = "/api/*"
    target_origin_id       = "custom-origin"
    viewer_protocol_policy = "https-only"
    cache_policy_id        = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad"
    compress               = %[4]t

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

  # Simplified tenant config
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

  tags = {
    Environment = "test"
    Name        = %[1]q
  }
}
`, rName, comment, defaultRootObject, compress)
}

func testAccMultiTenantDistributionConfig_customErrorResponseWithoutResponsePagePath() string {
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

  custom_error_response {
    error_code = 404
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

func testAccMultiTenantDistributionConfig_tags(tags map[string]string) string {
	var tagLines []string
	for key, value := range tags {
		tagLines = append(tagLines, fmt.Sprintf("    %q = %q", key, value))
	}
	tagConfig := fmt.Sprintf("  tags = {\n%s\n  }", strings.Join(tagLines, "\n"))

	return fmt.Sprintf(`
resource "aws_cloudfront_multitenant_distribution" "test" {
  enabled = false
  comment = "Test multi-tenant distribution for tags"

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

%s
}
`, tagConfig)
}

func testAccMultiTenantDistributionConfig_update(comment, httpVersion, defaultRootObject string) string {
	defaultRootObjectConfig := ""
	if defaultRootObject != "" {
		defaultRootObjectConfig = fmt.Sprintf("default_root_object = %q", defaultRootObject)
	}

	return fmt.Sprintf(`
resource "aws_cloudfront_multitenant_distribution" "test" {
  enabled      = false
  comment      = %[1]q
  http_version = %[2]q
  %[3]s

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
      cached_methods = ["GET", "HEAD"]
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
          comment  = "Updated origin domain parameter for tenants"
        }
      }
    }
  }
}
`, comment, httpVersion, defaultRootObjectConfig)
}

func testAccMultiTenantDistributionConfig_s3OriginWithOAC(rName string) string {
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
  enabled = true
  comment = "Test Distribution with S3 OAC"

  tenant_config {}

  origin {
    id                       = aws_s3_bucket.test.bucket_regional_domain_name
    domain_name              = aws_s3_bucket.test.bucket_regional_domain_name
    origin_access_control_id = aws_cloudfront_origin_access_control.test.id

    connection_attempts         = 3
    connection_timeout          = 10
    response_completion_timeout = 30
  }

  default_cache_behavior {
    target_origin_id       = aws_s3_bucket.test.bucket_regional_domain_name
    viewer_protocol_policy = "redirect-to-https"
    cache_policy_id        = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad"
    compress               = true

    allowed_methods {
      items          = ["GET", "HEAD"]
      cached_methods = ["GET", "HEAD"]
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
`, rName)
}
