---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_subnet"
description: |-
    Provides details about a specific VPC subnet
---

# Data Source: aws_subnet

`aws_subnet` provides details about a specific VPC subnet.

This resource can prove useful when a module accepts a subnet ID as an input variable and needs to, for example, determine the ID of the VPC that the subnet belongs to.

## Example Usage

The following example shows how one might accept a subnet ID as a variable and use this data source to obtain the data necessary to create a security group that allows connections from hosts in that subnet.

```terraform
variable "subnet_id" {}

data "aws_subnet" "selected" {
  id = var.subnet_id
}

resource "aws_security_group" "subnet" {
  vpc_id = data.aws_subnet.selected.vpc_id

  ingress {
    cidr_blocks = [data.aws_subnet.selected.cidr_block]
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
  }
}
```

### Filter Example

If you want to match against tag `Name`, use:

```terraform
data "aws_subnet" "selected" {
  filter {
    name   = "tag:Name"
    values = ["yakdriver"]
  }
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available subnets in the current region. The given filters must match exactly one subnet whose data will be exported as attributes.

The following arguments are optional:

* `availability_zone` - (Optional) Availability zone where the subnet must reside.
* `availability_zone_id` - (Optional) ID of the Availability Zone for the subnet. This argument is not supported in all regions or partitions. If necessary, use `availability_zone` instead.
* `cidr_block` - (Optional) CIDR block of the desired subnet.
* `default_for_az` - (Optional) Whether the desired subnet must be the default subnet for its associated availability zone.
* `filter` - (Optional) Configuration block. Detailed below.
* `id` - (Optional) ID of the specific subnet to retrieve.
* `ipv6_cidr_block` - (Optional) IPv6 CIDR block of the desired subnet.
* `state` - (Optional) State that the desired subnet must have.
* `tags` - (Optional) Map of tags, each pair of which must exactly match a pair on the desired subnet.
* `vpc_id` - (Optional) ID of the VPC that the desired subnet belongs to.

### filter

This block allows for complex filters. You can use one or more `filter` blocks.

The following arguments are required:

* `name` - (Required) Name of the field to filter by, as defined by [the underlying AWS API](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSubnets.html).
* `values` - (Required) Set of values that are accepted for the given field. A subnet will be selected if any one of the given values matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the subnet.
* `assign_ipv6_address_on_creation` - Whether an IPv6 address is assigned on creation.
* `available_ip_address_count` - Available IP addresses of the subnet.
* `customer_owned_ipv4_pool` - Identifier of customer owned IPv4 address pool.
* `enable_dns64` - Whether DNS queries made to the Amazon-provided DNS Resolver in this subnet return synthetic IPv6 addresses for IPv4-only destinations.
* `enable_lni_at_device_index` - Indicates the device position for local network interfaces in this subnet. For example, 1 indicates local network interfaces in this subnet are the secondary network interface (eth1). A local network interface cannot be the primary network interface (eth0).
* `enable_resource_name_dns_aaaa_record_on_launch` - Indicates whether to respond to DNS queries for instance hostnames with DNS AAAA records.
* `enable_resource_name_dns_a_record_on_launch` - Indicates whether to respond to DNS queries for instance hostnames with DNS A records.
* `ipv6_cidr_block_association_id` - Association ID of the IPv6 CIDR block.
* `ipv6_native` - Whether this is an IPv6-only subnet.
* `map_customer_owned_ip_on_launch` - Whether customer owned IP addresses are assigned on network interface creation.
* `map_public_ip_on_launch` - Whether public IP addresses are assigned on instance launch.
* `outpost_arn` - ARN of the Outpost.
* `owner_id` - ID of the AWS account that owns the subnet.
* `private_dns_hostname_type_on_launch` - The type of hostnames assigned to instances in the subnet at launch.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
