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
  opt_in_level = "ACCOUNT"
  opt_in_type  = "NotifyOnly"

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

* `opt_in_level` - (Required) The opt-in level for Automatic Management. Valid values: `ACCOUNT`.
* `opt_in_type` - (Required) The opt-in type for Automatic Management. Valid values: `NotifyOnly`, `NotifyAndAdjust`.

The following arguments are optional:

* `exclusion_list` - (Optional) Map of AWS services excluded from Automatic Management. You will need to include the AWS service code and one or more Service Quotas codes.
* `notification_arn` - (Optional) The AWS User Notifications ARN for Automatic Management notifications.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_servicequotas_auto_management.example
  identity = {
    region = "us-west-2"
  }
}

resource "aws_servicequotas_auto_management" "example" {
  # Additional attributes.
}
```

### Identity Schema

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Service Quotas Auto Management using the region name. For example:

```terraform
import {
  to = aws_servicequotas_auto_management.example
  id = "us-west-2"
}
```

Using `terraform import`, import Service Quotas Auto Management using the region name. For example:

```console
% terraform import aws_servicequotas_auto_management.example us-west-2
```
