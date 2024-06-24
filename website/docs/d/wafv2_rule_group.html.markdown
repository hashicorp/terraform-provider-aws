---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_rule_group"
description: |-
  Retrieves the summary of a WAFv2 Rule Group.
---

# Data Source: aws_wafv2_rule_group

Retrieves the summary of a WAFv2 Rule Group.

## Example Usage

```terraform
data "aws_wafv2_rule_group" "example" {
  name  = "some-rule-group"
  scope = "REGIONAL"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the WAFv2 Rule Group.
* `scope` - (Required) Specifies whether this is for an AWS CloudFront distribution or for a regional application. Valid values are `CLOUDFRONT` or `REGIONAL`. To work with CloudFront, you must also specify the region `us-east-1` (N. Virginia) on the AWS provider.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the entity.
* `description` - Description of the rule group that helps with identification.
* `id` - Unique identifier of the rule group.
