---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_web_acl"
description: |-
  Creates a WAFv2 Web ACL resource.
---

# Resource: aws_wafv2_web_acl

Creates a WAFv2 Web ACL resource.

~> **Note** In `field_to_match` blocks, _e.g._, in `byte_match_statement`, the `body` block includes an optional argument `oversize_handling`. AWS indicates this argument will be required starting February 2023. To avoid configurations breaking when that change happens, treat the `oversize_handling` argument as **required** as soon as possible.

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

    token_domains = ["mywebsite.com", "myotherwebsite.com"]

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

### Account Creation Fraud Prevention

```terraform
resource "aws_wafv2_web_acl" "acfp-example" {
  name        = "managed-acfp-example"
  description = "Example of a managed ACFP rule."
  scope       = "CLOUDFRONT"

  default_action {
    allow {}
  }

  rule {
    name     = "acfp-rule-1"
    priority = 1

    override_action {
      count {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesACFPRuleSet"
        vendor_name = "AWS"

        managed_rule_group_configs {
          aws_managed_rules_acfp_rule_set {
            creation_path          = "/signin"
            registration_page_path = "/register"

            request_inspection {
              email_field {
                identifier = "/email"
              }

              password_field {
                identifier = "/password"
              }

              payload_type = "JSON"

              username_field {
                identifier = "/username"
              }
            }

            response_inspection {
              status_code {
                failure_codes = ["403"]
                success_codes = ["200"]
              }
            }
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

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
```

### Account Takeover Protection

