---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_sites"
description: |-
  Provides details about existing Network Manager sites.
---

# Data Source: aws_networkmanager_sites

Provides details about existing Network Manager sites.

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

This data source supports the following arguments:

* `global_network_id` - (Required) ID of the Global Network of the sites to retrieve.
* `tags` - (Optional) Restricts the list to the sites with these tags.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - IDs of the sites.
