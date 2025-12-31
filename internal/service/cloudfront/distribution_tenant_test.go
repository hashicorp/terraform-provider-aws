// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	ret "github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontDistributionTenant_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var tenant awstypes.DistributionTenant
	resourceName := "aws_cloudfront_distribution_tenant.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionTenantDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantConfig_basic(t),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionTenantExists(ctx, resourceName, &tenant),
					resource.TestCheckResourceAttrSet(resourceName, "connection_group_id"),
					resource.TestCheckResourceAttr(resourceName, "domains.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "distribution_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "cloudfront", regexache.MustCompile(`distribution-tenant/[0-9A-Z]+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
				),
			},
			{
				Config:            testAccDistributionTenantConfig_basic(t),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"wait_for_deployment",
				},
			},
		},
	})
}
func TestAccCloudFrontDistributionTenant_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var tenant awstypes.DistributionTenant
	resourceName := "aws_cloudfront_distribution_tenant.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionTenantDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantConfig_basic(t),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionTenantExists(ctx, resourceName, &tenant),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceDistributionTenant(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontDistributionTenant_customCertificate(t *testing.T) {
	ctx := acctest.Context(t)
	var tenant awstypes.DistributionTenant
	resourceName := "aws_cloudfront_distribution_tenant.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionTenantDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantConfig_customCertificate(t),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionTenantExists(ctx, resourceName, &tenant),
					resource.TestCheckResourceAttr(resourceName, "customizations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "customizations.0.geo_restriction.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "customizations.0.geo_restriction.0.restriction_type", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "customizations.0.geo_restriction.0.locations.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "customizations.0.certificate.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "customizations.0.certificate.0.arn", "data.aws_acm_certificate.test", names.AttrARN),
				),
			},
		},
	})
}

func TestAccCloudFrontDistributionTenant_customCertificateWithWebACL(t *testing.T) {
	ctx := acctest.Context(t)
	var tenant awstypes.DistributionTenant
	resourceName := "aws_cloudfront_distribution_tenant.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionTenantDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantConfig_customCertificateWithWebACL(t),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionTenantExists(ctx, resourceName, &tenant),
					resource.TestCheckResourceAttr(resourceName, "customizations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "customizations.0.web_acl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "customizations.0.web_acl.0.action", "override"),
					resource.TestCheckResourceAttrPair(resourceName, "customizations.0.web_acl.0.arn", "aws_wafv2_web_acl.testacl", names.AttrARN),
				),
			},
		},
	})
}

func TestAccCloudFrontDistributionTenant_parameters(t *testing.T) {
	ctx := acctest.Context(t)
	var tenant awstypes.DistributionTenant
	resourceName := "aws_cloudfront_distribution_tenant.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionTenantDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantConfig_parameters(t),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionTenantExists(ctx, resourceName, &tenant),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.name", "place"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.value", "na"),
					resource.TestCheckResourceAttr(resourceName, "parameters.1.name", "tenantid"),
					resource.TestCheckResourceAttr(resourceName, "parameters.1.value", "tenant-123"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccCloudFrontDistributionTenant_managedCertificateRequest(t *testing.T) {
	ctx := acctest.Context(t)
	var tenant awstypes.DistributionTenant
	resourceName := "aws_cloudfront_distribution_tenant.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionTenantDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantConfig_managedCertificateRequest(t),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionTenantExists(ctx, resourceName, &tenant),
					resource.TestCheckResourceAttrSet(resourceName, "connection_group_id"),
					resource.TestCheckResourceAttr(resourceName, "managed_certificate_request.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "managed_certificate_request.0.primary_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "managed_certificate_request.0.validation_token_host", "cloudfront"),
					resource.TestCheckResourceAttr(resourceName, "managed_certificate_request.0.certificate_transparency_logging_preference", "disabled"),
				),
			},
		},
	})
}

func TestAccCloudFrontDistributionTenant_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var tenant awstypes.DistributionTenant
	resourceName := "aws_cloudfront_distribution_tenant.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionTenantDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantConfig_tags1(t, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionTenantExists(ctx, resourceName, &tenant),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccDistributionTenantConfig_tags2(t, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionTenantExists(ctx, resourceName, &tenant),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDistributionTenantConfig_tags1(t, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionTenantExists(ctx, resourceName, &tenant),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckDistributionTenantDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_distribution_tenant" {
				continue
			}

			_, err := tfcloudfront.FindDistributionTenantById(ctx, conn, rs.Primary.ID)

			if ret.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Distribution Tenant (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDistributionTenantExists(ctx context.Context, n string, v *awstypes.DistributionTenant) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		output, err := tfcloudfront.FindDistributionTenantById(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output.DistributionTenant

		return nil
	}
}

func testAccDistributionTenantConfig_basic(t *testing.T) string {
	certDomain := acctest.ACMCertificateDomainFromEnv(t)
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_acm_certificate" "test" {
  domain      = %[1]q
  region      = "data.aws_region.current.region"
  most_recent = true
}

resource "aws_cloudfront_cache_policy" "tf-policy" {
  name        = "tfpolicy"
  comment     = "test tenant cache policy"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1
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

resource "aws_cloudfront_distribution" "test" {
  connection_mode = "tenant-only"
  enabled         = true

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"
    cache_policy_id        = aws_cloudfront_cache_policy.tf-policy.id
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn = data.aws_acm_certificate.test.arn
    ssl_support_method  = "sni-only"
  }
}

resource "aws_cloudfront_distribution_tenant" "test" {
  distribution_id = aws_cloudfront_distribution.test.id
  domains         = [%[1]q]
  name            = "tftenant"
  enabled         = false
}
`, certDomain)
}

