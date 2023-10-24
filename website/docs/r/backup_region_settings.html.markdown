---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_region_settings"
description: |-
  Provides an AWS Backup Region Settings resource.
---

# Resource: aws_backup_region_settings

Provides an AWS Backup Region Settings resource.

## Example Usage

```terraform
resource "aws_backup_region_settings" "test" {
  resource_type_opt_in_preference = {
    "Aurora"          = true
    "DocumentDB"      = true
    "DynamoDB"        = true
    "EBS"             = true
    "EC2"             = true
    "EFS"             = true
    "FSx"             = true
    "Neptune"         = true
    "RDS"             = true
    "Storage Gateway" = true
    "VirtualMachine"  = true
  }

  resource_type_management_preference = {
    "DynamoDB" = true
    "EFS"      = true
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `resource_type_opt_in_preference` - (Required) A map of services along with the opt-in preferences for the Region.
* `resource_type_management_preference` - (Optional) A map of services along with the management preferences for the Region. For more information, see the [AWS Documentation](https://docs.aws.amazon.com/aws-backup/latest/devguide/API_UpdateRegionSettings.html#API_UpdateRegionSettings_RequestSyntax).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The AWS region.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Backup Region Settings using the `region`. For example:

```terraform
import {
  to = aws_backup_region_settings.test
  id = "us-west-2"
}
```

Using `terraform import`, import Backup Region Settings using the `region`. For example:

```console
% terraform import aws_backup_region_settings.test us-west-2
```
