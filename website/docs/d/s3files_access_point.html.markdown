---
subcategory: "S3 Files"
layout: "aws"
page_title: "Data Source: aws_s3files_access_point"
description: |-
  Data source for managing an S3 Files Access Point.
---

# Data Source: aws_s3files_access_point

Data source for managing an S3 Files Access Point.

## Example Usage

```terraform
data "aws_s3files_access_point" "example" {
  id = "fsap-1234567890abcdef0"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) Access point ID.

The following arguments are optional:

* `region` - (Optional) Region where this resource is [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the access point.
* `file_system_id` - File system ID.
* `name` - Access point name.
* `owner_id` - AWS account ID of the owner.
* `posix_user` - POSIX user configuration. See [`posix_user`](#posix_user) below.
* `root_directory` - Root directory configuration. See [`root_directory`](#root_directory) below.
* `status` - Access point status.
* `tags` - Map of tags assigned to the resource.

### posix_user

* `gid` - POSIX group ID.
* `secondary_gids` - Set of secondary POSIX group IDs.
* `uid` - POSIX user ID.

### root_directory

* `path` - Root directory path.
* `creation_permissions` - Permissions set when the root directory was created. See [`creation_permissions`](#creation_permissions) below.

### creation_permissions

* `owner_gid` - Owner group ID.
* `owner_uid` - Owner user ID.
* `permissions` - POSIX permissions in octal notation.
