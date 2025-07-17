---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_connect_peer"
description: |-
  Manages an AWS Network Manager Connect Peer.
---

# Resource: aws_networkmanager_connect_peer

Manages an AWS Network Manager Connect Peer.

Use this resource to create a Connect peer in AWS Network Manager. Connect peers establish BGP sessions with your on-premises networks through Connect attachments, enabling dynamic routing between your core network and external networks.

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
    aws_networkmanager_attachment_accepter.example
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
    aws_networkmanager_attachment_accepter.example2
  ]
}
```

### Usage with a Tunnel-less Connect attachment

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
    protocol = "NO_ENCAP"
  }
}

resource "aws_networkmanager_connect_peer" "example" {
  connect_attachment_id = aws_networkmanager_connect_attachment.example.id
  peer_address          = "127.0.0.1"
  bgp_options {
    peer_asn = 65000
  }
  subnet_arn = aws_subnet.example2.arn
}
```

## Argument Reference

The following arguments are required:

* `connect_attachment_id` - (Required) ID of the connection attachment.
* `peer_address` - (Required) Connect peer address.

The following arguments are optional:

* `bgp_options` - (Optional) Connect peer BGP options. See [bgp_options](#bgp_options) for more information.
* `core_network_address` - (Optional) Connect peer core network address.
* `inside_cidr_blocks` - (Optional) Inside IP addresses used for BGP peering. Required when the Connect attachment protocol is `GRE`. See [`aws_networkmanager_connect_attachment`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/networkmanager_connect_attachment) for details.
* `subnet_arn` - (Optional) Subnet ARN for the Connect peer. Required when the Connect attachment protocol is `NO_ENCAP`. See [`aws_networkmanager_connect_attachment`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/networkmanager_connect_attachment) for details.
* `tags` - (Optional) Key-value tags for the attachment. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### bgp_options

* `peer_asn` - (Optional) Peer ASN.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Connect peer.
* `configuration` - Configuration of the Connect peer.
* `connect_peer_id` - ID of the Connect peer.
* `core_network_id` - ID of a core network.
* `created_at` - Timestamp when the Connect peer was created.
* `edge_location` - Region where the peer is located.
* `state` - State of the Connect peer.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `15m`)

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
