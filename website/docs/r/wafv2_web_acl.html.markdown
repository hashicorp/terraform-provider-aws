---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_web_acl"
description: |-
  Creates a WAFv2 Web ACL resource.
---

# Resource: aws_wafv2_web_acl

Creates a WAFv2 Web ACL resource.

~> **Note:** In `field_to_match` blocks, _e.g._, in `byte_match_statement`, the `body` block includes an optional argument `oversize_handling`. AWS indicates this argument will be required starting February 2023. To avoid configurations breaking when that change happens, treat the `oversize_handling` argument as **required** as soon as possible.

## Example Usage

This resource is based on `aws_wafv2_rule_group`, check the documentation of the `aws_wafv2_rule_group` resource to see examples of the various available statements.

### Managed Rule

```terraform
resource "aws_wafv2_web_acl" "example" {
  name        = "managed-rule-example"
  description = "Example of a managed rule."
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      count {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesCommonRuleSet"
        vendor_name = "AWS"

        rule_action_override {
          action_to_use {
            count {}
          }

          name = "SizeRestrictions_QUERYSTRING"
        }

        rule_action_override {
          action_to_use {
            count {}
          }

          name = "NoUserAgent_HEADER"
        }

        scope_down_statement {
          geo_match_statement {
            country_codes = ["US", "NL"]
          }
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
```

### Rate Based
Rate-limit US and NL-based clients to 10,000 requests for every 5 minutes.

```terraform
resource "aws_wafv2_web_acl" "example" {
  name        = "rate-based-example"
  description = "Example of a Cloudfront rate based statement."
  scope       = "CLOUDFRONT"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      block {}
    }

    statement {
      rate_based_statement {
        limit              = 10000
        aggregate_key_type = "IP"

        scope_down_statement {
          geo_match_statement {
            country_codes = ["US", "NL"]
          }
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
```

### Rule Group Reference

```terraform
resource "aws_wafv2_rule_group" "example" {
  capacity = 10
  name     = "example-rule-group"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      geo_match_statement {
        country_codes = ["NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  rule {
    name     = "rule-to-exclude-a"
    priority = 10

    action {
      allow {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  rule {
    name     = "rule-to-exclude-b"
    priority = 15

    action {
      allow {}
    }

    statement {
      geo_match_statement {
        country_codes = ["GB"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}

resource "aws_wafv2_web_acl" "test" {
  name  = "rule-group-example"
  scope = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      count {}
    }

    statement {
      rule_group_reference_statement {
        arn = aws_wafv2_rule_group.example.arn

        excluded_rule {
          name = "rule-to-exclude-b"
        }

        excluded_rule {
          name = "rule-to-exclude-a"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
```

## Argument Reference

The following arguments are supported:

