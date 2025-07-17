---
subcategory: "FSx"
layout: "aws"
page_title: "AWS: aws_fsx_s3_access_point_attachment"
description: |-
  Manages an Amazon FSx S3 Access Point attachment.
---

# Resource: aws_fsx_s3_access_point_attachment

Manages an Amazon FSx S3 Access Point attachment.

## Example Usage

```terraform
resource "aws_fsx_s3_access_point_attachment" "example" {
  name = "example-attachment"
  type = "OPENZFS"

  openzfs_configuration {
    volume_id = aws_fsx_openzfs_volume.example.id

    file_system_identity {
      type = "POSIX"

      posix_user {
        uid = 1001
        gid = 1001
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the S3 access point.
* `openzfs_configuration` - (Required) Configuration to use when creating and attaching an S3 access point to an FSx for OpenZFS volume. See [`openzfs_configuration` Block](#openzfs_configuration-block) for details.
* `type` - (Required) Type of S3 access point. Valid values: `OpenZFS`.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `s3_access_point` - (Optional) S3 access point configuration. See [`s3_access_point` Block](#s3_access_point-block) for details.

### `openzfs_configuration` Block

The `openzfs_configuration` configuration block supports the following arguments:

* `file_system_identity` - (Required) File system user identity to use for authorizing file read and write requests that are made using the S3 access point. See [`file_system_identity` Block](#file_system_identity-block) for details.
* `volume_id` - (Required) ID of the FSx for OpenZFS volume to which the S3 access point is attached.

### `file_system_identity` Block

The `file_system_identity` configuration block supports the following arguments:

* `posix_user` - (Required) UID and GIDs of the file system POSIX user. See [`posix_user` Block](#posix_user-block) for details.
* `type` - (Required) FSx for OpenZFS user identity type. Valid values: `POSIX`.

### `posix_user` Block

The `posix_user` configuration block supports the following arguments:

* `gid` - (Required) GID of the file system user.
* `secondary_gids` - (Optional) List of secondary GIDs for the file system user..
* `uid` - (Required) UID of the file system user.

### `s3_access_point` Block

The `s3_access_point` configuration block supports the following arguments:

* `policy` - (Required) Access policy associated with the S3 access point configuration.
* `vpc_configuration` - (Optional) Amazon S3 restricts access to the S3 access point to requests made from the specified VPC. See [`vpc_configuration` Block](#vpc_configuration-block) for details.

### `vpc_configuration` Block

The `vpc_configuration` configuration block supports the following arguments:

* `vpc_id` - (Required) VPC ID.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `s3_access_point_alias` - S3 access point's alias.
* `s3_access_point_arn` - S3 access point's ARN.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import FSx S3 Access Point attachments using the `name`. For example:

```terraform
import {
  to = aws_fsx_s3_access_point_attachment.example
  id = "example-attachment"
}
```

Using `terraform import`, import FSx S3 Access Point attachments using the `name`. For example:

```console
% terraform import aws_fsx_s3_access_point_attachment.example example-attachment
```
