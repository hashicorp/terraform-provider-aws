---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_managed_rule_group"
description: |-
   High-level information for a managed rule group.
---

# Data Source: aws_wafv2_managed_rule_group

High-level information for a managed rule group.

## Example Usage

```terraform
data "aws_wafv2_managed_rule_group" "example" {
  name        = "AWSManagedRulesCommonRuleSet"
  scope       = "REGIONAL"
  vendor_name = "AWS"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Managed rule group name.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `scope` - (Required) Whether this is for a global resource type, such as a Amazon CloudFront distribution. For an AWS Amplify application, use `CLOUDFRONT`. Valid values: `CLOUDFRONT`, `REGIONAL`.
* `vendor_name` - (Required) Managed rule group vendor name.
* `version_name` - (Optional) Version of the rule group.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `available_labels` - Labels that one or more rules in this rule group add to matching web requests. See [Labels](#labels) below for details.
* `capacity` - WCUs required for this rule group.
* `consumed_labels` - Labels that one or more rules in this rule group match against in label match statements. See [Labels](#labels) below for details.
* `label_namespace` - Label namespace prefix for this rule group. All labels added by rules in this rule group have this prefix.
* `rules` - High-level information about the rules. See [`rules` Block](#rules-block) below for details.
* `sns_topic_arn` - ARN of the SNS topic that's used to provide notification of changes to the managed rule group.

### Labels

* `name` - Individual label specification.

### `rules` Block

* `action` - Action taken on a web request when it matches a rule's statement. See [`action` Block](#action-block) for details.
* `name` - Name of the rule.

### `action` Block

* `allow` - Rule action that allows the request. See [`allow` Block](#allow-block) for details.
* `block` - Rule action that blocks the request. See [`block` Block](#block-block) for details.
* `captcha` - Rule action that requires CAPTCHA verification. See [`captcha` Block](#captcha-block) for details.
* `challenge` - Rule action that requires challenge verification. See [`challenge` Block](#challenge-block) for details.
* `count` - Rule action that counts the request without taking other action. See [`count` Block](#count-block) for details.

### `allow` Block

* `custom_request_handling` - Custom handling for the allowed request. See [`custom_request_handling` Block](#custom_request_handling-block) for details.

### `block` Block

* `custom_response` - Custom response for the blocked request. See [`custom_response` Block](#custom_response-block) for details.

### `captcha` Block

* `custom_request_handling` - Custom handling for the CAPTCHA request. See [`custom_request_handling` Block](#custom_request_handling-block) for details.

### `challenge` Block

* `custom_request_handling` - Custom handling for the challenge request. See [`custom_request_handling` Block](#custom_request_handling-block) for details.

### `count` Block

* `custom_request_handling` - Custom handling for the counted request. See [`custom_request_handling` Block](#custom_request_handling-block) for details.

### `custom_request_handling` Block

* `insert_header` - Headers inserted into the request. See [`insert_header` Block](#insert_header-block) for details.

### `custom_response` Block

* `custom_response_body_key` - Key of the custom response body to use.
* `response_code` - HTTP response code returned.
* `response_header` - Headers included in the response. See [`response_header` Block](#response_header-block) for details.

### `insert_header` Block

* `name` - Name of the header.
* `value` - Value of the header.

### `response_header` Block

* `name` - Name of the header.
* `value` - Value of the header.
