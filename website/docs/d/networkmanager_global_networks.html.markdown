---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_global_networks"
description: |-
  Retrieve information about global networks.
---

# Data Source: aws_networkmanager_global_networks

Retrieve information about global networks.

## Example Usage

```terraform
data "aws_networkmanager_global_networks" "example" {
  tags = {
    Env = "test"
  }
}
```

## Argument Reference

* `tags` - (Optional) Restricts the list to the global networks with these tags.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - IDs of the global networks.
