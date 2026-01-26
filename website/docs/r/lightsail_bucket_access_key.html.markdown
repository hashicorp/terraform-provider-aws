---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_bucket_access_key"
description: |-
  Manages a Lightsail bucket access key for programmatic access.
---

# Resource: aws_lightsail_bucket_access_key

Manages a Lightsail bucket access key. Use this resource to create credentials that allow programmatic access to your Lightsail bucket via API requests.

## Example Usage

```terraform
resource "aws_lightsail_bucket" "example" {
  name      = "example-bucket"
  bundle_id = "small_1_0"
}

resource "aws_lightsail_bucket_access_key" "example" {
  bucket_name = aws_lightsail_bucket.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `bucket_name` - (Required) Name of the bucket that the access key will belong to and grant access to.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `access_key_id` - Access key ID.
* `created_at` - Date and time when the access key was created.
* `id` - Combination of attributes separated by a `,` to create a unique id: `bucket_name`,`access_key_id`.
* `secret_access_key` - Secret access key used to sign requests. This attribute is not available for imported resources. Note that this will be written to the state file.
* `status` - Status of the access key.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_lightsail_bucket_access_key` using the `id` attribute. For example:

```terraform
import {
  to = aws_lightsail_bucket_access_key.example
  id = "example-bucket,AKIAIOSFODNN7EXAMPLE"
}
```

Using `terraform import`, import `aws_lightsail_bucket_access_key` using the `id` attribute. For example:

```console
% terraform import aws_lightsail_bucket_access_key.example example-bucket,AKIAIOSFODNN7EXAMPLE
```
