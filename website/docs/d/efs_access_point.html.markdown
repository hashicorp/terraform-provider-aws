---
subcategory: "EFS"
layout: "aws"
page_title: "AWS: aws_efs_access_point"
description: |-
  Provides an Elastic File System (EFS) Access Point data source.
---

# Data Source: aws_efs_access_point

Provides information about an Elastic File System (EFS) Access Point.

## Example Usage

```hcl
resource "aws_efs_file_system" "test" {
  creation_token = "some-token"
}

resource "aws_efs_access_point" "test" {
  file_system_id = "${aws_efs_file_system.test.id}"
}

data "aws_efs_access_point" "test" {
  access_point_id = "${aws_efs_access_point.test.id}"
}
```

## Argument Reference

The following arguments are supported:

* `access_point_id` - (Required) The ID that identifies the file system.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the access point.
* `arn` - Amazon Resource Name of the file system.
* `file_system_arn` - Amazon Resource Name of the file system.
* `file_system_id` - The ID of the file system for which the access point is intended.
* `posix_user` - The operating system user and group applied to all file system requests made using the access point.
* `root_directory`- Specifies the directory on the Amazon EFS file system that the access point provides access to.
* `tags` - Key-value mapping of resource tags.
