---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_web_acl_rule"
description: |-
  Manages an individual rule within a WAFv2 Web ACL.
---

# Resource: aws_wafv2_web_acl_rule

Manages an individual rule within a WAFv2 Web ACL. This resource creates proper Terraform dependencies for safe deletion of referenced resources like IP sets, solving the `WAFAssociatedItemException` error that occurs when deleting IP sets that are still referenced by Web ACL rules.

~> **NOTE:** When using this resource, you must add `lifecycle { ignore_changes = [rule] }` to your `aws_wafv2_web_acl` resource to prevent conflicts. See the [`aws_wafv2_web_acl` documentation](/docs/providers/aws/r/wafv2_web_acl.html) for a full description of the limitations of inline rules that this resource addresses.

## Example Usage

### Migrating from Inline Rules

This resource supports a "create-or-adopt" pattern that allows seamless migration from inline Web ACL rules to separate `aws_wafv2_web_acl_rule` resources without infrastructure changes.

When you create an `aws_wafv2_web_acl_rule` resource with the same name as an existing inline rule in the Web ACL, the resource will automatically adopt the existing rule instead of creating a duplicate. This enables zero-downtime migration from inline rules to separate resources.

Starting with inline rules, update your configuration to use separate rule resources and apply:

```terraform
resource "aws_wafv2_web_acl" "example" {
  name  = "example"
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  # Remove inline rule block - it's now managed separately

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "example"
    sampled_requests_enabled   = false
  }

  # Prevent Terraform from managing inline rules
  lifecycle {
    ignore_changes = [rule]
  }
}

# Separate rule resource with identical configuration
resource "aws_wafv2_web_acl_rule" "block_countries" {
  name        = "block-countries" # Must match existing rule name
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

Apply the configuration:

```bash
terraform apply
```

The `aws_wafv2_web_acl_rule` resource will adopt the existing inline rule without making any changes to the actual Web ACL infrastructure. The rule continues to function identically, but is now managed as a separate Terraform resource.

- The rule name in the `aws_wafv2_web_acl_rule` resource must exactly match the existing inline rule name
- Add `lifecycle { ignore_changes = [rule] }` to your Web ACL resource to prevent conflicts
- The create-or-adopt behavior only applies when a rule with the same name already exists in the Web ACL

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

### IP Set Reference

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

### Logical AND Statement

Block requests that match multiple conditions (e.g., from a specific country AND containing a specific string):

```terraform
resource "aws_wafv2_web_acl_rule" "block_suspicious" {
  name        = "block-suspicious"
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.example.arn

  action {
    block {}
  }

  statement {
    and_statement {
      statement {
        geo_match_statement {
          country_codes = ["CN"]
        }
      }

      statement {
        byte_match_statement {
          search_string         = "admin"
          positional_constraint = "CONTAINS"

          field_to_match {
            uri_path {}
          }

          text_transformation {
            priority = 0
            type     = "LOWERCASE"
          }
        }
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "block-suspicious"
    sampled_requests_enabled   = true
  }
}
```

### Logical OR Statement

Block requests that match any of multiple conditions:

```terraform
resource "aws_wafv2_web_acl_rule" "block_countries" {
  name        = "block-countries"
  priority    = 2
  web_acl_arn = aws_wafv2_web_acl.example.arn

  action {
    block {}
  }

  statement {
    or_statement {
      statement {
        geo_match_statement {
          country_codes = ["CN"]
        }
      }

      statement {
        geo_match_statement {
          country_codes = ["RU"]
        }
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "block-countries"
    sampled_requests_enabled   = true
  }
}
```

### Logical NOT Statement

Allow requests only from specific countries by negating a geo match:

```terraform
resource "aws_wafv2_web_acl_rule" "allow_only_us" {
  name        = "allow-only-us"
  priority    = 3
  web_acl_arn = aws_wafv2_web_acl.example.arn

  action {
    block {}
  }

  statement {
    not_statement {
      statement {
        geo_match_statement {
          country_codes = ["US", "CA"]
        }
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "allow-only-us"
    sampled_requests_enabled   = true
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the rule. Must be unique within the Web ACL.
* `priority` - (Required) Rule priority. Rules with lower priority are evaluated first.
* `statement` - (Required) Rule statement. See [Statement](#statement-block) below.
* `visibility_config` - (Required) CloudWatch metrics configuration. See [Visibility Config](#visibility_config-block) below.
* `web_acl_arn` - (Required) ARN of the Web ACL to add the rule to.

The following arguments are optional:

* `action` - (Optional) Action to take when the rule matches. See [Action](#action-block) below. Conflicts with `override_action`.
* `captcha_config` - (Optional) CAPTCHA configuration that overrides the web ACL level setting. See [Captcha Config](#captcha_config-block) below.
* `challenge_config` - (Optional) Challenge configuration that overrides the web ACL level setting. See [Challenge Config](#challenge_config-block) below.
* `override_action` - (Optional) Override action for managed rule groups. See [Override Action](#override_action-block) below. Conflicts with `action`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `rule_label` - (Optional) Labels to apply to matching web requests. See [Rule Label](#rule_label-block) below.

### `action` Block

One of the following action blocks must be specified:

* `allow` - (Optional) Allow the request. See [Allow](#allow-block) below.
* `block` - (Optional) Block the request. See [Block](#block-block) below.
* `captcha` - (Optional) Present a CAPTCHA challenge. See [Captcha](#captcha-block) below.
* `challenge` - (Optional) Present a silent challenge. See [Challenge](#challenge-block) below.
* `count` - (Optional) Count the request without blocking. See [Count](#count-block) below.

#### `allow` Block

* `custom_request_handling` - (Optional) Custom request handling configuration. See [Custom Request Handling](#custom_request_handling-block) below.

#### `block` Block

* `custom_response` - (Optional) Custom response configuration. See [Custom Response](#custom_response-block) below.

#### `count` Block

* `custom_request_handling` - (Optional) Custom request handling configuration. See [Custom Request Handling](#custom_request_handling-block) below.

#### `captcha` Block

* `custom_request_handling` - (Optional) Custom request handling configuration. See [Custom Request Handling](#custom_request_handling-block) below.

#### `challenge` Block

* `custom_request_handling` - (Optional) Custom request handling configuration. See [Custom Request Handling](#custom_request_handling-block) below.

#### `custom_request_handling` Block

* `insert_header` - (Optional) Custom headers to insert into the request. See [Insert Header](#insert_header-block) below.

#### `insert_header` Block

* `name` - (Required) Header name.
* `value` - (Required) Header value.

#### `custom_response` Block

* `custom_response_body_key` - (Optional) Key of a custom response body defined in the Web ACL.
* `response_code` - (Required) HTTP status code to return (200-599).
* `response_header` - (Optional) Custom headers to include in the response. See [Response Header](#response_header-block) below.

#### `response_header` Block

* `name` - (Required) Header name.
* `value` - (Required) Header value.

### `statement` Block

Exactly one of the following statement blocks must be specified:

* `and_statement` - (Optional) Logical AND statement that combines multiple statements. See [And Statement](#and_statement-block) below.
* `asn_match_statement` - (Optional) Match requests based on Autonomous System Number (ASN). See [ASN Match Statement](#asn_match_statement-block) below.
* `byte_match_statement` - (Optional) Match requests based on byte patterns. See [Byte Match Statement](#byte_match_statement-block) below.
* `geo_match_statement` - (Optional) Match requests by geographic location. See [Geo Match Statement](#geo_match_statement-block) below.
* `ip_set_reference_statement` - (Optional) Reference to an IP set. See [IP Set Reference Statement](#ip_set_reference_statement-block) below.
* `label_match_statement` - (Optional) Match requests based on labels. See [Label Match Statement](#label_match_statement-block) below.
* `managed_rule_group_statement` - (Optional) Reference to a managed rule group. See [Managed Rule Group Statement](#managed_rule_group_statement-block) below.
* `not_statement` - (Optional) Logical NOT statement that negates a single statement. See [Not Statement](#not_statement-block) below.
* `or_statement` - (Optional) Logical OR statement that combines multiple statements. See [Or Statement](#or_statement-block) below.
* `rate_based_statement` - (Optional) Rate-based rule to track request rates. See [Rate Based Statement](#rate_based_statement-block) below.
* `regex_match_statement` - (Optional) Match requests using regex patterns. See [Regex Match Statement](#regex_match_statement-block) below.
* `regex_pattern_set_reference_statement` - (Optional) Reference to a regex pattern set. See [Regex Pattern Set Reference Statement](#regex_pattern_set_reference_statement-block) below.
* `rule_group_reference_statement` - (Optional) Reference to a rule group. See [Rule Group Reference Statement](#rule_group_reference_statement-block) below.
* `size_constraint_statement` - (Optional) Match requests based on size constraints. See [Size Constraint Statement](#size_constraint_statement-block) below.
* `sqli_match_statement` - (Optional) Match requests that appear to contain SQL injection attacks. See [SQL Injection Match Statement](#sql-injection-match-statement) below.
* `xss_match_statement` - (Optional) Match requests that appear to contain cross-site scripting attacks. See [Cross-Site Scripting Match Statement](#cross-site-scripting-match-statement) below.

~> **NOTE:** Logical statements (`and_statement`, `not_statement`, `or_statement`) can be nested up to 3 levels deep. This matches the nesting limit of the `aws_wafv2_web_acl` resource. As in AWS WAF, `managed_rule_group_statement`, `rate_based_statement` and `rule_group_reference_statement` can only be used at the top level of a rule's `statement`, not inside a logical statement.

#### `and_statement` Block

Combines multiple statements using logical AND. All nested statements must match for the AND statement to match.

* `statement` - (Required) List of statements to combine. At least two statements are required. Each nested statement supports the statement types listed above except `and_statement`, `managed_rule_group_statement`, `rate_based_statement` and `rule_group_reference_statement`.

#### `not_statement` Block

Negates a single statement. The NOT statement matches when the nested statement does not match.

* `statement` - (Required) Single statement to negate. Exactly one statement must be specified. The nested statement supports the statement types listed above except `managed_rule_group_statement`, `rate_based_statement` and `rule_group_reference_statement`.

#### `or_statement` Block

Combines multiple statements using logical OR. At least one nested statement must match for the OR statement to match.

* `statement` - (Required) List of statements to combine. At least two statements are required. Each nested statement supports the statement types listed above except `managed_rule_group_statement`, `or_statement`, `rate_based_statement` and `rule_group_reference_statement`.

#### `asn_match_statement` Block

* `asn_list` - (Required) List of Autonomous System Numbers (ASNs) to match against. ASNs are unique identifiers assigned to large internet networks managed by organizations such as internet service providers, enterprises, universities, or government agencies.
* `forwarded_ip_config` - (Optional) Configuration for inspecting IP addresses in an HTTP header instead of using the web request origin. See [Forwarded IP Config](#forwarded_ip_config-block) below.

#### `byte_match_statement` Block

* `field_to_match` - (Required) Part of the web request that you want WAF to inspect. See [Field to Match](#field_to_match-block) below.
* `positional_constraint` - (Required) Area within the portion of the web request that you want WAF to search for `search_string`. Valid values: `EXACTLY`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CONTAINS_WORD`.
* `search_string` - (Optional) String value to search for within the request (1-200 bytes). Exactly one of `search_string` or `search_string_base64` must be specified.
* `search_string_base64` - (Optional) Base64-encoded value to search for within the request (1-200 bytes when decoded), for a value that is not valid UTF-8 text, such as the raw digest that the `MD5` and `SHA256` text transformations produce. Exactly one of `search_string` or `search_string_base64` must be specified. On import, a value that is valid UTF-8 is read into `search_string` and any other value into `search_string_base64`.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. See [Text Transformation](#text_transformation-block) below.

#### `geo_match_statement` Block

* `country_codes` - (Required) List of two-character country codes (ISO 3166-1 alpha-2).
* `forwarded_ip_config` - (Optional) Configuration for inspecting forwarded IP headers. See [Forwarded IP Config](#forwarded_ip_config-block) below.

#### `ip_set_reference_statement` Block

* `arn` - (Required) ARN of the IP set to reference.
* `ip_set_forwarded_ip_config` - (Optional) Configuration for inspecting forwarded IP headers. See [IP Set Forwarded IP Config](#ip_set_forwarded_ip_config-block) below.

#### `ip_set_forwarded_ip_config` Block

* `fallback_behavior` - (Required) Action to take when the IP address in the header is invalid. Valid values: `MATCH`, `NO_MATCH`.
* `header_name` - (Required) Name of the header containing the forwarded IP address.
* `position` - (Required) Position in the header to use. Valid values: `FIRST`, `LAST`, `ANY`.

#### `forwarded_ip_config` Block

* `fallback_behavior` - (Required) Action to take when the IP address in the header is invalid. Valid values: `MATCH`, `NO_MATCH`.
* `header_name` - (Required) Name of the header containing the forwarded IP address.

#### `label_match_statement` Block

* `key` - (Required) String to match against. For `LABEL` scope, include the name and any preceding namespace specifications. For `NAMESPACE` scope, include namespace strings. Labels are case sensitive and components must be separated by colon (e.g., `NS1:NS2:name`).
* `scope` - (Required) Whether to match using the label name or namespace. Valid values: `LABEL`, `NAMESPACE`.

#### `managed_rule_group_statement` Block

* `name` - (Required) Name of the managed rule group.
* `rule_action_override` - (Optional) Override actions for specific rules within the managed rule group. See [Rule Action Override](#rule_action_override-block) below.
* `scope_down_statement` - (Optional) Additional statement to narrow the scope of requests that the managed rule group evaluates. See [Scope Down Statement](#scope_down_statement-block) below.
* `vendor_name` - (Required) Name of the managed rule group vendor (e.g., "AWS").
* `version` - (Optional) Version of the managed rule group.

#### `rule_group_reference_statement` Block

* `arn` - (Required) ARN of the rule group to reference.
* `excluded_rule` - (Optional) Rules to exclude from the rule group. See [Excluded Rule](#excluded_rule-block) below.
* `rule_action_override` - (Optional) Override actions for specific rules within the rule group. See [Rule Action Override](#rule_action_override-block) below.

#### `size_constraint_statement` Block

* `comparison_operator` - (Required) Operator to use to compare the request part to the size setting. Valid values: `EQ`, `NE`, `LE`, `LT`, `GE`, `GT`.
* `field_to_match` - (Required) Part of the web request that you want WAF to inspect. See [Field to Match](#field_to_match-block) below.
* `size` - (Required) Size, in bytes, to compare to the request part, after any transformations.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. See [Text Transformation](#text_transformation-block) below.

#### `regex_pattern_set_reference_statement` Block

* `arn` - (Required) ARN of the regex pattern set to reference.
* `field_to_match` - (Required) Part of the web request that you want WAF to inspect. See [Field to Match](#field_to_match-block) below.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. See [Text Transformation](#text_transformation-block) below.

#### `rate_based_statement` Block

* `aggregate_key_type` - (Required) Setting that indicates how to aggregate the request counts. Valid values: `IP`, `FORWARDED_IP`, `CUSTOM_KEYS`, `CONSTANT`.
* `custom_keys` - (Optional) Aggregate the request counts using one or more web request components as the aggregate keys. See [Custom Keys](#custom_keys-block) below.
* `evaluation_window_sec` - (Optional) Time window for which the rate limit applies, in seconds. Defaults to `300` (5 minutes). Valid values: `60`, `120`, `300`, `600`.
* `forwarded_ip_config` - (Optional) Configuration for inspecting IP addresses in an HTTP header instead of using the web request origin. See [Forwarded IP Config](#forwarded_ip_config-block) below.
* `limit` - (Required) Rate limit threshold (requests per evaluation window period).
* `scope_down_statement` - (Optional) Additional statement to narrow the scope of requests that the rate-based rule evaluates. See [Scope Down Statement](#scope_down_statement-block) below.

#### `regex_match_statement` Block

* `field_to_match` - (Required) Part of the web request that you want WAF to inspect. See [Field to Match](#field_to_match-block) below.
* `regex_string` - (Required) Regular expression pattern to match against the web request component.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. See [Text Transformation](#text_transformation-block) below.

#### SQL Injection Match Statement

* `field_to_match` - (Required) Part of the web request that you want WAF to inspect. See [Field to Match](#field_to_match-block) below.
* `sensitivity_level` - (Optional) Sensitivity level for detecting SQL injection attacks. Valid values: `HIGH`, `LOW`. Defaults to `LOW`. `HIGH` detects more attacks but might generate more false positives.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. See [Text Transformation](#text_transformation-block) below.

#### Cross-Site Scripting Match Statement

* `field_to_match` - (Required) Part of the web request that you want WAF to inspect. See [Field to Match](#field_to_match-block) below.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. See [Text Transformation](#text_transformation-block) below.

### `override_action` Block

One of the following override action blocks must be specified when using managed rule groups:

* `count` - (Optional) Override the rule action with count.
* `none` - (Optional) Don't override the rule action.

### `rule_label` Block

* `name` - (Required) Label string (1-1024 characters, alphanumeric, underscore, hyphen, and colon characters only).

### `captcha_config` Block

* `immunity_time_property` - (Optional) Immunity time configuration. See [Immunity Time Property](#immunity_time_property-block) below.

### `challenge_config` Block

* `immunity_time_property` - (Optional) Immunity time configuration. See [Immunity Time Property](#immunity_time_property-block) below.

#### `immunity_time_property` Block

* `immunity_time` - (Optional) Immunity time in seconds (60-259200).

### `visibility_config` Block

* `cloudwatch_metrics_enabled` - (Required) Whether to enable CloudWatch metrics.
* `metric_name` - (Required) Name of the CloudWatch metric.
* `sampled_requests_enabled` - (Required) Whether to store sampled requests.

### `field_to_match` Block

Exactly one of the following field to match blocks must be specified:

* `all_query_arguments` - (Optional) Inspect all query arguments.
* `body` - (Optional) Inspect the request body as plain text. See [Body](#body-block) below.
* `cookies` - (Optional) Inspect the request cookies. See [Cookies](#cookies-block) below.
* `header_order` - (Optional) Inspect a string containing the list of the request's header names, ordered as they appear in the web request. See [Header Order](#header_order-block) below.
* `headers` - (Optional) Inspect the request headers. See [Headers](#headers-block) below.
* `ja3_fingerprint` - (Optional) Match against the request's JA3 fingerprint (CloudFront and ALB only). See [JA3 Fingerprint](#ja3_fingerprint-block) below.
* `ja4_fingerprint` - (Optional) Match against the request's JA4 fingerprint (CloudFront and ALB only). See [JA4 Fingerprint](#ja4_fingerprint-block) below.
* `json_body` - (Optional) Inspect the request body as JSON. See [JSON Body](#json_body-block) below.
* `method` - (Optional) Inspect the HTTP method.
* `query_string` - (Optional) Inspect the query string.
* `single_header` - (Optional) Inspect a single header. See [Single Header](#single_header-block) below.
* `single_query_argument` - (Optional) Inspect a single query argument. See [Single Query Argument](#single_query_argument-block) below.
* `uri_fragment` - (Optional) Inspect fragments of the request URI. See [URI Fragment](#uri_fragment-block) below.
* `uri_path` - (Optional) Inspect the request URI path.

#### `body` Block

* `oversize_handling` - (Optional) How to handle requests with a body larger than the inspection limit. Valid values: `CONTINUE`, `MATCH`, `NO_MATCH`. Defaults to `CONTINUE`.

#### `cookies` Block

* `match_pattern` - (Required) Cookies to inspect. See [Cookies Match Pattern](#cookies-match-pattern) below.
* `match_scope` - (Required) Parts of the cookies to inspect. Valid values: `ALL`, `KEY`, `VALUE`.
* `oversize_handling` - (Required) How to handle requests with cookies larger than the inspection limit. Valid values: `CONTINUE`, `MATCH`, `NO_MATCH`.

#### Cookies Match Pattern

Exactly one of the following must be specified:

* `all` - (Optional) Inspect all cookies.
* `excluded_cookies` - (Optional) List of cookie names to exclude from inspection.
* `included_cookies` - (Optional) List of cookie names to inspect.

#### `header_order` Block

* `oversize_handling` - (Required) How to handle requests with headers larger than the inspection limit. Valid values: `CONTINUE`, `MATCH`, `NO_MATCH`.

#### `headers` Block

* `match_pattern` - (Required) Headers to inspect. See [Headers Match Pattern](#headers-match-pattern) below.
* `match_scope` - (Required) Parts of the headers to inspect. Valid values: `ALL`, `KEY`, `VALUE`.
* `oversize_handling` - (Required) How to handle requests with headers larger than the inspection limit. Valid values: `CONTINUE`, `MATCH`, `NO_MATCH`.

#### Headers Match Pattern

Exactly one of the following must be specified:

* `all` - (Optional) Inspect all headers.
* `excluded_headers` - (Optional) List of header names to exclude from inspection.
* `included_headers` - (Optional) List of header names to inspect.

#### `ja3_fingerprint` Block

* `fallback_behavior` - (Required) Action to take if WAF cannot calculate the fingerprint. Valid values: `MATCH`, `NO_MATCH`.

#### `ja4_fingerprint` Block

* `fallback_behavior` - (Required) Action to take if WAF cannot calculate the fingerprint. Valid values: `MATCH`, `NO_MATCH`.

#### `json_body` Block

* `invalid_fallback_behavior` - (Optional) How to handle requests with invalid JSON body. Valid values: `EVALUATE_AS_STRING`, `MATCH`, `NO_MATCH`.
* `match_pattern` - (Required) JSON content to inspect. See [JSON Body Match Pattern](#json-body-match-pattern) below.
* `match_scope` - (Required) Parts of the JSON to inspect. Valid values: `ALL`, `KEY`, `VALUE`.
* `oversize_handling` - (Optional) How to handle requests with a body larger than the inspection limit. Valid values: `CONTINUE`, `MATCH`, `NO_MATCH`. Defaults to `CONTINUE`.

#### JSON Body Match Pattern

Exactly one of the following must be specified:

* `all` - (Optional) Inspect all JSON content.
* `included_paths` - (Optional) List of JSON pointer expressions to inspect (e.g., `/foo/bar`).

#### `single_header` Block

* `name` - (Required) Name of the header to inspect (case insensitive).

#### `single_query_argument` Block

* `name` - (Required) Name of the query argument to inspect.

#### `uri_fragment` Block

* `fallback_behavior` - (Optional) How to handle requests with a URI fragment that is too large to inspect. Valid values: `MATCH`, `NO_MATCH`.

### `text_transformation` Block

* `priority` - (Required) Relative processing order for multiple transformations (0-based).
* `type` - (Required) Transformation to apply. Valid values: `NONE`, `COMPRESS_WHITE_SPACE`, `HTML_ENTITY_DECODE`, `LOWERCASE`, `CMD_LINE`, `URL_DECODE`, `BASE64_DECODE`, `HEX_DECODE`, `MD5`, `REPLACE_COMMENTS`, `ESCAPE_SEQ_DECODE`, `SQL_HEX_DECODE`, `CSS_DECODE`, `JS_DECODE`, `NORMALIZE_PATH`, `NORMALIZE_PATH_WIN`, `REMOVE_NULLS`, `REPLACE_NULLS`, `BASE64_DECODE_EXT`, `URL_DECODE_UNI`, `UTF8_TO_UNICODE`.

### `rule_action_override` Block

* `action_to_use` - (Required) Override action to use for the rule. See [Action](#action-block) below.
* `name` - (Required) Name of the rule to override.

### `scope_down_statement` Block

Exactly one of the following scope down statement blocks must be specified:

* `and_statement` - (Optional) Logical AND statement that combines multiple statements. See [And Statement](#and_statement-block) above.
* `asn_match_statement` - (Optional) Match requests based on Autonomous System Number (ASN). See [ASN Match Statement](#asn_match_statement-block) above.
* `byte_match_statement` - (Optional) Match requests based on byte patterns. See [Byte Match Statement](#byte_match_statement-block) above.
* `geo_match_statement` - (Optional) Match requests by geographic location. See [Geo Match Statement](#geo_match_statement-block) above.
* `ip_set_reference_statement` - (Optional) Reference to an IP set. See [IP Set Reference Statement](#ip_set_reference_statement-block) above.
* `label_match_statement` - (Optional) Match requests based on labels. See [Label Match Statement](#label_match_statement-block) above.
* `not_statement` - (Optional) Logical NOT statement that negates a single statement. See [Not Statement](#not_statement-block) above.
* `or_statement` - (Optional) Logical OR statement that combines multiple statements. See [Or Statement](#or_statement-block) above.
* `regex_match_statement` - (Optional) Match requests using regex patterns. See [Regex Match Statement](#regex_match_statement-block) above.
* `size_constraint_statement` - (Optional) Match requests based on size constraints. See [Size Constraint Statement](#size_constraint_statement-block) above.
* `sqli_match_statement` - (Optional) Match requests that appear to contain SQL injection attacks.
* `xss_match_statement` - (Optional) Match requests that appear to contain cross-site scripting attacks.

~> **NOTE:** Logical statements (`and_statement`, `not_statement`, `or_statement`) can be nested up to 2 levels deep. This matches the nesting limit of the `aws_wafv2_web_acl` resource.

### `custom_keys` Block

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
* `text_transformation` - (Required) Text transformations to apply to the cookie value. See [Text Transformation](#text_transformation-block) above.

#### Custom Key Header

* `name` - (Required) Name of the header.
* `text_transformation` - (Required) Text transformations to apply to the header value. See [Text Transformation](#text_transformation-block) above.

#### Custom Key Label Namespace

* `namespace` - (Required) Label namespace to use as the custom key.

#### Custom Key Query Argument

* `name` - (Required) Name of the query argument.
* `text_transformation` - (Required) Text transformations to apply to the query argument value. See [Text Transformation](#text_transformation-block) above.

### `excluded_rule` Block

* `name` - (Required) Name of the rule to exclude from the rule group.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

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

* `account_id` (String) Account ID where this resource is managed.
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
