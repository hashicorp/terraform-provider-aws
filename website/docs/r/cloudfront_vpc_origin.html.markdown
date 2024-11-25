---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_vpc_origin"
description: |-
  Provides a CloudFront VPC Origin
---

# Resource: aws_cloudfront_vpc_origin

Creates an Amazon CloudFront VPC Origin.

For information about CloudFront distributions, see the
[Amazon CloudFront Developer Guide][1].

## Example Usage

The following example below creates a CloudFront VPC Origin.

```terraform
resource "aws_cloudfront_vpc_origin" "example" {

}
```

## Argument Reference

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The identifier for the distribution. For example: `EDFDVBD632BHDS5`.
* `etag` - The current version of the VPC origin's information.
   For example: `E2QWRUHAPOMQZL`.

## Using With CloudFront

## Import

```terraform
import {
  to = 
  id = 
}
```

Using `terraform import`, import Cloudfront VPC Origins using the `id`. For example:

```console
% terraform import aws_cloudfront_vpc_origin
```
