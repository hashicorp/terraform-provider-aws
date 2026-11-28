---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_links"
description: |-
  Provides details about existing Network Manager links.
---

# Data Source: aws_networkmanager_links

Provides details about existing Network Manager links.

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

This data source supports the following arguments:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `global_network_id` - (Required) ID of the Global Network of the links to retrieve.
* `provider_name` - (Optional) Link provider to retrieve.
* `site_id` - (Optional) ID of the site of the links to retrieve.
* `tags` - (Optional) Restricts the list to the links with these tags.
* `type` - (Optional) Link type to retrieve.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - IDs of the links.
