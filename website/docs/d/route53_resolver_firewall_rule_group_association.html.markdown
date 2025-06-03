---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_firewall_rule_group_association"
description: |-
    Retrieves the specified firewall rule group association.
---

# Data Source: aws_route53_resolver_firewall_rule_group_association

`aws_route53_resolver_firewall_rule_group_association` Retrieves the specified firewall rule group association.

This data source allows to retrieve details about a specific a Route 53 Resolver DNS Firewall rule group association.

## Example Usage

The following example shows how to get a firewall rule group association from its id.

```terraform
data "aws_route53_resolver_firewall_rule_group_association" "example" {
  firewall_rule_group_association_id = "rslvr-frgassoc-example"
}
```

## Argument Reference

This data source supports the following arguments:

* `firewall_rule_group_association_id` - (Required) The identifier for the association.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the firewall rule group association.
* `creation_time` - The date and time that the association was created, in Unix time format and Coordinated Universal Time (UTC).
* `creator_request_id` - A unique string defined by you to identify the request.
* `firewall_rule_group_id` - The unique identifier of the firewall rule group.
* `managed_owner_name` - The owner of the association, used only for associations that are not managed by you.
* `modification_time` - The date and time that the association was last modified, in Unix time format and Coordinated Universal Time (UTC).
* `mutation_protection` - If enabled, this setting disallows modification or removal of the association, to help prevent against accidentally altering DNS firewall protections.
* `name` - The name of the association.
* `priority` - The setting that determines the processing order of the rule group among the rule groups that are associated with a single VPC.
* `status` - The current status of the association.
* `status_message` - Additional information about the status of the response, if available.
* `vpc_id` - The unique identifier of the VPC that is associated with the rule group.