```terraform
resource "aws_wafv2_web_acl" "atp-example" {
  name        = "managed-atp-example"
  description = "Example of a managed ATP rule."
  scope       = "CLOUDFRONT"

  default_action {
    allow {}
  }

  rule {
    name     = "atp-rule-1"
    priority = 1

    override_action {
      count {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesATPRuleSet"
        vendor_name = "AWS"

        managed_rule_group_configs {
          aws_managed_rules_atp_rule_set {
            login_path = "/api/1/signin"

            request_inspection {
              password_field {
                identifier = "/password"
              }

              payload_type = "JSON"

              username_field {
                identifier = "/email"
              }
            }

            response_inspection {
              status_code {
                failure_codes = ["403"]
                success_codes = ["200"]
              }
            }
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

        rule_action_override {
          action_to_use {
            count {}
          }

          name = "rule-to-exclude-b"
        }

        rule_action_override {
          action_to_use {
            count {}
          }

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

### Large Request Body Inspections for Regional Resources

```terraform
resource "aws_wafv2_web_acl" "example" {
  name  = "large-request-body-example"
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  association_config {
    request_body {
      api_gateway {
        default_size_inspection_limit = "KB_64"
      }
      app_runner_service {
        default_size_inspection_limit = "KB_64"
      }
      cognito_user_pool {
        default_size_inspection_limit = "KB_64"
      }
      verified_access_instance {
        default_size_inspection_limit = "KB_64"
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `association_config` - (Optional) Specifies custom configurations for the associations between the web ACL and protected resources. See [`association_config`](#association_config-block) below for details.
* `captcha_config` - (Optional) Specifies how AWS WAF should handle CAPTCHA evaluations on the ACL level (used by [AWS Bot Control](https://docs.aws.amazon.com/waf/latest/developerguide/aws-managed-rule-groups-bot.html)). See [`captcha_config`](#captcha_config-block) below for details.
* `challenge_config` - (Optional) Specifies how AWS WAF should handle Challenge evaluations on the ACL level (used by [AWS Bot Control](https://docs.aws.amazon.com/waf/latest/developerguide/aws-managed-rule-groups-bot.html)). See [`challenge_config`](#challenge_config-block) below for details.
* `custom_response_body` - (Optional) Defines custom response bodies that can be referenced by `custom_response` actions. See [`custom_response_body`](#custom_response_body-block) below for details.
* `default_action` - (Required) Action to perform if none of the `rules` contained in the WebACL match. See [`default_action`](#default_action-block) below for details.
* `description` - (Optional) Friendly description of the WebACL.
* `name` - (Required, Forces new resource) Friendly name of the WebACL.
* `rule` - (Optional) Rule blocks used to identify the web requests that you want to `allow`, `block`, or `count`. See [`rule`](#rule-block) below for details.
* `rule_json` (Optional) Raw JSON string to allow more than three nested statements. Conflicts with `rule` attribute. This is for advanced use cases where more than 3 levels of nested statements are required. **There is no drift detection at this time**. If you use this attribute instead of `rule`, you will be foregoing drift detection. See the AWS [documentation](https://docs.aws.amazon.com/waf/latest/APIReference/API_CreateWebACL.html) for the JSON structure.
* `scope` - (Required, Forces new resource) Specifies whether this is for an AWS CloudFront distribution or for a regional application. Valid values are `CLOUDFRONT` or `REGIONAL`. To work with CloudFront, you must also specify the region `us-east-1` (N. Virginia) on the AWS provider.
* `tags` - (Optional) Map of key-value pairs to associate with the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `token_domains` - (Optional) Specifies the domains that AWS WAF should accept in a web request token. This enables the use of tokens across multiple protected websites. When AWS WAF provides a token, it uses the domain of the AWS resource that the web ACL is protecting. If you don't specify a list of token domains, AWS WAF accepts tokens only for the domain of the protected resource. With a token domain list, AWS WAF accepts the resource's host domain plus all domains in the token domain list, including their prefixed subdomains.
* `visibility_config` - (Required) Defines and enables Amazon CloudWatch metrics and web request sample collection. See [`visibility_config`](#visibility_config-block) below for details.

### `association_config` Block

The `association_config` block supports the following arguments:

* `request_body` - (Optional) Customizes the request body that your protected resource forward to AWS WAF for inspection. See [`request_body`](#request_body-block) below for details.

### `custom_response_body` Block

Each `custom_response_body` block supports the following arguments:

* `key` - (Required) Unique key identifying the custom response body. This is referenced by the `custom_response_body_key` argument in the [`custom_response`](#custom_response-block) block.
* `content` - (Required) Payload of the custom response.
* `content_type` - (Required) Type of content in the payload that you are defining in the `content` argument. Valid values are `TEXT_PLAIN`, `TEXT_HTML`, or `APPLICATION_JSON`.

### `default_action` Block

The `default_action` block supports the following arguments:

~> **Note** One of `allow` or `block` is required when specifying a `default_action`

* `allow` - (Optional) Specifies that AWS WAF should allow requests by default. See [`allow`](#allow-block) below for details.
* `block` - (Optional) Specifies that AWS WAF should block requests by default. See [`block`](#block-block) below for details.

### `rule` Block

~> **Note** One of `action` or `override_action` is required when specifying a rule

Each `rule` supports the following arguments:

* `action` - (Optional) Action that AWS WAF should take on a web request when it matches the rule's statement. This is used only for rules whose **statements do not reference a rule group**. See [`action`](#action-block) for details.
* `captcha_config` - (Optional) Specifies how AWS WAF should handle CAPTCHA evaluations. See [`captcha_config`](#captcha_config-block) below for details.
* `name` - (Required) Friendly name of the rule. Note that the provider assumes that rules with names matching this pattern, `^ShieldMitigationRuleGroup_<account-id>_<web-acl-guid>_.*`, are AWS-added for [automatic application layer DDoS mitigation activities](https://docs.aws.amazon.com/waf/latest/developerguide/ddos-automatic-app-layer-response-rg.html). Such rules will be ignored by the provider unless you explicitly include them in your configuration (for example, by using the AWS CLI to discover their properties and creating matching configuration). However, since these rules are owned and managed by AWS, you may get permission errors.
* `override_action` - (Optional) Override action to apply to the rules in a rule group. Used only for rule **statements that reference a rule group**, like `rule_group_reference_statement` and `managed_rule_group_statement`. See [`override_action`](#override_action-block) below for details.
* `priority` - (Required) If you define more than one Rule in a WebACL, AWS WAF evaluates each request against the `rules` in order based on the value of `priority`. AWS WAF processes rules with lower priority first.
* `rule_label` - (Optional) Labels to apply to web requests that match the rule match statement. See [`rule_label`](#rule_label-block) below for details.
* `statement` - (Required) The AWS WAF processing statement for the rule, for example `byte_match_statement` or `geo_match_statement`. See [`statement`](#statement-block) below for details.
* `visibility_config` - (Required) Defines and enables Amazon CloudWatch metrics and web request sample collection. See [`visibility_config`](#visibility_config-block) below for details.

### `action` Block

The `action` block supports the following arguments:

~> **Note** One of `allow`, `block`, or `count`, is required when specifying an `action`.

* `allow` - (Optional) Instructs AWS WAF to allow the web request. See [`allow`](#allow-block) below for details.
* `block` - (Optional) Instructs AWS WAF to block the web request. See [`block`](#block-block) below for details.
* `captcha` - (Optional) Instructs AWS WAF to run a Captcha check against the web request. See [`captcha`](#captcha-block) below for details.
* `challenge` - (Optional) Instructs AWS WAF to run a check against the request to verify that the request is coming from a legitimate client session. See [`challenge`](#challenge-block) below for details.
* `count` - (Optional) Instructs AWS WAF to count the web request and allow it. See [`count`](#count-block) below for details.

### `override_action` Block

The `override_action` block supports the following arguments:

~> **Note** One of `count` or `none`, expressed as an empty configuration block `{}`, is required when specifying an `override_action`

* `count` - (Optional) Override the rule action setting to count (i.e., only count matches). Configured as an empty block `{}`.
* `none` - (Optional) Don't override the rule action setting. Configured as an empty block `{}`.

### `allow` Block

The `allow` block supports the following arguments:

* `custom_request_handling` - (Optional) Defines custom handling for the web request. See [`custom_request_handling`](#custom_request_handling-block) below for details.

### `block` Block

The `block` block supports the following arguments:

* `custom_response` - (Optional) Defines a custom response for the web request. See [`custom_response`](#custom_response-block) below for details.

### `captcha` Block

The `captcha` block supports the following arguments:

* `custom_request_handling` - (Optional) Defines custom handling for the web request. See [`custom_request_handling`](#custom_request_handling-block) below for details.

### `challenge` Block

The `challenge` block supports the following arguments:

* `custom_request_handling` - (Optional) Defines custom handling for the web request. See [`custom_request_handling`](#custom_request_handling-block) below for details.

### `count` Block

The `count` block supports the following arguments:

* `custom_request_handling` - (Optional) Defines custom handling for the web request. See [`custom_request_handling`](#custom_request_handling-block) below for details.

### `custom_request_handling` Block

The `custom_request_handling` block supports the following arguments:

* `insert_header` - (Required) The `insert_header` blocks used to define HTTP headers added to the request. See [`insert_header`](#insert_header-block) below for details.

### `insert_header` Block

Each `insert_header` block supports the following arguments. Duplicate header names are not allowed:

* `name` - Name of the custom header. For custom request header insertion, when AWS WAF inserts the header into the request, it prefixes this name `x-amzn-waf-`, to avoid confusion with the headers that are already in the request. For example, for the header name `sample`, AWS WAF inserts the header `x-amzn-waf-sample`.
* `value` - Value of the custom header.

### `custom_response` Block

The `custom_response` block supports the following arguments:

* `custom_response_body_key` - (Optional) References the response body that you want AWS WAF to return to the web request client. This must reference a `key` defined in a `custom_response_body` block of this resource.
* `response_code` - (Required) The HTTP status code to return to the client.
* `response_header` - (Optional) The `response_header` blocks used to define the HTTP response headers added to the response. See [`response_header`](#response_header-block) below for details.

### `response_header` Block

Each `response_header` block supports the following arguments. Duplicate header names are not allowed:

* `name` - Name of the custom header. For custom request header insertion, when AWS WAF inserts the header into the request, it prefixes this name `x-amzn-waf-`, to avoid confusion with the headers that are already in the request. For example, for the header name `sample`, AWS WAF inserts the header `x-amzn-waf-sample`.
* `value` - Value of the custom header.

### `rule_label` Block

Each block supports the following arguments:

* `name` - Label string.

### `statement` Block

The processing guidance for a Rule, used by AWS WAF to determine whether a web request matches the rule. See the [documentation](https://docs.aws.amazon.com/waf/latest/developerguide/waf-rule-statements-list.html) for more information.

-> **Note** Although the `statement` block is recursive, currently only 3 levels are supported.

The `statement` block supports the following arguments:

* `and_statement` - (Optional) Logical rule statement used to combine other rule statements with AND logic. See [`and_statement`](#and_statement-block) below for details.
* `byte_match_statement` - (Optional) Rule statement that defines a string match search for AWS WAF to apply to web requests. See [`byte_match_statement`](#byte_match_statement-block) below for details.
* `geo_match_statement` - (Optional) Rule statement used to identify web requests based on country of origin. See [`geo_match_statement`](#geo_match_statement-block) below for details.
* `ip_set_reference_statement` - (Optional) Rule statement used to detect web requests coming from particular IP addresses or address ranges. See [`ip_set_reference_statement`](#ip_set_reference_statement-block) below for details.
* `label_match_statement` - (Optional) Rule statement that defines a string match search against labels that have been added to the web request by rules that have already run in the web ACL. See [`label_match_statement`](#label_match_statement-block) below for details.
* `managed_rule_group_statement` - (Optional) Rule statement used to run the rules that are defined in a managed rule group.  This statement can not be nested. See [`managed_rule_group_statement`](#managed_rule_group_statement-block) below for details.
* `not_statement` - (Optional) Logical rule statement used to negate the results of another rule statement. See [`not_statement`](#not_statement-block) below for details.
* `or_statement` - (Optional) Logical rule statement used to combine other rule statements with OR logic. See [`or_statement`](#or_statement-block) below for details.
* `rate_based_statement` - (Optional) Rate-based rule tracks the rate of requests for each originating `IP address`, and triggers the rule action when the rate exceeds a limit that you specify on the number of requests in any `5-minute` time span. This statement can not be nested. See [`rate_based_statement`](#rate_based_statement-block) below for details.
* `regex_match_statement` - (Optional) Rule statement used to search web request components for a match against a single regular expression. See [`regex_match_statement`](#regex_match_statement-block) below for details.
* `regex_pattern_set_reference_statement` - (Optional) Rule statement used to search web request components for matches with regular expressions. See [`regex_pattern_set_reference_statement`](#regex_pattern_set_reference_statement-block) below for details.
* `rule_group_reference_statement` - (Optional) Rule statement used to run the rules that are defined in an WAFv2 Rule Group. See [`rule_group_reference_statement`](#rule_group_reference_statement-block) below for details.
* `size_constraint_statement` - (Optional) Rule statement that compares a number of bytes against the size of a request component, using a comparison operator, such as greater than (>) or less than (<). See [`size_constraint_statement`](#size_constraint_statement-block) below for more details.
* `sqli_match_statement` - (Optional) An SQL injection match condition identifies the part of web requests, such as the URI or the query string, that you want AWS WAF to inspect. See [`sqli_match_statement`](#sqli_match_statement-block) below for details.
* `xss_match_statement` - (Optional) Rule statement that defines a cross-site scripting (XSS) match search for AWS WAF to apply to web requests. See [`xss_match_statement`](#xss_match_statement-block) below for details.

### `and_statement` Block

A logical rule statement used to combine other rule statements with `AND` logic. You provide more than one `statement` within the `and_statement`.

The `and_statement` block supports the following arguments:

* `statement` - (Required) Statements to combine with `AND` logic. You can use any statements that can be nested. See [`statement`](#statement-block) above for details.

### `byte_match_statement` Block

The byte match statement provides the bytes to search for, the location in requests that you want AWS WAF to search, and other settings. The bytes to search for are typically a string that corresponds with ASCII characters.

The `byte_match_statement` block supports the following arguments:

* `field_to_match` - (Optional) Part of a web request that you want AWS WAF to inspect. See [`field_to_match`](#field_to_match-block) below for details.
* `positional_constraint` - (Required) Area within the portion of a web request that you want AWS WAF to search for `search_string`. Valid values include the following: `EXACTLY`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CONTAINS_WORD`. See the AWS [documentation](https://docs.aws.amazon.com/waf/latest/APIReference/API_ByteMatchStatement.html) for more information.
* `search_string` - (Required) String value that you want AWS WAF to search for. AWS WAF searches only in the part of web requests that you designate for inspection in `field_to_match`. The maximum length of the value is 50 bytes.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. At least one transformation is required. See [`text_transformation`](#text_transformation-block) below for details.

### `geo_match_statement` Block

The `geo_match_statement` block supports the following arguments:

* `country_codes` - (Required) Array of two-character country codes, for example, [ "US", "CN" ], from the alpha-2 country ISO codes of the `ISO 3166` international standard. See the [documentation](https://docs.aws.amazon.com/waf/latest/APIReference/API_GeoMatchStatement.html) for valid values.
* `forwarded_ip_config` - (Optional) Configuration for inspecting IP addresses in an HTTP header that you specify, instead of using the IP address that's reported by the web request origin. See [`forwarded_ip_config`](#forwarded_ip_config-block) below for details.

### `ip_set_reference_statement` Block

A rule statement used to detect web requests coming from particular IP addresses or address ranges. To use this, create an `aws_wafv2_ip_set` that specifies the addresses you want to detect, then use the `ARN` of that set in this statement.

The `ip_set_reference_statement` block supports the following arguments:

* `arn` - (Required) The Amazon Resource Name (ARN) of the IP Set that this statement references.
* `ip_set_forwarded_ip_config` - (Optional) Configuration for inspecting IP addresses in an HTTP header that you specify, instead of using the IP address that's reported by the web request origin. See [`ip_set_forwarded_ip_config`](#ip_set_forwarded_ip_config-block) below for more details.

### `label_match_statement` Block

The `label_match_statement` block supports the following arguments:

* `scope` - (Required) Specify whether you want to match using the label name or just the namespace. Valid values are `LABEL` or `NAMESPACE`.
* `key` - (Required) String to match against.

### `managed_rule_group_statement` Block

A rule statement used to run the rules that are defined in a managed rule group.

You can't nest a `managed_rule_group_statement`, for example for use inside a `not_statement` or `or_statement`. It can only be referenced as a `top-level` statement within a `rule`.

The `managed_rule_group_statement` block supports the following arguments:

* `name` - (Required) Name of the managed rule group.
* `rule_action_override` - (Optional) Action settings to use in the place of the rule actions that are configured inside the rule group. You specify one override for each rule whose action you want to change. See [`rule_action_override`](#rule_action_override-block) below for details.
* `managed_rule_group_configs`- (Optional) Additional information that's used by a managed rule group. Only one rule attribute is allowed in each config. See [`managed_rule_group_configs`](#managed_rule_group_configs-block) for more details
* `scope_down_statement` - Narrows the scope of the statement to matching web requests. This can be any nestable statement, and you can nest statements at any level below this scope-down statement. See [`statement`](#statement-block) above for details.
* `vendor_name` - (Required) Name of the managed rule group vendor.
* `version` - (Optional) Version of the managed rule group. You can set `Version_1.0` or `Version_1.1` etc. If you want to use the default version, do not set anything.

### `not_statement` Block

A logical rule statement used to negate the results of another rule statement. You provide one `statement` within the `not_statement`.

The `not_statement` block supports the following arguments:

* `statement` - (Required) Statement to negate. You can use any statement that can be nested. See [`statement`](#statement-block) above for details.

### `or_statement` Block

A logical rule statement used to combine other rule statements with `OR` logic. You provide more than one `statement` within the `or_statement`.

The `or_statement` block supports the following arguments:

* `statement` - (Required) Statements to combine with `OR` logic. You can use any statements that can be nested. See [`statement`](#statement-block) above for details.

### `rate_based_statement` Block

A rate-based rule tracks the rate of requests for each originating IP address, and triggers the rule action when the rate exceeds a limit that you specify on the number of requests in any 5-minute time span. You can use this to put a temporary block on requests from an IP address that is sending excessive requests. See the [documentation](https://docs.aws.amazon.com/waf/latest/APIReference/API_RateBasedStatement.html) for more information.

You can't nest a `rate_based_statement`, for example for use inside a `not_statement` or `or_statement`. It can only be referenced as a `top-level` statement within a `rule`.

The `rate_based_statement` block supports the following arguments:

* `aggregate_key_type` - (Optional) Setting that indicates how to aggregate the request counts. Valid values include: `CONSTANT`, `CUSTOM_KEYS`, `FORWARDED_IP`, or `IP`. Default: `IP`.
* `custom_key` - (Optional) Aggregate the request counts using one or more web request components as the aggregate keys. See [`custom_key`](#custom_key-block) below for details.
* `evaluation_window_sec` - (Optional) The amount of time, in seconds, that AWS WAF should include in its request counts, looking back from the current time. Valid values are `60`, `120`, `300`, and `600`. Defaults to `300` (5 minutes).

  **NOTE:** This setting doesn't determine how often AWS WAF checks the rate, but how far back it looks each time it checks. AWS WAF checks the rate about every 10 seconds.
* `forwarded_ip_config` - (Optional) Configuration for inspecting IP addresses in an HTTP header that you specify, instead of using the IP address that's reported by the web request origin. If `aggregate_key_type` is set to `FORWARDED_IP`, this block is required. See [`forwarded_ip_config`](#forwarded_ip_config-block) below for details.
* `limit` - (Required) Limit on requests per 5-minute period for a single originating IP address.
* `scope_down_statement` - (Optional) Optional nested statement that narrows the scope of the rate-based statement to matching web requests. This can be any nestable statement, and you can nest statements at any level below this scope-down statement. See [`statement`](#statement-block) above for details. If `aggregate_key_type` is set to `CONSTANT`, this block is required.

### `regex_match_statement` Block

A rule statement used to search web request components for a match against a single regular expression.

The `regex_match_statement` block supports the following arguments:

* `regex_string` - (Required) String representing the regular expression. Minimum of `1` and maximum of `512` characters.
* `field_to_match` - (Required) The part of a web request that you want AWS WAF to inspect. See [`field_to_match`](#field_to_match-block) below for details.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. At least one transformation is required. See [`text_transformation`](#text_transformation-block) below for details.

### `regex_pattern_set_reference_statement` Block

A rule statement used to search web request components for matches with regular expressions. To use this, create a `aws_wafv2_regex_pattern_set` that specifies the expressions that you want to detect, then use the `ARN` of that set in this statement. A web request matches the pattern set rule statement if the request component matches any of the patterns in the set.

The `regex_pattern_set_reference_statement` block supports the following arguments:

* `arn` - (Required) The Amazon Resource Name (ARN) of the Regex Pattern Set that this statement references.
* `field_to_match` - (Optional) Part of a web request that you want AWS WAF to inspect. See [`field_to_match`](#field_to_match-block) below for details.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. At least one transformation is required. See [`text_transformation`](#text_transformation-block) below for details.

### `rule_group_reference_statement` Block

A rule statement used to run the rules that are defined in an WAFv2 Rule Group or `aws_wafv2_rule_group` resource.

You can't nest a `rule_group_reference_statement`, for example for use inside a `not_statement` or `or_statement`. It can only be referenced as a `top-level` statement within a `rule`.

The `rule_group_reference_statement` block supports the following arguments:

* `arn` - (Required) The Amazon Resource Name (ARN) of the `aws_wafv2_rule_group` resource.
* `rule_action_override` - (Optional) Action settings to use in the place of the rule actions that are configured inside the rule group. You specify one override for each rule whose action you want to change. See [`rule_action_override`](#rule_action_override-block) below for details.

### `size_constraint_statement` Block

A rule statement that uses a comparison operator to compare a number of bytes against the size of a request component. AWS WAFv2 inspects up to the first 8192 bytes (8 KB) of a request body, and when inspecting the request URI Path, the slash `/` in
the URI counts as one character.

The `size_constraint_statement` block supports the following arguments:

* `comparison_operator` - (Required) Operator to use to compare the request part to the size setting. Valid values include: `EQ`, `NE`, `LE`, `LT`, `GE`, or `GT`.
* `field_to_match` - (Optional) Part of a web request that you want AWS WAF to inspect. See [`field_to_match`](#field_to_match-block) below for details.
* `size` - (Required) Size, in bytes, to compare to the request part, after any transformations. Valid values are integers between 0 and 21474836480, inclusive.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. At least one transformation is required. See [`text_transformation`](#text_transformation-block) below for details.

### `sqli_match_statement` Block

An SQL injection match condition identifies the part of web requests, such as the URI or the query string, that you want AWS WAF to inspect. Later in the process, when you create a web ACL, you specify whether to allow or block requests that appear to contain malicious SQL code.

The `sqli_match_statement` block supports the following arguments:

* `field_to_match` - (Optional) Part of a web request that you want AWS WAF to inspect. See [`field_to_match`](#field_to_match-block) below for details.
* `sensitivity_level` - (Optional) Sensitivity that you want AWS WAF to use to inspect for SQL injection attacks. Valid values include: `LOW`, `HIGH`.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. At least one transformation is required. See [`text_transformation`](#text_transformation-block) below for details.

### `xss_match_statement` Block

The XSS match statement provides the location in requests that you want AWS WAF to search and text transformations to use on the search area before AWS WAF searches for character sequences that are likely to be malicious strings.

The `xss_match_statement` block supports the following arguments:

* `field_to_match` - (Optional) Part of a web request that you want AWS WAF to inspect. See [`field_to_match`](#field_to_match-block) below for details.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. At least one transformation is required. See [`text_transformation`](#text_transformation-block) below for details.

### `rule_action_override` Block

The `rule_action_override` block supports the following arguments:

* `action_to_use` - (Required) Override action to use, in place of the configured action of the rule in the rule group. See [`action`](#action-block) for details.
* `name` - (Required) Name of the rule to override. See the [documentation](https://docs.aws.amazon.com/waf/latest/developerguide/aws-managed-rule-groups-list.html) for a list of names in the appropriate rule group in use.

### `managed_rule_group_configs` Block

The `managed_rule_group_configs` block support the following arguments:

* `aws_managed_rules_bot_control_rule_set` - (Optional) Additional configuration for using the Bot Control managed rule group. Use this to specify the inspection level that you want to use. See [`aws_managed_rules_bot_control_rule_set`](#aws_managed_rules_bot_control_rule_set-block) for more details
* `aws_managed_rules_acfp_rule_set` - (Optional) Additional configuration for using the Account Creation Fraud Prevention managed rule group. Use this to specify information such as the registration page of your application and the type of content to accept or reject from the client.
* `aws_managed_rules_atp_rule_set` - (Optional) Additional configuration for using the Account Takeover Protection managed rule group. Use this to specify information such as the sign-in page of your application and the type of content to accept or reject from the client.
* `login_path` - (Optional, **Deprecated**) The path of the login endpoint for your application.
* `password_field` - (Optional, **Deprecated**) Details about your login page password field. See [`password_field`](#password_field-block) for more details.
* `payload_type`- (Optional, **Deprecated**) The payload type for your login endpoint, either JSON or form encoded.
* `username_field` - (Optional, **Deprecated**) Details about your login page username field. See [`username_field`](#username_field-block) for more details.

### `aws_managed_rules_bot_control_rule_set` Block

* `enable_machine_learning` - (Optional) Applies only to the targeted inspection level. Determines whether to use machine learning (ML) to analyze your web traffic for bot-related activity. Defaults to `true`.
* `inspection_level` - (Optional) The inspection level to use for the Bot Control rule group.

### `aws_managed_rules_acfp_rule_set` Block

* `creation_path` - (Required) The path of the account creation endpoint for your application. This is the page on your website that accepts the completed registration form for a new user. This page must accept POST requests.
* `enable_regex_in_path` - (Optional) Whether or not to allow the use of regular expressions in the login page path.
* `registration_page_path` - (Required) The path of the account registration endpoint for your application. This is the page on your website that presents the registration form to new users. This page must accept GET text/html requests.
* `request_inspection` - (Optional) The criteria for inspecting login requests, used by the ATP rule group to validate credentials usage. See [`request_inspection`](#request_inspection-block) for more details.
* `response_inspection` - (Optional) The criteria for inspecting responses to login requests, used by the ATP rule group to track login failure rates. Note that Response Inspection is available only on web ACLs that protect CloudFront distributions. See [`response_inspection`](#response_inspection-block) for more details.

### `aws_managed_rules_atp_rule_set` Block

* `enable_regex_in_path` - (Optional) Whether or not to allow the use of regular expressions in the login page path.
* `login_path` - (Required) The path of the login endpoint for your application.
* `request_inspection` - (Optional) The criteria for inspecting login requests, used by the ATP rule group to validate credentials usage. See [`request_inspection`](#request_inspection-block) for more details.
* `response_inspection` - (Optional) The criteria for inspecting responses to login requests, used by the ATP rule group to track login failure rates. Note that Response Inspection is available only on web ACLs that protect CloudFront distributions. See [`response_inspection`](#response_inspection-block) for more details.

### `request_inspection` Block

* `address_fields` (Optional) The names of the fields in the request payload that contain your customer's primary physical address. See [`address_fields`](#address_fields-block) for more details.
* `email_field` (Optional) The name of the field in the request payload that contains your customer's email. See [`email_field`](#email_field-block) for more details.
* `password_field` (Optional) Details about your login page password field. See [`password_field`](#password_field-block) for more details.
* `payload_type` (Required) The payload type for your login endpoint, either JSON or form encoded.
* `phone_number_fields` (Optional) The names of the fields in the request payload that contain your customer's primary phone number. See [`phone_number_fields`](#phone_number_fields-block) for more details.
* `username_field` (Optional) Details about your login page username field. See [`username_field`](#username_field-block) for more details.

### `address_fields` Block

* `identifier` - (Required) The name of a single primary address field.

### `email_field` Block

* `identifier` - (Required) The name of the field in the request payload that contains your customer's email.

### `password_field` Block

* `identifier` - (Required) The name of the password field.

### `phone_number_fields` Block

* `identifier` - (Required) The name of a single primary phone number field.

### `username_field` Block

* `identifier` - (Required) The name of the username field.

### `response_inspection` Block

* `body_contains` (Optional) Configures inspection of the response body. See [`body_contains`](#body_contains-block) for more details.
* `header` (Optional) Configures inspection of the response header.See [`header`](#header-block) for more details.
* `json` (Optional) Configures inspection of the response JSON. See [`json`](#json-block) for more details.
* `status_code` (Optional) Configures inspection of the response status code.See [`status_code`](#status_code-block) for more details.

### `body_contains` Block

* `success_strings` (Required) Strings in the body of the response that indicate a successful login attempt.
* `failure_strings` (Required) Strings in the body of the response that indicate a failed login attempt.

### `header` Block

* `name` (Required) The name of the header to match against. The name must be an exact match, including case.
* `success_values` (Required) Values in the response header with the specified name that indicate a successful login attempt.
* `failure_values` (Required) Values in the response header with the specified name that indicate a failed login attempt.

### `json` Block

* `identifier` (Required) The identifier for the value to match against in the JSON.
* `success_strings` (Required) Strings in the body of the response that indicate a successful login attempt.
* `failure_strings` (Required) Strings in the body of the response that indicate a failed login attempt.

### `status_code` Block

* `success_codes` (Required) Status codes in the response that indicate a successful login attempt.
* `failure_codes` (Required) Status codes in the response that indicate a failed login attempt.

### `field_to_match` Block

The part of a web request that you want AWS WAF to inspect. Include the single `field_to_match` type that you want to inspect, with additional specifications as needed, according to the type. You specify a single request component in `field_to_match` for each rule statement that requires it. To inspect more than one component of a web request, create a separate rule statement for each component. See the [documentation](https://docs.aws.amazon.com/waf/latest/developerguide/waf-rule-statement-fields.html#waf-rule-statement-request-component) for more details.

The `field_to_match` block supports the following arguments:

~> **Note** Only one of `all_query_arguments`, `body`, `cookies`, `header_order`, `headers`, `ja3_fingerprint`, `json_body`, `method`, `query_string`, `single_header`, `single_query_argument`, or `uri_path` can be specified. An empty configuration block `{}` should be used when specifying `all_query_arguments`, `method`, or `query_string` attributes.

* `all_query_arguments` - (Optional) Inspect all query arguments.
* `body` - (Optional) Inspect the request body, which immediately follows the request headers. See [`body`](#body-block) below for details.
* `cookies` - (Optional) Inspect the cookies in the web request. See [`cookies`](#cookies-block) below for details.
* `header_order` - (Optional) Inspect a string containing the list of the request's header names, ordered as they appear in the web request that AWS WAF receives for inspection. See [`header_order`](#header_order-block) below for details.
* `headers` - (Optional) Inspect the request headers. See [`headers`](#headers-block) below for details.
* `ja3_fingerprint` - (Optional) Inspect the JA3 fingerprint. See [`ja3_fingerprint`](#ja3_fingerprint-block) below for details.
* `json_body` - (Optional) Inspect the request body as JSON. See [`json_body`](#json_body-block) for details.
* `method` - (Optional) Inspect the HTTP method. The method indicates the type of operation that the request is asking the origin to perform.
* `query_string` - (Optional) Inspect the query string. This is the part of a URL that appears after a `?` character, if any.
* `single_header` - (Optional) Inspect a single header. See [`single_header`](#single_header-block) below for details.
* `single_query_argument` - (Optional) Inspect a single query argument. See [`single_query_argument`](#single_query_argument-block) below for details.
* `uri_path` - (Optional) Inspect the request URI path. This is the part of a web request that identifies a resource, for example, `/images/daily-ad.jpg`.

### `forwarded_ip_config` Block

The configuration for inspecting IP addresses in an HTTP header that you specify, instead of using the IP address that's reported by the web request origin. Commonly, this is the X-Forwarded-For (XFF) header, but you can specify any header name. If the specified header isn't present in the request, AWS WAFv2 doesn't apply the rule to the web request at all. AWS WAFv2 only evaluates the first IP address found in the specified HTTP header.

The `forwarded_ip_config` block supports the following arguments:

* `fallback_behavior` - (Required) - Match status to assign to the web request if the request doesn't have a valid IP address in the specified position. Valid values include: `MATCH` or `NO_MATCH`.
* `header_name` - (Required) - Name of the HTTP header to use for the IP address.

### `ip_set_forwarded_ip_config` Block

The configuration for inspecting IP addresses in an HTTP header that you specify, instead of using the IP address that's reported by the web request origin. Commonly, this is the X-Forwarded-For (XFF) header, but you can specify any header name.

The `ip_set_forwarded_ip_config` block supports the following arguments:

* `fallback_behavior` - (Required) - Match status to assign to the web request if the request doesn't have a valid IP address in the specified position. Valid values include: `MATCH` or `NO_MATCH`.
* `header_name` - (Required) - Name of the HTTP header to use for the IP address.
* `position` - (Required) - Position in the header to search for the IP address. Valid values include: `FIRST`, `LAST`, or `ANY`. If `ANY` is specified and the header contains more than 10 IP addresses, AWS WAFv2 inspects the last 10.

### `header_order` Block

Inspect a string containing the list of the request's header names, ordered as they appear in the web request that AWS WAF receives for inspection. AWS WAF generates the string and then uses that as the field to match component in its inspection. AWS WAF separates the header names in the string using colons and no added spaces, for example `host:user-agent:accept:authorization:referer`.

The `header_order` block supports the following arguments:

* `oversize_handling` - (Required) Oversize handling tells AWS WAF what to do with a web request when the request component that the rule inspects is over the limits. Valid values include the following: `CONTINUE`, `MATCH`, `NO_MATCH`. See the AWS [documentation](https://docs.aws.amazon.com/waf/latest/developerguide/waf-rule-statement-oversize-handling.html) for more information.

### `headers` Block

Inspect the request headers.

The `headers` block supports the following arguments:

* `match_pattern` - (Required) The filter to use to identify the subset of headers to inspect in a web request. The `match_pattern` block supports only one of the following arguments:
    * `all` - An empty configuration block that is used for inspecting all headers.
    * `included_headers` - An array of strings that will be used for inspecting headers that have a key that matches one of the provided values.
    * `excluded_headers` - An array of strings that will be used for inspecting headers that do not have a key that matches one of the provided values.
* `match_scope` - (Required) The parts of the headers to inspect with the rule inspection criteria. If you specify `All`, AWS WAF inspects both keys and values. Valid values include the following: `ALL`, `Key`, `Value`.
* `oversize_handling` - (Required) Oversize handling tells AWS WAF what to do with a web request when the request component that the rule inspects is over the limits. Valid values include the following: `CONTINUE`, `MATCH`, `NO_MATCH`. See the AWS [documentation](https://docs.aws.amazon.com/waf/latest/developerguide/waf-rule-statement-oversize-handling.html) for more information.

### `ja3_fingerprint` Block

The `ja3_fingerprint` block supports the following arguments:

* `fallback_behavior` - (Required) The match status to assign to the web request if the request doesn't have a JA3 fingerprint. Valid values include: `MATCH` or `NO_MATCH`.

### `json_body` Block

The `json_body` block supports the following arguments:

* `invalid_fallback_behavior` - (Optional) What to do when JSON parsing fails. Defaults to evaluating up to the first parsing failure. Valid values are `EVALUATE_AS_STRING`, `MATCH` and `NO_MATCH`.
* `match_pattern` - (Required) The patterns to look for in the JSON body. You must specify exactly one setting: either `all` or `included_paths`. See [JsonMatchPattern](https://docs.aws.amazon.com/waf/latest/APIReference/API_JsonMatchPattern.html) for details.
* `match_scope` - (Required) The parts of the JSON to match against using the `match_pattern`. Valid values are `ALL`, `KEY` and `VALUE`.
* `oversize_handling` - (Optional) What to do if the body is larger than can be inspected. Valid values are `CONTINUE` (default), `MATCH` and `NO_MATCH`.

### `single_header` Block

Inspect a single header. Provide the name of the header to inspect, for example, `User-Agent` or `Referer` (provided as lowercase strings).

The `single_header` block supports the following arguments:

* `name` - (Required) Name of the query header to inspect. This setting must be provided as lower case characters.

### `single_query_argument` Block

Inspect a single query argument. Provide the name of the query argument to inspect, such as `UserName` or `SalesRegion` (provided as lowercase strings).

The `single_query_argument` block supports the following arguments:

* `name` - (Required) Name of the query header to inspect. This setting must be provided as lower case characters.

### `body` Block

The `body` block supports the following arguments:

* `oversize_handling` - (Optional) What WAF should do if the body is larger than WAF can inspect. WAF does not support inspecting the entire contents of the body of a web request when the body exceeds 8 KB (8192 bytes). Only the first 8 KB of the request body are forwarded to WAF by the underlying host service. Valid values: `CONTINUE`, `MATCH`, `NO_MATCH`.

### `cookies` Block

Inspect the cookies in the web request. You can specify the parts of the cookies to inspect and you can narrow the set of cookies to inspect by including or excluding specific keys. This is used to indicate the web request component to inspect, in the [FieldToMatch](https://docs.aws.amazon.com/waf/latest/APIReference/API_FieldToMatch.html) specification.

The `cookies` block supports the following arguments:

* `match_pattern` - (Required) The filter to use to identify the subset of cookies to inspect in a web request. You must specify exactly one setting: either `all`, `included_cookies` or `excluded_cookies`. More details: [CookieMatchPattern](https://docs.aws.amazon.com/waf/latest/APIReference/API_CookieMatchPattern.html)
* `match_scope` - (Required) The parts of the cookies to inspect with the rule inspection criteria. If you specify All, AWS WAF inspects both keys and values. Valid values: `ALL`, `KEY`, `VALUE`
* `oversize_handling` - (Required) What AWS WAF should do if the cookies of the request are larger than AWS WAF can inspect. AWS WAF does not support inspecting the entire contents of request cookies when they exceed 8 KB (8192 bytes) or 200 total cookies. The underlying host service forwards a maximum of 200 cookies and at most 8 KB of cookie contents to AWS WAF. Valid values: `CONTINUE`, `MATCH`, `NO_MATCH`.

### `text_transformation` Block

The `text_transformation` block supports the following arguments:

* `priority` - (Required) Relative processing order for multiple transformations that are defined for a rule statement. AWS WAF processes all transformations, from lowest priority to highest, before inspecting the transformed content.
* `type` - (Required) Transformation to apply, please refer to the Text Transformation [documentation](https://docs.aws.amazon.com/waf/latest/APIReference/API_TextTransformation.html) for more details.

### `visibility_config` Block

The `visibility_config` block supports the following arguments:

* `cloudwatch_metrics_enabled` - (Required) Whether the associated resource sends metrics to CloudWatch. For the list of available metrics, see [AWS WAF Metrics](https://docs.aws.amazon.com/waf/latest/developerguide/monitoring-cloudwatch.html#waf-metrics).
* `metric_name` - (Required) A friendly name of the CloudWatch metric. The name can contain only alphanumeric characters (A-Z, a-z, 0-9) hyphen(-) and underscore (\_), with length from one to 128 characters. It can't contain whitespace or metric names reserved for AWS WAF, for example `All` and `Default_Action`.
* `sampled_requests_enabled` - (Required) Whether AWS WAF should store a sampling of the web requests that match the rules. You can view the sampled requests through the AWS WAF console.

### `captcha_config` Block

The `captcha_config` block supports the following arguments:

* `immunity_time_property` - (Optional) Defines custom immunity time. See [`immunity_time_property`](#immunity_time_property-block) below for details.

### `challenge_config` Block

The `challenge_config` block supports the following arguments:

* `immunity_time_property` - (Optional) Defines custom immunity time. See [`immunity_time_property`](#immunity_time_property-block) below for details.

### `immunity_time_property` Block

The `immunity_time_property` block supports the following arguments:

* `immunity_time` - (Optional) The amount of time, in seconds, that a CAPTCHA or challenge timestamp is considered valid by AWS WAF. The default setting is 300.

### `request_body` Block

The `request_body` block supports the following arguments:

* `api_gateway` - (Optional) Customizes the request body that your protected Amazon API Gateway REST APIs forward to AWS WAF for inspection. Applicable only when `scope` is set to `CLOUDFRONT`. See [`api_gateway`](#api_gateway-block) below for details.
* `app_runner_service` - (Optional) Customizes the request body that your protected Amazon App Runner services forward to AWS WAF for inspection. Applicable only when `scope` is set to `REGIONAL`. See [`app_runner_service`](#app_runner_service-block) below for details.
* `cloudfront` - (Optional) Customizes the request body that your protected Amazon CloudFront distributions forward to AWS WAF for inspection. Applicable only when `scope` is set to `REGIONAL`. See [`cloudfront`](#cloudfront-block) below for details.
* `cognito_user_pool` - (Optional) Customizes the request body that your protected Amazon Cognito user pools forward to AWS WAF for inspection. Applicable only when `scope` is set to `REGIONAL`. See [`cognito_user_pool`](#cognito_user_pool-block) below for details.
* `verified_access_instance` - (Optional) Customizes the request body that your protected AWS Verfied Access instances forward to AWS WAF for inspection. Applicable only when `scope` is set to `REGIONAL`. See [`verified_access_instance`](#verified_access_instance-block) below for details.

### `api_gateway` Block

The `api_gateway` block supports the following arguments:

* `default_size_inspection_limit` - (Required) Specifies the maximum size of the web request body component that an associated Amazon API Gateway REST APIs should send to AWS WAF for inspection. This applies to statements in the web ACL that inspect the body or JSON body. Valid values are `KB_16`, `KB_32`, `KB_48` and `KB_64`.

### `app_runner_service` Block

The `app_runner_service` block supports the following arguments:

* `default_size_inspection_limit` - (Required) Specifies the maximum size of the web request body component that an associated Amazon App Runner services should send to AWS WAF for inspection. This applies to statements in the web ACL that inspect the body or JSON body. Valid values are `KB_16`, `KB_32`, `KB_48` and `KB_64`.

### `cloudfront` Block

The `cloudfront` block supports the following arguments:

* `default_size_inspection_limit` - (Required) Specifies the maximum size of the web request body component that an associated Amazon CloudFront distribution should send to AWS WAF for inspection. This applies to statements in the web ACL that inspect the body or JSON body. Valid values are `KB_16`, `KB_32`, `KB_48` and `KB_64`.

### `cognito_user_pool` Block

The `cognito_user_pool` block supports the following arguments:

* `default_size_inspection_limit` - (Required) Specifies the maximum size of the web request body component that an associated Amazon Cognito user pools should send to AWS WAF for inspection. This applies to statements in the web ACL that inspect the body or JSON body. Valid values are `KB_16`, `KB_32`, `KB_48` and `KB_64`.

### `verified_access_instance` Block

The `verified_access_instance` block supports the following arguments:

* `default_size_inspection_limit` - (Required) Specifies the maximum size of the web request body component that an associated AWS Verified Access instances should send to AWS WAF for inspection. This applies to statements in the web ACL that inspect the body or JSON body. Valid values are `KB_16`, `KB_32`, `KB_48` and `KB_64`.

### `custom_key` Block

Aggregate the request counts using one or more web request components as the aggregate keys. With this option, you must specify the aggregate keys in the `custom_keys` block. To aggregate on only the IP address or only the forwarded IP address, don't use custom keys. Instead, set the `aggregate_key_type` to `IP` or `FORWARDED_IP`.

The `custom_key` block supports the following arguments:

* `cookie` - (Optional) Use the value of a cookie in the request as an aggregate key. See [RateLimit `cookie`](#ratelimit-cookie-block) below for details.
* `forwarded_ip` - (Optional) Use the first IP address in an HTTP header as an aggregate key. See [`forwarded_ip`](#ratelimit-forwarded_ip-block) below for details.
* `http_method` - (Optional) Use the request's HTTP method as an aggregate key. See [RateLimit `http_method`](#ratelimit-http_method-block) below for details.
* `header` - (Optional) Use the value of a header in the request as an aggregate key. See [RateLimit `header`](#ratelimit-header-block) below for details.
* `ip` - (Optional) Use the request's originating IP address as an aggregate key. See [`RateLimit ip`](#ratelimit-ip-block) below for details.
* `label_namespace` - (Optional) Use the specified label namespace as an aggregate key. See [RateLimit `label_namespace`](#ratelimit-label_namespace-block) below for details.
* `query_argument` - (Optional) Use the specified query argument as an aggregate key. See [RateLimit `query_argument`](#ratelimit-query_argument-block) below for details.
* `query_string` - (Optional) Use the request's query string as an aggregate key. See [RateLimit `query_string`](#ratelimit-query_string-block) below for details.
* `uri_path` - (Optional) Use the request's URI path as an aggregate key. See [RateLimit `uri_path`](#ratelimit-uri_path-block) below for details.

### RateLimit `cookie` Block

Use the value of a cookie in the request as an aggregate key. Each distinct value in the cookie contributes to the aggregation instance. If you use a single cookie as your custom key, then each value fully defines an aggregation instance.

The `cookie` block supports the following arguments:

* `name`: The name of the cookie to use.
* `text_transformation`: Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. They are used in rate-based rule statements, to transform request components before using them as custom aggregation keys. Atleast one transformation is required. See [`text_transformation`](#text_transformation-block) above for details.

### RateLimit `forwarded_ip` Block

Use the first IP address in an HTTP header as an aggregate key. Each distinct forwarded IP address contributes to the aggregation instance. When you specify an IP or forwarded IP in the custom key settings, you must also specify at least one other key to use. You can aggregate on only the forwarded IP address by specifying `FORWARDED_IP` in your rate-based statement's `aggregate_key_type`. With this option, you must specify the header to use in the rate-based rule's [`forwarded_ip_config`](#forwarded_ip_config-block) block.

The `forwarded_ip` block is configured as an empty block `{}`.

### RateLimit `http_method` Block

Use the request's HTTP method as an aggregate key. Each distinct HTTP method contributes to the aggregation instance. If you use just the HTTP method as your custom key, then each method fully defines an aggregation instance.

The `http_method` block is configured as an empty block `{}`.

### RateLimit `header` Block

Use the value of a header in the request as an aggregate key. Each distinct value in the header contributes to the aggregation instance. If you use a single header as your custom key, then each value fully defines an aggregation instance.

The `header` block supports the following arguments:

* `name`: The name of the header to use.
* `text_transformation`: Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. They are used in rate-based rule statements, to transform request components before using them as custom aggregation keys. Atleast one transformation is required. See [`text_transformation`](#text_transformation-block) above for details.

### RateLimit `ip` Block

Use the request's originating IP address as an aggregate key. Each distinct IP address contributes to the aggregation instance. When you specify an IP or forwarded IP in the custom key settings, you must also specify at least one other key to use. You can aggregate on only the IP address by specifying `IP` in your rate-based statement's `aggregate_key_type`.

The `ip` block is configured as an empty block `{}`.

### RateLimit `label_namespace` Block

Use the specified label namespace as an aggregate key. Each distinct fully qualified label name that has the specified label namespace contributes to the aggregation instance. If you use just one label namespace as your custom key, then each label name fully defines an aggregation instance. This uses only labels that have been added to the request by rules that are evaluated before this rate-based rule in the web ACL. For information about label namespaces and names, see Label syntax and naming requirements (https://docs.aws.amazon.com/waf/latest/developerguide/waf-rule-label-requirements.html) in the WAF Developer Guide.

The `label_namespace` block supports the following arguments:

* `namespace`: The namespace to use for aggregation

### RateLimit `query_argument` Block

Use the specified query argument as an aggregate key. Each distinct value for the named query argument contributes to the aggregation instance. If you use a single query argument as your custom key, then each value fully defines an aggregation instance.

The `query_argument` block supports the following arguments:

* `name`: The name of the query argument to use.
* `text_transformation`: Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. They are used in rate-based rule statements, to transform request components before using them as custom aggregation keys. Atleast one transformation is required. See [`text_transformation`](#text_transformation-block) above for details.

### RateLimit `query_string` Block

Use the request's query string as an aggregate key. Each distinct string contributes to the aggregation instance. If you use just the query string as your custom key, then each string fully defines an aggregation instance.

The `query_string` block supports the following arguments:

* `text_transformation`: Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. They are used in rate-based rule statements, to transform request components before using them as custom aggregation keys. Atleast one transformation is required. See [`text_transformation`](#text_transformation-block) above for details.

### RateLimit `uri_path` Block

Use the request's URI path as an aggregate key. Each distinct URI path contributes to the aggregation instance. If you use just the URI path as your custom key, then each URI path fully defines an aggregation instance.

The `uri_path` block supports the following arguments:

* `text_transformation`: Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. They are used in rate-based rule statements, to transform request components before using them as custom aggregation keys. Atleast one transformation is required. See [`text_transformation`](#text_transformation-block) above for details.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `application_integration_url` - The URL to use in SDK integrations with managed rule groups.
* `arn` - The ARN of the WAF WebACL.
* `capacity` - Web ACL capacity units (WCUs) currently being used by this web ACL.
* `id` - The ID of the WAF WebACL.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAFv2 Web ACLs using `ID/Name/Scope`. For example:

```terraform
import {
  to = aws_wafv2_web_acl.example
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc/example/REGIONAL"
}
```

Using `terraform import`, import WAFv2 Web ACLs using `ID/Name/Scope`. For example:

```console
% terraform import aws_wafv2_web_acl.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc/example/REGIONAL
```
