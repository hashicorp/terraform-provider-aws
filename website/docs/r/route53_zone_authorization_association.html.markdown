---
layout: "aws"
page_title: "AWS: aws_route53_vpc_association_authorization"
sidebar_current: "aws_route53_vpc_association_authorization"
description: |-
  Authorizes a VPC in a peer account to associated with a local Route53 Hosted Zone
---

# aws\_route53\_vpc\_association\_authorization

Authorizes a VPC in a peer account to associated with a local Route53 Hosted Zone

~> **NOTE:** Currently the route53\_zone\_assocation resource does not work across accounts. If you use this resource you will need to do the association the VPC to the Zone outside of Terraform.

## Example Usage

```hcl
provider "aws" {
    region = "us-west-1"
    // Requester's credentials.
}
provider "aws" {
    alias = "peer"
    region = "us-east-2"
}
resource "aws_vpc" "foo" {
	cidr_block = "10.6.0.0/16"
	enable_dns_hostnames = true
	enable_dns_support = true
}
resource "aws_vpc" "peer" {
    provider = "aws.peer"
	cidr_block = "10.7.0.0/16"
	enable_dns_hostnames = true
	enable_dns_support = true
}
resource "aws_route53_zone" "foo" {
	name = "foo.com"
	vpc_id = "${aws_vpc.foo.id}"
}
resource "aws_route53_vpc_association_authorization" "peer" {
    zone_id = "${aws_route53_zone.foo.id}"
    vpc_id  = "${aws_vpc.peer.id}"
}
```

## Argument Reference

The following arguments are supported:

* `zone_id` - (Required) The private hosted zone to associate.
* `vpc_id` - (Required) The VPC to associate with the private hosted zone.
* `vpc_region` - (Optional) The VPC's region. Defaults to the region of the AWS provider.

## Attributes Reference

The following attributes are exported:

* `id` - The calculated unique identifier for the association.
* `zone_id` - The ID of the hosted zone for the association.
* `vpc_id` - The ID of the VPC for the association.
* `vpc_region` - The region in which the VPC identified by `vpc_id` was created.
