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

```hcl
data "aws_networkmanager_devices" "example" {
  global_network_id = var.global_network_id

  tags = {
    Env = "test"
  }
}
```

## Argument Reference

* `global_network_id` - (Required) The ID of the Global Network of the devices to retrieve.
* `tags` - (Optional) Restricts the list to the devices with these tags.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `ids` - The IDs of the devices.
