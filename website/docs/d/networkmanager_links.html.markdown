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

* `global_network_id` - (Required) The ID of the Global Network of the links to retrieve.
* `provider_name` - (Optional) The link provider to retrieve.
* `site_id` - (Optional) The ID of the site of the links to retrieve.
* `tags` - (Optional) Restricts the list to the links with these tags.
* `type` - (Optional) The link type to retrieve.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `ids` - The IDs of the links.
