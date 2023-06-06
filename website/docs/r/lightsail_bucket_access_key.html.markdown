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

The following arguments are supported:

* `bucket_name` - (Required) The name of the bucket that the new access key will belong to, and grant access to.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A combination of attributes separated by a `,` to create a unique id: `bucket_name`,`access_key_id`
* `access_key_id` - The ID of the access key.
* `created_at` - The timestamp when the access key was created.
* `secret_access_key` - The secret access key used to sign requests. This attribute is not available for imported resources. Note that this will be written to the state file.
* `status` - The status of the access key.

## Import

`aws_lightsail_bucket_access_key` can be imported by using the `id` attribute, e.g.,

```
$ terraform import aws_lightsail_bucket_access_key.test example-bucket,AKIA47VOQ2KPR7LLRZ6D
```
