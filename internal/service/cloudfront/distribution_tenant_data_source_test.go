// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontDistributionTenantDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudfront_distribution_tenant.test"
	resourceName := "aws_cloudfront_distribution_tenant.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "connection_group_id", resourceName, "connection_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "customizations.#", resourceName, "customizations.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution_id", resourceName, "distribution_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domains.#", resourceName, "domains.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEnabled, resourceName, names.AttrEnabled),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_modified_time", resourceName, "last_modified_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "managed_certificate_request.#", resourceName, "managed_certificate_request.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "parameters.#", resourceName, "parameters.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
				),
			},
		},
	})
}

func TestAccCloudFrontDistributionTenantDataSource_byARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudfront_distribution_tenant.test"
	resourceName := "aws_cloudfront_distribution_tenant.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantDataSourceConfig_byARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "connection_group_id", resourceName, "connection_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "customizations.#", resourceName, "customizations.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution_id", resourceName, "distribution_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domains.#", resourceName, "domains.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEnabled, resourceName, names.AttrEnabled),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_modified_time", resourceName, "last_modified_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "managed_certificate_request.#", resourceName, "managed_certificate_request.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "parameters.#", resourceName, "parameters.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
				),
			},
		},
	})
}

func TestAccCloudFrontDistributionTenantDataSource_byName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudfront_distribution_tenant.test"
	resourceName := "aws_cloudfront_distribution_tenant.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantDataSourceConfig_byName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "connection_group_id", resourceName, "connection_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "customizations.#", resourceName, "customizations.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution_id", resourceName, "distribution_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domains.#", resourceName, "domains.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEnabled, resourceName, names.AttrEnabled),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_modified_time", resourceName, "last_modified_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "managed_certificate_request.#", resourceName, "managed_certificate_request.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "parameters.#", resourceName, "parameters.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
				),
			},
		},
	})
}

func TestAccCloudFrontDistributionTenantDataSource_byDomain(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudfront_distribution_tenant.test"
	resourceName := "aws_cloudfront_distribution_tenant.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantDataSourceConfig_byDomain(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "connection_group_id", resourceName, "connection_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "customizations.#", resourceName, "customizations.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution_id", resourceName, "distribution_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domains.#", resourceName, "domains.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEnabled, resourceName, names.AttrEnabled),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_modified_time", resourceName, "last_modified_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "managed_certificate_request.#", resourceName, "managed_certificate_request.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "parameters.#", resourceName, "parameters.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
				),
			},
		},
	})
}

func testAccDistributionTenantDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain      = "example.com"
  most_recent = true
}

resource "aws_cloudfront_connection_group" "test" {
  name = %[1]q
}

resource "aws_cloudfront_cache_policy" "test" {
  name        = %[1]q
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
    cache_policy_id        = aws_cloudfront_cache_policy.test.id
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
  distribution_id     = aws_cloudfront_distribution.test.id
  domains             = ["example.com"]
  name                = %[1]q
  enabled             = false
  connection_group_id = aws_cloudfront_connection_group.test.id
}

data "aws_cloudfront_distribution_tenant" "test" {
  id = aws_cloudfront_distribution_tenant.test.id
}
`, rName))
}

func testAccDistributionTenantDataSourceConfig_byARN(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain      = "example.com"
  most_recent = true
}

resource "aws_cloudfront_connection_group" "test" {
  name = %[1]q
}

resource "aws_cloudfront_cache_policy" "test" {
  name        = %[1]q
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
    cache_policy_id        = aws_cloudfront_cache_policy.test.id
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
  distribution_id     = aws_cloudfront_distribution.test.id
  domains             = ["example.com"]
  name                = %[1]q
  enabled             = false
  connection_group_id = aws_cloudfront_connection_group.test.id
}

data "aws_cloudfront_distribution_tenant" "test" {
  arn = aws_cloudfront_distribution_tenant.test.arn
}
`, rName))
}

func testAccDistributionTenantDataSourceConfig_byName(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain      = "example.com"
  most_recent = true
}

resource "aws_cloudfront_connection_group" "test" {
  name = %[1]q
}

resource "aws_cloudfront_cache_policy" "test" {
  name        = %[1]q
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
    cache_policy_id        = aws_cloudfront_cache_policy.test.id
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
  distribution_id     = aws_cloudfront_distribution.test.id
  domains             = ["example.com"]
  name                = %[1]q
  enabled             = false
  connection_group_id = aws_cloudfront_connection_group.test.id
}

data "aws_cloudfront_distribution_tenant" "test" {
  name = aws_cloudfront_distribution_tenant.test.name
}
`, rName))
}

func testAccDistributionTenantDataSourceConfig_byDomain(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain      = "example.com"
  most_recent = true
}

resource "aws_cloudfront_connection_group" "test" {
  name = %[1]q
}

resource "aws_cloudfront_cache_policy" "test" {
  name        = %[1]q
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
    cache_policy_id        = aws_cloudfront_cache_policy.test.id
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
  distribution_id     = aws_cloudfront_distribution.test.id
  domains             = ["example-%[1]s.com"]
  name                = %[1]q
  enabled             = false
  connection_group_id = aws_cloudfront_connection_group.test.id
}

data "aws_cloudfront_distribution_tenant" "test" {
  domain = "example-%[1]s.com"
}
`, rName))
}
