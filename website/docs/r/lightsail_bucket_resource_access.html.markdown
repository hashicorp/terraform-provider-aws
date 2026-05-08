---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_bucket_resource_access"
description: |-
  Manages access permissions between Lightsail resources and buckets.
---

# Resource: aws_lightsail_bucket_resource_access

Manages a Lightsail bucket resource access. Use this resource to grant a Lightsail resource (such as an instance) access to a specific bucket.

## Example Usage

```terraform
resource "aws_lightsail_bucket" "example" {
  name      = "example-bucket"
  bundle_id = "small_1_0"
}

resource "aws_lightsail_instance" "example" {
  name              = "example-instance"
  availability_zone = "us-east-1b"
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "nano_3_0"
}

resource "aws_lightsail_bucket_resource_access" "example" {
  bucket_name   = aws_lightsail_bucket.example.id
  resource_name = aws_lightsail_instance.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `bucket_name` - (Required) Name of the bucket to grant access to.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resource_name` - (Required) Name of the resource to grant bucket access.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Combination of attributes separated by a `,` to create a unique id: `bucket_name`,`resource_name`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_lightsail_bucket_resource_access` using the `id` attribute. For example:

```terraform
import {
  to = aws_lightsail_bucket_resource_access.example
  id = "example-bucket,example-instance"
}
```

Using `terraform import`, import `aws_lightsail_bucket_resource_access` using the `id` attribute. For example:

```console
% terraform import aws_lightsail_bucket_resource_access.example example-bucket,example-instance
```
