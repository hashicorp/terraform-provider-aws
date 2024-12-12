---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_cidr_collection"
description: |-
  Provides a Route53 CIDR collection resource.
---

# Resource: aws_route53_cidr_collection

Provides a Route53 CIDR collection resource.

## Example Usage

```terraform
resource "aws_route53_cidr_collection" "example" {
  name = "collection-1"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Unique name for the CIDR collection.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the CIDR collection.
* `id` - The CIDR collection ID.
* `version` - The lastest version of the CIDR collection.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CIDR collections using their ID. For example:

```terraform
import {
  to = aws_route53_cidr_collection.example
  id = "9ac32814-3e67-0932-6048-8d779cc6f511"
}
```

Using `terraform import`, import CIDR collections using their ID. For example:

```console
% terraform import aws_route53_cidr_collection.example 9ac32814-3e67-0932-6048-8d779cc6f511
```
