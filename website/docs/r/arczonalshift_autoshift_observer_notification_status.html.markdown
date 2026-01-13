---
subcategory: "Application Recovery Controller Zonal Shift"
layout: "aws"
page_title: "AWS: aws_arczonalshift_autoshift_observer_notification_status"
description: |-
  Manages an AWS Application Recovery Controller Zonal Shift Autoshift Observer Notification Status.
---
<!---
Documentation guidelines:
- Begin resource descriptions with "Manages..."
- Use simple language and avoid jargon
- Focus on brevity and clarity
- Use present tense and active voice
- Don't begin argument/attribute descriptions with "An", "The", "Defines", "Indicates", or "Specifies"
- Boolean arguments should begin with "Whether to"
- Use "example" instead of "test" in examples
--->

# Resource: aws_arczonalshift_autoshift_observer_notification_status

Manages an AWS Application Recovery Controller Zonal Shift Autoshift Observer Notification Status.

## Example Usage

### Basic Usage

```terraform
resource "aws_arczonalshift_autoshift_observer_notification_status" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Autoshift Observer Notification Status.
* `example_attribute` - Brief description of the attribute.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Application Recovery Controller Zonal Shift Autoshift Observer Notification Status using the `example_id_arg`. For example:

```terraform
import {
  to = aws_arczonalshift_autoshift_observer_notification_status.example
  id = "autoshift_observer_notification_status-id-12345678"
}
```

Using `terraform import`, import Application Recovery Controller Zonal Shift Autoshift Observer Notification Status using the `example_id_arg`. For example:

```console
% terraform import aws_arczonalshift_autoshift_observer_notification_status.example autoshift_observer_notification_status-id-12345678
```
