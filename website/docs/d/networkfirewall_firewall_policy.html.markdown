---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_firewall_policy"
description: |-
  Retrieve information about a firewall policy.
---

# Data Source: aws_networkfirewall_firewall_policy

Retrieve information about a firewall policy.

## Example Usage

### Find firewall policy by name

```terraform
data "aws_networkfirewall_firewall_policy" "example" {
  name = var.firewall_policy_name
}
```

### Find firewall policy by ARN

```terraform
data "aws_networkfirewall_firewall_policy" "example" {
  arn = var.firewall_policy_arn
}
```

### Find firewall policy by name and ARN

```terraform
data "aws_networkfirewall_firewall_policy" "example" {
  arn  = var.firewall_policy_arn
  name = var.firewall_policy_name
}
```

AWS Network Firewall does not allow multiple firewall policies with the same name to be created in an account. It is possible, however, to have multiple firewall policies available in a single account with identical `name` values but distinct `arn` values, e.g. firewall policies shared via a [Resource Access Manager (RAM) share][1]. In that case specifying `arn`, or `name` and `arn`, is recommended.

~> **Note:** If there are multiple firewall policies in an account with the same `name`, and `arn` is not specified, the default behavior will return the firewall policy with `name` that was created in the account.

## Argument Reference

One or more of the following arguments are required:

* `arn` - ARN of the firewall policy.
* `name` - Descriptive name of the firewall policy.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `description` - Description of the firewall policy.
* `firewall_policy` - The [policy][2] for the specified firewall policy.
* `tags` - Key-value tags for the firewall policy.
* `update_token` - Token used for optimistic locking.

[1]: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ram_resource_share
[2]: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/networkfirewall_firewall_policy
