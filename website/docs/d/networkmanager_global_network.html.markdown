---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_global_network"
description: |-
  Provides details about an existing Network Manager global network.
---

# Data Source: aws_networkmanager_global_network

Provides details about an existing Network Manager global network.

## Example Usage

```terraform
data "aws_networkmanager_global_network" "example" {
  global_network_id = var.global_network_id
}
```

## Argument Reference

This data source supports the following arguments:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `global_network_id` - (Required) ID of the specific global network to retrieve.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the global network.
* `description` - Description of the global network.
* `tags` - Map of resource tags.
