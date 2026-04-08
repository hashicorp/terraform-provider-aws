---
subcategory: "S3 Files"
layout: "aws"
page_title: "AWS: aws_s3files_access_point"
description: |-
  Terraform resource for managing an Amazon S3 Files Access Point.
---

# Resource: aws_s3files_access_point

Terraform resource for managing an Amazon S3 Files Access Point.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3files_access_point" "example" {
  file_system_id = aws_s3files_file_system.example.file_system_id
}
```

## Argument Reference

The following arguments are required:

* `file_system_id` - (Required, Forces new resource) The ID of the file system.

The following arguments are optional:

* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `access_point_arn` - The ARN of the access point.
* `access_point_id` - The ID of the access point.
* `owner_id` - The AWS account ID of the access point owner.
* `status` - The lifecycle state of the access point.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Files Access Point using the `access_point_id`. For example:

```terraform
import {
  to = aws_s3files_access_point.example
  id = "fsap-0123456789abcdef0"
}
```

Using `terraform import`, import S3 Files Access Point using the `access_point_id`. For example:

```console
% terraform import aws_s3files_access_point.example fsap-0123456789abcdef0
```
