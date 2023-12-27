---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_web_acl"
description: |-
  Retrieves the summary of a WAFv2 Web ACL.
---

# Data Source: aws_wafv2_web_acl

Retrieves the summary of a WAFv2 Web ACL.

## Example Usage

```terraform
data "aws_wafv2_web_acl" "example" {
  name  = "some-web-acl"
  scope = "REGIONAL"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the WAFv2 Web ACL.
* `scope` - (Required) Specifies whether this is for an AWS CloudFront distribution or for a regional application. Valid values are `CLOUDFRONT` or `REGIONAL`. To work with CloudFront, you must also specify the region `us-east-1` (N. Virginia) on the AWS provider.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the entity.
* `description` - Description of the WebACL that helps with identification.
* `id` - Unique identifier of the WebACL.
