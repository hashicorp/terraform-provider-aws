---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_resource_policy"
description: |-
  Retrieve information about a Network Firewall resource policy.
---

# Data Source: aws_networkfirewall_resource_policy

Retrieve information about a Network Firewall resource policy.

## Example Usage

```terraform
data "aws_networkfirewall_resource_policy" "example" {
  resource_arn = var.resource_policy_arn
}
```

## Argument Reference

* `resource_arn` - (Required) The Amazon Resource Name (ARN) that identifies the resource policy.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - The Amazon Resource Name (ARN) that identifies the resource policy.
* `policy` - The [policy][1] for the resource.

[1]: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/networkfirewall_resource_policy
