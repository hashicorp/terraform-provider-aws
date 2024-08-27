---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_devices"
description: |-
  Retrieve information about devices.
---

# Data Source: aws_networkmanager_devices

Retrieve information about devices.

## Example Usage

```terraform
data "aws_networkmanager_devices" "example" {
  global_network_id = var.global_network_id

  tags = {
    Env = "test"
  }
}
```

## Argument Reference

* `global_network_id` - (Required) ID of the Global Network of the devices to retrieve.
* `site_id` - (Optional) ID of the site of the devices to retrieve.
* `tags` - (Optional) Restricts the list to the devices with these tags.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - IDs of the devices.
