---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_firewall_rule_group"
description: |-
    Retrieves the specified firewall rule group.
---

# Data Source: aws_route53_resolver_firewall_rule_group

`aws_route53_resolver_firewall_rule_group` Retrieves the specified firewall rule group.

This data source allows to retrieve details about a specific a Route 53 Resolver DNS Firewall rule group.

## Example Usage

The following example shows how to get a firewall rule group from its ID.

```terraform
data "aws_route53_resolver_firewall_rule_group" "example" {
  firewall_rule_group_id = "rslvr-frg-example"
}
```

## Argument Reference

* `firewall_rule_group_id` - (Required) The ID of the rule group.

The following attribute is additionally exported:

* `arn` - The ARN (Amazon Resource Name) of the rule group.
* `creation_time` - The date and time that the rule group was created, in Unix time format and Coordinated Universal Time (UTC).
* `creator_request_id` - A unique string defined by you to identify the request.
* `name` - The name of the rule group.
* `modification_time` - The date and time that the rule group was last modified, in Unix time format and Coordinated Universal Time (UTC).
* `owner_id` - The Amazon Web Services account ID for the account that created the rule group. When a rule group is shared with your account, this is the account that has shared the rule group with you.
* `rule_count` - The number of rules in the rule group.
* `share_status` - Whether the rule group is shared with other Amazon Web Services accounts, or was shared with the current account by another Amazon Web Services account.
* `status` - The status of the rule group.
* `status_message` - Additional information about the status of the rule group, if available.
