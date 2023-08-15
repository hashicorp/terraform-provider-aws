---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_links"
description: |-
  Retrieve information about links.
---

# Data Source: aws_networkmanager_links

Retrieve information about link.

## Example Usage

```terraform
data "aws_networkmanager_links" "example" {
  global_network_id = var.global_network_id

  tags = {
    Env = "test"
  }
}
```

## Argument Reference

* `global_network_id` - (Required) ID of the Global Network of the links to retrieve.
* `provider_name` - (Optional) Link provider to retrieve.
* `site_id` - (Optional) ID of the site of the links to retrieve.
* `tags` - (Optional) Restricts the list to the links with these tags.
* `type` - (Optional) Link type to retrieve.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - IDs of the links.
