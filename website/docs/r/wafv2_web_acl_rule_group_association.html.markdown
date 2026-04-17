---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_web_acl_rule_group_association"
description: |-
  Associates a WAFv2 Rule Group with a Web ACL by adding a rule that references the Rule Group.
---

# Resource: aws_wafv2_web_acl_rule_group_association

Associates a WAFv2 Rule Group (custom or managed) with a Web ACL by adding a rule that references the Rule Group. Use this resource to apply the rules defined in a Rule Group to a Web ACL without duplicating rule definitions.

This resource supports both:

- **Custom Rule Groups**: User-created rule groups that you manage within your AWS account
- **Managed Rule Groups**: Pre-configured rule groups provided by AWS or third-party vendors

!> **Warning:** Verify the rule names in your `rule_action_override`s carefully. With managed rule groups, WAF silently ignores any override that uses an invalid rule name. With customer-owned rule groups, invalid rule names in your overrides will cause web ACL updates to fail. An invalid rule name is any name that doesn't exactly match the case-sensitive name of an existing rule in the rule group.

!> **Warning:** Using this resource will cause the associated Web ACL resource to show configuration drift in the `rule` argument unless you add `lifecycle { ignore_changes = [rule] }` to the Web ACL resource configuration. This is because this resource modifies the Web ACL's rules outside of the Web ACL resource's direct management.

~> **Note:** This resource creates a rule within the Web ACL that references the entire Rule Group. The rule group's individual rules are evaluated as a unit when requests are processed by the Web ACL.

## Example Usage

### Basic Usage

```terraform
# Web ACL must use lifecycle.ignore_changes to prevent drift from this resource
resource "aws_wafv2_web_acl" "example" {
  name  = "example-web-acl"
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "example-web-acl"
    sampled_requests_enabled   = true
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

# Associate a custom rule group
resource "aws_wafv2_web_acl_rule_group_association" "example" {
  rule_name   = "example-rule-group-rule"
  priority    = 100
  web_acl_arn = aws_wafv2_web_acl.example.arn

  rule_group_reference {
    arn = aws_wafv2_rule_group.example.arn
  }
}
```

### Managed Rule Group

```terraform
resource "aws_wafv2_web_acl_rule_group_association" "example" {
  rule_name   = "aws-common-rule-set"
  priority    = 50
  web_acl_arn = aws_wafv2_web_acl.example.arn

  managed_rule_group {
    name        = "AWSManagedRulesCommonRuleSet"
    vendor_name = "AWS"
  }
}
```

### Managed Rule Group With Version

```terraform
resource "aws_wafv2_web_acl_rule_group_association" "example" {
  rule_name   = "aws-common-rule-set-versioned"
  priority    = 60
  web_acl_arn = aws_wafv2_web_acl.example.arn

  managed_rule_group {
    name        = "AWSManagedRulesCommonRuleSet"
    vendor_name = "AWS"
    version     = "Version_1.0"
  }
}
```

### Managed Rule Group With Rule Action Overrides

```terraform
resource "aws_wafv2_web_acl_rule_group_association" "example" {
  rule_name   = "aws-common-rule-set-with-overrides"
  priority    = 70
  web_acl_arn = aws_wafv2_web_acl.example.arn

  managed_rule_group {
    name        = "AWSManagedRulesCommonRuleSet"
    vendor_name = "AWS"

    rule_action_override {
      name = "GenericRFI_BODY"
      action_to_use {
        count {
          custom_request_handling {
            insert_header {
              name  = "X-RFI-Override"
              value = "counted"
            }
          }
        }
      }
    }

    rule_action_override {
      name = "SizeRestrictions_BODY"
      action_to_use {
        captcha {}
      }
    }
  }
}
```

### Managed Rule Group With Managed Rule Group Configs

