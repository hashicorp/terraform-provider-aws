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

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `subnet_id` - (Required) The ID of the subnet to associate with the transit gateway multicast domain.
* `transit_gateway_attachment_id` - (Required) The ID of the transit gateway attachment.
* `transit_gateway_multicast_domain_id` - (Required) The ID of the transit gateway multicast domain.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - EC2 Transit Gateway Multicast Domain Association identifier.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `delete` - (Default `10m`)
