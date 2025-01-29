---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_zones"
description: |-
    Provides a list of Route53 Hosted Zone IDs in a Region
---

# Data Source: aws_route53_zones

This resource can be useful for getting back a list of Route53 Hosted Zone IDs for a Region.

## Example Usage

The following example retrieves a list of all Hosted Zone IDs.

```hcl
data "aws_route53_zones" "all" {}

output "example" {
  value = data.aws_route53_zones.all.ids
}
```

## Attributes Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - A list of all the Route53 Hosted Zone IDs found.
