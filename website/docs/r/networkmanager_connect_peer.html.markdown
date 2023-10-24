---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_connect_peer"
description: |-
  Terraform resource for managing an AWS NetworkManager Connect Peer.
---

# Resource: aws_networkmanager_connect_peer

Terraform resource for managing an AWS NetworkManager Connect Peer.

## Example Usage

### Basic Usage

```terraform
resource "aws_networkmanager_vpc_attachment" "example" {
  subnet_arns     = aws_subnet.example[*].arn
  core_network_id = awscc_networkmanager_core_network.example.id
  vpc_arn         = aws_vpc.example.arn
}

resource "aws_networkmanager_connect_attachment" "example" {
  core_network_id         = awscc_networkmanager_core_network.example.id
  transport_attachment_id = aws_networkmanager_vpc_attachment.example.id
  edge_location           = aws_networkmanager_vpc_attachment.example.edge_location
  options {
    protocol = "GRE"
  }
}

resource "aws_networkmanager_connect_peer" "example" {
  connect_attachment_id = aws_networkmanager_connect_attachment.example.id
  peer_address          = "127.0.0.1"
  bgp_options {
    peer_asn = 65000
  }
  inside_cidr_blocks = ["172.16.0.0/16"]
}
```

### Usage with attachment accepter

```terraform
resource "aws_networkmanager_vpc_attachment" "example" {
  subnet_arns     = aws_subnet.example[*].arn
  core_network_id = awscc_networkmanager_core_network.example.id
  vpc_arn         = aws_vpc.example.arn
}

resource "aws_networkmanager_attachment_accepter" "example" {
  attachment_id   = aws_networkmanager_vpc_attachment.example.id
  attachment_type = aws_networkmanager_vpc_attachment.example.attachment_type
}

resource "aws_networkmanager_connect_attachment" "example" {
  core_network_id         = awscc_networkmanager_core_network.example.id
  transport_attachment_id = aws_networkmanager_vpc_attachment.example.id
  edge_location           = aws_networkmanager_vpc_attachment.example.edge_location
  options {
    protocol = "GRE"
  }
  depends_on = [
    "aws_networkmanager_attachment_accepter.test"
  ]
}

resource "aws_networkmanager_attachment_accepter" "example2" {
  attachment_id   = aws_networkmanager_connect_attachment.example.id
  attachment_type = aws_networkmanager_connect_attachment.example.attachment_type
}

resource "aws_networkmanager_connect_peer" "example" {
  connect_attachment_id = aws_networkmanager_connect_attachment.example.id
  peer_address          = "127.0.0.1"
  bgp_options {
    peer_asn = 65500
  }
  inside_cidr_blocks = ["172.16.0.0/16"]
  depends_on = [
    "aws_networkmanager_attachment_accepter.example2"
  ]
}
```

## Argument Reference

The following arguments are required:

- `connect_attachment_id` - (Required) The ID of the connection attachment.
- `inside_cidr_blocks` - (Required) The inside IP addresses used for BGP peering.
- `peer_address` - (Required) The Connect peer address.

The following arguments are optional:

- `bgp_options` (Optional) The Connect peer BGP options.
- `core_network_address` (Optional) A Connect peer core network address.
- `tags` - (Optional) Key-value tags for the attachment. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

- `arn` - The ARN of the attachment.
- `configuration` - The configuration of the Connect peer.
- `core_network_id` - The ID of a core network.
- `edge_location` - The Region where the peer is located.
- `id` - The ID of the Connect peer.
- `state` - The state of the Connect peer.
- `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmanager_connect_peer` using the connect peer ID. For example:

```terraform
import {
  to = aws_networkmanager_connect_peer.example
  id = "connect-peer-061f3e96275db1acc"
}
```

Using `terraform import`, import `aws_networkmanager_connect_peer` using the connect peer ID. For example:

```console
% terraform import aws_networkmanager_connect_peer.example connect-peer-061f3e96275db1acc
```
