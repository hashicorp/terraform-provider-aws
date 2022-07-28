---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_device"
description: |-
  Retrieve information about a device.
---

# Data Source: aws_networkmanager_device

Retrieve information about a device.

## Example Usage

```terraform
data "aws_networkmanager_device" "example" {
  global_network_id_id = var.global_network_id
  device_id            = var.device_id
}
```

## Argument Reference

* `device_id` - (Required) The ID of the device.
* `global_network_id` - (Required) The ID of the global network.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the device.
* `aws_location` - The AWS location of the device. Documented below.
* `description` - A description of the device.
* `location` - The location of the device. Documented below.
* `model` - The model of device.
* `serial_number` - The serial number of the device.
* `site_id` - The ID of the site.
* `tags` - Key-value tags for the device.
* `type` - The type of device.
* `vendor` - The vendor of the device.

The `aws_location` object supports the following:

* `subnet_arn` - The Amazon Resource Name (ARN) of the subnet that the device is located in.
* `zone` - The Zone that the device is located in.

The `location` object supports the following:

* `address` - The physical address.
* `latitude` - The latitude.
* `longitude` - The longitude.
