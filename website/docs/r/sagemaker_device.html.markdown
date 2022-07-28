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

The following arguments are supported:

* `device_fleet_name` - (Required) The name of the Device Fleet.
* `role_arn` - (Required) The Amazon Resource Name (ARN) that has access to AWS Internet of Things (IoT).
* `device` - (Required) The device to register with SageMaker Edge Manager. See [Device](#device) details below.

### Device

* `description` - (Required) A description for the device.
* `device_name` - (Optional) The name of the device.
* `iot_thing_name` - (Optional) Amazon Web Services Internet of Things (IoT) object name.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The id is constructed from `device-fleet-name/device-name`.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Device.

## Import

SageMaker Devices can be imported using the `device-fleet-name/device-name`, e.g.,

```
$ terraform import aws_sagemaker_device.example my-fleet/my-device
```
