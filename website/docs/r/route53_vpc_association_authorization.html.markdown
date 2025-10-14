---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_vpc_association_authorization"
description: |-
  Authorizes a VPC in a different account to be associated with a local Route53 Hosted Zone
---

# Resource: aws_route53_vpc_association_authorization

Authorizes a VPC in a different account to be associated with a local Route53 Hosted Zone.

## Example Usage

```terraform
provider "aws" {
}

provider "aws" {
  alias = "alternate"
}

resource "aws_vpc" "example" {
  cidr_block           = "10.6.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
}

resource "aws_route53_zone" "example" {
  name = "example.com"

  vpc {
    vpc_id = aws_vpc.example.id
  }

  # Prevent the deletion of associated VPCs after
  # the initial creation. See documentation on
  # aws_route53_zone_association for details
  lifecycle {
    ignore_changes = [vpc]
  }
}

resource "aws_vpc" "alternate" {
  provider = aws.alternate

  cidr_block           = "10.7.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
}

resource "aws_route53_vpc_association_authorization" "example" {
  vpc_id  = aws_vpc.alternate.id
  zone_id = aws_route53_zone.example.id
}

resource "aws_route53_zone_association" "example" {
  provider = aws.alternate

  vpc_id  = aws_route53_vpc_association_authorization.example.vpc_id
  zone_id = aws_route53_vpc_association_authorization.example.zone_id
}
```

## Argument Reference

This resource supports the following arguments:

* `zone_id` - (Required) The ID of the private hosted zone that you want to authorize associating a VPC with.
* `vpc_id` - (Required) The VPC to authorize for association with the private hosted zone.
* `vpc_region` - (Optional) The VPC's region. Defaults to the region of the AWS provider.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The calculated unique identifier for the association.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `20m`)
* `read` - (Default `5m`)
* `delete` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Route 53 VPC Association Authorizations using the Hosted Zone ID and VPC ID, separated by a colon (`:`). For example:

```terraform
import {
  to = aws_route53_vpc_association_authorization.example
  id = "Z123456ABCDEFG:vpc-12345678"
}
```

Using `terraform import`, import Route 53 VPC Association Authorizations using the Hosted Zone ID and VPC ID, separated by a colon (`:`). For example:

```console
% terraform import aws_route53_vpc_association_authorization.example Z123456ABCDEFG:vpc-12345678
```
