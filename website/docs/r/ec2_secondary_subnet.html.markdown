---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_secondary_subnet"
description: |-
  Provides an EC2 Secondary Subnet resource.
---

# Resource: aws_ec2_secondary_subnet

Provides an EC2 Secondary Subnet resource.

A secondary subnet is a subnet within a secondary network that provides high-performance networking capabilities for specialized workloads such as RDMA (Remote Direct Memory Access) applications.

## Example Usage

### Basic Usage

```terraform
resource "aws_ec2_secondary_network" "example" {
  ipv4_cidr_block = "10.0.0.0/16"
  network_type    = "rdma"

  tags = {
    Name = "example-secondary-network"
  }
}

resource "aws_ec2_secondary_subnet" "example" {
  secondary_network_id = aws_ec2_secondary_network.example.id
  ipv4_cidr_block      = "10.0.1.0/24"
  availability_zone    = "us-west-2a"

  tags = {
    Name = "example-secondary-subnet"
  }
}
```

### Using Availability Zone ID

```terraform
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ec2_secondary_network" "example" {
  ipv4_cidr_block = "10.0.0.0/16"
  network_type    = "rdma"

  tags = {
    Name = "example-secondary-network"
  }
}

resource "aws_ec2_secondary_subnet" "example" {
  secondary_network_id = aws_ec2_secondary_network.example.id
  ipv4_cidr_block      = "10.0.1.0/24"
  availability_zone_id = data.aws_availability_zones.available.zone_ids[0]

  tags = {
    Name = "example-secondary-subnet"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `secondary_network_id` - (Required) ID of the secondary network in which to create the secondary subnet.
* `ipv4_cidr_block` - (Required) IPv4 CIDR block for the secondary subnet. The CIDR block size must be between `/12` and `/28`.
* `availability_zone` - (Optional) Availability Zone for the secondary subnet. Cannot be specified with `availability_zone_id`.
* `availability_zone_id` - (Optional) ID of the Availability Zone for the secondary subnet. This option is preferred over `availability_zone` as it provides a consistent identifier across AWS accounts. Cannot be specified with `availability_zone`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the secondary subnet.
* `id` - ID of the secondary subnet.
* `ipv4_cidr_block_associations` - A list of IPv4 CIDR block associations for the secondary network.
* `owner_id` - ID of the AWS account that owns the secondary subnet.
* `secondary_network_type` - Type of the secondary network (e.g., `rdma`).
* `secondary_subnet_id` - ID of the secondary subnet.
* `state` - State of the secondary subnet.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

The following attributes are exported in the `ipv4_cidr_block_associations` block:

* `association_id` - Association ID for the IPv4 CIDR block.
* `cidr_block` - IPv4 CIDR block.
* `state` - State of the IPv4 CIDR block association.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ec2_secondary_subnet.example
  identity = {
    id = "ss-0123456789abcdef0"
  }
}

resource "aws_ec2_secondary_subnet" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` - (String) ID of the secondary subnet.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 Secondary Subnets using the secondary subnet ID. For example:

```terraform
import {
  to = aws_ec2_secondary_subnet.example
  id = "ss-0123456789abcdef0"
}
```

Using `terraform import`, import EC2 Secondary Subnets using the secondary subnet ID. For example:

```console
% terraform import aws_ec2_secondary_subnet.example ss-0123456789abcdef0
```
