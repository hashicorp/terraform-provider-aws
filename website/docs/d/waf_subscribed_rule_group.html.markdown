---
subcategory: "WAF Classic"
layout: "aws"
page_title: "AWS: aws_waf_subscribed_rule_group"
description: |-
  Retrieves information about a Managed WAF Rule Group from AWS Marketplace.
---

# Data Source: aws_waf_subscribed_rule_group

`aws_waf_subscribed_rule_group` retrieves information about a Managed WAF Rule Group from AWS Marketplace (needs to be subscribed to first).

## Example Usage

```terraform
data "aws_waf_subscribed_rule_group" "by_name" {
  name = "F5 Bot Detection Signatures For AWS WAF"
}

data "aws_waf_subscribed_rule_group" "by_metric_name" {
  metric_name = "F5BotDetectionSignatures"
}

resource "aws_waf_web_acl" "acl" {
  # ...

  rules {
    priority = 1
    rule_id  = data.aws_waf_subscribed_rule_group.by_name.id
    type     = "GROUP"
  }

  rules {
    priority = 2
    rule_id  = data.aws_waf_subscribed_rule_group.by_metric_name.id
    type     = "GROUP"
  }
}
```

## Argument Reference

This data source supports the following arguments:

At least one of `name` or `metric_name` must be configured.

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `metric_name` - (Optional) Name of the WAF rule group.
* `name` - (Optional) Name of the WAF rule group.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the WAF rule group.
