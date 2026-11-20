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
  distribution_id = aws_cloudfront_multitenant_distribution.example.id
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
  distribution_id = aws_cloudfront_multitenant_distribution.example.id
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

### Move a Domain from Another Distribution Tenant

`on_domain_takeover = "allow"` moves a domain that is already associated with another standard distribution or distribution tenant to this distribution tenant in the same request. A DNS record must prove that you control the domain and point it at this distribution tenant, for example a CNAME record to the routing endpoint of its connection group.

```terraform
data "aws_cloudfront_connection_group" "default" {
  id = aws_cloudfront_distribution_tenant.existing.connection_group_id
}

resource "aws_route53_record" "tenant" {
  zone_id = data.aws_route53_zone.example.zone_id
  name    = "tenant.example.com"
  type    = "CNAME"
  ttl     = 300
  records = [data.aws_cloudfront_connection_group.default.routing_endpoint]
}

resource "aws_cloudfront_distribution_tenant" "example" {
  name               = "example-tenant"
  distribution_id    = aws_cloudfront_multitenant_distribution.example.id
  on_domain_takeover = "allow"

  domain {
    domain = "tenant.example.com"
  }

  parameter {
    name  = "origin_domain"
    value = "origin.example.com"
  }

  depends_on = [aws_route53_record.tenant]
}
```

~> **NOTE:** CloudFront removes a moved domain from its source. If Terraform also manages the source distribution or distribution tenant, remove the domain from its configuration too, or Terraform plans to add it back.

## Argument Reference

This resource supports the following arguments:

* `name` (Required) - Name of the distribution tenant.
* `distribution_id` (Required) - ID of the multi-tenant distribution.
* `domain` (Required) - Set of [domains](#domain-arguments) associated with the distribution tenant.
* `enabled` (Optional) - Whether the distribution tenant is enabled to serve traffic. Defaults to `true`.
* `connection_group_id` (Optional) - ID of the connection group for the distribution tenant. If not specified, CloudFront uses the default connection group.
* `customizations` (Optional) - [Customizations](#customizations-arguments) for the distribution tenant (maximum one).
* `on_domain_takeover` (Optional) - Whether CloudFront moves domains in `domain` that are already associated with another standard distribution or distribution tenant to this distribution tenant when it creates or updates it. Valid values: `allow`, `reject`. CloudFront defaults to `reject`, which fails the request with `CNAMEAlreadyExists`. CloudFront does not return this value, so Terraform does not detect drift for it.
* `managed_certificate_request` (Optional) - [Managed certificate request](#managed-certificate-request-arguments) for CloudFront managed ACM certificate (maximum one).
* `parameter` (Optional) - Set of [parameter](#parameter-arguments) values for the distribution tenant.
* `wait_for_deployment` (Optional) - If enabled, the resource will wait for the distribution tenant status to change from `InProgress` to `Deployed`. Setting this to `false` will skip the process. Default: `true`.
* `tags` (Optional) - Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

#### Domain Arguments

* `domain` (Required) - Domain name.
* `serve_on` (Optional) - Which requests for this domain CloudFront routes to the distribution tenant. Valid values: `any` (requests that arrive on any IP address), `anycast-ip-list-only` (only requests that arrive on the Anycast static IPs of the tenant's connection group; the domain must be a lowercase wildcard domain such as `*.example.com`). If not set, CloudFront uses `any`.

#### Customizations Arguments

* `certificate` (Optional) - [Certificate](#certificate-arguments) configuration for the tenant (maximum one).
* `geo_restriction` (Optional) - [Geographic restrictions](#geo-restriction-arguments) configuration for the tenant (maximum one).
* `web_acl` (Optional) - [Web ACL](#web-acl-arguments) configuration for the tenant (maximum one).

#### Certificate Arguments

* `arn` (Optional) - ARN of the AWS Certificate Manager certificate to use with this distribution tenant.

#### Geo Restriction Arguments

* `restriction_type` (Optional) - Method to restrict distribution by country: `none`, `whitelist`, or `blacklist`.
* `locations` (Optional) - Set of ISO 3166-1-alpha-2 country codes for the restriction. Required if `restriction_type` is `whitelist` or `blacklist`.

#### Web ACL Arguments

* `action` (Optional) - Action to take for the web ACL. Valid values: `allow`, `block`.
* `arn` (Optional) - ARN of the AWS WAF web ACL to associate with this distribution tenant.

#### Managed Certificate Request Arguments

* `certificate_transparency_logging_preference` (Optional) - Certificate transparency logging preference. Valid values: `enabled`, `disabled`.
* `primary_domain_name` (Optional) - Primary domain name for the certificate.
* `validation_token_host` (Optional) - Host for validation token. Valid values: `cloudfront`, `self-hosted`.

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

## Timeouts

~> **NOTE:** In versions prior to `v6.51.0`, resource timeouts were accepted but not enforced. Timeouts are now fully enforced. If you have previously configured custom timeouts, they will now take effect.

~> **NOTE:** When using `managed_certificate_request`, the managed certificate issuance uses a fixed 3 hour timeout regardless of the configured resource timeout. This behavior will be updated in a future major version.

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `45m`)
* `update` - (Default `45m`)
* `delete` - (Default `45m`)

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