func testAccDistributionTenantConfig_customCertificate(t *testing.T) string {
	certDomain := acctest.ACMCertificateDomainFromEnv(t)
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_acm_certificate" "test" {
  domain      = %[1]q
  region      = "data.aws_region.current.region"
  most_recent = true
}

resource "aws_cloudfront_cache_policy" "tf-policy" {
  name        = "tfpolicy"
  comment     = "test tenant cache policy"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1
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

resource "aws_cloudfront_distribution" "test" {
  connection_mode = "tenant-only"
  enabled         = true

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"
    cache_policy_id        = aws_cloudfront_cache_policy.tf-policy.id
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

resource "aws_cloudfront_distribution_tenant" "test" {
  distribution_id = aws_cloudfront_distribution.test.id
  domains         = [%[1]q]
  name            = "tftenant"
  enabled         = false

  customizations {
    geo_restriction {
      locations        = ["US", "CA"]
      restriction_type = "whitelist"
    }

    certificate {
      arn = data.aws_acm_certificate.test.arn
    }
  }
}
`, certDomain)
}

func testAccDistributionTenantConfig_customCertificateWithWebACL(t *testing.T) string {
	certDomain := acctest.ACMCertificateDomainFromEnv(t)
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_acm_certificate" "test" {
  domain      = %[1]q
  region      = "data.aws_region.current.region"
  most_recent = true
}

resource "aws_wafv2_web_acl" "testacl" {
  name        = "tftest"
  description = "tftest"
  scope       = "CLOUDFRONT"
  region      = "data.aws_region.current.region"

  default_action {
    allow {
      custom_request_handling {
        insert_header {
          name  = "X-WebACL-Test"
          value = "test"
        }
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}

resource "aws_cloudfront_cache_policy" "tf-policy" {
  name        = "tfpolicy"
  comment     = "test tenant cache policy"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1
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

resource "aws_cloudfront_distribution" "test" {
  connection_mode = "tenant-only"
  enabled         = true

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"
    cache_policy_id        = aws_cloudfront_cache_policy.tf-policy.id
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

resource "aws_cloudfront_distribution_tenant" "test" {
  distribution_id = aws_cloudfront_distribution.test.id
  domains         = [%[1]q]
  name            = "tftenant"
  enabled = false

  customizations {
    geo_restriction {
      locations        = ["US", "CA"]
      restriction_type = "whitelist"
    }

    certificate {
      arn = data.aws_acm_certificate.test.arn
    }

  	web_acl {
	    action = "override"
	    arn    = aws_wafv2_web_acl.testacl.arn
	  }
  }
}
`, certDomain)
}

func testAccDistributionTenantConfig_parameters(t *testing.T) string {
	certDomain := acctest.ACMCertificateDomainFromEnv(t)
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_acm_certificate" "test" {
  domain      = %[1]q
  region      = "data.aws_region.current.region"
  most_recent = true
}

