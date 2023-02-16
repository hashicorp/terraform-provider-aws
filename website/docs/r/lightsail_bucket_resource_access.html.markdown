---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_bucket_resource_access"
description: |-
  Provides a lightsail resource access to a bucket.
---

# Resource: aws_lightsail_bucket_resource_access

Provides a lightsail resource access to a bucket.

## Example Usage

```terraform
resource "aws_lightsail_bucket" "test" {
  name      = "mytestbucket"
  bundle_id = "small_1_0"
}

resource "aws_lightsail_instance" "test" {
  name              = "mytestinstance"
  availability_zone = "us-east-1b"
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "nano_1_0"
}

resource "aws_lightsail_bucket_resource_access" "test" {
  bucket_name   = aws_lightsail_bucket_resource_access.test.id
  resource_name = aws_lightsail_instance.id
}
```

## Argument Reference

The following arguments are supported:

* `bucket_name` - (Required) The name of the bucket to grant access to.
* `resource_name` - (Required) The name of the resource to be granted bucket access.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A combination of attributes separated by a `,` to create a unique id: `bucket_name`,`resource_name`

## Import

`aws_lightsail_bucket_resource_access` can be imported by using the `id` attribute, e.g.,

```
$ terraform import aws_lightsail_bucket_resource_access.test example-bucket,example-instance
```
