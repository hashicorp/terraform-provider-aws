---
subcategory: "WAF Classic Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_rule"
description: |-
  Provides an AWS WAF Regional rule resource for use with ALB.
---

# Resource: aws_wafregional_rule

Provides an WAF Regional Rule Resource for use with Application Load Balancer.

## Example Usage

```terraform
resource "aws_wafregional_ipset" "ipset" {
  name = "tfIPSet"

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rule" "wafrule" {
  name        = "tfWAFRule"
  metric_name = "tfWAFRule"

  predicate {
    type    = "IPMatch"
    data_id = aws_wafregional_ipset.ipset.id
    negated = false
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `account` - (Optional) AWS account ID where this resource will be managed. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `metric_name` - (Required) Name or description for the Amazon CloudWatch metric of this rule.
* `name` - (Required) Name or description of the rule.
* `predicate` - (Optional) Objects to include in a rule. See [`predicate` Block](#predicate-block) below for details.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `predicate` Block

See the [WAF Documentation](https://docs.aws.amazon.com/waf/latest/APIReference/API_Predicate.html) for more information.

The `predicate` block supports the following arguments:

* `data_id` - (Required) Unique identifier of a predicate, such as the ID of a `ByteMatchSet` or `IPSet`.
* `negated` - (Required) Whether to use the settings or the negated settings that you specified in the objects.
* `type` - (Required) Type of predicate in a rule. Valid values: `ByteMatch`, `GeoMatch`, `IPMatch`, `RegexMatch`, `SizeConstraint`, `SqlInjectionMatch`, or `XssMatch`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the WAF Regional Rule.
* `id` - ID of the WAF Regional Rule.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAF Regional Rule using the ID. For example:

```terraform
import {
  to = aws_wafregional_rule.wafrule
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc"
}
```

Using `terraform import`, import WAF Regional Rule using the ID. For example:

```console
% terraform import aws_wafregional_rule.wafrule a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
