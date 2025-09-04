---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_route_server_peer"
description: |-
  Terraform resource for managing a VPC (Virtual Private Cloud) Route Server Peer.
---
# Resource: aws_vpc_route_server_peer

  Provides a resource for managing a VPC (Virtual Private Cloud) Route Server Peer.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpc_route_server_peer" "test" {
  route_server_endpoint_id = aws_vpc_route_server_endpoint.example.route_server_endpoint_id
  peer_address             = "10.0.1.250"
  bgp_options {
    peer_asn = 65200
  }

  tags = {
    Name = "Appliance 1"
  }
}
```

### Complete Configuration

```terraform
resource "aws_vpc_route_server" "test" {
  amazon_side_asn = 4294967294

  tags = {
    Name = "Test"
  }
}

resource "aws_vpc_route_server_association" "test" {
  route_server_id = aws_vpc_route_server.test.route_server_id
  vpc_id          = aws_vpc.test.id
}

resource "aws_vpc_route_server_endpoint" "test" {
  route_server_id = aws_vpc_route_server.test.route_server_id
  subnet_id       = aws_subnet.test.id

  tags = {
    Name = "Test Endpoint"
  }

  depends_on = [aws_vpc_route_server_association.test]
}

resource "aws_vpc_route_server_propagation" "test" {
  route_server_id = aws_vpc_route_server.test.route_server_id
  route_table_id  = aws_route_table.test.id

  depends_on = [aws_vpc_route_server_association.test]
}

resource "aws_vpc_route_server_peer" "test" {
  route_server_endpoint_id = aws_vpc_route_server_endpoint.test.route_server_endpoint_id
  peer_address             = "10.0.1.250"
  bgp_options {
    peer_asn                = 65000
    peer_liveness_detection = "bgp-keepalive"
  }

  tags = {
    Name = "Test Appliance"
  }
}

```

## Argument Reference

The following arguments are required:

* `bgp_options` - (Required) The BGP options for the peer, including ASN (Autonomous System Number) and BFD (Bidrectional Forwarding Detection) settings. Configuration block with BGP Options configuration Detailed below
* `peer_address` - (Required) The IPv4 address of the peer device.
* `route_server_endpoint_id` - (Required) The ID of the route server endpoint for which to create a peer.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### bgp_options

* `peer_asn` - (Required) The Border Gateway Protocol (BGP) Autonomous System Number (ASN) for the appliance. Valid values are from 1 to 4294967295. We recommend using a private ASN in the 64512–65534 (16-bit ASN) or 4200000000–4294967294 (32-bit ASN) range.
* `peer_liveness_detection` (Optional) The requested liveness detection protocol for the BGP peer. Valid values are `bgp-keepalive` and `bfd`. Default value is `bgp-keepalive`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the route server peer.
* `route_server_peer_id` - The unique identifier of the route server peer.
* `route_server_id` - The ID of the route server associated with this peer.
* `endpoint_eni_address` - The IP address of the Elastic network interface for the route server endpoint.
* `endpoint_eni_id` - The ID of the Elastic network interface for the route server endpoint.
* `subnet_id` - The ID of the subnet containing the route server peer.
* `vpc_id` - The ID of the VPC containing the route server peer.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC (Virtual Private Cloud) Route Server using the `route_server_peer_id`. For example:

```terraform
import {
  to = aws_vpc_route_server_peer.example
  id = "rsp-12345678"
}
```

Using `terraform import`, import VPC (Virtual Private Cloud) Route Server using the `route_server_peer_id`. For example:

```console
% terraform import aws_vpc_route_server_peer.example rsp-12345678
```
