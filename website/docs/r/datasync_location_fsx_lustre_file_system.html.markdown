---
subcategory: "DataSync"
layout: "aws"
page_title: "AWS: aws_datasync_location_fsx_lustre_file_system"
description: |-
  Manages an FSx Lustre Location within AWS DataSync.
---

# Resource: aws_datasync_location_fsx_lustre_file_system

Manages an AWS DataSync FSx Lustre Location.

## Example Usage

```terraform
resource "aws_datasync_location_fsx_lustre_file_system" "example" {
  fsx_filesystem_arn  = aws_fsx_lustre_file_system.example.arn
  security_group_arns = [aws_security_group.example.arn]
}
```

## Argument Reference

This resource supports the following arguments:

* `fsx_filesystem_arn` - (Required) The Amazon Resource Name (ARN) for the FSx for Lustre file system.
* `security_group_arns` - (Optional) The Amazon Resource Names (ARNs) of the security groups that are to use to configure the FSx for Lustre file system.
* `subdirectory` - (Optional) Subdirectory to perform actions as source or destination.
* `tags` - (Optional) Key-value pairs of resource tags to assign to the DataSync Location. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Amazon Resource Name (ARN) of the DataSync Location.
* `arn` - Amazon Resource Name (ARN) of the DataSync Location.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `uri` - The URL of the FSx for Lustre location that was described.
* `creation_time` - The time that the FSx for Lustre location was created.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_datasync_location_fsx_lustre_file_system` using the `DataSync-ARN#FSx-Lustre-ARN`. For example:

```terraform
import {
  to = aws_datasync_location_fsx_lustre_file_system.example
  id = "arn:aws:datasync:us-west-2:123456789012:location/loc-12345678901234567#arn:aws:fsx:us-west-2:476956259333:file-system/fs-08e04cd442c1bb94a"
}
```

Using `terraform import`, import `aws_datasync_location_fsx_lustre_file_system` using the `DataSync-ARN#FSx-Lustre-ARN`. For example:

```console
% terraform import aws_datasync_location_fsx_lustre_file_system.example arn:aws:datasync:us-west-2:123456789012:location/loc-12345678901234567#arn:aws:fsx:us-west-2:476956259333:file-system/fs-08e04cd442c1bb94a
```
