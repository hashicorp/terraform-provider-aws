---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_public_ipv4_pool"
description: |-
  Terraform resource for managing an AWS EC2 (Elastic Compute Cloud) Public IPv4 Pool.
---

# Resource: aws_ec2_public_ipv4_pool

Terraform resource for managing an AWS EC2 (Elastic Compute Cloud) Public I Pv4 Pool.

## Example Usage

```terraform
resource "aws_ec2_public_ipv4_pool" "example" {}
```

## Argument Reference

This resource has no arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `description` - Description of the pool, if any.
* `network_border_group` - Name of the location from which the address pool is advertised.
* `pool_address_ranges` - List of Address Ranges in the Pool; each address range record contains:
    * `address_count` - Number of addresses in the range.
    * `available_address_count` - Number of available addresses in the range.
    * `first_address` - First address in the range.
    * `last_address` - Last address in the range.
* `tags` - Any tags for the address pool.
* `total_address_count` - Total number of addresses in the pool.
* `total_available_address_count` - Total number of available addresses in the pool.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 (Elastic Compute Cloud) Public I Pv4 Pool using the `example_id_arg`. For example:

```terraform
import {
  to = aws_ec2_public_ipv4_pool.example
  id = "ipv4pool-ec2-12345678901234567"
}
```

Using `terraform import`, import EC2 (Elastic Compute Cloud) Public I Pv4 Pool using the `example_id_arg`. For example:

```console
% terraform import aws_ec2_public_ipv4_pool.example public_ipv4_pool-id-12345678
```
