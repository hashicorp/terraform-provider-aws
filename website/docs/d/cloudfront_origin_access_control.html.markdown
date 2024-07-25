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

The below example retrieves a CloudFront origin access control config.

```terraform
data "aws_cloudfront_origin_access_identity" "example" {
  id = "E2T5VTFBZJ3BJB"
}
```

## Argument Reference

* `id` (Required) -  The identifier for the origin access control settings. For example: `E2T5VTFBZJ3BJB`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `description` - A description of the origin access control.
* `etag` - Current version of the origin access control's information. For example: `E2QWRUHAPOMQZL`.
* `name` - A name to identify the origin access control.
* `origin_access_control_origin_type` - The type of origin that this origin access control is for.
* `signing_behavior` - Specifies which requests CloudFront signs.
* `signing_protocol` - The signing protocol of the origin access control, which determines how CloudFront signs (authenticates) requests.
