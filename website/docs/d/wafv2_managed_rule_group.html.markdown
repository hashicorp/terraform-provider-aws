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
* `rules` - High-level information about the rules. See [Rules](#rules) below for details.
* `sns_topic_arn` - ARN of the SNS topic that's used to provide notification of changes to the managed rule group.

### Labels

* `name` - Individual label specification.

### Rules

* `action` - Action taken on a web request when it matches a rule's statement. See [`action_to_use`](../r/wafv2_web_acl_rule_group_association.html#action_to_use) for details.
* `name` - Name of the rule.
