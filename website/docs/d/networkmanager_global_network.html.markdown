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

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the global network.
* `description` - Description of the global network.
* `tags` - Map of resource tags.
