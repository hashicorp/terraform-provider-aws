---
subcategory: "S3 Files"
layout: "aws"
page_title: "AWS: aws_s3files_access_point"
description: |-
  Manages an S3 Files Access Point.
---

# Resource: aws_s3files_access_point

Manages an S3 Files Access Point.

## Example Usage

```terraform
resource "aws_s3files_access_point" "example" {
  file_system_id = aws_s3files_file_system.example.id

  posix_user {
    gid = 1001
    uid = 1001
  }
}
```

## Argument Reference

The following arguments are required:

* `file_system_id` - (Required) File system ID. Changing this value forces replacement.
* `posix_user` - (Required) POSIX user configuration. See [`posix_user`](#posix_user) below. Changing this value forces replacement.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `root_directory` - (Optional) Root directory configuration. See [`root_directory`](#root_directory) below. Changing this value forces replacement.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### posix_user

* `gid` - (Required) POSIX group ID. Changing this value forces replacement.
* `uid` - (Required) POSIX user ID. Changing this value forces replacement.
* `secondary_gids` - (Optional) Set of secondary POSIX group IDs. Changing this value forces replacement.

### root_directory

* `path` - (Optional) Root directory path. Changing this value forces replacement.
* `creation_permissions` - (Optional) Permissions to set when creating the root directory. See [`creation_permissions`](#creation_permissions) below. Changing this value forces replacement.

### creation_permissions

* `owner_gid` - (Required) Owner group ID. Changing this value forces replacement.
* `owner_uid` - (Required) Owner user ID. Changing this value forces replacement.
* `permissions` - (Required) POSIX permissions in octal notation. Changing this value forces replacement.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the access point.
* `id` - Identifier of the access point.
* `name` - Access point name.
* `owner_id` - AWS account ID of the owner.
* `status` - Access point status.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_s3files_access_point.example
  identity = {
    id = "fsap-1234567890abcdef0"
  }
}

resource "aws_s3files_access_point" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` - Identifier of the access point.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Files Access Point using the resource ID. For example:

```terraform
import {
  to = aws_s3files_access_point.example
  id = "fsap-1234567890abcdef0"
}
```

Using `terraform import`, import S3 Files Access Point using `id`. For example:

```console
% terraform import aws_s3files_access_point.example fsap-1234567890abcdef0
```
