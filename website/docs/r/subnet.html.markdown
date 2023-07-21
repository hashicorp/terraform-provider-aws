---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_subnet"
description: |-
  Provides an VPC subnet resource.
---

# Resource: aws_subnet

Provides an VPC subnet resource.

~> **NOTE:** Due to [AWS Lambda improved VPC networking changes that began deploying in September 2019](https://aws.amazon.com/blogs/compute/announcing-improved-vpc-networking-for-aws-lambda-functions/), subnets associated with Lambda Functions can take up to 45 minutes to successfully delete. Terraform AWS Provider version 2.31.0 and later automatically handles this increased timeout, however prior versions require setting the [customizable deletion timeout](#timeouts) to 45 minutes (`delete = "45m"`). AWS and HashiCorp are working together to reduce the amount of time required for resource deletion and updates can be tracked in this [GitHub issue](https://github.com/hashicorp/terraform-provider-aws/issues/10329).

## Example Usage

### Basic Usage

```terraform
resource "aws_subnet" "main" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = "Main"
  }
}
```

### Subnets In Secondary VPC CIDR Blocks

When managing subnets in one of a VPC's secondary CIDR blocks created using a [`aws_vpc_ipv4_cidr_block_association`](vpc_ipv4_cidr_block_association.html)
resource, it is recommended to reference that resource's `vpc_id` attribute to ensure correct dependency ordering.

```terraform
resource "aws_vpc_ipv4_cidr_block_association" "secondary_cidr" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "172.2.0.0/16"
}

resource "aws_subnet" "in_secondary_cidr" {
  vpc_id     = aws_vpc_ipv4_cidr_block_association.secondary_cidr.vpc_id
  cidr_block = "172.2.0.0/24"
}
```

## Argument Reference

This resource supports the following arguments:

* `assign_ipv6_address_on_creation` - (Optional) Specify true to indicate
    that network interfaces created in the specified subnet should be
    assigned an IPv6 address. Default is `false`
* `availability_zone` - (Optional) AZ for the subnet.
* `availability_zone_id` - (Optional) AZ ID of the subnet. This argument is not supported in all regions or partitions. If necessary, use `availability_zone` instead.
* `cidr_block` - (Optional) The IPv4 CIDR block for the subnet.
* `customer_owned_ipv4_pool` - (Optional) The customer owned IPv4 address pool. Typically used with the `map_customer_owned_ip_on_launch` argument. The `outpost_arn` argument must be specified when configured.
* `enable_dns64` - (Optional) Indicates whether DNS queries made to the Amazon-provided DNS Resolver in this subnet should return synthetic IPv6 addresses for IPv4-only destinations. Default: `false`.
* `enable_lni_at_device_index` - (Optional) Indicates the device position for local network interfaces in this subnet. For example, 1 indicates local network interfaces in this subnet are the secondary network interface (eth1). A local network interface cannot be the primary network interface (eth0).
* `enable_resource_name_dns_aaaa_record_on_launch` - (Optional) Indicates whether to respond to DNS queries for instance hostnames with DNS AAAA records. Default: `false`.
* `enable_resource_name_dns_a_record_on_launch` - (Optional) Indicates whether to respond to DNS queries for instance hostnames with DNS A records. Default: `false`.
* `ipv6_cidr_block` - (Optional) The IPv6 network range for the subnet,
    in CIDR notation. The subnet size must use a /64 prefix length.
* `ipv6_native` - (Optional) Indicates whether to create an IPv6-only subnet. Default: `false`.
* `map_customer_owned_ip_on_launch` -  (Optional) Specify `true` to indicate that network interfaces created in the subnet should be assigned a customer owned IP address. The `customer_owned_ipv4_pool` and `outpost_arn` arguments must be specified when set to `true`. Default is `false`.
* `map_public_ip_on_launch` -  (Optional) Specify true to indicate
    that instances launched into the subnet should be assigned
    a public IP address. Default is `false`.
* `outpost_arn` - (Optional) The Amazon Resource Name (ARN) of the Outpost.
* `private_dns_hostname_type_on_launch` - (Optional) The type of hostnames to assign to instances in the subnet at launch. For IPv6-only subnets, an instance DNS name must be based on the instance ID. For dual-stack and IPv4-only subnets, you can specify whether DNS names use the instance IPv4 address or the instance ID. Valid values: `ip-name`, `resource-name`.
* `vpc_id` - (Required) The VPC ID.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the subnet
* `arn` - The ARN of the subnet.
* `ipv6_cidr_block_association_id` - The association ID for the IPv6 CIDR block.
* `owner_id` - The ID of the AWS account that owns the subnet.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `delete` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import subnets using the subnet `id`. For example:

```terraform
import {
  to = aws_subnet.public_subnet
  id = "subnet-9d4a7b6c"
}
```

Using `terraform import`, import subnets using the subnet `id`. For example:

```console
% terraform import aws_subnet.public_subnet subnet-9d4a7b6c
```
