---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_default_vpc_dhcp_options"
description: |-
  Manage the default VPC DHCP Options resource.
---

# Resource: aws_default_vpc_dhcp_options

Provides a resource to manage the [default AWS DHCP Options Set](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_DHCP_Options.html#AmazonDNS)
in the current region.

Each AWS region comes with a default set of DHCP options.
**This is an advanced resource**, and has special caveats to be aware of when
using it. Please read this document in its entirety before using this resource.

The `aws_default_vpc_dhcp_options` behaves differently from normal resources, in that
Terraform does not _create_ this resource, but instead "adopts" it
into management.

## Example Usage

Basic usage with tags:

```terraform
resource "aws_default_vpc_dhcp_options" "default" {
  tags = {
    Name = "Default DHCP Option Set"
  }
}
```

## Argument Reference

This resource supports the following arguments:

The arguments of an `aws_default_vpc_dhcp_options` differ slightly from `aws_vpc_dhcp_options`  resources.
Namely, the `domain_name`, `domain_name_servers` and `ntp_servers` arguments are computed.
The following arguments are still supported:

* `netbios_name_servers` - (Optional) List of NETBIOS name servers.
* `netbios_node_type` - (Optional) The NetBIOS node type (1, 2, 4, or 8). AWS recommends to specify 2 since broadcast and multicast are not supported in their network. For more information about these node types, see [RFC 2132](http://www.ietf.org/rfc/rfc2132.txt).
* `owner_id` - The ID of the AWS account that owns the DHCP options set.
* `tags` - (Optional) A map of tags to assign to the resource.

### Removing `aws_default_vpc_dhcp_options` from your configuration

The `aws_default_vpc_dhcp_options` resource allows you to manage a region's default DHCP Options Set,
but Terraform cannot destroy it. Removing this resource from your configuration
will remove it from your statefile and management, but will not destroy the DHCP Options Set.
You can resume managing the DHCP Options Set via the AWS Console.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the DHCP Options Set.
* `arn` - The ARN of the DHCP Options Set.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC DHCP Options using the DHCP Options `id`. For example:

```terraform
import {
  to = aws_default_vpc_dhcp_options.default_options
  id = "dopt-d9070ebb"
}
```

Using `terraform import`, import VPC DHCP Options using the DHCP Options `id`. For example:

```console
% terraform import aws_default_vpc_dhcp_options.default_options dopt-d9070ebb
```
