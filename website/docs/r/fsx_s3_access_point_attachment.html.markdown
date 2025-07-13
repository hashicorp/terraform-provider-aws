---
subcategory: "FSx"
layout: "aws"
page_title: "AWS: aws_fsx_s3_access_point_attachment"
description: |-
  Manages an Amazon FSx S3 Access Point attachment.
---

# Resource: aws_fsx_openzfs_file_system

Manages an Amazon FSx S3 Access Point attachment.

## Example Usage

```terraform
resource "aws_fsx_s3_access_point_attachment" "example" {
  name = "example-attachment"
  type = "OPENZFS"

  openzfs_configuration {
    volume_id = aws_fsx_openzfs_volume.example.id
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

* TODO
* `volume_id` - (Required) ID of the FSx for OpenZFS volume to which the S3 access point is attached.

### `s3_access_point` Block

The `s3_access_point` configuration block supports the following arguments:

* `policy` - (Required) Access policy associated with the S3 access point configuration.
* TODO

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `s3_access_point.0.alias` - S3 access point's alias.
* `s3_access_point.0.resource_arn` - S3 access point's ARN.

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
