---
subcategory: "SageMaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_device"
description: |-
  Provides a SageMaker Device resource.
---

# Resource: aws_sagemaker_device

Provides a SageMaker Device resource.

## Example Usage

### Basic usage

```terraform
resource "aws_sagemaker_device" "example" {
  device_fleet_name = aws_sagemaker_device_fleet.example.device_fleet_name

  device {
    device_name = "example"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `device_fleet_name` - (Required) The name of the Device Fleet.
* `device` - (Required) The device to register with SageMaker Edge Manager. See [Device](#device) details below.

### Device

* `description` - (Required) A description for the device.
* `device_name` - (Optional) The name of the device.
* `iot_thing_name` - (Optional) Amazon Web Services Internet of Things (IoT) object name.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The id is constructed from `device-fleet-name/device-name`.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Device.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker Devices using the `device-fleet-name/device-name`. For example:

```terraform
import {
  to = aws_sagemaker_device.example
  id = "my-fleet/my-device"
}
```

Using `terraform import`, import SageMaker Devices using the `device-fleet-name/device-name`. For example:

```console
% terraform import aws_sagemaker_device.example my-fleet/my-device
```
