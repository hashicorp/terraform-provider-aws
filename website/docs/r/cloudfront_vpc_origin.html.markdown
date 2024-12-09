---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_vpc_origin"
description: |-
  Provides a CloudFront VPC Origin
---

# Resource: aws_cloudfront_vpc_origin

Creates an Amazon CloudFront VPC origin.

For information about CloudFront VPC origins, see
[Amazon CloudFront Developer Guide - Restrict access with VPC origins][1].

## Example Usage

### Application Load Balancer

The following example below creates a CloudFront VPC origin for a Application Load Balancer.

```terraform
resource "aws_cloudfront_vpc_origin" "alb" {
  vpc_origin_endpoint_config {
    name = "Example VPC Origin"
    origin_arn = aws_lb.this.arn
    http_port = 8080
    https_port = 8443
    origin_protocol_policy = "https-only"
    origin_ssl_protocols {
      items = ["TLSv1.2"]
      quantity = 1
    }
  }
```

## Argument Reference

### Top Level Arguments

* `vpc_origin_endpoint_config` (Required) - The VPC origin endpoint configuration.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### VPC Origin Endpoint Config Arguments

* `http_port` (Required) - The HTTP port for the CloudFront VPC origin endpoint configuration.
* `https_port` (Required) - The HTTPS port for the CloudFront VPC origin endpoint configuration.
* `name` (Required) - The name of the CloudFront VPC origin endpoint configuration.
* `origin_protocol_policy` (Required) - The origin protocol policy for the CloudFront VPC origin endpoint configuration.
* `origin_ssl_protocols` (Required) - A complex type that contains information about the SSL/TLS protocols that CloudFront can use when establishing an HTTPS connection with your origin.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The VPC origin ARN.
* `created_time` - The VPC origin created time.
* `etag` - The current version of the origin. For example: `<todo>`.
* `id` - The VPC origin ID. For example: `<todo>`.
* `last_modified_time` - The VPC origin last modified time.
* `status` - The VPC origin status.

[1]: https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-vpc-origins.html


## Import

```terraform
import {
  to = aws_cloudfront_vpc_origin.origin.alb
  id = 
}
```

Using `terraform import`, import Cloudfront VPC origins using the `id`. For example:

```console
% terraform import aws_cloudfront_vpc_origin
```
