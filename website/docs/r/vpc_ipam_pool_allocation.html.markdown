---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_vpc_ipam_pool_allocation"
description: |-
  Allocates (reserves) a CIDR from an IPAM address pool, preventing usage by IPAM.
---

# Resource: aws_vpc_ipam_pool_allocation

Allocates (reserves) a CIDR from an IPAM address pool, preventing usage by IPAM. Only works for private IPv4.

## Example Usage

Basic usage:

```terraform
resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = "172.2.0.0/16"
}
```

## Argument Reference

The following arguments are supported:

* `cidr` - (Optional) The CIDR you want to assign to the pool.
* `description` - (Optional) The description for the allocation.
* `ipam_pool_id` - (Required) The ID of the pool to which you want to assign a CIDR.
* `netmask_length` - (Optional) The netmask length of the CIDR you would like to allocate to the IPAM pool.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the allocation.
* `resource_id` - The ID of the resource.
* `resource_owner` - The owner of the resource.
* `resource_type` - The type of the resource.

## Import

IPAMs can be imported using the `allocation id`, e.g.

```
$ terraform import aws_vpc_ipam_pool_allocation.test
```
