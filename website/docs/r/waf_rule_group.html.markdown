---
subcategory: "WAF Classic"
layout: "aws"
page_title: "AWS: aws_waf_rule_group"
description: |-
  Provides a AWS WAF rule group resource.
---

# Resource: aws_waf_rule_group

Provides a WAF Rule Group Resource

## Example Usage

```terraform
resource "aws_waf_rule" "example" {
  name        = "example"
  metric_name = "example"
}

resource "aws_waf_rule_group" "example" {
  name        = "example"
  metric_name = "example"

  activated_rule {
    action {
      type = "COUNT"
    }

    priority = 50
    rule_id  = aws_waf_rule.example.id
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Optional) Name of the rule group. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `metric_name` - (Required) A friendly name for the metrics from the rule group
* `activated_rule` - (Optional) A list of activated rules, see below
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Nested Blocks

### `activated_rule`

#### Arguments

* `action` - (Required) Specifies the action that CloudFront or AWS WAF takes when a web request matches the conditions in the rule.
    * `type` - (Required) e.g., `BLOCK`, `ALLOW`, or `COUNT`
* `priority` - (Required) Specifies the order in which the rules are evaluated. Rules with a lower value are evaluated before rules with a higher value.
* `rule_id` - (Required) The ID of a [rule](/docs/providers/aws/r/waf_rule.html)
* `type` - (Optional) The rule type, either [`REGULAR`](/docs/providers/aws/r/waf_rule.html), [`RATE_BASED`](/docs/providers/aws/r/waf_rate_based_rule.html), or `GROUP`. Defaults to `REGULAR`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the WAF rule group.
* `arn` - The ARN of the WAF rule group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAF Rule Group using the id. For example:

```terraform
import {
  to = aws_waf_rule_group.example
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc"
}
```

Using `terraform import`, import WAF Rule Group using the id. For example:

```console
% terraform import aws_waf_rule_group.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