```terraform
resource "aws_wafv2_web_acl_rule_group_association" "example" {
  rule_name   = "acfp-ruleset-with-rule-config"
  priority    = 70
  web_acl_arn = aws_wafv2_web_acl.example.arn

  managed_rule_group {
    name        = "AWSManagedRulesACFPRuleSet"
    vendor_name = "AWS"

    managed_rule_group_configs {
      aws_managed_rules_acfp_rule_set {
        creation_path          = "/creation"
        registration_page_path = "/registration"
        request_inspection {
          email_field {
            identifier = "/email"
          }
          password_field {
            identifier = "/password"
          }
          phone_number_fields {
            identifiers = ["/phone1", "/phone2"]
          }
          address_fields {
            identifiers = ["home", "work"]
          }
          payload_type = "JSON"
          username_field {
            identifier = "/username"
          }
        }
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = true
  }
}
```

### Custom Rule Group With Override Action

```terraform
resource "aws_wafv2_web_acl_rule_group_association" "example" {
  rule_name       = "example-rule-group-rule"
  priority        = 100
  web_acl_arn     = aws_wafv2_web_acl.example.arn
  override_action = "count"

  rule_group_reference {
    arn = aws_wafv2_rule_group.example.arn
  }
}
```

### Custom Rule Group With Rule Action Overrides

```terraform
resource "aws_wafv2_web_acl_rule_group_association" "example" {
  rule_name   = "example-rule-group-rule"
  priority    = 100
  web_acl_arn = aws_wafv2_web_acl.example.arn

  rule_group_reference {
    arn = aws_wafv2_rule_group.example.arn

    rule_action_override {
      name = "geo-block-rule"
      action_to_use {
        count {
          custom_request_handling {
            insert_header {
              name  = "X-Geo-Block-Override"
              value = "counted"
            }
          }
        }
      }
    }

    rule_action_override {
      name = "rate-limit-rule"
      action_to_use {
        captcha {
          custom_request_handling {
            insert_header {
              name  = "X-Rate-Limit-Override"
              value = "captcha-required"
            }
          }
        }
      }
    }
  }
}
```

### CloudFront Web ACL

```terraform
resource "aws_wafv2_web_acl_rule_group_association" "example" {
  rule_name   = "cloudfront-rule-group-rule"
  priority    = 50
  web_acl_arn = aws_wafv2_web_acl.example.arn

  rule_group_reference {
    arn = aws_wafv2_rule_group.example.arn
  }
}
```

## Argument Reference

The following arguments are required:

* `rule_name` - (Required) Name of the rule to create in the Web ACL that references the rule group. Must be between 1 and 128 characters.
* `priority` - (Required) Priority of the rule within the Web ACL. Rules are evaluated in order of priority, with lower numbers evaluated first.
* `web_acl_arn` - (Required) ARN of the Web ACL to associate the Rule Group with.

The following arguments are optional:

