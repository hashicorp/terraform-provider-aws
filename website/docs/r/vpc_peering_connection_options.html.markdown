---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_peering_connection_options"
description: |-
  Provides a resource to manage VPC peering connection options.
---

# Resource: aws_vpc_peering_connection_options

Provides a resource to manage VPC peering connection options.

~> **NOTE on VPC Peering Connections and VPC Peering Connection Options:** Terraform provides
both a standalone VPC Peering Connection Options and a [VPC Peering Connection](vpc_peering_connection.html)
resource with `accepter` and `requester` attributes. Do not manage options for the same VPC peering
connection in both a VPC Peering Connection resource and a VPC Peering Connection Options resource.
Doing so will cause a conflict of options and will overwrite the options.
Using a VPC Peering Connection Options resource decouples management of the connection options from
management of the VPC Peering Connection and allows options to be set correctly in cross-region and
cross-account scenarios.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc" "bar" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_vpc_peering_connection" "foo" {
  vpc_id      = aws_vpc.foo.id
  peer_vpc_id = aws_vpc.bar.id
  auto_accept = true
}

resource "aws_vpc_peering_connection_options" "foo" {
  vpc_peering_connection_id = aws_vpc_peering_connection.foo.id

  accepter {
    allow_remote_vpc_dns_resolution = true
  }
}
```

### Cross-Account Usage

```terraform
provider "aws" {
  alias = "requester"

  # Requester's credentials.
}

provider "aws" {
  alias = "accepter"

  # Accepter's credentials.
}

resource "aws_vpc" "main" {
  provider = aws.requester

  cidr_block = "10.0.0.0/16"

  enable_dns_support   = true
  enable_dns_hostnames = true
}

resource "aws_vpc" "peer" {
  provider = aws.accepter

  cidr_block = "10.1.0.0/16"

  enable_dns_support   = true
  enable_dns_hostnames = true
}

data "aws_caller_identity" "peer" {
  provider = aws.accepter
}

# Requester's side of the connection.
resource "aws_vpc_peering_connection" "peer" {
  provider = aws.requester

  vpc_id        = aws_vpc.main.id
  peer_vpc_id   = aws_vpc.peer.id
  peer_owner_id = data.aws_caller_identity.peer.account_id
  auto_accept   = false

  tags = {
    Side = "Requester"
  }
}

# Accepter's side of the connection.
resource "aws_vpc_peering_connection_accepter" "peer" {
  provider = aws.accepter

  vpc_peering_connection_id = aws_vpc_peering_connection.peer.id
  auto_accept               = true

  tags = {
    Side = "Accepter"
  }
}

resource "aws_vpc_peering_connection_options" "requester" {
  provider = aws.requester

  # As options can't be set until the connection has been accepted
  # create an explicit dependency on the accepter.
  vpc_peering_connection_id = aws_vpc_peering_connection_accepter.peer.id

  requester {
    allow_remote_vpc_dns_resolution = true
  }
}

resource "aws_vpc_peering_connection_options" "accepter" {
  provider = aws.accepter

  vpc_peering_connection_id = aws_vpc_peering_connection_accepter.peer.id

  accepter {
    allow_remote_vpc_dns_resolution = true
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `vpc_peering_connection_id` - (Required) The ID of the requester VPC peering connection.
* `accepter` (Optional) - An optional configuration block that allows for [VPC Peering Connection](https://docs.aws.amazon.com/vpc/latest/peering/what-is-vpc-peering.html) options to be set for the VPC that acceptsthe peering connection (a maximum of one).
* `requester` (Optional) - A optional configuration block that allows for [VPC Peering Connection](https://docs.aws.amazon.com/vpc/latest/peering/what-is-vpc-peering.html) options to be set for the VPC that requeststhe peering connection (a maximum of one).

#### Accepter and Requester Arguments

-> **Note:** When enabled, the DNS resolution feature requires that VPCs participating in the peering must have support for the DNS hostnames enabled. This can be done using the [`enable_dns_hostnames`](vpc.html#enable_dns_hostnames) attribute in the [`aws_vpc`](vpc.html) resource. See [Using DNS with Your VPC](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/vpc-dns.html) user guide for more information.

* `allow_remote_vpc_dns_resolution` - (Optional) Allow a local VPC to resolve public DNS hostnames to private IP addresses when queried from instances in the peer VPC.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the VPC Peering Connection Options.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Peering Connection Options using the VPC peering `id`. For example:

```terraform
import {
  to = aws_vpc_peering_connection_options.foo
  id = "pcx-111aaa111"
}
```

Using `terraform import`, import VPC Peering Connection Options using the VPC peering `id`. For example:

```console
% terraform import aws_vpc_peering_connection_options.foo pcx-111aaa111
```
