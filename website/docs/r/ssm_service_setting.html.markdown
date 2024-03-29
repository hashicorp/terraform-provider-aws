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
  setting_id    = "arn:aws:ssm:us-east-1:123456789012:servicesetting/ssm/parameter-store/high-throughput-enabled"
  setting_value = "true"
}
```

## Argument Reference

This resource supports the following arguments:

* `setting_id` - (Required) ID of the service setting.
* `setting_value` - (Required) Value of the service setting.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the service setting.
* `status` - Status of the service setting. Value can be `Default`, `Customized` or `PendingUpdate`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS SSM Service Setting using the `setting_id`. For example:

```terraform
import {
  to = aws_ssm_service_setting.example
  id = "arn:aws:ssm:us-east-1:123456789012:servicesetting/ssm/parameter-store/high-throughput-enabled"
}
```

Using `terraform import`, import AWS SSM Service Setting using the `setting_id`. For example:

```console
% terraform import aws_ssm_service_setting.example arn:aws:ssm:us-east-1:123456789012:servicesetting/ssm/parameter-store/high-throughput-enabled
```
