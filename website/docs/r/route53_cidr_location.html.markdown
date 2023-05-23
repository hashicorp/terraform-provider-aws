---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_cidr_location"
description: |-
  Provides a Route53 CIDR location resource.
---

# Resource: aws_route53_cidr_location

Provides a Route53 CIDR location resource.

## Example Usage

```terraform
resource "aws_route53_cidr_collection" "example" {
  name = "collection-1"
}

resource "aws_route53_cidr_location" "example" {
  cidr_collection_id = aws_route53_cidr_collection.example.id
  name               = "office"
  cidr_blocks        = ["200.5.3.0/24", "200.6.3.0/24"]
}
```

## Argument Reference

The following arguments are supported:

* `cidr_blocks` - (Required) CIDR blocks for the location.
* `cidr_collection_id` - (Required) The ID of the CIDR collection to update.
* `name` - (Required) Name for the CIDR location.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The CIDR location ID.

## Import

CIDR locations can be imported using their the CIDR collection ID and location name, e.g.,

```
$ terraform import aws_route53_cidr_location.example 9ac32814-3e67-0932-6048-8d779cc6f511,office
```
