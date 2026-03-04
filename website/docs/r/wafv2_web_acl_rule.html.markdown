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

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "block-countries"
    sampled_requests_enabled   = false
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

### Rate-Based Rule

```terraform
resource "aws_wafv2_web_acl_rule" "rate_limit" {
  name        = "rate-limit"
  priority    = 2
  web_acl_arn = aws_wafv2_web_acl.example.arn

  action {
    block {}
  }

  statement {
    rate_based_statement {
      limit              = 2000
      aggregate_key_type = "IP"
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "rate-limit"
    sampled_requests_enabled   = true
  }
}
```

### Managed Rule Group with Override Action

```terraform
resource "aws_wafv2_web_acl_rule" "aws_managed_rules" {
  name        = "aws-managed-rules"
  priority    = 3
  web_acl_arn = aws_wafv2_web_acl.example.arn

  override_action {
    none {}
  }

  statement {
    managed_rule_group_statement {
      name        = "AWSManagedRulesCommonRuleSet"
      vendor_name = "AWS"
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "aws-managed-rules"
    sampled_requests_enabled   = true
  }
}
```

### Custom Request Handling

```terraform
resource "aws_wafv2_web_acl_rule" "captcha_with_headers" {
  name        = "captcha-with-headers"
  priority    = 4
  web_acl_arn = aws_wafv2_web_acl.example.arn

  action {
    captcha {
      custom_request_handling {
        insert_header {
          name  = "x-captcha-rule"
          value = "triggered"
        }
      }
    }
  }

  statement {
    geo_match_statement {
      country_codes = ["US"]
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "captcha-with-headers"
    sampled_requests_enabled   = true
  }
}
```

## Example Usage - IP Set Reference

