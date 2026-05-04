---
subcategory: "ARC (Application Recovery Controller) Zonal Shift"
layout: "aws"
page_title: "AWS: aws_arczonalshift_autoshift_observer_notification_status"
description: |-
  Manages the autoshift observer notification status for AWS Application Recovery Controller Zonal Shift.
---

# Resource: aws_arczonalshift_autoshift_observer_notification_status

Manages the autoshift observer notification status for AWS Application Recovery Controller Zonal Shift. This controls whether autoshift observer notifications are enabled or disabled.

## Example Usage

```terraform
resource "aws_arczonalshift_autoshift_observer_notification_status" "example" {
  status = "ENABLED"
}
```

## Argument Reference

The following arguments are required:

* `status` - (Required) Autoshift observer notification status. Valid values are `ENABLED` or `DISABLED`.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The AWS region (e.g. `us-east-1`).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_arczonalshift_autoshift_observer_notification_status.example
  identity = {
    region = "us-east-1"
  }
}

resource "aws_arczonalshift_autoshift_observer_notification_status" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ARC Zonal Shift Autoshift Observer Notification Status using the AWS region. For example:

```terraform
import {
  to = aws_arczonalshift_autoshift_observer_notification_status.example
  id = "us-east-1"
}
```

Using `terraform import`, import ARC Zonal Shift Autoshift Observer Notification Status using the AWS region. For example:

```console
% terraform import aws_arczonalshift_autoshift_observer_notification_status.example us-east-1
```
