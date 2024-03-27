---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_origin_access_control"
description: |-
  Use this data source to retrieve information for an Amazon CloudFront origin access control config.
---

# Data Source: aws_cloudfront_origin_access_control

Use this data source to retrieve information for an Amazon CloudFront origin access control config.

## Example Usage

The below example retrivies a CloudFront origin access control config.

```terraform
data "aws_cloudfront_origin_access_identity" "example" {
  id = "E2T5VTFBZJ3BJB"
}
```

## Argument Reference

* `id` (Required) -  The identifier for the origin access control settings. For example: `E2T5VTFBZJ3BJB`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - The unique identifier of the origin access control.
* `description` - A description of the origin access control.
* `etag` - Current version of the origin access control's information.
  For example: `E2QWRUHAPOMQZL`.
* `name` - A name to identify the origin access control.
* `origin_access_control_origin_type` - The type of origin that this origin access control is for.
* `signing_behavior` - Specifies which requests CloudFront signs. See [Signing Behavior](#signing-behavior) for more information.
* `signing_protocol` - The signing protocol of the origin access control, which determines how CloudFront signs (authenticates) requests.
  The only valid value is `sigv4`.

### Signing Behavior

Specify always for the most common use case. For more information, see origin access control
advanced settings in the Amazon CloudFront Developer Guide.

This field can have one of the following values:

* `always` – CloudFront signs all origin requests, overwriting the Authorization header from the viewer request if one exists.
* `never` – CloudFront doesn't sign any origin requests. This value turns off origin access control for all origins in all
  distributions that use this origin access control.
* `no-override` – If the viewer request doesn't contain the Authorization header, then CloudFront signs the origin request.
  If the viewer request contains the Authorization header, then CloudFront doesn't sign the origin request and instead passes
  along the Authorization header from the viewer request. WARNING: To pass along the `Authorization` header from the viewer
  request, you *must* add the `Authorization` header to a cache policy for all cache behaviors that use origins associated with
  this origin access control. See https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/controlling-the-cache-key.html
