---
subcategory: "WAFv2"
layout: "aws"
page_title: "AWS: aws_wafv2_ip_set"
description: |-
  Retrieves the summary of a WAFv2 IP Set.
---

# Data Source: aws_wafv2_ip_set

Retrieves the summary of a WAFv2 IP Set.

## Example Usage

```hcl
data "aws_wafv2_ip_set" "example" {
  name  = "some-ip-set"
  scope = "REGIONAL"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the WAFv2 IP Set.
* `scope` - (Required) Specifies whether this is for an AWS CloudFront distribution or for a regional application. To work with CloudFront, you must also specify the Region US East (N. Virginia).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the entity.
* `description` - The description of the set that helps with identification.
* `id` - The unique identifier for the set.
