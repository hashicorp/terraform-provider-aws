---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_distribution_tenant"
description: |-
  Provides a CloudFront distribution tenant resource.
---

# Resource: aws_cloudfront_distribution_tenant

Creates an Amazon CloudFront distribution tenant.

Distribution tenants allow you to create isolated configurations within a multi-tenant CloudFront distribution. Each tenant can have its own domains, customizations, and parameters while sharing the underlying distribution infrastructure.

For information about CloudFront distribution tenants, see the [Amazon CloudFront Developer Guide](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/distribution-tenants.html).

## Example Usage

### Basic Distribution Tenant

```terraform
resource "aws_cloudfront_distribution_tenant" "example" {
  name            = "example-tenant"
  distribution_id = aws_cloudfront_distribution.multi_tenant.id
  enabled         = true

  domain {
    domain = "tenant.example.com"
  }

  tags = {
    Environment = "production"
  }
}
```

### Distribution Tenant with Customizations

```terraform
resource "aws_cloudfront_distribution_tenant" "example" {
  name            = "example-tenant"
  distribution_id = aws_cloudfront_distribution.multi_tenant.id
  enabled         = false

  domain {
    domain = "tenant.example.com"
  }

  customizations {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA"]
    }

    certificate {
      arn = aws_acm_certificate.tenant_cert.arn
    }

    web_acl {
      action = "override"
      arn    = aws_wafv2_web_acl.tenant_waf.arn
    }
  }

  tags = {
    Environment = "production"
    Tenant      = "example"
  }
}
```

### Distribution Tenant with Managed Certificate

```terraform
resource "aws_cloudfront_distribution_tenant" "main" {
  distribution_id     = aws_cloudfront_distribution.main.id
  name                = "main-tenant"
  enabled             = false
  connection_group_id = aws_cloudfront_connection_group.main_group.id

  domain {
    domain = "tenant.example.com"
  }

  managed_certificate_request {
    primary_domain_name                         = "app.example.com"
    validation_token_host                       = "cloudfront"
    certificate_transparency_logging_preference = "disabled"
  }
}

data "aws_route53_zone" "main" {
  name         = "example.com"
  private_zone = false
}

resource "aws_cloudfront_connection_group" "main_group" {
  name = "main-group"
}

resource "aws_route53_record" "domain_record" {
  zone_id = data.aws_route53_zone.main.id
  type    = "CNAME"
  ttl     = 300
  name    = "app.example.com"
  records = [aws_cloudfront_connection_group.main_group.routing_endpoint]
}

resource "aws_cloudfront_cache_policy" "main_policy" {
  name        = "main-policy"
  comment     = "tenant cache policy"
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

resource "aws_cloudfront_distribution" "main" {
  connection_mode = "tenant-only"
  enabled         = true

  origin {
    domain_name = "www.example.com"
    origin_id   = "main"

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
    target_origin_id       = "main"
    viewer_protocol_policy = "allow-all"
    cache_policy_id        = aws_cloudfront_cache_policy.main_policy.id
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
```

## Argument Reference

This resource supports the following arguments:

* `name` (Required) - Name of the distribution tenant.
* `distribution_id` (Required) - ID of the multi-tenant distribution.
* `domain` (Required) - Set of domains associated with the distribution tenant.
* `enabled` (Optional) - Whether the distribution tenant is enabled to serve traffic. Defaults to `true`.
* `connection_group_id` (Optional) - ID of the connection group for the distribution tenant. If not specified, CloudFront uses the default connection group.
* `customizations` (Optional) - [Customizations](#customizations-arguments) for the distribution tenant (maximum one).
* `managed_certificate_request` (Optional) - [Managed certificate request](#managed-certificate-request-arguments) for CloudFront managed ACM certificate (maximum one).
* `parameter` (Optional) - Set of [parameter](#parameter-arguments) values for the distribution tenant.
* `wait_for_deployment` (Optional) - If enabled, the resource will wait for the distribution tenant status to change from `InProgress` to `Deployed`. Setting this to `false` will skip the process. Default: `true`.
* `tags` (Optional) - Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

#### Customizations Arguments

* `certificate` (Optional) - [Certificate](#certificate-arguments) configuration for the tenant (maximum one).
* `geo_restriction` (Optional) - [Geographic restrictions](#geo-restriction-arguments) configuration for the tenant (maximum one).
* `web_acl` (Optional) - [Web ACL](#web-acl-arguments) configuration for the tenant (maximum one).

##### Certificate Arguments

* `arn` (Optional) - ARN of the AWS Certificate Manager certificate to use with this distribution tenant.

##### Geo Restriction Arguments

* `restriction_type` (Optional) - Method to restrict distribution by country: `none`, `whitelist`, or `blacklist`.
* `locations` (Optional) - Set of ISO 3166-1-alpha-2 country codes for the restriction. Required if `restriction_type` is `whitelist` or `blacklist`.

##### Web ACL Arguments

* `action` (Optional) - Action to take for the web ACL. Valid values: `allow`, `block`.
* `arn` (Optional) - ARN of the AWS WAF web ACL to associate with this distribution tenant.

#### Managed Certificate Request Arguments

* `certificate_transparency_logging_preference` (Optional) - Certificate transparency logging preference. Valid values: `enabled`, `disabled`.
* `primary_domain_name` (Optional) - Primary domain name for the certificate.
* `validation_token_host` (Optional) - Host for validation token. Valid values: `cloudfront`, `domain`.

#### Parameter Arguments

* `name` (Required) - Name of the parameter.
* `value` (Required) - Value of the parameter.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the distribution tenant.
* `arn` - ARN of the distribution tenant.
* `status` - Current status of the distribution tenant.
* `etag` - Current version of the distribution tenant.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudFront Distribution Tenants using the `id`. For example:

```terraform
import {
  to = aws_cloudfront_distribution_tenant.example
  id = "TENANT123EXAMPLE"
}
```

Using `terraform import`, import CloudFront Distribution Tenants using the `id`. For example:

```console
% terraform import aws_cloudfront_distribution_tenant.example TENANT123EXAMPLE
```
