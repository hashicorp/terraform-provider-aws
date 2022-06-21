---
subcategory: "DataSync"
layout: "aws"
page_title: "AWS: aws_datasync_location_efs"
description: |-
  Manages an EFS Location within AWS DataSync.
---

# Resource: aws_datasync_location_efs

Manages an AWS DataSync EFS Location.

~> **NOTE:** The EFS File System must have a mounted EFS Mount Target before creating this resource.

## Example Usage

```terraform
resource "aws_datasync_location_efs" "example" {
  # The below example uses aws_efs_mount_target as a reference to ensure a mount target already exists when resource creation occurs.
  # You can accomplish the same behavior with depends_on or an aws_efs_mount_target data source reference.
  efs_file_system_arn = aws_efs_mount_target.example.file_system_arn

  ec2_config {
    security_group_arns = [aws_security_group.example.arn]
    subnet_arn          = aws_subnet.example.arn
  }
}
```

## Argument Reference

The following arguments are supported:

* `access_point_arn` - (Optional) Specifies the Amazon Resource Name (ARN) of the access point that DataSync uses to access the Amazon EFS file system.
* `ec2_config` - (Required) Configuration block containing EC2 configurations for connecting to the EFS File System.
* `efs_file_system_arn` - (Required) Amazon Resource Name (ARN) of EFS File System.
* `file_system_access_role_arn` - (Optional)  Specifies an Identity and Access Management (IAM) role that DataSync assumes when mounting the Amazon EFS file system.
* `in_transit_encryption` - (Optional) Specifies whether you want DataSync to use TLS encryption when transferring data to or from your Amazon EFS file system. Valid values are `NONE` and `TLS1_2`.
* `subdirectory` - (Optional) Subdirectory to perform actions as source or destination. Default `/`.
* `tags` - (Optional) Key-value pairs of resource tags to assign to the DataSync Location. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### ec2_config Argument Reference

The following arguments are supported inside the `ec2_config` configuration block:

* `security_group_arns` - (Required) List of Amazon Resource Names (ARNs) of the EC2 Security Groups that are associated with the EFS Mount Target.
* `subnet_arn` - (Required) Amazon Resource Name (ARN) of the EC2 Subnet that is associated with the EFS Mount Target.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the DataSync Location.
* `arn` - Amazon Resource Name (ARN) of the DataSync Location.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

`aws_datasync_location_efs` can be imported by using the DataSync Task Amazon Resource Name (ARN), e.g.,

```
$ terraform import aws_datasync_location_efs.example arn:aws:datasync:us-east-1:123456789012:location/loc-12345678901234567
```
