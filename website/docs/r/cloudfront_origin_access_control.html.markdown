---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_origin_access_control"
description: |-
  Terraform resource for managing an AWS CloudFront Origin Access Control.
---

# Resource: aws_cloudfront_origin_access_control

Manages an AWS CloudFront Origin Access Control, which is used by CloudFront Distributions with an Amazon S3 bucket as the origin.

Read more about Origin Access Control in the [CloudFront Developer Guide](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-restricting-access-to-s3.html).

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudfront_origin_access_control" "example" {
  name                              = "example"
  description                       = "Example Policy"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) A name that identifies the Origin Access Control.
* `description` - (Required) The description of the Origin Access Control. It may be empty.
* `origin_access_control_origin_type` - (Required) The type of origin that this Origin Access Control is for. The only valid value is `s3`.
* `signing_behavior` - (Required) Specifies which requests CloudFront signs. Specify `always` for the most common use case. Allowed values: `always`, `never`, `no-override`.
* `signing_protocol` - (Required) Determines how CloudFront signs (authenticates) requests. Valid values: `sigv4`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of this Origin Access Control.
* `etag` - The current version of this Origin Access Control.

## Import

CloudFront Origin Access Control can be imported using the `id`. For example:

```
$ terraform import aws_cloudfront_origin_access_control.example E327GJI25M56DG
```
