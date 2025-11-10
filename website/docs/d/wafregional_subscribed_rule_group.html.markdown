---
subcategory: "WAF Classic Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_subscribed_rule_group"
description: |-
  retrieves information about a Managed WAF Rule Group from AWS Marketplace for use in WAF Regional.
---

# Data Source: aws_wafregional_subscribed_rule_group

`aws_wafregional_subscribed_rule_group` retrieves information about a Managed WAF Rule Group from AWS Marketplace for use in WAF Regional (needs to be subscribed to first).

## Example Usage

```terraform
data "aws_wafregional_subscribed_rule_group" "by_name" {
  name = "F5 Bot Detection Signatures For AWS WAF"
}

data "aws_wafregional_subscribed_rule_group" "by_metric_name" {
  metric_name = "F5BotDetectionSignatures"
}

resource "aws_wafregional_web_acl" "acl" {
  # ...

  rules {
    priority = 1
    rule_id  = data.aws_wafregional_subscribed_rule_group.by_name.id
    type     = "GROUP"
  }

  rules {
    priority = 2
    rule_id  = data.aws_wafregional_subscribed_rule_group.by_metric_name.id
    type     = "GROUP"
  }
}

```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Optional) Name of the WAF rule group.
* `metric_name` - (Optional) Name of the WAF rule group.

At least one of `name` or `metric_name` must be configured.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the WAF rule group.
