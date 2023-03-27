---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_bucket"
description: |-
  Provides a lightsail bucket
---

# Resource: aws_lightsail_bucket

Provides a lightsail bucket.

## Example Usage

```terraform
resource "aws_lightsail_bucket" "test" {
  name      = "mytestbucket"
  bundle_id = "small_1_0"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name for the bucket.
* `bundle_id` - (Required) - The ID of the bundle to use for the bucket. A bucket bundle specifies the monthly cost, storage space, and data transfer quota for a bucket. Use the [get-bucket-bundles](https://docs.aws.amazon.com/cli/latest/reference/lightsail/get-bucket-bundles.html) cli command to get a list of bundle IDs that you can specify.
* `tags` - (Optional) A map of tags to assign to the resource. To create a key-only tag, use an empty string as the value. If configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name used for this bucket (matches `name`).
* `arn` - The ARN of the lightsail bucket.
* `availability_zone` - The resource Availability Zone. Follows the format us-east-2a (case-sensitive).
* `created_at` - The timestamp when the bucket was created.
* `region` - The Amazon Web Services Region name.
* `support_code` - The support code for the resource. Include this code in your email to support when you have questions about a resource in Lightsail. This code enables our support team to look up your Lightsail information more easily.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider `default_tags` configuration block.

## Import

`aws_lightsail_bucket` can be imported by using the `name` attribute, e.g.,

```
$ terraform import aws_lightsail_bucket.test example-bucket
```
