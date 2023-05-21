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

* `global_network_id` - (Required) ID of the specific global network to retrieve.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the global network.
* `description` - Description of the global network.
* `tags` - Map of resource tags.
