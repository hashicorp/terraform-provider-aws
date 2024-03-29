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

* `device_id` - (Required) ID of the device.
* `global_network_id` - (Required) ID of the global network.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the device.
* `aws_location` - AWS location of the device. Documented below.
* `description` - Description of the device.
* `location` - Location of the device. Documented below.
* `model` - Model of device.
* `serial_number` - Serial number of the device.
* `site_id` - ID of the site.
* `tags` - Key-value tags for the device.
* `type` - Type of device.
* `vendor` - Vendor of the device.

The `aws_location` object supports the following:

* `subnet_arn` - ARN of the subnet that the device is located in.
* `zone` - Zone that the device is located in.

The `location` object supports the following:

* `address` - Physical address.
* `latitude` - Latitude.
* `longitude` - Longitude.
