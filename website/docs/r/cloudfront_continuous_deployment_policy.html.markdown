---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_continuous_deployment_policy"
description: |-
  Terraform resource for managing an AWS CloudFront Continuous Deployment Policy.
---
# Resource: aws_cloudfront_continuous_deployment_policy

Terraform resource for managing an AWS CloudFront Continuous Deployment Policy.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudfront_distribution" "staging" {
  enabled = true
  staging = true

  # ... other configuration ...
}

resource "aws_cloudfront_continuous_deployment_policy" "example" {
  enabled = true

  staging_distribution_dns_names {
    items    = [aws_cloudfront_distribution.staging.domain_name]
    quantity = 1
  }

  traffic_config {
    type = "SingleWeight"
    single_weight_config {
      weight = "0.01"
    }
  }
}

resource "aws_cloudfront_distribution" "production" {
  enabled = true

  # NOTE: A continuous deployment policy cannot be associated to distribution
  # on creation. Set this argument once the resource exists.
  continuous_deployment_policy_id = aws_cloudfront_continuous_deployment_policy.example.id

  # ... other configuration ...
}
```

### Single Weight Config with Session Stickiness

```terraform
resource "aws_cloudfront_continuous_deployment_policy" "example" {
  enabled = true

  staging_distribution_dns_names {
    items    = [aws_cloudfront_distribution.staging.domain_name]
    quantity = 1
  }

  traffic_config {
    type = "SingleWeight"
    single_weight_config {
      weight = "0.01"
      session_stickiness_config {
        idle_ttl    = 300
        maximum_ttl = 600
      }
    }
  }
}
```

### Single Header Config

```terraform
resource "aws_cloudfront_continuous_deployment_policy" "example" {
  enabled = true

  staging_distribution_dns_names {
    items    = [aws_cloudfront_distribution.staging.domain_name]
    quantity = 1
  }

  traffic_config {
    type = "SingleHeader"
    single_header_config {
      header = "aws-cf-cd-example"
      value  = "example"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `enabled` - (Required) Whether this continuous deployment policy is enabled.
* `staging_distribution_dns_names` - (Required) CloudFront domain name of the staging distribution. See [`staging_distribution_dns_names`](#staging_distribution_dns_names).
* `traffic_config` - (Required) Parameters for routing production traffic from primary to staging distributions. See [`traffic_config`](#traffic_config).

### `staging_distribution_dns_names`

* `items` - (Required) A list of CloudFront domain names for the staging distribution.
* `quantity` - (Required) Number of CloudFront domain names in the staging distribution.

### `traffic_config`

* `type` - (Required) Type of traffic configuration. Valid values are `SingleWeight` and `SingleHeader`.
* `single_header_config` - (Optional) Determines which HTTP requests are sent to the staging distribution. See [`single_header_config`](#single_header_config).
* `single_weight_config` - (Optional) Contains the percentage of traffic to send to the staging distribution. See [`single_weight_config`](#single_weight_config).

### `single_header_config`

* `header` - (Required) Request header name to send to the staging distribution. The header must contain the prefix `aws-cf-cd-`.
* `value` - (Required) Request header value.

### `single_weight_config`

* `weight` - (Required) The percentage of traffic to send to a staging distribution, expressed as a decimal number between `0` and `.15`.
* `session_stickiness_config` - (Optional) Session stickiness provides the ability to define multiple requests from a single viewer as a single session. This prevents the potentially inconsistent experience of sending some of a given user's requests to the staging distribution, while others are sent to the primary distribution. Define the session duration using TTL values. See [`session_stickiness_config`](#session_stickiness_config).

### `session_stickiness_config`

* `idle_ttl` - (Required) The amount of time in seconds after which sessions will cease if no requests are received. Valid values are `300` – `3600` (5–60 minutes). The value must be less than or equal to `maximum_ttl`.
* `maximum_ttl` - (Required) The maximum amount of time in seconds to consider requests from the viewer as being part of the same session. Valid values are `300` – `3600` (5–60 minutes). The value must be greater than or equal to `idle_ttl`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The continuous deployment policy ARN.
* `etag` - Current version of the continuous distribution policy.
* `id` - Identifier of the continuous deployment policy.
* `last_modified_time` - Date and time the continuous deployment policy was last modified.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudFront Continuous Deployment Policy using the `id`. For example:

```terraform
import {
  to = aws_cloudfront_continuous_deployment_policy.example
  id = "abcd-1234"
}
```

Using `terraform import`, import CloudFront Continuous Deployment Policy using the `id`. For example:

```console
% terraform import aws_cloudfront_continuous_deployment_policy.example abcd-1234 
```
