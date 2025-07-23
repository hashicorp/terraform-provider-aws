---
subcategory: "WAF Classic Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_rate_based_rule"
description: |-
  Provides a AWS WAF Regional rate based rule resource.
---

# Resource: aws_wafregional_rate_based_rule

Provides a WAF Rate Based Rule Resource

## Example Usage

```terraform
resource "aws_wafregional_ipset" "ipset" {
  name = "tfIPSet"

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rate_based_rule" "wafrule" {
  depends_on  = [aws_wafregional_ipset.ipset]
  name        = "tfWAFRule"
  metric_name = "tfWAFRule"

  rate_key   = "IP"
  rate_limit = 100

  predicate {
    data_id = aws_wafregional_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `metric_name` - (Required) The name or description for the Amazon CloudWatch metric of this rule.
* `name` - (Required) The name or description of the rule.
* `rate_key` - (Required) Valid value is IP.
* `rate_limit` - (Required) The maximum number of requests, which have an identical value in the field specified by the RateKey, allowed in a five-minute period. Minimum value is 100.
* `predicate` - (Optional) The objects to include in a rule (documented below).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Nested Blocks

### `predicate`

See the [WAF Documentation](https://docs.aws.amazon.com/waf/latest/APIReference/API_Predicate.html) for more information.

#### Arguments

* `negated` - (Required) Set this to `false` if you want to allow, block, or count requests
  based on the settings in the specified `ByteMatchSet`, `IPSet`, `SqlInjectionMatchSet`, `XssMatchSet`, or `SizeConstraintSet`.
  For example, if an IPSet includes the IP address `192.0.2.44`, AWS WAF will allow or block requests based on that IP address.
  If set to `true`, AWS WAF will allow, block, or count requests based on all IP addresses _except_ `192.0.2.44`.
* `data_id` - (Required) A unique identifier for a predicate in the rule, such as Byte Match Set ID or IPSet ID.
* `type` - (Required) The type of predicate in a rule. Valid values: `ByteMatch`, `GeoMatch`, `IPMatch`, `RegexMatch`, `SizeConstraint`, `SqlInjectionMatch`, or `XssMatch`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the WAF Regional Rate Based Rule.
* `arn` - The ARN of the WAF Regional Rate Based Rule.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAF Regional Rate Based Rule using the id. For example:

```terraform
import {
  to = aws_wafregional_rate_based_rule.wafrule
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc"
}
```

Using `terraform import`, import WAF Regional Rate Based Rule using the id. For example:

```console
% terraform import aws_wafregional_rate_based_rule.wafrule a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
