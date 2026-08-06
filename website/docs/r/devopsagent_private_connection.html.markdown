---
subcategory: "DevOps Agent"
layout: "aws"
page_title: "AWS: aws_devopsagent_private_connection"
description: |-
  Manages an AWS DevOps Agent Private Connection.
---

# Resource: aws_devopsagent_private_connection

Manages an AWS DevOps Agent Private Connection.

A Private Connection enables AWS DevOps Agent to securely connect to resources in your VPC or on-premises environment.

## Example Usage

### Self Managed

```terraform
resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "example" {
  vpc_id     = aws_vpc.example.id
  cidr_block = "10.0.1.0/24"
}

resource "aws_vpclattice_resource_gateway" "example" {
  name       = "example"
  vpc_id     = aws_vpc.example.id
  subnet_ids = [aws_subnet.example.id]
}

resource "aws_vpclattice_resource_configuration" "example" {
  name                        = "example"
  resource_gateway_identifier = aws_vpclattice_resource_gateway.example.id
  protocol                    = "TCP"
  port_ranges                 = ["443"]

  resource_configuration_definition {
    dns_resource {
      domain_name     = "example.com"
      ip_address_type = "IPV4"
    }
  }
}

resource "aws_devopsagent_private_connection" "example" {
  name                      = "example-connection"
  mode                      = "SELF_MANAGED"
  resource_configuration_id = aws_vpclattice_resource_configuration.example.id
  certificate               = var.certificate_pem
}
```

### Service Managed

```terraform
resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "example" {
  vpc_id     = aws_vpc.example.id
  cidr_block = "10.0.1.0/24"
}

resource "aws_devopsagent_private_connection" "example" {
  name         = "example-connection"
  mode         = "SERVICE_MANAGED"
  host_address = "10.0.0.1"
  vpc_id       = aws_vpc.example.id
  subnet_ids   = [aws_subnet.example.id]
}
```

## Argument Reference

The following arguments are required:

* `mode` - (Required, Forces new resource) Mode of the Private Connection. Valid values: `SELF_MANAGED`, `SERVICE_MANAGED`.
* `name` - (Required, Forces new resource) Unique name for the Private Connection within the account. Must be between 3 and 30 characters.

The following arguments are optional:

* `certificate` - (Optional, Sensitive) Certificate to associate with the Private Connection. This is the only field that can be updated in-place.
* `host_address` - (Optional, Forces new resource) IP address or DNS name of the target resource. Only applicable for `SERVICE_MANAGED` connections.
* `region` - (Optional) AWS region for the Private Connection. If not specified, uses the provider's default region.
* `resource_configuration_id` - (Optional, Forces new resource) ID or ARN of the VPC Lattice resource configuration. Only applicable for `SELF_MANAGED` connections.
* `subnet_ids` - (Optional, Forces new resource) Subnets that the service-managed Resource Gateway will span. Only applicable for `SERVICE_MANAGED` connections.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_id` - (Optional, Forces new resource) VPC to create the service-managed Resource Gateway in. Only applicable for `SERVICE_MANAGED` connections.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Private Connection.
* `status` - Current status of the Private Connection. Values include `ACTIVE`, `CREATE_IN_PROGRESS`, `CREATE_FAILED`, `DELETE_IN_PROGRESS`, `DELETE_FAILED`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_devopsagent_private_connection.example
  identity = {
    name = "example-connection"
  }
}

resource "aws_devopsagent_private_connection" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` - Name of the Private Connection.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DevOps Agent Private Connection using `name`. For example:

```terraform
import {
  to = aws_devopsagent_private_connection.example
  id = "example-connection"
}
```

Using `terraform import`, import DevOps Agent Private Connection using `name`. For example:

```console
% terraform import aws_devopsagent_private_connection.example example-connection
```
