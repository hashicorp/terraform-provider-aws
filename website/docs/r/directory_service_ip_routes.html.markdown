---
subcategory: "Directory Service"
layout: "aws"
page_title: "AWS: aws_directory_service_ip_routes"
description: |-
  Manages IP routes for an AWS Directory Service directory.
---

# Resource: aws_directory_service_ip_routes

Manages IP routes for an AWS Directory Service directory. IP routes are used to route traffic from an AWS Managed Microsoft AD or AD Connector directory to a CIDR block, such as an on-premises network reachable over a VPN or AWS Direct Connect connection, or a peered VPC.

~> **Note:** This resource manages the complete set of IP routes for a directory. Any IP routes added outside of Terraform are removed on the next apply.

## Example Usage

### Basic Usage

```terraform
resource "aws_directory_service_directory" "example" {
  name     = "corp.example.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = aws_vpc.example.id
    subnet_ids = aws_subnet.example[*].id
  }
}

resource "aws_directory_service_ip_routes" "example" {
  directory_id = aws_directory_service_directory.example.id

  update_security_group_for_directory_controllers = true

  ip_route {
    cidr_ip     = "10.0.0.0/24"
    description = "On-premises network"
  }

  ip_route {
    cidr_ip     = "192.168.100.0/24"
    description = "Peered VPC"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `directory_id` - (Required) Identifier of the directory to which to add the IP routes. Changing this forces a new resource to be created.
* `ip_route` - (Required) Set of IP routes to add to the directory. Detailed below.
* `region` - (Optional) Region where this resource is managed. Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `update_security_group_for_directory_controllers` - (Optional) Whether to update the security group of the directory controllers to allow traffic to and from the added CIDR blocks. Changing this forces a new resource to be created. Defaults to `false`.

### ip_route

* `cidr_ip` - (Required) IP address block in CIDR format, such as `10.0.0.0/24`. For a single IP address, use a `/32` CIDR block, such as `10.0.0.0/32`.
* `description` - (Optional) Description of the address block.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_directory_service_ip_routes.example
  identity = {
    directory_id = "d-1234567890"
  }
}

resource "aws_directory_service_ip_routes" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `directory_id` (String) Identifier of the directory.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IP routes using the directory ID. For example:

```terraform
import {
  to = aws_directory_service_ip_routes.example
  id = "d-1234567890"
}
```

Using `terraform import`, import IP routes using the directory ID. For example:

```console
% terraform import aws_directory_service_ip_routes.example d-1234567890
```
