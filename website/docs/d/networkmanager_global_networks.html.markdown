---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_global_networks"
description: |-
  Provides details about existing Network Manager global networks.
---

# Data Source: aws_networkmanager_global_networks

Provides details about existing Network Manager global networks.

## Example Usage

```terraform
data "aws_networkmanager_global_networks" "example" {
  tags = {
    Env = "test"
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `tags` - (Optional) Restricts the list to the global networks with these tags.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - IDs of the global networks.