* `managed_rule_group` - (Optional) Managed Rule Group configuration. One of `rule_group_reference` or `managed_rule_group` is required. Conflicts with `rule_group_reference`. [See below](#managed_rule_group).
* `override_action` - (Optional) Override action for the rule group. Valid values are `none` and `count`. Defaults to `none`. When set to `count`, the actions defined in the rule group rules are overridden to count matches instead of blocking or allowing requests.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `rule_group_reference` - (Optional) Custom Rule Group reference configuration. One of `rule_group_reference` or `managed_rule_group` is required. Conflicts with `managed_rule_group`. [See below](#rule_group_reference).
* `visibility_config` - (Optional) Defines and enables Amazon CloudWatch metrics and web request sample collection. [See below](#visibility_config).

### visibility_config

* `cloudwatch_metrics_enabled` - (Required) Whether the associated resource sends metrics to CloudWatch. For the list of available metrics, see [AWS WAF Metrics](https://docs.aws.amazon.com/waf/latest/developerguide/monitoring-cloudwatch.html#waf-metrics).
* `metric_name` - (Required) Friendly name of the CloudWatch metric. The name can contain only alphanumeric characters (A-Z, a-z, 0-9) hyphen(-) and underscore (\_), with length from one to 128 characters. It can't contain whitespace or metric names reserved for AWS WAF, for example `All` and `Default_Action`.
* `sampled_requests_enabled` - (Required) Whether AWS WAF should store a sampling of the web requests that match the rules. You can view the sampled requests through the AWS WAF console.

### rule_group_reference

* `arn` - (Required) ARN of the Rule Group to associate with the Web ACL.
* `rule_action_override` - (Optional) Override actions for specific rules within the rule group. [See below](#rule_action_override).

### managed_rule_group

* `managed_rule_group_configs` - (Optional) Additional information that's used by a managed rule group. Only one rule attribute is allowed in each config. [See below](#managed_rule_group_configs).
* `name` - (Required) Name of the managed rule group.
* `rule_action_override` - (Optional) Override actions for specific rules within the rule group. [See below](#rule_action_override).
* `vendor_name` - (Required) Name of the managed rule group vendor. For AWS managed rule groups, this is `AWS`.
* `version` - (Optional) Version of the managed rule group. If not specified, the default version is used.

### rule_action_override

* `name` - (Required) Name of the rule to override within the rule group. Verify the name carefully. With managed rule groups, WAF silently ignores any override that uses an invalid rule name. With customer-owned rule groups, invalid rule names in your overrides will cause web ACL updates to fail. An invalid rule name is any name that doesn't exactly match the case-sensitive name of an existing rule in the rule group.
* `action_to_use` - (Required) Action to use instead of the rule's original action. [See below](#action_to_use).

### action_to_use

Exactly one of the following action blocks must be specified:

* `allow` - (Optional) Allow the request. [See below](#allow).
* `block` - (Optional) Block the request. [See below](#block).
* `captcha` - (Optional) Require CAPTCHA verification. [See below](#captcha).
* `challenge` - (Optional) Require challenge verification. [See below](#challenge).
* `count` - (Optional) Count the request without taking action. [See below](#count).

### allow

* `custom_request_handling` - (Optional) Custom handling for allowed requests. [See below](#custom_request_handling).

### block

* `custom_response` - (Optional) Custom response for blocked requests. [See below](#custom_response).

### captcha

* `custom_request_handling` - (Optional) Custom handling for CAPTCHA requests. [See below](#custom_request_handling).

### challenge

* `custom_request_handling` - (Optional) Custom handling for challenge requests. [See below](#custom_request_handling).

### count

* `custom_request_handling` - (Optional) Custom handling for counted requests. [See below](#custom_request_handling).

### custom_request_handling

* `insert_header` - (Required) Headers to insert into the request. [See below](#insert_header).

### custom_response

* `custom_response_body_key` - (Optional) Key of a custom response body to use.
* `response_code` - (Required) HTTP response code to return (200-599).
* `response_header` - (Optional) Headers to include in the response. [See below](#response_header).

### insert_header

* `name` - (Required) Name of the header to insert.
* `value` - (Required) Value of the header to insert.

### response_header

* `name` - (Required) Name of the response header.
* `value` - (Required) Value of the response header.

### managed_rule_group_configs

* `aws_managed_rules_acfp_rule_set` - (Optional) Additional configuration for using the Account Creation Fraud Prevention managed rule group. Use this to specify information such as the registration page of your application and the type of content to accept or reject from the client. [See below](#aws_managed_rules_acfp_rule_set).
* `aws_managed_rules_anti_ddos_rule_set` - (Optional) Configuration for using the anti-DDoS managed rule group. [See below](#aws_managed_rules_anti_ddos_rule_set).
* `aws_managed_rules_atp_rule_set` - (Optional) Additional configuration for using the Account Takeover Protection managed rule group. Use this to specify information such as the sign-in page of your application and the type of content to accept or reject from the client. [See below](#aws_managed_rules_atp_rule_set).
* `aws_managed_rules_bot_control_rule_set` - (Optional) Additional configuration for using the Bot Control managed rule group. Use this to specify the inspection level that you want to use. [See below](#aws_managed_rules_bot_control_rule_set).

### aws_managed_rules_bot_control_rule_set

* `enable_machine_learning` - (Optional) Applies only to the targeted inspection level. Determines whether to use machine learning (ML) to analyze your web traffic for bot-related activity. Defaults to `false`.
* `inspection_level` - (Optional) Inspection level to use for the Bot Control rule group.

### aws_managed_rules_acfp_rule_set

* `creation_path` - (Required) Path of the account creation endpoint for your application. This is the page on your website that accepts the completed registration form for a new user. This page must accept POST requests.
* `enable_regex_in_path` - (Optional) Whether or not to allow the use of regular expressions in the login page path.
* `registration_page_path` - (Required) Path of the account registration endpoint for your application. This is the page on your website that presents the registration form to new users. This page must accept GET text/html requests.
* `request_inspection` - (Optional) Criteria for inspecting login requests, used by the ATP rule group to validate credentials usage. [See below](#request_inspection_acfp).
* `response_inspection` - (Optional) Criteria for inspecting responses to login requests, used by the ATP rule group to track login failure rates. Note that Response Inspection is available only on web ACLs that protect CloudFront distributions. [See below](#response_inspection).

### request_inspection_acfp

* `address_fields` - (Optional) Names of the fields in the request payload that contain your customer's primary physical address. [See below](#address_fields).
* `email_field` - (Optional) Name of the field in the request payload that contains your customer's email. [See below](#email_field).
* `password_field` - (Optional) Details about your login page password field. [See below](#password_field).
* `payload_type` - (Required) Payload type for your login endpoint, either JSON or form encoded.
* `phone_number_fields` - (Optional) Names of the fields in the request payload that contain your customer's primary phone number. [See below](#phone_number_fields).
* `username_field` - (Optional) Details about your login page username field. [See below](#username_field).

### aws_managed_rules_anti_ddos_rule_set

* `client_side_action_config` - (Required) Configuration for the request handling that's applied by the managed rule group rules `ChallengeAllDuringEvent` and `ChallengeDDoSRequests` during a distributed denial of service (DDoS) attack. [See below](#client_side_action_config).
* `sensitivity_to_block` - (Optional) Sensitivity that the rule group rule DDoSRequests uses when matching against the DDoS suspicion labeling on a request. Valid values are `LOW` (Default), `MEDIUM`, and `HIGH`.

### client_side_action_config

* `challenge` - (Required) Configuration for the use of the `AWSManagedRulesAntiDDoSRuleSet` rules `ChallengeAllDuringEvent` and `ChallengeDDoSRequests`. [See below](#challenge_config).

### challenge_config

* `exempt_uri_regular_expression` - (Optional) Block for the list of the regular expressions to match against the web request URI, used to identify requests that can't handle a silent browser challenge. [See below](#exempt_uri_regular_expression).
* `sensitivity` - (Optional) Sensitivity that the rule group rule ChallengeDDoSRequests uses when matching against the DDoS suspicion labeling on a request. Valid values are `LOW`, `MEDIUM` and `HIGH` (Default).
* `usage_of_action` - (Required) Configuration whether to use the `AWSManagedRulesAntiDDoSRuleSet` rules `ChallengeAllDuringEvent` and `ChallengeDDoSRequests` in the rule group evaluation. Valid values are `ENABLED` and `DISABLED`.

### exempt_uri_regular_expression

* `regex_string` - (Optional) Regular expression string.

### aws_managed_rules_atp_rule_set

* `enable_regex_in_path` - (Optional) Whether or not to allow the use of regular expressions in the login page path.
* `login_path` - (Required) Path of the login endpoint for your application.
* `request_inspection` - (Optional) Criteria for inspecting login requests, used by the ATP rule group to validate credentials usage. [See below](#request_inspection).
* `response_inspection` - (Optional) Criteria for inspecting responses to login requests, used by the ATP rule group to track login failure rates. Note that Response Inspection is available only on web ACLs that protect CloudFront distributions. [See below](#response_inspection).

### request_inspection

* `password_field` - (Optional) Details about your login page password field. [See below](#password_field).
* `payload_type` - (Required) Payload type for your login endpoint, either JSON or form encoded.
* `username_field` - (Optional) Details about your login page username field. [See below](#username_field).

### address_fields

* `identifiers` - (Required) Names of the address fields.

### email_field

* `identifier` - (Required) Name of the field in the request payload that contains your customer's email.

### password_field

* `identifier` - (Required) Name of the password field.

### phone_number_fields

* `identifiers` - (Required) Names of the phone number fields.

### username_field

* `identifier` - (Required) Name of the username field.

### response_inspection

* `body_contains` - (Optional) Configures inspection of the response body. [See below](#body_contains).
* `header` - (Optional) Configures inspection of the response header. [See below](#header).
* `json` - (Optional) Configures inspection of the response JSON. [See below](#json).
* `status_code` - (Optional) Configures inspection of the response status code. [See below](#status_code).

### body_contains

* `failure_strings` - (Required) Strings in the body of the response that indicate a failed login attempt.
* `success_strings` - (Required) Strings in the body of the response that indicate a successful login attempt.

### header

* `failure_values` - (Required) Values in the response header with the specified name that indicate a failed login attempt.
* `name` - (Required) Name of the header to match against. The name must be an exact match, including case.
* `success_values` - (Required) Values in the response header with the specified name that indicate a successful login attempt.

### json

* `failure_strings` - (Required) Strings in the body of the response that indicate a failed login attempt.
* `identifier` - (Required) Identifier for the value to match against in the JSON.
* `success_strings` - (Required) Strings in the body of the response that indicate a successful login attempt.

### status_code

* `success_codes` (Required) Status codes in the response that indicate a successful login attempt.
* `failure_codes` (Required) Status codes in the response that indicate a failed login attempt.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

None.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAFv2 web ACL custom rule group associations using `WebACLARN,RuleGroupARN,RuleName`. For example:

```terraform
import {
  to = aws_wafv2_web_acl_rule_group_association.example
  id = "arn:aws:wafv2:us-east-1:123456789012:regional/webacl/example-web-acl/12345678-1234-1234-1234-123456789012,arn:aws:wafv2:us-east-1:123456789012:regional/rulegroup/example-rule-group/87654321-4321-4321-4321-210987654321,example-rule-group-rule"
}
```

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAFv2 web ACL managed rule group associations using `WebACLARN,VendorName:RuleGroupName[:Version],RuleName`. For example:

```terraform
import {
  to = aws_wafv2_web_acl_rule_group_association.managed_example
  id = "arn:aws:wafv2:us-east-1:123456789012:regional/webacl/example-web-acl/12345678-1234-1234-1234-123456789012,AWS:AWSManagedRulesCommonRuleSet,aws-common-rule-set"
}
```

Using `terraform import`, import WAFv2 web ACL custom rule group associations using `WebACLARN,RuleGroupARN,RuleName`. For example:

```console
% terraform import aws_wafv2_web_acl_rule_group_association.example "arn:aws:wafv2:us-east-1:123456789012:regional/webacl/example-web-acl/12345678-1234-1234-1234-123456789012,arn:aws:wafv2:us-east-1:123456789012:regional/rulegroup/example-rule-group/87654321-4321-4321-4321-210987654321,example-rule-group-rule"
```

Using `terraform import`, import WAFv2 web ACL managed rule group associations using `WebACLARN,VendorName:RuleGroupName[:Version],RuleName`. For example:

```console
% terraform import aws_wafv2_web_acl_rule_group_association.managed_example "arn:aws:wafv2:us-east-1:123456789012:regional/webacl/example-web-acl/12345678-1234-1234-1234-123456789012,AWS:AWSManagedRulesCommonRuleSet,aws-common-rule-set"
```
