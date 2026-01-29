---
subcategory: "Service Quotas"
layout: "aws"
page_title: "AWS: aws_servicequotas_auto_management"
description: |-
  Manages AWS Service Quotas Automatic Management.
---
# Resource: aws_servicequotas_auto_management

Manages AWS Service Quotas Automatic Management.

~> **Note:** Due to AWS API limitations, the `notification_arn` attribute cannot be removed once set without recreating the resource. Removing this value from your configuration will trigger resource replacement.

## Example Usage

```terraform
resource "aws_servicequotas_auto_management" "example" {
  opt_in_type = "NotifyOnly"

  exclusion_list = {
    "dynamodb" = [
      "L-F98FE922"
    ]
  }
  notification_arn = aws_notifications_notification_configuration.example.arn
}

resource "aws_notifications_notification_configuration" "example" {
  name        = "example"
  description = "example configuration"
}
```

## Argument Reference

The following arguments are required:

* `opt_in_type` - (Required) Sets the opt-in type for Automatic Management. There are two modes: `NotifyOnly` and `NotifyAndAdjust`.

The following arguments are optional:

* `exclusion_list` - (Optional) Map of AWS services excluded from Automatic Management. You will need to include the AWS service code and one or more Service Quotas codes.
* `notification_arn` - (Optional) The AWS User Notifications ARN for Automatic Management notifications.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Service Quotas Auto Management using the account ID. For example:

```terraform
import {
  to = aws_servicequotas_auto_management.example
  id = "123456789012"
}
```

Using `terraform import`, import Service Quotas Auto Management using the account ID. For example:

```console
% terraform import aws_servicequotas_auto_management.example 123456789012
```
