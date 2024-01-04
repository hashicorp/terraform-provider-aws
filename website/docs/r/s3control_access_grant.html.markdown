---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_access_grant"
description: |-
  Provides a resource to manage an S3 Access Grant.
---

# Resource: aws_s3control_access_grant

Provides a resource to manage an S3 Access Grant.
Each access grant has its own ID and gives an IAM user or role or a directory user, or group (the grantee) access to a registered location. You determine the level of access, such as `READ` or `READWRITE`.
Before you can create a grant, you must have an S3 Access Grants instance in the same Region as the S3 data.

## Example Usage

```terraform
resource "aws_s3control_access_grants_instance" "example" {}

resource "aws_s3control_access_grants_location" "example" {
  depends_on = [aws_s3control_access_grants_instance.example]

  iam_role_arn   = aws_iam_role.example.arn
  location_scope = "s3://${aws_s3_bucket.example.bucket}/prefixA*"
}

resource "aws_s3control_access_grant" "example" {
  access_grants_location_id = aws_s3control_access_grants_location.example.access_grants_location_id
  permission                = "READ"

  access_grants_location_configuration {
    s3_sub_prefix = "prefixB*"
  }

  grantee {
    grantee_type       = "IAM"
    grantee_identifier = aws_iam_user.example.arn
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `access_grants_location_configuration` - (Optional) See [Location Configuration](#location-configuration) below for more details.
* `access_grants_location_id` - (Required) The ID of the S3 Access Grants location to with the access grant is giving access.
* `account_id` - (Optional) The AWS account ID for the S3 Access Grants location. Defaults to automatically determined account ID of the Terraform AWS provider.
* `grantee` - (Optional) See [Grantee](#grantee) below for more details.
* `permission` - (Required) The access grant's level of access. Valid values: `READ`, `WRITE`, `READWRITE`.
* `s3_prefix_type` - (Optional) If you are creating an access grant that grants access to only one object, set this to `Object`. Valid values: `Object`.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Location Configuration

The `access_grants_location_configuration` block supports the following:

* `s3_sub_prefix` - (Optional) Sub-prefix.

### Grantee

The `grantee` block supports the following:

* `grantee_identifier` - (Required) Grantee identifier.
* `grantee_type` - (Required) Grantee types. Valid values: `DIRECTORY_USER`, `DIRECTORY_GROUP`, `IAM`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `access_grant_arn` - Amazon Resource Name (ARN) of the S3 Access Grant.
* `access_grant_id` - Unique ID of the S3 Access Grant.
* `grant_scope` - The access grant's scope.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Access Grants using the `account_id` and `access_grant_id`, separated by a comma (`,`). For example:

```terraform
import {
  to = aws_s3control_access_grant.example
  id = "123456789012,04549c5e-2f3c-4a07-824d-2cafe720aa22"
}
```

Using `terraform import`, import S3 Access Grants using the `account_id` and `access_grant_id`, separated by a comma (`,`). For example:

```console
% terraform import aws_s3control_access_grants_location.example 123456789012,04549c5e-2f3c-4a07-824d-2cafe720aa22
```
