---
subcategory: "Directory Service"
layout: "aws"
page_title: "AWS: aws_directory_service_ip_route"
description: |-
  Manages an IP route for an AWS Directory Service directory.
---

# Resource: aws_directory_service_ip_route

Manages an IP route for an AWS Directory Service directory. IP routes are used to route traffic from an AWS Managed Microsoft AD or AD Connector directory to an IPv4 or IPv6 CIDR block, such as an on-premises network reachable over a VPN or AWS Direct Connect connection, or a peered VPC.

~> To manage the complete set of IP routes for a directory, and remove any not configured in Terraform, use [`aws_directory_service_ip_routes_exclusive`](directory_service_ip_routes_exclusive.html) instead. Using both resources for the same directory causes persistent drift unless every `aws_directory_service_ip_route` has an equivalent `ip_route` block.

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

resource "aws_directory_service_ip_route" "example" {
  directory_id = aws_directory_service_directory.example.id
  cidr_ip      = "10.0.0.0/24"
  description  = "On-premises network"
}
```

### IPv6

IPv6 routes require an existing directory that has been updated to dual-stack.

```terraform
resource "aws_directory_service_ip_route" "example" {
  directory_id = "d-1234567890"
  cidr_ipv6    = "2001:db8::/64"
  description  = "On-premises IPv6 network"
}
```

## Argument Reference

The following arguments are required:

* `directory_id` - (Required) Identifier of the directory to which to add the IP route. Changing this forces a new resource to be created.

The following arguments are optional:

* `cidr_ip` - (Optional) IPv4 CIDR block, such as `10.0.0.0/24`. For a single address, use a `/32` block, such as `10.0.0.0/32`. Exactly one of `cidr_ip` or `cidr_ipv6` must be set. Changing this forces a new resource to be created.
* `cidr_ipv6` - (Optional) IPv6 CIDR block, such as `2001:db8::/64`. For a single address, use a `/128` block. Exactly one of `cidr_ip` or `cidr_ipv6` must be set. Changing this forces a new resource to be created.
* `description` - (Optional) Description of the address block. Changing this forces a new resource to be created.
* `region` - (Optional) Region where this resource is managed. Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `update_security_group_for_directory_controllers` - (Optional) Whether to update the inbound and outbound rules of the security group for the directory controllers. Changing this forces a new resource to be created. Defaults to `false`.

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
  to = aws_directory_service_ip_route.example
  identity = {
    directory_id = "d-1234567890"
    cidr_ip      = "10.0.0.0/24"
  }
}

resource "aws_directory_service_ip_route" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `directory_id` (String) Identifier of the directory.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `cidr_ip` (String) IPv4 CIDR block. Exactly one of `cidr_ip` or `cidr_ipv6` must be set.
* `cidr_ipv6` (String) IPv6 CIDR block. Exactly one of `cidr_ip` or `cidr_ipv6` must be set.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an IP route using the directory ID and CIDR block separated by a comma (`,`). For example:

```terraform
import {
  to = aws_directory_service_ip_route.example
  id = "d-1234567890,10.0.0.0/24"
}
```

Using `terraform import`, import an IP route using the directory ID and CIDR block separated by a comma (`,`). For example:

```console
% terraform import aws_directory_service_ip_route.example d-1234567890,10.0.0.0/24
```
