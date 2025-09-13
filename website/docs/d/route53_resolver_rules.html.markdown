---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_rules"
description: |-
    Provides details about a set of Route53 Resolver rules
---

# Data Source: aws_route53_resolver_rules

`aws_route53_resolver_rules` provides details about a set of Route53 Resolver rules.

## Example Usage

### Retrieving the default resolver rule

```terraform
data "aws_route53_resolver_rules" "example" {
  owner_id     = "Route 53 Resolver"
  rule_type    = "RECURSIVE"
  share_status = "NOT_SHARED"
}
```

### Retrieving forward rules shared with me

```terraform
data "aws_route53_resolver_rules" "example" {
  rule_type    = "FORWARD"
  share_status = "SHARED_WITH_ME"
}
```

### Retrieving rules by name regex

Resolver rules whose name contains `abc`.

```terraform
data "aws_route53_resolver_rules" "example" {
  name_regex = ".*abc.*"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name_regex` - (Optional) Regex string to filter resolver rule names.
  The filtering is done locally, so could have a performance impact if the result is large.
  This argument should be used along with other arguments to limit the number of results returned.
* `owner_id` (Optional) When the desired resolver rules are shared with another AWS account, the account ID of the account that the rules are shared with.
* `resolver_endpoint_id` (Optional) ID of the outbound resolver endpoint for the desired resolver rules.
* `rule_type` (Optional) Rule type of the desired resolver rules. Valid values are `FORWARD`, `SYSTEM` and `RECURSIVE`.
* `share_status` (Optional) Whether the desired resolver rules are shared and, if so, whether the current account is sharing the rules with another account, or another account is sharing the rules with the current account. Valid values are `NOT_SHARED`, `SHARED_BY_ME` or `SHARED_WITH_ME`

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `resolver_rule_ids` - IDs of the matched resolver rules.
