---
subcategory: "Directory Service"
layout: "aws"
page_title: "AWS: aws_directory_service_ip_routes_exclusive"
description: |-
  Manages an exclusive set of IP routes for an AWS Directory Service directory.
---

# Resource: aws_directory_service_ip_routes_exclusive

Manages an exclusive set of IP routes for an AWS Directory Service directory. IP routes are used to route traffic from an AWS Managed Microsoft AD or AD Connector directory to an IPv4 or IPv6 CIDR block, such as an on-premises network reachable over a VPN or AWS Direct Connect connection, or a peered VPC.

!> This resource takes exclusive ownership over the IP routes of a directory. This includes removal of IP routes which are not explicitly configured. To prevent persistent drift, ensure any `aws_directory_service_ip_route` resources managed alongside this resource have an equivalent `ip_route` block.

~> Destruction of this resource means Terraform will no longer manage reconciliation of the configured IP routes. It __will not__ remove the configured IP routes from the directory.

~> Adding an IPv6 route (`cidr_ipv6`) requires the directory's network type to be dual-stack (IPv4 and IPv6). Enabling IPv6 support on a directory is a one-way operation performed outside of Terraform; see [Updating directory network type](https://docs.aws.amazon.com/directoryservice/latest/admin-guide/ms_ad_update-directory-type.html).

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

resource "aws_directory_service_ip_routes_exclusive" "example" {
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

### IPv6

IPv6 routes require an existing directory that has been updated to dual-stack.

```terraform
resource "aws_directory_service_ip_routes_exclusive" "example" {
  directory_id = "d-1234567890"

  ip_route {
    cidr_ipv6   = "2001:db8::/64"
    description = "On-premises IPv6 network"
  }
}
```

### Disallow IP Routes

To remove all IP routes from a directory, omit all `ip_route` blocks. Any IP routes added outside of Terraform are detected on the next refresh and removed on the next apply.

```terraform
resource "aws_directory_service_ip_routes_exclusive" "example" {
  directory_id = aws_directory_service_directory.example.id
}
```

## Argument Reference

The following arguments are required:

* `directory_id` - (Required) Identifier of the directory. Changing this forces a new resource to be created.

The following arguments are optional:

* `ip_route` - (Optional) Set of IP routes the directory should have. Omit all `ip_route` blocks to remove every IP route from the directory. See [`ip_route` Block](#ip_route-block) below.
* `region` - (Optional) Region where this resource is managed. Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `update_security_group_for_directory_controllers` - (Optional) Whether to update the inbound and outbound rules of the security group for the directory controllers when adding IP routes. Applies only to IP routes added by this resource. Defaults to `false`.

### `ip_route` Block

Exactly one of `cidr_ip` or `cidr_ipv6` must be set.

* `cidr_ip` - (Optional) IPv4 CIDR block, such as `10.0.0.0/24`. For a single address, use a `/32` block, such as `10.0.0.0/32`. Must be unique across all `ip_route` blocks.
* `cidr_ipv6` - (Optional) IPv6 CIDR block, such as `2001:db8::/64`. For a single address, use a `/128` block. Must be unique across all `ip_route` blocks.
* `description` - (Optional) Description of the address block.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_directory_service_ip_routes_exclusive.example
  identity = {
    directory_id = "d-1234567890"
  }
}

resource "aws_directory_service_ip_routes_exclusive" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `directory_id` (String) Identifier of the directory.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import exclusive management of IP routes using the directory ID. For example:

```terraform
import {
  to = aws_directory_service_ip_routes_exclusive.example
  id = "d-1234567890"
}
```

Using `terraform import`, import exclusive management of IP routes using the directory ID. For example:

```console
% terraform import aws_directory_service_ip_routes_exclusive.example d-1234567890
```
