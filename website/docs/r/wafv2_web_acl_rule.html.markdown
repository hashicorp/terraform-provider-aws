---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_web_acl_rule"
description: |-
  Manages an individual rule within a WAFv2 Web ACL.
---

# Resource: aws_wafv2_web_acl_rule

Manages an individual rule within a WAFv2 Web ACL. This resource creates proper Terraform dependencies for safe deletion of referenced resources like IP sets, solving the `WAFAssociatedItemException` error that occurs when deleting IP sets that are still referenced by Web ACL rules.

~> **NOTE:** When using this resource, you must add `lifecycle { ignore_changes = [rule] }` to your `aws_wafv2_web_acl` resource to prevent conflicts.

## Example Usage

### Basic Geo Match Rule

```terraform
resource "aws_wafv2_web_acl" "example" {
  name  = "example"
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "example"
    sampled_requests_enabled   = false
  }

  # Required when using aws_wafv2_web_acl_rule
  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "block_countries" {
  name        = "block-countries"
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.example.arn

  action {
    block {}
  }

  statement {
    geo_match_statement {
      country_codes = ["CN", "RU"]
    }
  }
}
```

### IP Set Reference (Solves Deletion Ordering)

This example demonstrates the primary use case: referencing an IP set in a way that allows safe deletion.

```terraform
resource "aws_wafv2_ip_set" "blocked_ips" {
  name               = "blocked-ips"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.2.3.4/32", "5.6.7.8/32"]
}

resource "aws_wafv2_web_acl" "example" {
  name  = "example"
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "example"
    sampled_requests_enabled   = true
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "block_ips" {
  name        = "block-bad-ips"
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.example.arn

  action {
    block {}
  }

  statement {
    ip_set_reference_statement {
      arn = aws_wafv2_ip_set.blocked_ips.arn
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "block-bad-ips"
    sampled_requests_enabled   = true
  }
}
```

With this configuration, when you remove both the `aws_wafv2_web_acl_rule` and `aws_wafv2_ip_set` resources, Terraform will:

1. Delete the rule first (removing the reference from the Web ACL)
2. Delete the IP set second (now safe because it's no longer referenced)

This prevents the `WAFAssociatedItemException` error.

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the rule. Must be unique within the Web ACL.
* `priority` - (Required) Rule priority. Rules with lower priority are evaluated first.
* `web_acl_arn` - (Required) ARN of the Web ACL to add the rule to.
* `action` - (Required) Action to take when the rule matches. See [Action](#action) below.
* `statement` - (Required) Rule statement. See [Statement](#statement) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `visibility_config` - (Optional) CloudWatch metrics configuration. If not specified, defaults to enabled with the rule name as the metric name. See [Visibility Config](#visibility-config) below.

### Action

One of the following action blocks must be specified:

* `allow` - (Optional) Allow the request.
* `block` - (Optional) Block the request. See [Block](#block) below.
* `count` - (Optional) Count the request without blocking.
* `captcha` - (Optional) Present a CAPTCHA challenge.
* `challenge` - (Optional) Present a silent challenge.

#### Block

* `custom_response` - (Optional) Custom response configuration. See [Custom Response](#custom-response) below.

#### Custom Response

* `response_code` - (Required) HTTP status code to return (200-599).
* `custom_response_body_key` - (Optional) Key of a custom response body defined in the Web ACL.
* `response_header` - (Optional) Custom headers to include in the response. See [Response Header](#response-header) below.

#### Response Header

* `name` - (Required) Header name.
* `value` - (Required) Header value.

### Statement

Exactly one of the following statement blocks must be specified:

* `ip_set_reference_statement` - (Optional) Reference to an IP set. See [IP Set Reference Statement](#ip-set-reference-statement) below.
* `geo_match_statement` - (Optional) Match requests by geographic location. See [Geo Match Statement](#geo-match-statement) below.

#### IP Set Reference Statement

* `arn` - (Required) ARN of the IP set to reference.
* `ip_set_forwarded_ip_config` - (Optional) Configuration for inspecting forwarded IP headers. See [IP Set Forwarded IP Config](#ip-set-forwarded-ip-config) below.

#### IP Set Forwarded IP Config

* `fallback_behavior` - (Required) Action to take when the IP address in the header is invalid. Valid values: `MATCH`, `NO_MATCH`.
* `header_name` - (Required) Name of the header containing the forwarded IP address.
* `position` - (Required) Position in the header to use. Valid values: `FIRST`, `LAST`, `ANY`.

#### Geo Match Statement

* `country_codes` - (Required) List of two-character country codes (ISO 3166-1 alpha-2).
* `forwarded_ip_config` - (Optional) Configuration for inspecting forwarded IP headers. See [Forwarded IP Config](#forwarded-ip-config) below.

#### Forwarded IP Config

* `fallback_behavior` - (Required) Action to take when the IP address in the header is invalid. Valid values: `MATCH`, `NO_MATCH`.
* `header_name` - (Required) Name of the header containing the forwarded IP address.

### Visibility Config

* `cloudwatch_metrics_enabled` - (Optional) Whether to enable CloudWatch metrics. Defaults to `true`.
* `metric_name` - (Optional) Name of the CloudWatch metric. Defaults to the rule name.
* `sampled_requests_enabled` - (Optional) Whether to store sampled requests. Defaults to `true`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Composite ID in the format `web_acl_arn/rule_name`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAFv2 Web ACL Rules using the Web ACL ARN and rule name separated by `/`. For example:

```terraform
import {
  to = aws_wafv2_web_acl_rule.example
  id = "arn:aws:wafv2:us-east-1:123456789012:regional/webacl/example/abc123def456/my-rule"
}
```

Using `terraform import`, import WAFv2 Web ACL Rules using the Web ACL ARN and rule name separated by `/`. For example:

```console
% terraform import aws_wafv2_web_acl_rule.example arn:aws:wafv2:us-east-1:123456789012:regional/webacl/example/abc123def456/my-rule
```