```terraform
resource "aws_wafv2_web_acl_rule" "blocked_ips" {
  name        = "blocked-ips"
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
* `statement` - (Required) Rule statement. See [Statement](#statement) below.
* `visibility_config` - (Required) CloudWatch metrics configuration. See [Visibility Config](#visibility-config) below.

The following arguments are optional:

* `action` - (Optional) Action to take when the rule matches. See [Action](#action) below. Conflicts with `override_action`.
* `override_action` - (Optional) Override action for managed rule groups. See [Override Action](#override-action) below. Conflicts with `action`.
* `captcha_config` - (Optional) CAPTCHA configuration that overrides the web ACL level setting. See [Captcha Config](#captcha-config) below.
* `challenge_config` - (Optional) Challenge configuration that overrides the web ACL level setting. See [Challenge Config](#challenge-config) below.
* `rule_label` - (Optional) Labels to apply to matching web requests. See [Rule Label](#rule-label) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### Action

One of the following action blocks must be specified:

* `allow` - (Optional) Allow the request. See [Allow](#allow) below.
* `block` - (Optional) Block the request. See [Block](#block) below.
* `count` - (Optional) Count the request without blocking. See [Count](#count) below.
* `captcha` - (Optional) Present a CAPTCHA challenge. See [Captcha](#captcha) below.
* `challenge` - (Optional) Present a silent challenge. See [Challenge](#challenge) below.

#### Allow

* `custom_request_handling` - (Optional) Custom request handling configuration. See [Custom Request Handling](#custom-request-handling) below.

#### Block

* `custom_response` - (Optional) Custom response configuration. See [Custom Response](#custom-response) below.

#### Count

* `custom_request_handling` - (Optional) Custom request handling configuration. See [Custom Request Handling](#custom-request-handling) below.

#### Captcha

* `custom_request_handling` - (Optional) Custom request handling configuration. See [Custom Request Handling](#custom-request-handling) below.

#### Challenge

* `custom_request_handling` - (Optional) Custom request handling configuration. See [Custom Request Handling](#custom-request-handling) below.

#### Custom Request Handling

* `insert_header` - (Optional) Custom headers to insert into the request. See [Insert Header](#insert-header) below.

#### Insert Header

* `name` - (Required) Header name.
* `value` - (Required) Header value.

#### Custom Response

* `response_code` - (Required) HTTP status code to return (200-599).
* `custom_response_body_key` - (Optional) Key of a custom response body defined in the Web ACL.
* `response_header` - (Optional) Custom headers to include in the response. See [Response Header](#response-header) below.

#### Response Header

* `name` - (Required) Header name.
* `value` - (Required) Header value.

### Statement

Exactly one of the following statement blocks must be specified:

* `asn_match_statement` - (Optional) Match requests based on Autonomous System Number (ASN). See [ASN Match Statement](#asn-match-statement) below.
* `byte_match_statement` - (Optional) Match requests based on byte patterns. See [Byte Match Statement](#byte-match-statement) below.
* `geo_match_statement` - (Optional) Match requests by geographic location. See [Geo Match Statement](#geo-match-statement) below.
* `ip_set_reference_statement` - (Optional) Reference to an IP set. See [IP Set Reference Statement](#ip-set-reference-statement) below.
* `label_match_statement` - (Optional) Match requests based on labels. See [Label Match Statement](#label-match-statement) below.
* `managed_rule_group_statement` - (Optional) Reference to a managed rule group. See [Managed Rule Group Statement](#managed-rule-group-statement) below.
* `rate_based_statement` - (Optional) Rate-based rule to track request rates. See [Rate Based Statement](#rate-based-statement) below.
* `regex_match_statement` - (Optional) Match requests using regex patterns. See [Regex Match Statement](#regex-match-statement) below.
* `regex_pattern_set_reference_statement` - (Optional) Reference to a regex pattern set. See [Regex Pattern Set Reference Statement](#regex-pattern-set-reference-statement) below.
* `rule_group_reference_statement` - (Optional) Reference to a rule group. See [Rule Group Reference Statement](#rule-group-reference-statement) below.

#### ASN Match Statement

* `asn_list` - (Required) List of Autonomous System Numbers (ASNs) to match against. ASNs are unique identifiers assigned to large internet networks managed by organizations such as internet service providers, enterprises, universities, or government agencies.
* `forwarded_ip_config` - (Optional) Configuration for inspecting IP addresses in an HTTP header instead of using the web request origin. See [Forwarded IP Config](#forwarded-ip-config) below.

#### Byte Match Statement

* `search_string` - (Required) String value to search for within the request (1-200 characters).
* `positional_constraint` - (Required) Area within the portion of the web request that you want WAF to search for `search_string`. Valid values: `EXACTLY`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CONTAINS_WORD`.
* `field_to_match` - (Required) Part of the web request that you want WAF to inspect. See [Field to Match](#field-to-match) below.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. See [Text Transformation](#text-transformation) below.

#### Geo Match Statement

* `country_codes` - (Required) List of two-character country codes (ISO 3166-1 alpha-2).
* `forwarded_ip_config` - (Optional) Configuration for inspecting forwarded IP headers. See [Forwarded IP Config](#forwarded-ip-config) below.

#### IP Set Reference Statement

* `arn` - (Required) ARN of the IP set to reference.
* `ip_set_forwarded_ip_config` - (Optional) Configuration for inspecting forwarded IP headers. See [IP Set Forwarded IP Config](#ip-set-forwarded-ip-config) below.

#### IP Set Forwarded IP Config

* `fallback_behavior` - (Required) Action to take when the IP address in the header is invalid. Valid values: `MATCH`, `NO_MATCH`.
* `header_name` - (Required) Name of the header containing the forwarded IP address.
* `position` - (Required) Position in the header to use. Valid values: `FIRST`, `LAST`, `ANY`.

#### Label Match Statement

* `key` - (Required) String to match against. For `LABEL` scope, include the name and any preceding namespace specifications. For `NAMESPACE` scope, include namespace strings. Labels are case sensitive and components must be separated by colon (e.g., `NS1:NS2:name`).
* `scope` - (Required) Whether to match using the label name or namespace. Valid values: `LABEL`, `NAMESPACE`.

#### Managed Rule Group Statement

* `name` - (Required) Name of the managed rule group.
* `vendor_name` - (Required) Name of the managed rule group vendor (e.g., "AWS").
* `version` - (Optional) Version of the managed rule group.
* `rule_action_override` - (Optional) Override actions for specific rules within the managed rule group. See [Rule Action Override](#rule-action-override) below.
* `scope_down_statement` - (Optional) Additional statement to narrow the scope of requests that the managed rule group evaluates. See [Scope Down Statement](#scope-down-statement) below.

#### Rule Group Reference Statement

* `arn` - (Required) ARN of the rule group to reference.

#### Regex Pattern Set Reference Statement

* `arn` - (Required) ARN of the regex pattern set to reference.
* `field_to_match` - (Required) Part of the web request that you want WAF to inspect. See [Field to Match](#field-to-match) below.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. See [Text Transformation](#text-transformation) below.

#### Rate Based Statement

* `limit` - (Required) Rate limit threshold (requests per evaluation window period).
* `aggregate_key_type` - (Optional) Setting that indicates how to aggregate the request counts. Defaults to `IP`. Valid values: `IP`, `FORWARDED_IP`, `CUSTOM_KEYS`, `CONSTANT`.
* `custom_keys` - (Optional) Aggregate the request counts using one or more web request components as the aggregate keys. See [Custom Keys](#custom-keys) below.
* `evaluation_window_sec` - (Optional) Time window for which the rate limit applies, in seconds. Defaults to `300` (5 minutes). Valid values: `60`, `120`, `300`, `600`.
* `forwarded_ip_config` - (Optional) Configuration for inspecting IP addresses in an HTTP header instead of using the web request origin. See [Forwarded IP Config](#forwarded-ip-config) below.
* `scope_down_statement` - (Optional) Additional statement to narrow the scope of requests that the rate-based rule evaluates. See [Scope Down Statement](#scope-down-statement) below.

#### Regex Match Statement

* `regex_string` - (Required) Regular expression pattern to match against the web request component.
* `field_to_match` - (Required) Part of the web request that you want WAF to inspect. See [Field to Match](#field-to-match) below.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. See [Text Transformation](#text-transformation) below.

### Override Action

One of the following override action blocks must be specified when using managed rule groups:

* `count` - (Optional) Override the rule action with count.
* `none` - (Optional) Don't override the rule action.

### Rule Label

* `name` - (Required) Label string (1-1024 characters, alphanumeric, underscore, hyphen, and colon characters only).

### Captcha Config

* `immunity_time_property` - (Optional) Immunity time configuration. See [Immunity Time Property](#immunity-time-property) below.

### Challenge Config

* `immunity_time_property` - (Optional) Immunity time configuration. See [Immunity Time Property](#immunity-time-property) below.

#### Immunity Time Property

* `immunity_time` - (Optional) Immunity time in seconds (60-259200).

### Visibility Config

* `cloudwatch_metrics_enabled` - (Optional) Whether to enable CloudWatch metrics. Defaults to `true`.
* `metric_name` - (Optional) Name of the CloudWatch metric. Defaults to the rule name.
* `sampled_requests_enabled` - (Optional) Whether to store sampled requests. Defaults to `true`.

## Attribute Reference

This resource exports no additional attributes.

### Field to Match

Exactly one of the following field to match blocks must be specified:

* `all_query_arguments` - (Optional) Inspect all query arguments.
* `body` - (Optional) Inspect the request body as plain text.
* `method` - (Optional) Inspect the HTTP method.
* `query_string` - (Optional) Inspect the query string.
* `single_header` - (Optional) Inspect a single header. See [Single Header](#single-header) below.
* `single_query_argument` - (Optional) Inspect a single query argument. See [Single Query Argument](#single-query-argument) below.
* `uri_path` - (Optional) Inspect the request URI path.

#### Single Header

* `name` - (Required) Name of the header to inspect (case insensitive).

#### Single Query Argument

* `name` - (Required) Name of the query argument to inspect.

### Text Transformation

* `priority` - (Required) Relative processing order for multiple transformations (0-based).
* `type` - (Required) Transformation to apply. Valid values: `NONE`, `COMPRESS_WHITE_SPACE`, `HTML_ENTITY_DECODE`, `LOWERCASE`, `CMD_LINE`, `URL_DECODE`, `BASE64_DECODE`, `HEX_DECODE`, `MD5`, `REPLACE_COMMENTS`, `ESCAPE_SEQ_DECODE`, `SQL_HEX_DECODE`, `CSS_DECODE`, `JS_DECODE`, `NORMALIZE_PATH`, `NORMALIZE_PATH_WIN`, `REMOVE_NULLS`, `REPLACE_NULLS`, `BASE64_DECODE_EXT`, `URL_DECODE_UNI`, `UTF8_TO_UNICODE`.

### Rule Action Override

* `name` - (Required) Name of the rule to override.
* `action_to_use` - (Required) Override action to use for the rule. See [Action](#action) below.

### Scope Down Statement

Exactly one of the following scope down statement blocks must be specified:

* `asn_match_statement` - (Optional) Match requests based on Autonomous System Number (ASN). See [ASN Match Statement](#asn-match-statement) above.
* `byte_match_statement` - (Optional) Match requests based on byte patterns. See [Byte Match Statement](#byte-match-statement) above.
* `geo_match_statement` - (Optional) Match requests by geographic location. See [Geo Match Statement](#geo-match-statement) above.
* `ip_set_reference_statement` - (Optional) Reference to an IP set. See [IP Set Reference Statement](#ip-set-reference-statement) above.
* `label_match_statement` - (Optional) Match requests based on labels. See [Label Match Statement](#label-match-statement) above.
* `regex_match_statement` - (Optional) Match requests using regex patterns. See [Regex Match Statement](#regex-match-statement) above.
* `size_constraint_statement` - (Optional) Match requests based on size constraints. See [Size Constraint Statement](#size-constraint-statement) above.
* `sqli_match_statement` - (Optional) Match requests that appear to contain SQL injection attacks.
* `xss_match_statement` - (Optional) Match requests that appear to contain cross-site scripting attacks.

### Custom Keys

Exactly one of the following custom key blocks must be specified:

* `cookie` - (Optional) Use a cookie as an aggregate key. See [Custom Key Cookie](#custom-key-cookie) below.
* `forwarded_ip` - (Optional) Use the forwarded IP address as an aggregate key.
* `header` - (Optional) Use a header as an aggregate key. See [Custom Key Header](#custom-key-header) below.
* `http_method` - (Optional) Use the HTTP method as an aggregate key.
* `ip` - (Optional) Use the IP address as an aggregate key.
* `label_namespace` - (Optional) Use a label namespace as an aggregate key. See [Custom Key Label Namespace](#custom-key-label-namespace) below.
* `query_argument` - (Optional) Use a query argument as an aggregate key. See [Custom Key Query Argument](#custom-key-query-argument) below.
* `query_string` - (Optional) Use the query string as an aggregate key.
* `uri_path` - (Optional) Use the URI path as an aggregate key.

#### Custom Key Cookie

* `name` - (Required) Name of the cookie.
* `text_transformation` - (Required) Text transformations to apply to the cookie value. See [Text Transformation](#text-transformation) above.

#### Custom Key Header

* `name` - (Required) Name of the header.
* `text_transformation` - (Required) Text transformations to apply to the header value. See [Text Transformation](#text-transformation) above.

#### Custom Key Label Namespace

* `namespace` - (Required) Label namespace to use as the custom key.

#### Custom Key Query Argument

* `name` - (Required) Name of the query argument.
* `text_transformation` - (Required) Text transformations to apply to the query argument value. See [Text Transformation](#text-transformation) above.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_wafv2_web_acl_rule.example
  identity = {
    web_acl_arn = "arn:aws:wafv2:us-east-1:123456789012:regional/webacl/example/abc123def456"
    name        = "my-rule"
  }
}

resource "aws_wafv2_web_acl_rule" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` (String) Rule name, unique within the Web ACL.
* `web_acl_arn` (String) ARN of the Web ACL.

#### Optional

* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAFv2 Web ACL Rules using the `web_acl_arn` and `name` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_wafv2_web_acl_rule.example
  id = "arn:aws:wafv2:us-east-1:123456789012:regional/webacl/example/abc123def456,my-rule"
}
```

Using `terraform import`, import WAFv2 Web ACL Rules using the `web_acl_arn` and `name` separated by a comma (`,`). For example:

```console
% terraform import aws_wafv2_web_acl_rule.example arn:aws:wafv2:us-east-1:123456789012:regional/webacl/example/abc123def456,my-rule
```
