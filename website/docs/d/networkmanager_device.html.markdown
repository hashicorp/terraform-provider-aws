---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_device"
description: |-
  Provides details about an existing Network Manager device.
---

# Data Source: aws_networkmanager_device

Provides details about an existing Network Manager device.

## Example Usage

```terraform
data "aws_networkmanager_device" "example" {
  global_network_id = var.global_network_id
  device_id         = var.device_id
}
```

## Argument Reference

This data source supports the following arguments:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
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

### `aws_location` Block

* `subnet_arn` - ARN of the subnet that the device is located in.
* `zone` - Zone that the device is located in.

### `location` Block

* `address` - Physical address.
* `latitude` - Latitude.
* `longitude` - Longitude.