resource "aws_cloudfront_cache_policy" "tf-policy" {
  name        = "tfpolicy"
  comment     = "test tenant cache policy"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1
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

resource "aws_cloudfront_distribution" "test" {
  connection_mode = "tenant-only"
  enabled         = true

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"
    cache_policy_id        = aws_cloudfront_cache_policy.tf-policy.id
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn = data.aws_acm_certificate.test.arn
    ssl_support_method  = "sni-only"
  }
}

resource "aws_cloudfront_distribution_tenant" "test" {
  distribution_id = aws_cloudfront_distribution.test.id
  domains         = [%[1]q]
  name            = "tftenant"
  enabled         = false

  parameters {
		name  = "tenantid"
		value = "tenant-123"
  }

  parameters {
	  name  = "place"
	  value = "na"
  }
}
`, certDomain)
}

func testAccDistributionTenantConfig_tags1(t *testing.T, tagKey1, tagValue1 string) string {
	certDomain := acctest.ACMCertificateDomainFromEnv(t)
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_acm_certificate" "test" {
  domain      = %[1]q
  region      = "data.aws_region.current.region"
  most_recent = true
}

resource "aws_cloudfront_cache_policy" "tf-policy" {
  name        = "tfpolicy"
  comment     = "test tenant cache policy"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1
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

resource "aws_cloudfront_distribution" "test" {
  connection_mode = "tenant-only"
  enabled         = true

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"
    cache_policy_id        = aws_cloudfront_cache_policy.tf-policy.id
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn = data.aws_acm_certificate.test.arn
    ssl_support_method  = "sni-only"
  }
}

resource "aws_cloudfront_distribution_tenant" "test" {
  distribution_id = aws_cloudfront_distribution.test.id
  domains         = [%[1]q]
  name            = "tftenant"
  enabled         = false

  tags = {
    %[2]q = %[3]q
  }
}
`, certDomain, tagKey1, tagValue1)
}

func testAccDistributionTenantConfig_tags2(t *testing.T, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	certDomain := acctest.ACMCertificateDomainFromEnv(t)
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_acm_certificate" "test" {
  domain      = %[1]q
  region      = "data.aws_region.current.region"
  most_recent = true
}

resource "aws_cloudfront_cache_policy" "tf-policy" {
  name        = "tfpolicy"
  comment     = "test tenant cache policy"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1
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

resource "aws_cloudfront_distribution" "test" {
  connection_mode = "tenant-only"
  enabled         = true

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"
    cache_policy_id        = aws_cloudfront_cache_policy.tf-policy.id
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn = data.aws_acm_certificate.test.arn
    ssl_support_method  = "sni-only"
  }
}

resource "aws_cloudfront_distribution_tenant" "test" {
  distribution_id = aws_cloudfront_distribution.test.id
  domains         = [%[1]q]
  name            = "tftenant"
  enabled         = false

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, certDomain, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDistributionTenantConfig_managedCertificateRequest(t *testing.T) string {
	zoneDomain := acctest.ACMCertificateDomainFromEnv(t)
	domainName := "tf." + zoneDomain
	return fmt.Sprintf(`
resource "aws_cloudfront_cache_policy" "tf-policy" {
  name        = "tfpolicy"
  comment     = "test tenant cache policy"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1
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

resource "aws_cloudfront_distribution" "test" {
  connection_mode = "tenant-only"
  enabled         = true

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"
    cache_policy_id        = aws_cloudfront_cache_policy.tf-policy.id
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  depends_on = [aws_route53_record.testrecord]
}

data "aws_route53_zone" "test" {
	name         = %[1]q
	private_zone = false
}

resource "aws_cloudfront_connection_group" "testgroup" {
  name = "testgroup"
}

resource "aws_route53_record" "testrecord" {
	zone_id = data.aws_route53_zone.test.id
	type    = "CNAME"
	ttl     = 300
	name    = %[2]q
	records = [aws_cloudfront_connection_group.testgroup.routing_endpoint]
}

resource "aws_cloudfront_distribution_tenant" "test" {
  distribution_id     = aws_cloudfront_distribution.test.id
  domains             = [%[2]q]
  name                = "tftenant"
  enabled             = false
  connection_group_id = aws_cloudfront_connection_group.testgroup.id

  managed_certificate_request {
	  primary_domain_name                         = %[2]q
	  validation_token_host                       = "cloudfront"
	  certificate_transparency_logging_preference = "disabled"
  }
}
`, zoneDomain, domainName)
}
