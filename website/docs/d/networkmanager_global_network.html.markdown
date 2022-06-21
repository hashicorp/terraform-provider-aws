---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_global_network"
description: |-
  Retrieve information about a global network.
---

# Data Source: aws_networkmanager_global_network

Retrieve information about a global network.

## Example Usage

```terraform
data "aws_networkmanager_global_network" "example" {
  global_network_id = var.global_network_id
}
```

## Argument Reference

* `global_network_id` - (Required) The id of the specific global network to retrieve.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the global network.
* `description` - The description of the global network.
* `tags` - A map of resource tags.
