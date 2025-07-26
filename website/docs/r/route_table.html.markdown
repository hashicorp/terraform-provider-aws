---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_route_table"
description: |-
  Provides a resource to create a VPC routing table.
---

# Resource: aws_route_table

Provides a resource to create a VPC routing table.

~> **NOTE on Route Tables and Routes:** Terraform currently
provides both a standalone [Route resource](route.html) and a Route Table resource with routes
defined in-line. At this time you cannot use a Route Table with in-line routes
in conjunction with any Route resources. Doing so will cause
a conflict of rule settings and will overwrite rules.

~> **NOTE on `gateway_id` and `nat_gateway_id`:** The AWS API is very forgiving with these two
attributes and the `aws_route_table` resource can be created with a NAT ID specified as a Gateway ID attribute.
This _will_ lead to a permanent diff between your configuration and statefile, as the API returns the correct
parameters in the returned route table. If you're experiencing constant diffs in your `aws_route_table` resources,
the first thing to check is whether or not you're specifying a NAT ID instead of a Gateway ID, or vice-versa.

~> **NOTE on `propagating_vgws` and the `aws_vpn_gateway_route_propagation` resource:**
If the `propagating_vgws` argument is present, it's not supported to _also_
define route propagations using `aws_vpn_gateway_route_propagation`, since
this resource will delete any propagating gateways not explicitly listed in
`propagating_vgws`. Omit this argument when defining route propagation using
the separate resource.

## Example Usage

### Basic example

```terraform
resource "aws_route_table" "example" {
  vpc_id = aws_vpc.example.id

  route {
    cidr_block = "10.0.1.0/24"
    gateway_id = aws_internet_gateway.example.id
  }

  route {
    ipv6_cidr_block        = "::/0"
    egress_only_gateway_id = aws_egress_only_internet_gateway.example.id
  }

  tags = {
    Name = "example"
  }
}
```

To subsequently remove all managed routes:

```terraform
resource "aws_route_table" "example" {
  vpc_id = aws_vpc.example.id

  route = []

  tags = {
    Name = "example"
  }
}
```

### Adopting an existing local route

AWS creates certain routes that the AWS provider mostly ignores. You can manage them by importing or adopting them. See [Import](#import) below for information on importing. This example shows adopting a route and then updating its target.

First, adopt an existing AWS-created route:

```terraform
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  # since this is exactly the route AWS will create, the route will be adopted
  route {
    cidr_block = "10.1.0.0/16"
    gateway_id = "local"
  }
}
```

Next, update the target of the route:

```terraform
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block           = aws_vpc.test.cidr_block
    network_interface_id = aws_network_interface.test.id
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id
}
```

The target could then be updated again back to `local`.

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `vpc_id` - (Required) The VPC ID.
* `route` - (Optional) A list of route objects. Their keys are documented below. This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).
This means that omitting this argument is interpreted as ignoring any existing routes. To remove all managed routes an empty list should be specified. See the example above.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `propagating_vgws` - (Optional) A list of virtual gateways for propagation.

### route Argument Reference

This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).

One of the following destination arguments must be supplied:

* `cidr_block` - (Required) The CIDR block of the route.
* `ipv6_cidr_block` - (Optional) The Ipv6 CIDR block of the route.
* `destination_prefix_list_id` - (Optional) The ID of a [managed prefix list](ec2_managed_prefix_list.html) destination of the route.

One of the following target arguments must be supplied:

* `carrier_gateway_id` - (Optional) Identifier of a carrier gateway. This attribute can only be used when the VPC contains a subnet which is associated with a Wavelength Zone.
* `core_network_arn` - (Optional) The Amazon Resource Name (ARN) of a core network.
* `egress_only_gateway_id` - (Optional) Identifier of a VPC Egress Only Internet Gateway.
* `gateway_id` - (Optional) Identifier of a VPC internet gateway, virtual private gateway, or `local`. `local` routes cannot be created but can be adopted or imported. See the [example](#adopting-an-existing-local-route) above.
* `local_gateway_id` - (Optional) Identifier of a Outpost local gateway.
* `nat_gateway_id` - (Optional) Identifier of a VPC NAT gateway.
* `network_interface_id` - (Optional) Identifier of an EC2 network interface.
* `transit_gateway_id` - (Optional) Identifier of an EC2 Transit Gateway.
* `vpc_endpoint_id` - (Optional) Identifier of a VPC Endpoint.
* `vpc_peering_connection_id` - (Optional) Identifier of a VPC peering connection.

Note that the default route, mapping the VPC's CIDR block to "local", is created implicitly and cannot be specified.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

~> **NOTE:** Only the target that is entered is exported as a readable
attribute once the route resource is created.

* `id` - The ID of the routing table.
* `arn` - The ARN of the route table.
* `owner_id` - The ID of the AWS account that owns the route table.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `5m`)
- `update` - (Default `2m`)
- `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Route Tables using the route table `id`. For example:

```terraform
import {
  to = aws_route_table.public_rt
  id = "rtb-4e616f6d69"
}
```

Using `terraform import`, import Route Tables using the route table `id`. For example:

```console
% terraform import aws_route_table.public_rt rtb-4e616f6d69
```
