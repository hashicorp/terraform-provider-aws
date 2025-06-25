---
subcategory: "VPC IPAM (IP Address Manager)"
layout: "aws"
page_title: "AWS: aws_vpc_ipam_pool_cidrs"
description: |-
    Returns cidrs provisioned into an IPAM pool.
---

# Data Source: aws_vpc_ipam_pool_cidrs

`aws_vpc_ipam_pool_cidrs` provides details about an IPAM pool.

This resource can prove useful when an ipam pool was shared to your account and you want to know all (or a filtered list) of the CIDRs that are provisioned into the pool.

## Example Usage

Basic usage:

```terraform
data "aws_vpc_ipam_pool_cidrs" "c" {
  ipam_pool_id = data.aws_vpc_ipam_pool.p.id
}

data "aws_vpc_ipam_pool" "p" {
  filter {
    name   = "description"
    values = ["*mypool*"]
  }

  filter {
    name   = "address-family"
    values = ["ipv4"]
  }
}
```

Filtering:

```terraform
data "aws_vpc_ipam_pool_cidrs" "c" {
  ipam_pool_id = "ipam-pool-123"
  filter {
    name   = "cidr"
    values = ["10.*"]
  }
}

locals {
  mycidrs = [for cidr in data.aws_vpc_ipam_pool_cidrs.c.ipam_pool_cidrs :
    cidr.cidr if
  cidr.state == "provisioned"]
}

resource "aws_ec2_managed_prefix_list" "pls" {
  name           = "IPAM Pool (${aws_vpc_ipam_pool.test.id}) Cidrs"
  address_family = "IPv4"
  max_entries    = length(local.mycidrs)

  dynamic "entry" {
    for_each = local.mycidrs
    content {
      cidr        = entry.value
      description = entry.value
    }
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `ipam_pool_id` - ID of the IPAM pool you would like the list of provisioned CIDRs.
* `filter` - Custom filter block as described below.

### `filter`

More complex filters can be expressed using one or more `filter` sub-blocks, which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_GetIpamPoolCidrs.html).
* `values` - (Required) Set of values that are accepted for the given field.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ipam_pool_cidrs` - The CIDRs provisioned into the IPAM pool, described below.

### ipam_pool_cidrs

* `cidr` - A network CIDR.
* `state` - The provisioning state of that CIDR.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `1m`)
