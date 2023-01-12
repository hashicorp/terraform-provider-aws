---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_sites"
description: |-
  Retrieve information about sites.
---

# Data Source: aws_networkmanager_sites

Retrieve information about sites.

## Example Usage

```terraform
data "aws_networkmanager_sites" "example" {
  global_network_id = var.global_network_id

  tags = {
    Env = "test"
  }
}
```

## Argument Reference

* `global_network_id` - (Required) ID of the Global Network of the sites to retrieve.
* `tags` - (Optional) Restricts the list to the sites with these tags.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `ids` - IDs of the sites.
