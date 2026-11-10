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

### With custom domain

```terraform
resource "aws_vpclattice_domain_verification" "example" {
  domain_name = "example.com"
}

resource "aws_vpclattice_resource_configuration" "example" {
  name = "Example"

  resource_gateway_identifier = aws_vpclattice_resource_gateway.example.id
  custom_domain_name          = "custom.example.com"
  domain_verification_id      = aws_vpclattice_domain_verification.example.id

  port_ranges = ["443"]
  protocol    = "TCP"

  resource_configuration_definition {
    dns_resource {
      domain_name     = "test.example.com"
      ip_address_type = "IPV4"
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

### CIDR Example

A `CIDR` resource configuration must use a resource gateway with `resource_config_dns_resolution` set to `IN_VPC`. It cannot be associated with a service network or a resource endpoint and is consumed by a VPC endpoint of type `Tunnel`. See [Access CIDR ranges using AWS PrivateLink](https://docs.aws.amazon.com/vpc/latest/privatelink/privatelink-access-cidr-ranges.html) for details.

```terraform
resource "aws_vpclattice_resource_gateway" "example" {
  name       = "example"
  vpc_id     = aws_vpc.example.id
  subnet_ids = [aws_subnet.example.id]

  resource_config_dns_resolution = "IN_VPC"
}

resource "aws_vpclattice_resource_configuration" "example" {
  name = "example"

  resource_gateway_identifier = aws_vpclattice_resource_gateway.example.id

  type        = "CIDR"
  protocol    = "TCP_UDP"
  port_ranges = ["53", "443"]

  resource_configuration_definition {
    cidr_resource {
      cidr_ranges = ["10.0.0.0/16", "192.168.0.0/24"]
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name for the Resource Configuration.
* `resource_configuration_definition` - (Required) Details of the Resource Configuration. See [`resource_configuration_definition` Block](#resource_configuration_definition-block) for details.

The following arguments are optional:

* `allow_association_to_shareable_service_network` - (Optional) Allow or Deny the association of this resource to a shareable service network.
* `custom_domain_name` - (Optional) Custom domain name for your resource configuration. Additionally, provide a `domain_verification_id` to prove your ownership of a domain.
* `domain_verification_id` - (Optional) Domain verification ID of your verified custom domain name. If you don't provide an ID, you must configure the DNS settings yourself.
* `port_ranges` - (Optional) Port ranges to access the Resource either single port `80` or range `80-81` range. For `CIDR` resource configurations, include port `53` to allow DNS.
* `protocol` - (Optional) Protocol for the Resource. Valid values are `TCP` and `TCP_UDP`. `TCP_UDP` is only supported when `type` is `CIDR` (UDP is only supported for DNS). MUST be specified if `resource_configuration_group_id` is not.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resource_configuration_group_id` - (Optional) ID of Resource Configuration where `type` is `CHILD`.
* `resource_gateway_identifier` - (Optional) ID of the Resource Gateway used to access the resource. MUST be specified if `resource_configuration_group_id` is not.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `type` - (Optional) Type of Resource Configuration. Must be one of `GROUP`, `CHILD`, `SINGLE`, `ARN`, `CIDR`. Use `CIDR` with a `cidr_resource` block.

### `resource_configuration_definition` Block

One of `arn_resource`, `cidr_resource`, `dns_resource`, `ip_resource` must be specified.

The following arguments are optional:

* `arn_resource` - (Optional) Resource DNS Configuration. See [`arn_resource` Block](#arn_resource-block) for details.
* `cidr_resource` - (Optional) Resource CIDR configuration. Only valid when `type` is `CIDR`. See [`cidr_resource` Block](#cidr_resource-block) for details.
* `dns_resource` - (Optional) Resource DNS Configuration. See [`dns_resource` Block](#dns_resource-block) for details.
* `ip_resource` - (Optional) Resource DNS Configuration. See [`ip_resource` Block](#ip_resource-block) for details.

### `arn_resource` Block

The following arguments are required:

* `arn` - (Required) ARN of the Resource for this configuration.

### `cidr_resource` Block

The following arguments are required:

* `cidr_ranges` - (Required) Set of 1 to 10 IPv4 and/or IPv6 CIDR ranges (prefix length required), for example `10.0.0.0/16`. `0.0.0.0/0` and `::/0` are only allowed as the sole range. Reserved ranges such as `169.254.0.0/16`, `100.64.0.0/10`, `224.0.0.0/4`, `fe80::/10` and `ff00::/8` are rejected by AWS.

### `dns_resource` Block

The following arguments are required:

* `domain_name` - (Required) Hostname of the Resource for this configuration.
* `ip_address_type` - (Required) IP Address type either `IPV4` or `IPV6`

### `ip_resource` Block

The following arguments are required:

* `ip_address` - (Required) IP Address of the Resource for this configuration.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the resource gateway.
* `domain_verification_arn` - ARN of the domain verification.
* `domain_verification_status` - Domain verification status.
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
