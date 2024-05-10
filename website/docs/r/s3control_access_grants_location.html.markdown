---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_access_grants_location"
description: |-
  Provides a resource to manage an S3 Access Grants location.
---

# Resource: aws_s3control_access_grants_location

Provides a resource to manage an S3 Access Grants location.
A location is an S3 resource (bucket or prefix) in a permission grant that the grantee can access.
The S3 data must be in the same Region as your S3 Access Grants instance.
When you register a location, you must include the IAM role that has permission to manage the S3 location that you are registering.

## Example Usage

```terraform
resource "aws_s3control_access_grants_instance" "example" {}

resource "aws_s3control_access_grants_location" "example" {
  depends_on = [aws_s3control_access_grants_instance.example]

  iam_role_arn   = aws_iam_role.example.arn
  location_scope = "s3://" # Default scope.
}
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Optional) The AWS account ID for the S3 Access Grants location. Defaults to automatically determined account ID of the Terraform AWS provider.
* `iam_role_arn` - (Required) The ARN of the IAM role that S3 Access Grants should use when fulfilling runtime access
requests to the location.
* `location_scope` - (Required) The default S3 URI `s3://` or the URI to a custom location, a specific bucket or prefix.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `access_grants_location_arn` - Amazon Resource Name (ARN) of the S3 Access Grants location.
* `access_grants_location_id` - Unique ID of the S3 Access Grants location.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Access Grants locations using the `account_id` and `access_grants_location_id`, separated by a comma (`,`). For example:

```terraform
import {
  to = aws_s3control_access_grants_location.example
  id = "123456789012,default"
}
```

Using `terraform import`, import S3 Access Grants locations using the `account_id` and `access_grants_location_id`, separated by a comma (`,`). For example:

```console
% terraform import aws_s3control_access_grants_location.example 123456789012,default
```
