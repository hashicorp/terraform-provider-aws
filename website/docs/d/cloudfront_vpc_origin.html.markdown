---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_vpc_origin"
description: |-
  Use this data source to retrieve information for an Amazon CloudFront VPC origin.
---

# Data Source: aws_cloudfront_vpc_origin

Use this data source to retrieve information for an Amazon CloudFront VPC origin.

## Example Usage

The below example retrieves a CloudFront VPC origin.

```terraform
data "aws_cloudfront_vpc_origin" "example" {
  id = "vo_Ffua5QO1LTlCYJyo1GLcJc"
}
```

## Argument Reference

This data source supports the following arguments:

* `id` - (Required) The identifier for the VPC origin. 

## Attribute Reference
This data source exports the following attributes in addition to the arguments above:

* `arn` - The VPC origin ARN.
* `etag` - Current version of the VPC origin's information. 