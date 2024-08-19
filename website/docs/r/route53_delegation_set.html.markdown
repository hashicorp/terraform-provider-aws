---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_delegation_set"
description: |-
  Provides a Route53 Delegation Set resource.
---

# Resource: aws_route53_delegation_set

Provides a [Route53 Delegation Set](https://docs.aws.amazon.com/Route53/latest/APIReference/API-actions-by-function.html#actions-by-function-reusable-delegation-sets) resource.

## Example Usage

```terraform
resource "aws_route53_delegation_set" "main" {
  reference_name = "DynDNS"
}

resource "aws_route53_zone" "primary" {
  name              = "hashicorp.com"
  delegation_set_id = aws_route53_delegation_set.main.id
}

resource "aws_route53_zone" "secondary" {
  name              = "terraform.io"
  delegation_set_id = aws_route53_delegation_set.main.id
}
```

## Argument Reference

This resource supports the following arguments:

* `reference_name` - (Optional) This is a reference name used in Caller Reference
  (helpful for identifying single delegation set amongst others)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the Delegation Set.
* `id` - The delegation set ID
* `name_servers` - A list of authoritative name servers for the hosted zone
  (effectively a list of NS records).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Route53 Delegation Sets using the delegation set `id`. For example:

```terraform
import {
  to = aws_route53_delegation_set.set1
  id = "N1PA6795SAMPLE"
}
```

Using `terraform import`, import Route53 Delegation Sets using the delegation set `id`. For example:

```console
% terraform import aws_route53_delegation_set.set1 N1PA6795SAMPLE
```
