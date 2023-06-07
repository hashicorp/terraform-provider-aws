---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_multicast_domain_association"
description: |-
  Manages an EC2 Transit Gateway Multicast Domain Association
---

# Resource: aws_ec2_transit_gateway_multicast_domain_association

Associates the specified subnet and transit gateway attachment with the specified transit gateway multicast domain.

## Example Usage

```terraform
resource "aws_ec2_transit_gateway" "example" {
  multicast_support = "enable"
}

resource "aws_ec2_transit_gateway_vpc_attachment" "example" {
  subnet_ids         = [aws_subnet.example.id]
  transit_gateway_id = aws_ec2_transit_gateway.example.id
  vpc_id             = aws_vpc.example.id
}

resource "aws_ec2_transit_gateway_multicast_domain" "example" {
  transit_gateway_id = aws_ec2_transit_gateway.example.id
}

resource "aws_ec2_transit_gateway_multicast_domain_association" "example" {
  subnet_id                           = aws_subnet.example.id
  transit_gateway_attachment_id       = aws_ec2_transit_gateway_vpc_attachment.example.id
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain.example.id
}
```

## Argument Reference

The following arguments are supported:

* `subnetId` - (Required) The ID of the subnet to associate with the transit gateway multicast domain.
* `transitGatewayAttachmentId` - (Required) The ID of the transit gateway attachment.
* `transitGatewayMulticastDomainId` - (Required) The ID of the transit gateway multicast domain.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Transit Gateway Multicast Domain Association identifier.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10M`)
- `delete` - (Default `10M`)

<!-- cache-key: cdktf-0.17.0-pre.15 input-0aab754cf2f3ec5b55918c480d40a25c05ca4acac47f8f8d44dabc61eeb6bf26 -->