* `custom_response_body` - (Optional) Defines custom response bodies that can be referenced by `custom_response` actions. See [`custom_response_body`](#custom_response_body) below for details.
* `default_action` - (Required) Action to perform if none of the `rules` contained in the WebACL match. See [`default_ action`](#default_action) below for details.
* `description` - (Optional) Friendly description of the WebACL.
* `name` - (Required) Friendly name of the WebACL.
* `rule` - (Optional) Rule blocks used to identify the web requests that you want to `allow`, `block`, or `count`. See [`rule`](#rule) below for details.
* `scope` - (Required) Specifies whether this is for an AWS CloudFront distribution or for a regional application. Valid values are `CLOUDFRONT` or `REGIONAL`. To work with CloudFront, you must also specify the region `us-east-1` (N. Virginia) on the AWS provider.
* `tags` - (Optional) Map of key-value pairs to associate with the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `visibility_config` - (Required) Defines and enables Amazon CloudWatch metrics and web request sample collection. See [`visibility_config`](#visibility_config) below for details.

### `custom_response_body`

Each `custom_response_body` block supports the following arguments:

* `key` - (Required) Unique key identifying the custom response body. This is referenced by the `custom_response_body_key` argument in the [`custom_response`](#custom_response) block.
* `content` - (Required) Payload of the custom response.
* `content_type` - (Required) Type of content in the payload that you are defining in the `content` argument. Valid values are `TEXT_PLAIN`, `TEXT_HTML`, or `APPLICATION_JSON`.

### `default_action`

The `default_action` block supports the following arguments:

~> **NOTE:** One of `allow` or `block`, expressed as an empty configuration block `{}`, is required when specifying a `default_action`

* `allow` - (Optional) Specifies that AWS WAF should allow requests by default. See [`allow`](#allow) below for details.
* `block` - (Optional) Specifies that AWS WAF should block requests by default. See [`block`](#block) below for details.

### `rule`

~> **NOTE:** One of `action` or `override_action` is required when specifying a rule

Each `rule` supports the following arguments:

* `action` - (Optional) Action that AWS WAF should take on a web request when it matches the rule's statement. This is used only for rules whose **statements do not reference a rule group**. See [`action`](#action) below for details.
* `name` - (Required) Friendly name of the rule.
* `override_action` - (Optional) Override action to apply to the rules in a rule group. Used only for rule **statements that reference a rule group**, like `rule_group_reference_statement` and `managed_rule_group_statement`. See [`override_action`](#override_action) below for details.
* `priority` - (Required) If you define more than one Rule in a WebACL, AWS WAF evaluates each request against the `rules` in order based on the value of `priority`. AWS WAF processes rules with lower priority first.
* `rule_label` - (Optional) Labels to apply to web requests that match the rule match statement. See [`rule_label`](#rule_label) below for details.
* `statement` - (Required) The AWS WAF processing statement for the rule, for example `byte_match_statement` or `geo_match_statement`. See [`statement`](#statement) below for details.
* `visibility_config` - (Required) Defines and enables Amazon CloudWatch metrics and web request sample collection. See [`visibility_config`](#visibility_config) below for details.

#### `action`

The `action` block supports the following arguments:

~> **NOTE:** One of `allow`, `block`, or `count`, is required when specifying an `action`.

* `allow` - (Optional) Instructs AWS WAF to allow the web request. See [`allow`](#allow) below for details.
* `block` - (Optional) Instructs AWS WAF to block the web request. See [`block`](#block) below for details.
* `captcha` - (Optional) Instructs AWS WAF to run a Captcha check against the web request. See [`captcha`](#captcha) below for details.
* `challenge` - (Optional) Instructs AWS WAF to run a check against the request to verify that the request is coming from a legitimate client session. See [`challenge`](#challenge) below for details.
* `count` - (Optional) Instructs AWS WAF to count the web request and allow it. See [`count`](#count) below for details.

#### `override_action`

The `override_action` block supports the following arguments:

~> **NOTE:** One of `count` or `none`, expressed as an empty configuration block `{}`, is required when specifying an `override_action`

* `count` - (Optional) Override the rule action setting to count (i.e., only count matches). Configured as an empty block `{}`.
* `none` - (Optional) Don't override the rule action setting. Configured as an empty block `{}`.

#### `allow`

The `allow` block supports the following arguments:

* `custom_request_handling` - (Optional) Defines custom handling for the web request. See [`custom_request_handling`](#custom_request_handling) below for details.

#### `block`

The `block` block supports the following arguments:

* `custom_response` - (Optional) Defines a custom response for the web request. See [`custom_response`](#custom_response) below for details.

#### `captcha`

The `captcha` block supports the following arguments:

* `custom_request_handling` - (Optional) Defines custom handling for the web request. See [`custom_request_handling`](#custom_request_handling) below for details.

#### `challenge`

The `challenge` block supports the following arguments:

* `custom_request_handling` - (Optional) Defines custom handling for the web request. See [`custom_request_handling`](#custom_request_handling) below for details.

#### `count`

The `count` block supports the following arguments:

* `custom_request_handling` - (Optional) Defines custom handling for the web request. See [`custom_request_handling`](#custom_request_handling) below for details.

#### `custom_request_handling`

The `custom_request_handling` block supports the following arguments:

* `insert_header` - (Required) The `insert_header` blocks used to define HTTP headers added to the request. See [`insert_header`](#insert_header) below for details.

#### `insert_header`

Each `insert_header` block supports the following arguments. Duplicate header names are not allowed:

* `name` - Name of the custom header. For custom request header insertion, when AWS WAF inserts the header into the request, it prefixes this name `x-amzn-waf-`, to avoid confusion with the headers that are already in the request. For example, for the header name `sample`, AWS WAF inserts the header `x-amzn-waf-sample`.
* `value` - Value of the custom header.

#### `custom_response`

The `custom_response` block supports the following arguments:

* `custom_response_body_key` - (Optional) References the response body that you want AWS WAF to return to the web request client. This must reference a `key` defined in a `custom_response_body` block of this resource.
* `response_code` - (Required) The HTTP status code to return to the client.
* `response_header` - (Optional) The `response_header` blocks used to define the HTTP response headers added to the response. See [`response_header`](#response_header) below for details.

#### `response_header`

Each `response_header` block supports the following arguments. Duplicate header names are not allowed:

* `name` - Name of the custom header. For custom request header insertion, when AWS WAF inserts the header into the request, it prefixes this name `x-amzn-waf-`, to avoid confusion with the headers that are already in the request. For example, for the header name `sample`, AWS WAF inserts the header `x-amzn-waf-sample`.
* `value` - Value of the custom header.

#### `rule_label`

Each block supports the following arguments:

* `name` - Label string.

#### `statement`

The processing guidance for a Rule, used by AWS WAF to determine whether a web request matches the rule. See the [documentation](https://docs.aws.amazon.com/waf/latest/developerguide/waf-rule-statements-list.html) for more information.

-> **NOTE:** Although the `statement` block is recursive, currently only 3 levels are supported.

The `statement` block supports the following arguments:

* `and_statement` - (Optional) Logical rule statement used to combine other rule statements with AND logic. See [`and_statement`](#and_statement) below for details.
* `byte_match_statement` - (Optional) Rule statement that defines a string match search for AWS WAF to apply to web requests. See [`byte_match_statement`](#byte_match_statement) below for details.
* `geo_match_statement` - (Optional) Rule statement used to identify web requests based on country of origin. See [`geo_match_statement`](#geo_match_statement) below for details.
* `ip_set_reference_statement` - (Optional) Rule statement used to detect web requests coming from particular IP addresses or address ranges. See [IP Set Reference Statement](#ip_set_reference_statement) below for details.
* `label_match_statement` - (Optional) Rule statement that defines a string match search against labels that have been added to the web request by rules that have already run in the web ACL. See [`label_match_statement`](#label_match_statement) below for details.
* `managed_rule_group_statement` - (Optional) Rule statement used to run the rules that are defined in a managed rule group.  This statement can not be nested. See [Managed Rule Group Statement](#managed_rule_group_statement) below for details.
* `not_statement` - (Optional) Logical rule statement used to negate the results of another rule statement. See [`not_statement`](#not_statement) below for details.
* `or_statement` - (Optional) Logical rule statement used to combine other rule statements with OR logic. See [`or_statement`](#or_statement) below for details.
* `rate_based_statement` - (Optional) Rate-based rule tracks the rate of requests for each originating `IP address`, and triggers the rule action when the rate exceeds a limit that you specify on the number of requests in any `5-minute` time span. This statement can not be nested. See [`rate_based_statement`](#rate_based_statement) below for details.
* `regex_match_statement` - (Optional) Rule statement used to search web request components for a match against a single regular expression. See [`regex_match_statement`](#regex_match_statement) below for details.
* `regex_pattern_set_reference_statement` - (Optional) Rule statement used to search web request components for matches with regular expressions. See [Regex Pattern Set Reference Statement](#regex_pattern_set_reference_statement) below for details.
* `rule_group_reference_statement` - (Optional) Rule statement used to run the rules that are defined in an WAFv2 Rule Group. See [Rule Group Reference Statement](#rule_group_reference_statement) below for details.
* `size_constraint_statement` - (Optional) Rule statement that compares a number of bytes against the size of a request component, using a comparison operator, such as greater than (>) or less than (<). See [`size_constraint_statement`](#size_constraint_statement) below for more details.
* `sqli_match_statement` - (Optional) An SQL injection match condition identifies the part of web requests, such as the URI or the query string, that you want AWS WAF to inspect. See [`sqli_match_statement`](#sqli_match_statement) below for details.
* `xss_match_statement` - (Optional) Rule statement that defines a cross-site scripting (XSS) match search for AWS WAF to apply to web requests. See [`xss_match_statement`](#xss_match_statement) below for details.

#### `and_statement`

A logical rule statement used to combine other rule statements with `AND` logic. You provide more than one `statement` within the `and_statement`.

The `and_statement` block supports the following arguments:

* `statement` - (Required) Statements to combine with `AND` logic. You can use any statements that can be nested. See [`statement`](#statement) above for details.

#### `byte_match_statement`

The byte match statement provides the bytes to search for, the location in requests that you want AWS WAF to search, and other settings. The bytes to search for are typically a string that corresponds with ASCII characters.

The `byte_match_statement` block supports the following arguments:

* `field_to_match` - (Optional) Part of a web request that you want AWS WAF to inspect. See [`field_to_match`](#field_to_match) below for details.
* `positional_constraint` - (Required) Area within the portion of a web request that you want AWS WAF to search for `search_string`. Valid values include the following: `EXACTLY`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CONTAINS_WORD`. See the AWS [documentation](https://docs.aws.amazon.com/waf/latest/APIReference/API_ByteMatchStatement.html) for more information.
* `search_string` - (Required) String value that you want AWS WAF to search for. AWS WAF searches only in the part of web requests that you designate for inspection in `field_to_match`. The maximum length of the value is 50 bytes.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection.
  At least one required.
  See [`text_transformation`](#text_transformation) below for details.

#### `geo_match_statement`

The `geo_match_statement` block supports the following arguments:

* `country_codes` - (Required) Array of two-character country codes, for example, [ "US", "CN" ], from the alpha-2 country ISO codes of the `ISO 3166` international standard. See the [documentation](https://docs.aws.amazon.com/waf/latest/APIReference/API_GeoMatchStatement.html) for valid values.
* `forwarded_ip_config` - (Optional) Configuration for inspecting IP addresses in an HTTP header that you specify, instead of using the IP address that's reported by the web request origin. See [`forwarded_ip_config`](#forwarded_ip_config) below for details.

#### `ip_set_reference_statement`

A rule statement used to detect web requests coming from particular IP addresses or address ranges. To use this, create an `aws_wafv2_ip_set` that specifies the addresses you want to detect, then use the `ARN` of that set in this statement.

The `ip_set_reference_statement` block supports the following arguments:

* `arn` - (Required) The Amazon Resource Name (ARN) of the IP Set that this statement references.
* `ip_set_forwarded_ip_config` - (Optional) Configuration for inspecting IP addresses in an HTTP header that you specify, instead of using the IP address that's reported by the web request origin. See [`ip_set_forwarded_ip_config`](#ip_set_forwarded_ip_config) below for more details.

#### `label_match_statement`

The `label_match_statement` block supports the following arguments:

* `scope` - (Required) Specify whether you want to match using the label name or just the namespace. Valid values are `LABEL` or `NAMESPACE`.
* `key` - (Required) String to match against.

#### `managed_rule_group_statement`

A rule statement used to run the rules that are defined in a managed rule group.

You can't nest a `managed_rule_group_statement`, for example for use inside a `not_statement` or `or_statement`. It can only be referenced as a `top-level` statement within a `rule`.

The `managed_rule_group_statement` block supports the following arguments:

* `excluded_rule` - (Optional, **Deprecated**) The `rules` whose actions are set to `COUNT` by the web ACL, regardless of the action that is set on the rule. See [`excluded_rule`](#excluded_rule) below for details. Use `rule_action_override` instead. (See the [documentation](https://docs.aws.amazon.com/waf/latest/APIReference/API_ManagedRuleGroupStatement.html#WAF-Type-ManagedRuleGroupStatement-ExcludedRules))
* `name` - (Required) Name of the managed rule group.
* `rule_action_override` - (Optional) Action settings to use in the place of the rule actions that are configured inside the rule group. You specify one override for each rule whose action you want to change. See [`rule_action_override`](#rule_action_override) below for details.
* `managed_rule_group_configs`- (Optional) Additional information that's used by a managed rule group. Only one rule attribute is allowed in each config. See [Managed Rule Group Configs](#managed_rule_group_configs) for more details
* `scope_down_statement` - Narrows the scope of the statement to matching web requests. This can be any nestable statement, and you can nest statements at any level below this scope-down statement. See [`statement`](#statement) above for details.
* `vendor_name` - (Required) Name of the managed rule group vendor.
* `version` - (Optional) Version of the managed rule group. You can set `Version_1.0` or `Version_1.1` etc. If you want to use the default version, do not set anything.

#### `not_statement`

A logical rule statement used to negate the results of another rule statement. You provide one `statement` within the `not_statement`.

The `not_statement` block supports the following arguments:

* `statement` - (Required) Statement to negate. You can use any statement that can be nested. See [`statement`](#statement) above for details.

#### `or_statement`

A logical rule statement used to combine other rule statements with `OR` logic. You provide more than one `statement` within the `or_statement`.

The `or_statement` block supports the following arguments:

* `statement` - (Required) Statements to combine with `OR` logic. You can use any statements that can be nested. See [`statement`](#statement) above for details.

#### `rate_based_statement`

A rate-based rule tracks the rate of requests for each originating IP address, and triggers the rule action when the rate exceeds a limit that you specify on the number of requests in any 5-minute time span. You can use this to put a temporary block on requests from an IP address that is sending excessive requests. See the [documentation](https://docs.aws.amazon.com/waf/latest/APIReference/API_RateBasedStatement.html) for more information.

You can't nest a `rate_based_statement`, for example for use inside a `not_statement` or `or_statement`. It can only be referenced as a `top-level` statement within a `rule`.

The `rate_based_statement` block supports the following arguments:

* `aggregate_key_type` - (Optional) Setting that indicates how to aggregate the request counts. Valid values include: `FORWARDED_IP` or `IP`. Default: `IP`.
* `forwarded_ip_config` - (Optional) Configuration for inspecting IP addresses in an HTTP header that you specify, instead of using the IP address that's reported by the web request origin. If `aggregate_key_type` is set to `FORWARDED_IP`, this block is required. See [`forwarded_ip_config`](#forwarded_ip_config) below for details.
* `limit` - (Required) Limit on requests per 5-minute period for a single originating IP address.
* `scope_down_statement` - (Optional) Optional nested statement that narrows the scope of the rate-based statement to matching web requests. This can be any nestable statement, and you can nest statements at any level below this scope-down statement. See [`statement`](#statement) above for details.

#### `regex_match_statement`

A rule statement used to search web request components for a match against a single regular expression.

The `regex_match_statement` block supports the following arguments:

* `regex_string` - (Required) String representing the regular expression. Minimum of `1` and maximum of `512` characters.
* `field_to_match` - (Required) The part of a web request that you want AWS WAF to inspect. See [`field_to_match`](#field_to_match) below for details.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection.
  At least one required.
  See [`text_transformation`](#text_transformation) below for details.

#### `regex_pattern_set_reference_statement`

A rule statement used to search web request components for matches with regular expressions. To use this, create a `aws_wafv2_regex_pattern_set` that specifies the expressions that you want to detect, then use the `ARN` of that set in this statement. A web request matches the pattern set rule statement if the request component matches any of the patterns in the set.

The `regex_pattern_set_reference_statement` block supports the following arguments:

* `arn` - (Required) The Amazon Resource Name (ARN) of the Regex Pattern Set that this statement references.
* `field_to_match` - (Optional) Part of a web request that you want AWS WAF to inspect. See [`field_to_match`](#field_to_match) below for details.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection.
  At least one required.
  See [`text_transformation`](#text_transformation) below for details.

#### `rule_group_reference_statement`

A rule statement used to run the rules that are defined in an WAFv2 Rule Group or `aws_wafv2_rule_group` resource.

You can't nest a `rule_group_reference_statement`, for example for use inside a `not_statement` or `or_statement`. It can only be referenced as a `top-level` statement within a `rule`.

The `rule_group_reference_statement` block supports the following arguments:

* `arn` - (Required) The Amazon Resource Name (ARN) of the `aws_wafv2_rule_group` resource.
* `excluded_rule` - (Optional) The `rules` whose actions are set to `COUNT` by the web ACL, regardless of the action that is set on the rule. See [`excluded_rule`](#excluded_rule) below for details.

#### `size_constraint_statement`

A rule statement that uses a comparison operator to compare a number of bytes against the size of a request component. AWS WAFv2 inspects up to the first 8192 bytes (8 KB) of a request body, and when inspecting the request URI Path, the slash `/` in
the URI counts as one character.

The `size_constraint_statement` block supports the following arguments:

* `comparison_operator` - (Required) Operator to use to compare the request part to the size setting. Valid values include: `EQ`, `NE`, `LE`, `LT`, `GE`, or `GT`.
* `field_to_match` - (Optional) Part of a web request that you want AWS WAF to inspect. See [`field_to_match`](#field_to_match) below for details.
* `size` - (Required) Size, in bytes, to compare to the request part, after any transformations. Valid values are integers between 0 and 21474836480, inclusive.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection.
  At least one required.
  See [`text_transformation`](#text_transformation) below for details.

#### `sqli_match_statement`

An SQL injection match condition identifies the part of web requests, such as the URI or the query string, that you want AWS WAF to inspect. Later in the process, when you create a web ACL, you specify whether to allow or block requests that appear to contain malicious SQL code.

The `sqli_match_statement` block supports the following arguments:

* `field_to_match` - (Optional) Part of a web request that you want AWS WAF to inspect. See [`field_to_match`](#field_to_match) below for details.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection.
  At least one required.
  See [`text_transformation`](#text_transformation) below for details.

#### `xss_match_statement`

The XSS match statement provides the location in requests that you want AWS WAF to search and text transformations to use on the search area before AWS WAF searches for character sequences that are likely to be malicious strings.

The `xss_match_statement` block supports the following arguments:

* `field_to_match` - (Optional) Part of a web request that you want AWS WAF to inspect. See [`field_to_match`](#field_to_match) below for details.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection.
  At least one required.
  See [`text_transformation`](#text_transformation) below for details.

#### `excluded_rule`

The `excluded_rule` block supports the following arguments:

* `name` - (Required) Name of the rule to exclude. If the rule group is managed by AWS, see the [documentation](https://docs.aws.amazon.com/waf/latest/developerguide/aws-managed-rule-groups-list.html) for a list of names in the appropriate rule group in use.

#### `rule_action_override`

The `rule_action_override` block supports the following arguments:

* `action_to_use` - (Required) Override action to use, in place of the configured action of the rule in the rule group. See [`action`](#action) below for details.
* `name` - (Required) Name of the rule to override. See the [documentation](https://docs.aws.amazon.com/waf/latest/developerguide/aws-managed-rule-groups-list.html) for a list of names in the appropriate rule group in use.

#### `managed_rule_group_configs`

The `managed_rule_group_configs` block support the following arguments:

* `aws_managed_rules_bot_control_rule_set` - (Optional) Additional configuration for using the Bot Control managed rule group. Use this to specify the inspection level that you want to use. See [`aws_managed_rules_bot_control_rule_set`](#aws_managed_rules_bot_control_rule_set) for more details
* `login_path` - (Optional) The path of the login endpoint for your application.
* `password_field` - (Optional) Details about your login page password field. See [`password_field`](#password_field) for more details.
* `payload_type`- (Optional) The payload type for your login endpoint, either JSON or form encoded.
* `username_field` - (Optional) Details about your login page username field. See [`username_field`](#username_field) for more details.

#### `aws_managed_rules_bot_control_rule_set`

* `inspection_level` - (Optional) The inspection level to use for the Bot Control rule group.

#### `password_field`

* `identifier` - (Optional) The name of the password field.

#### `username_field`

* `identifier` - (Optional) The name of the username field.

#### `field_to_match`

The part of a web request that you want AWS WAF to inspect. Include the single `field_to_match` type that you want to inspect, with additional specifications as needed, according to the type. You specify a single request component in `field_to_match` for each rule statement that requires it. To inspect more than one component of a web request, create a separate rule statement for each component. See the [documentation](https://docs.aws.amazon.com/waf/latest/developerguide/waf-rule-statement-fields.html#waf-rule-statement-request-component) for more details.

The `field_to_match` block supports the following arguments:

~> **NOTE:** Only one of `all_query_arguments`, `body`, `cookies`, `headers`, `json_body`, `method`, `query_string`, `single_header`, `single_query_argument`, or `uri_path` can be specified.
An empty configuration block `{}` should be used when specifying `all_query_arguments`, `method`, or `query_string` attributes.

* `all_query_arguments` - (Optional) Inspect all query arguments.
* `body` - (Optional) Inspect the request body, which immediately follows the request headers. See [`body`](#body) below for details.
* `cookies` - (Optional) Inspect the cookies in the web request. See [`cookies`](#cookies) below for details.
* `headers` - (Optional) Inspect the request headers. See [`headers`](#headers) below for details.
* `json_body` - (Optional) Inspect the request body as JSON. See [`json_body`](#json_body) for details.
* `method` - (Optional) Inspect the HTTP method. The method indicates the type of operation that the request is asking the origin to perform.
* `query_string` - (Optional) Inspect the query string. This is the part of a URL that appears after a `?` character, if any.
* `single_header` - (Optional) Inspect a single header. See [`single_header`](#single_header) below for details.
* `single_query_argument` - (Optional) Inspect a single query argument. See [`single_query_argument`](#single_query_argument) below for details.
* `uri_path` - (Optional) Inspect the request URI path. This is the part of a web request that identifies a resource, for example, `/images/daily-ad.jpg`.

#### `forwarded_ip_config`

The configuration for inspecting IP addresses in an HTTP header that you specify, instead of using the IP address that's reported by the web request origin. Commonly, this is the X-Forwarded-For (XFF) header, but you can specify
any header name. If the specified header isn't present in the request, AWS WAFv2 doesn't apply the rule to the web request at all.
AWS WAFv2 only evaluates the first IP address found in the specified HTTP header.

The `forwarded_ip_config` block supports the following arguments:

* `fallback_behavior` - (Required) - Match status to assign to the web request if the request doesn't have a valid IP address in the specified position. Valid values include: `MATCH` or `NO_MATCH`.
* `header_name` - (Required) - Name of the HTTP header to use for the IP address.

#### `ip_set_forwarded_ip_config`

The configuration for inspecting IP addresses in an HTTP header that you specify, instead of using the IP address that's reported by the web request origin. Commonly, this is the X-Forwarded-For (XFF) header, but you can specify any header name.

The `ip_set_forwarded_ip_config` block supports the following arguments:

* `fallback_behavior` - (Required) - Match status to assign to the web request if the request doesn't have a valid IP address in the specified position. Valid values include: `MATCH` or `NO_MATCH`.
* `header_name` - (Required) - Name of the HTTP header to use for the IP address.
* `position` - (Required) - Position in the header to search for the IP address. Valid values include: `FIRST`, `LAST`, or `ANY`. If `ANY` is specified and the header contains more than 10 IP addresses, AWS WAFv2 inspects the last 10.

#### `headers`

Inspect the request headers.

The `headers` block supports the following arguments:

* `match_pattern` - (Required) The filter to use to identify the subset of headers to inspect in a web request. The `match_pattern` block supports only one of the following arguments:
    * `all` - An empty configuration block that is used for inspecting all headers.
    * `included_headers` - An array of strings that will be used for inspecting headers that have a key that matches one of the provided values.
    * `excluded_headers` - An array of strings that will be used for inspecting headers that do not have a key that matches one of the provided values.
* `match_scope` - (Required) The parts of the headers to inspect with the rule inspection criteria. If you specify `All`, AWS WAF inspects both keys and values. Valid values include the following: `ALL`, `Key`, `Value`.
* `oversize_handling` - (Required) Oversize handling tells AWS WAF what to do with a web request when the request component that the rule inspects is over the limits. Valid values include the following: `CONTINUE`, `MATCH`, `NO_MATCH`. See the AWS [documentation](https://docs.aws.amazon.com/waf/latest/developerguide/waf-rule-statement-oversize-handling.html) for more information.

#### `json_body`

The `json_body` block supports the following arguments:

* `invalid_fallback_behavior` - (Optional) What to do when JSON parsing fails. Defaults to evaluating up to the first parsing failure. Valid values are `EVALUATE_AS_STRING`, `MATCH` and `NO_MATCH`.
* `match_pattern` - (Required) The patterns to look for in the JSON body. You must specify exactly one setting: either `all` or `included_paths`. See [JsonMatchPattern](https://docs.aws.amazon.com/waf/latest/APIReference/API_JsonMatchPattern.html) for details.
* `match_scope` - (Required) The parts of the JSON to match against using the `match_pattern`. Valid values are `ALL`, `KEY` and `VALUE`.
* `oversize_handling` - (Optional) What to do if the body is larger than can be inspected. Valid values are `CONTINUE` (default), `MATCH` and `NO_MATCH`.

#### `single_header`

Inspect a single header. Provide the name of the header to inspect, for example, `User-Agent` or `Referer` (provided as lowercase strings).

The `single_header` block supports the following arguments:

* `name` - (Optional) Name of the query header to inspect. This setting must be provided as lower case characters.

#### `single_query_argument`

Inspect a single query argument. Provide the name of the query argument to inspect, such as `UserName` or `SalesRegion` (provided as lowercase strings).

The `single_query_argument` block supports the following arguments:

* `name` - (Optional) Name of the query header to inspect. This setting must be provided as lower case characters.

#### `body`

The `body` block supports the following arguments:

* `oversize_handling` - (Optional) What WAF should do if the body is larger than WAF can inspect. WAF does not support inspecting the entire contents of the body of a web request when the body exceeds 8 KB (8192 bytes). Only the first 8 KB of the request body are forwarded to WAF by the underlying host service. Valid values: `CONTINUE`, `MATCH`, `NO_MATCH`.

#### `cookies`

Inspect the cookies in the web request. You can specify the parts of the cookies to inspect and you can narrow the set of cookies to inspect by including or excluding specific keys.
This is used to indicate the web request component to inspect, in the [FieldToMatch](https://docs.aws.amazon.com/waf/latest/APIReference/API_FieldToMatch.html) specification.

The `cookies` block supports the following arguments:

* `match_pattern` - (Required) The filter to use to identify the subset of cookies to inspect in a web request. You must specify exactly one setting: either `all`, `included_cookies` or `excluded_cookies`. More details: [CookieMatchPattern](https://docs.aws.amazon.com/waf/latest/APIReference/API_CookieMatchPattern.html)
* `match_scope` - (Required) The parts of the cookies to inspect with the rule inspection criteria. If you specify All, AWS WAF inspects both keys and values. Valid values: `ALL`, `KEY`, `VALUE`
* `oversize_handling` - (Required) What AWS WAF should do if the cookies of the request are larger than AWS WAF can inspect. AWS WAF does not support inspecting the entire contents of request cookies when they exceed 8 KB (8192 bytes) or 200 total cookies. The underlying host service forwards a maximum of 200 cookies and at most 8 KB of cookie contents to AWS WAF. Valid values: `CONTINUE`, `MATCH`, `NO_MATCH`.

#### `text_transformation`

The `text_transformation` block supports the following arguments:

* `priority` - (Required) Relative processing order for multiple transformations that are defined for a rule statement. AWS WAF processes all transformations, from lowest priority to highest, before inspecting the transformed content.
* `type` - (Required) Transformation to apply, please refer to the Text Transformation [documentation](https://docs.aws.amazon.com/waf/latest/APIReference/API_TextTransformation.html) for more details.

### `visibility_config`

The `visibility_config` block supports the following arguments:

* `cloudwatch_metrics_enabled` - (Required) Whether the associated resource sends metrics to CloudWatch. For the list of available metrics, see [AWS WAF Metrics](https://docs.aws.amazon.com/waf/latest/developerguide/monitoring-cloudwatch.html#waf-metrics).
* `metric_name` - (Required) A friendly name of the CloudWatch metric. The name can contain only alphanumeric characters (A-Z, a-z, 0-9) hyphen(-) and underscore (\_), with length from one to 128 characters. It can't contain whitespace or metric names reserved for AWS WAF, for example `All` and `Default_Action`.
* `sampled_requests_enabled` - (Required) Whether AWS WAF should store a sampling of the web requests that match the rules. You can view the sampled requests through the AWS WAF console.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the WAF WebACL.
* `capacity` - Web ACL capacity units (WCUs) currently being used by this web ACL.
* `id` - The ID of the WAF WebACL.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

WAFv2 Web ACLs can be imported using `ID/Name/Scope` e.g.,

```
$ terraform import aws_wafv2_web_acl.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc/example/REGIONAL
```
