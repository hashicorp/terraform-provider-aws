---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_firewall_policy"
description: |-
  Retrieve information about a firewall policy.
---

# Data Source:  aws_networkfirewall_firewall_policy

Retrieve information about a firewall policy.

## Example Usage

```terraform
data "aws_networkfirewall_firewall_policy" "example" {
  arn = var.firewall_policy_arn
  name = var.firewall_policy_name
}
```

## Argument Reference
One or more of the following arguments are required:

* `arn` - The Amazon Resource Name (ARN) of the firewall policy. 
* `name` - The descriptive name of the firewall policy. 


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the firewall policy.
* `description` - A description of the firewall policy.
* `firewall_policy` - The policy for the specified firewall policy.
* `name` - The descriptive name of the firewall policy.
* `tags` - Key-value tags for the firewall policy.
* `update_token` - A token used for optimistic locking.
