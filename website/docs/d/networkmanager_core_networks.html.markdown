---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_core_networks"
description: |-
  Provides details about existing Network Manager core networks.
---

# Data Source: aws_networkmanager_core_networks

Provides details about existing Network Manager core networks.

## Example Usage

### Basic Usage

```terraform
data "aws_networkmanager_core_networks" "example" {
  tags = {
    Env = "example"
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `tags` - (Optional) Restricts the list to the core networks with these tags.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `core_networks` - Core networks. See [`core_networks` Attribute Reference](#core_networks-attribute-reference) below.
* `ids` - IDs of the core networks.

### `core_networks` Attribute Reference

The `core_networks` attribute exports the following:

* `arn` - ARN of the core network.
* `core_network_id` - ID of the core network.
* `description` - Description of the core network.
* `global_network_id` - ID of the global network that the core network is a part of.
* `owner_account_id` - ID of the account that owns the core network.
* `state` - Current state of the core network.
