---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_rule"
description: |-
    Provides details about a specific Route53 Resolver rule
---

# Data Source: aws_route53_resolver_rule

`aws_route53_resolver_rule` provides details about a specific Route53 Resolver rule.

## Example Usage

The following example shows how to get a Route53 Resolver rule based on its associated domain name and rule type.

```terraform
data "aws_route53_resolver_rule" "example" {
  domain_name = "subdomain.example.com"
  rule_type   = "SYSTEM"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `domain_name` - (Optional) Domain name the desired resolver rule forwards DNS queries for. Conflicts with `resolver_rule_id`.
* `name` - (Optional) Friendly name of the desired resolver rule. Conflicts with `resolver_rule_id`.
* `resolver_endpoint_id` (Optional) ID of the outbound resolver endpoint of the desired resolver rule. Conflicts with `resolver_rule_id`.
* `resolver_rule_id` (Optional) ID of the desired resolver rule. Conflicts with `domain_name`, `name`, `resolver_endpoint_id` and `rule_type`.
* `rule_type` - (Optional) Rule type of the desired resolver rule. Valid values are `FORWARD`, `SYSTEM` and `RECURSIVE`. Conflicts with `resolver_rule_id`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the resolver rule.
* `arn` - ARN (Amazon Resource Name) for the resolver rule.
* `owner_id` - When a rule is shared with another AWS account, the account ID of the account that the rule is shared with.
* `share_status` - Whether the rules is shared and, if so, whether the current account is sharing the rule with another account, or another account is sharing the rule with the current account.
Values are `NOT_SHARED`, `SHARED_BY_ME` or `SHARED_WITH_ME`
* `tags` - Map of tags assigned to the resolver rule.
