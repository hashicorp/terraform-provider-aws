---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_resource_gateway"
description: |-
  Terraform resource for managing an AWS VPC Lattice Resource Gateway.
---
# Resource: aws_vpclattice_resource_gateway

Terraform resource for managing an AWS VPC Lattice Resource Gateway.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpclattice_resource_gateway" "example" {
  name       = "Example"
  vpc_id     = aws_vpc.example.id
  subnet_ids = [aws_subnet.example.id]

  tags = {
    Environment = "Example"
  }
}
```

### Specifying IP address type

```terraform
resource "aws_vpclattice_resource_gateway" "example" {
  name            = "Example"
  vpc_id          = aws_vpc.example.id
  subnet_ids      = [aws_subnet.example.id]
  ip_address_type = "DUALSTACK"

  tags = {
    Environment = "Example"
  }
}
```

### With security groups

```terraform
resource "aws_vpclattice_resource_gateway" "example" {
  name               = "Example"
  vpc_id             = aws_vpc.example.id
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = [aws_subnet.example.id]
}
```

## Argument Reference

The following arguments are required:

* `name` - Name of the resource gateway.
* `subnet_ids` - IDs of the VPC subnets in which to create the resource gateway.
* `vpc_id` - ID of the VPC for the resource gateway.

The following arguments are optional:

* `ip_address_type` - (Optional) IP address type used by the resource gateway. Valid values are `IPV4`, `IPV6`, and `DUALSTACK`. The IP address type of a resource gateway must be compatible with the subnets of the resource gateway and the IP address type of the resource.
* `security_group_ids` - (Optional) Security group IDs associated with the resource gateway. The security groups must be in the same VPC.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the resource gateway.
* `id` - ID of the resource gateway.
* `status` - Status of the resource gateway.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Lattice Resource Gateway using the `id`. For example:

```terraform
import {
  to = aws_vpclattice_resource_gateway.example
  id = "rgw-0a1b2c3d4e5f"
}
```

Using `terraform import`, import VPC Lattice Resource Gateway using the `id`. For example:

```console
% terraform import aws_vpclattice_resource_gateway.example rgw-0a1b2c3d4e5f
```
