---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_rule_associations"
description: |-
    Lists the Route53 Resolver rule assocations for current account
---

# Data Source: aws_route53_resolver_rule_associations

`aws_route53_resolver_rule_associations` lists the Route53 Resolver rule assocations for current account as defined in the
[AWS Docs](https://docs.aws.amazon.com/Route53/latest/APIReference/API_route53resolver_ListResolverRuleAssociations.html).

Note that rule associations take time to propagate through AWS and be exposed through the API. For this reason, it's advisable to use a `depends_on` block if defining `aws_route53_resolver_rule_association` resources and fetching/iterating through this data source to avoid a race condition.

## Example Usage

### Retrieving all resolver rule assocations for VPC

```terraform
data "aws_route53_resolver_rule_associations" "examples" {
  filter {
    name   = "VPCId"
    values = [aws_vpc.example.id]
  }

  filter {
    name   = "Status"
    values = ["COMPLETE", "CREATING"]
  }
}
```

### Retrieving resolver rule assocations by resolver rule id

```terraform
data "aws_route53_resolver_rule_associations" "examples" {
  filter {
    name   = "ResolverRuleId"
    values = [aws_route53_resolver_rule_association.example.resolver_rule_id]
  }
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available resolver rules in the current region.

* `filter` - (Optional) A set of name/values pairs, such as Resolver rules or VPC ID.

Valid `filter` names include the following:

* `Name` - The name of the Resolver rule association.
* `ResolverRuleId` : The ID of the Resolver rule that is associated with one or more VPCs.
* `Status` : The status of the Resolver rule association. If you specify Status for Name , specify one of the following status codes for the Value : `CREATING`, `COMPLETE`, `DELETING`, or `FAILED`.
* `VPCId` : The ID of the VPC that the Resolver rule is associated with.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `assocations` - List of resolver rule assocations.

#### `associations` Attribute Reference

* `id` - Id of the resolver rule association.
* `name` - Name of the resolver rule association.
* `resolver_rule_id` - Id of the Resolver rule associated with this VPC.
* `status` - The current status of the assiocation - `CREATING`, `COMPLETE`, `DELETING`, or `FAILED`
* `status_message` - Description of the status of the association.
* `vpc_id` - Id of the VPC associated with this resolver rule.
