---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_service_setting"
description: |-
  Defines how a user interacts with or uses a service or a feature of a service.
---

# Resource: aws_ssm_service_setting

This setting defines how a user interacts with or uses a service or a feature of a service.

## Example Usage

```terraform
resource "aws_ssm_service_setting" "test_setting" {
  service_id    = "arn:aws:ssm:us-east-1:123456789012:servicesetting/ssm/parameter-store/high-throughput-enabled"
  service_value = "true"
}
```

## Argument Reference

The following arguments are supported:

* `service_id` - (Required) ID of the service setting.
* `service_value` - (Required) Value of the service setting.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the service setting.
* `status` - Status of the service setting. Value can be `Default`, `Customized` or `PendingUpdate`.

## Import

AWS SSM Service Setting can be imported using the `setting_id`, e.g.

```sh
$ terraform import aws_ssm_service_setting.example arn:aws:ssm:us-east-1:123456789012:servicesetting/ssm/parameter-store/high-throughput-enabled
```
