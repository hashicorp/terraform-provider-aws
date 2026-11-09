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

### Lambda Web Function Origin

Lambda Web Function origins support only the `always-amz-auth` signing behavior. Use `sigv4` for a regional endpoint (hostname ending in `.lambda-web.<region>.on.aws`) and `sigv4a` for a global endpoint (hostname ending in `.lambda-web.global.on.aws`). See [Restrict access to an AWS Lambda Web Function origin](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-restricting-access-to-lambda-web.html) in the CloudFront Developer Guide.

```terraform
resource "aws_cloudfront_origin_access_control" "example" {
  name                              = "example"
  origin_access_control_origin_type = "lambda"
  signing_behavior                  = "always-amz-auth"
  signing_protocol                  = "sigv4a"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) A name that identifies the Origin Access Control.
* `description` - (Optional) The description of the Origin Access Control. Defaults to "Managed by Terraform" if omitted.
* `origin_access_control_origin_type` - (Required) The type of origin that this Origin Access Control is for. Valid values are `lambda`, `mediapackagev2`, `mediastore`, and `s3`.
* `signing_behavior` - (Required) Specifies which requests CloudFront signs. Specify `always` for the most common use case. Allowed values: `always`, `always-amz-auth`, `never`, and `no-override`. `always-amz-auth` is supported only for Lambda Web Function origins, and is the only value those origins support.
* `signing_protocol` - (Required) Determines how CloudFront signs (authenticates) requests. Valid values are `sigv4` and `sigv4a`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Origin Access Control ARN.
* `id` - The unique identifier of this Origin Access Control.
* `etag` - The current version of this Origin Access Control.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudFront Origin Access Control using the `id`. For example:

```terraform
import {
  to = aws_cloudfront_origin_access_control.example
  id = "E327GJI25M56DG"
}
```

Using `terraform import`, import CloudFront Origin Access Control using the `id`. For example:

```console
% terraform import aws_cloudfront_origin_access_control.example E327GJI25M56DG
```
