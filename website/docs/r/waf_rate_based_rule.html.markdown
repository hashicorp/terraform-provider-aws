---
subcategory: "WAF Classic"
layout: "aws"
page_title: "AWS: aws_waf_rate_based_rule"
description: |-
  Provides a AWS WAF rule resource.
---

# Resource: aws_waf_rate_based_rule

Provides a WAF Rate Based Rule Resource

## Example Usage

```terraform
resource "aws_waf_ipset" "ipset" {
  name = "tfIPSet"

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rate_based_rule" "wafrule" {
  depends_on  = [aws_waf_ipset.ipset]
  name        = "tfWAFRule"
  metric_name = "tfWAFRule"

  rate_key   = "IP"
  rate_limit = 100

  predicates {
    data_id = aws_waf_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `account` - (Optional) AWS account ID where this resource will be managed. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `metric_name` - (Required) Name or description for the Amazon CloudWatch metric of this rule.
* `name` - (Required) Name or description of the rule.
* `predicates` - (Optional) Objects to include in a rule. See [`predicates` Block](#predicates-block) below for details.
* `rate_key` - (Required) Valid value is IP.
* `rate_limit` - (Required) Maximum number of requests, which have an identical value in the field specified by the RateKey, allowed in a five-minute period. Minimum value is 100.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `predicates` Block

See the [WAF Documentation](https://docs.aws.amazon.com/waf/latest/APIReference/API_Predicate.html) for more information.

* `data_id` - (Required) Unique identifier for a predicate in the rule, such as Byte Match Set ID or IPSet ID.
* `negated` - (Required) Set this to `false` if you want to allow, block, or count requests based on the settings in the specified `ByteMatchSet`, `IPSet`, `SqlInjectionMatchSet`, `XssMatchSet`, or `SizeConstraintSet`. For example, if an IPSet includes the IP address `192.0.2.44`, AWS WAF will allow or block requests based on that IP address. If set to `true`, AWS WAF will allow, block, or count requests based on all IP addresses _except_ `192.0.2.44`.
* `type` - (Required) Type of predicate in a rule. Valid values: `ByteMatch`, `GeoMatch`, `IPMatch`, `RegexMatch`, `SizeConstraint`, `SqlInjectionMatch`, or `XssMatch`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN.
* `id` - ID of the WAF rule.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAF Rated Based Rule using the ID. For example:

```terraform
import {
  to = aws_waf_rate_based_rule.wafrule
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc"
}
```

Using `terraform import`, import WAF Rated Based Rule using the ID. For example:

```console
% terraform import aws_waf_rate_based_rule.wafrule a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
