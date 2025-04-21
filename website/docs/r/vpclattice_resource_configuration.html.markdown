---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_resource_configuration"
description: |-
  Terraform resource for managing an AWS VPC Lattice Resource Configuration.
---
# Resource: aws_vpclattice_resource_configuration

Terraform resource for managing an AWS VPC Lattice Resource Configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpclattice_resource_configuration" "example" {
  name = "Example"

  resource_gateway_identifier = aws_vpclattice_resource_gateway.example.id

  port_ranges = ["80"]
  protocol    = "TCP"

  resource_configuration_definition {
    dns_resource {
      domain_name     = "example.com"
      ip_address_type = "IPV4"
    }
  }

  tags = {
    Environment = "Example"
  }
}
```

### IP Address Example

```terraform
resource "aws_vpclattice_resource_configuration" "example" {
  name = "Example"

  resource_gateway_identifier = aws_vpclattice_resource_gateway.example.id

  port_ranges = ["80"]
  protocol    = "TCP"

  resource_configuration_definition {
    ip_resource {
      ip_address = "10.0.0.1"
    }
  }

  tags = {
    Environment = "Example"
  }
}
```

### ARN Example

```terraform
resource "aws_vpclattice_resource_configuration" "test" {
  name = "Example"

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  type = "ARN"

  resource_configuration_definition {
    arn_resource {
      arn = aws_rds_cluster_instance.example.arn
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name for the Resource Configuration.
* `port_ranges` - (Required) Port ranges to access the Resource either single port `80` or range `80-81` range.
* `resource_configuration_definition` - (Required) Details of the Resource Configuration. See [`resource_configuration_definition` Block](#resource_configuration_definition-block) for details.

The following arguments are optional:

* `allow_association_to_shareable_service_network` (Optional) Allow or Deny the association of this resource to a shareable service network.
* `protocol` - (Optional) Protocol for the Resource `TCP` is currently the only supported value.  MUST be specified if `resource_configuration_group_id` is not.
* `resource_configuration_group_id` (Optional) ID of Resource Configuration where `type` is `CHILD`.
* `resource_gateway_identifier` - (Optional) ID of the Resource Gateway used to access the resource. MUST be specified if `resource_configuration_group_id` is not.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `type` - (Optional) Type of Resource Configuration. Must be one of `GROUP`, `CHILD`, `SINGLE`, `ARN`.

### `resource_configuration_definition` Block

One of `dns_resource`, `ip_resource`, `arn_resource` must be specified.

The following arguments are optional:

* `arn_resource` - (Optional) Resource DNS Configuration. See [`arn_resource` Block](#arn_resource-block) for details.
* `dns_resource` - (Optional) Resource DNS Configuration. See [`dns_resource` Block](#dns_resource-block) for details.
* `ip_resource` - (Optional) Resource DNS Configuration. See [`ip_resource` Block](#ip_resource-block) for details.

### `arn_resource` Block

The following arguments are required:

* `arn` - (Required) The ARN of the Resource for this configuration.

### `dns_resource` Block

The following arguments are required:

* `domain_name` - (Required) The hostname of the Resource for this configuration.
* `ip_address_type` - (Required) The IP Address type either `IPV4` or `IPV6`

### `ip_resource` Block

The following arguments are required:

* `ip_address` - (Required) The IP Address of the Resource for this configuration.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the resource gateway.
* `id` - ID of the resource gateway.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Lattice Resource Configuration using the `id`. For example:

```terraform
import {
  to = aws_vpclattice_resource_configuration.example
  id = "rcfg-1234567890abcdef1"
}
```

Using `terraform import`, import VPC Lattice Resource Configuration using the `id`. For example:

```console
% terraform import aws_vpclattice_resource_configuration.example rcfg-1234567890abcdef1
```
