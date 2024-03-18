---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_bucket_access_key"
description: |-
  Provides a lightsail bucket access key. This is a set of credentials that allow API requests to be made to the lightsail bucket.
---

# Resource: aws_lightsail_bucket_access_key

Provides a lightsail bucket access key. This is a set of credentials that allow API requests to be made to the lightsail bucket.

## Example Usage

```terraform
resource "aws_lightsail_bucket" "test" {
  name      = "mytestbucket"
  bundle_id = "small_1_0"
}

resource "aws_lightsail_bucket_access_key_access_key" "test" {
  bucket_name = aws_lightsail_bucket_access_key.test.id
}
```

## Argument Reference

This resource supports the following arguments:

* `bucket_name` - (Required) The name of the bucket that the new access key will belong to, and grant access to.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A combination of attributes separated by a `,` to create a unique id: `bucket_name`,`access_key_id`
* `access_key_id` - The ID of the access key.
* `created_at` - The timestamp when the access key was created.
* `secret_access_key` - The secret access key used to sign requests. This attribute is not available for imported resources. Note that this will be written to the state file.
* `status` - The status of the access key.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_lightsail_bucket_access_key` using the `id` attribute. For example:

```terraform
import {
  to = aws_lightsail_bucket_access_key.test
  id = "example-bucket,AKIAIOSFODNN7EXAMPLE"
}
```

Using `terraform import`, import `aws_lightsail_bucket_access_key` using the `id` attribute. For example:

```console
% terraform import aws_lightsail_bucket_access_key.test example-bucket,AKIAIOSFODNN7EXAMPLE
```
