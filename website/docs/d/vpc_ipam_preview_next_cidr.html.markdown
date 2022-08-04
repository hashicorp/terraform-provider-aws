---
subcategory: "VPC IPAM (IP Address Manager)"
layout: "aws"
page_title: "AWS: aws_vpc_ipam_preview_next_cidr"
description: |-
  Previews a CIDR from an IPAM address pool.
---

# Data Source: aws_vpc_ipam_preview_next_cidr

Previews a CIDR from an IPAM address pool. Only works for private IPv4.

~> **NOTE:** This functionality is also encapsulated in a resource sharing the same name. The data source can be used when you need to use the cidr in a calculation of the same Root module, `count` for example. However, once a cidr range has been allocated that was previewed, the next refresh will find a **new** cidr and may force new resources downstream. Make sure to use Terraform's lifecycle `ignore_changes` policy if this is undesirable.

## Example Usage

Basic usage:

```terraform
data "aws_vpc_ipam_preview_next_cidr" "test" {
  ipam_pool_id   = aws_vpc_ipam_pool.test.id
  netmask_length = 28

  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}

resource "aws_vpc_ipam_pool_cidr_allocation" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = data.aws_vpc_ipam_preview_next_cidr.test.cidr

  lifecycle {
    ignore_changes = [cidr]
  }
}
```

## Argument Reference

The following arguments are supported:

* `disallowed_cidrs` - (Optional) Exclude a particular CIDR range from being returned by the pool.
* `ipam_pool_id` - (Required) The ID of the pool to which you want to assign a CIDR.
* `netmask_length` - (Optional) The netmask length of the CIDR you would like to preview from the IPAM pool.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `cidr` - The previewed CIDR from the pool.
* `id` - The ID of the preview.

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

- `read` - (Default `20m`)
