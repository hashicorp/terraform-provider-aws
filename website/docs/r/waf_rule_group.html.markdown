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

* `activated_rule` - (Optional) Set of activated rules. See [`activated_rule` Block](#activated_rule-block) below for details.
* `metric_name` - (Required) Friendly name for the metrics from the rule group.
* `name` - (Required) Name of the rule group.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `activated_rule` Block

* `action` - (Required) Action that CloudFront or AWS WAF takes when a web request matches the conditions in the rule. See [`action` Block](#action-block) below for details.
* `priority` - (Required) Order in which the rules are evaluated. Rules with a lower value are evaluated before rules with a higher value.
* `rule_id` - (Required) ID of a [rule](/docs/providers/aws/r/waf_rule.html).
* `type` - (Optional) Rule type, either [`REGULAR`](/docs/providers/aws/r/waf_rule.html), [`RATE_BASED`](/docs/providers/aws/r/waf_rate_based_rule.html), or `GROUP`. Defaults to `REGULAR`.

### `action` Block

* `type` - (Required) e.g., `BLOCK`, `ALLOW`, or `COUNT`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the WAF rule group.
* `id` - ID of the WAF rule group.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAF Rule Group using the ID. For example:

```terraform
import {
  to = aws_waf_rule_group.example
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc"
}
```

Using `terraform import`, import WAF Rule Group using the ID. For example:

```console
% terraform import aws_waf_rule_group.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
