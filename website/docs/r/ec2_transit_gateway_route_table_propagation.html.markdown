---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_route_table_propagation"
description: |-
  Manages an EC2 Transit Gateway Route Table propagation
---

# Resource: aws_ec2_transit_gateway_route_table_propagation

Manages an EC2 Transit Gateway Route Table propagation.

## Example Usage

```terraform
resource "aws_ec2_transit_gateway_route_table_propagation" "example" {
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.example.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.example.id
}
```

### Direct Connect Gateway Propagation

When propagating routes from a Direct Connect Gateway attachment, reference the `transit_gateway_attachment_id` attribute directly from the `aws_dx_gateway_association` resource (available in v6.5.0+):

```terraform
resource "aws_dx_gateway" "example" {
  name            = "example"
  amazon_side_asn = 64512
}

resource "aws_ec2_transit_gateway" "example" {
  description = "example"
}

resource "aws_dx_gateway_association" "example" {
  dx_gateway_id         = aws_dx_gateway.example.id
  associated_gateway_id = aws_ec2_transit_gateway.example.id

  allowed_prefixes = [
    "10.0.0.0/16",
  ]
}

resource "aws_ec2_transit_gateway_route_table" "example" {
  transit_gateway_id = aws_ec2_transit_gateway.example.id
}

# Correct: Reference the attachment ID directly from the association resource
resource "aws_ec2_transit_gateway_route_table_propagation" "example" {
  transit_gateway_attachment_id  = aws_dx_gateway_association.example.transit_gateway_attachment_id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.example.id
}
```

~> **NOTE:** Avoid using the `aws_ec2_transit_gateway_dx_gateway_attachment` data source to retrieve the attachment ID, as this can cause unnecessary resource recreation when unrelated attributes of the Direct Connect Gateway association change (such as `allowed_prefixes`). Always reference the `transit_gateway_attachment_id` attribute directly from the `aws_dx_gateway_association` resource when available.

### VPC Attachment Propagation

For VPC attachments, always reference the attachment resource's `id` attribute directly. Avoid using data sources or lifecycle rules that might cause the attachment ID to become unknown during planning:

```terraform
resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "example" {
  vpc_id     = aws_vpc.example.id
  cidr_block = "10.0.1.0/24"
}

resource "aws_ec2_transit_gateway" "example" {
  description = "example"
}

resource "aws_ec2_transit_gateway_vpc_attachment" "example" {
  subnet_ids         = [aws_subnet.example.id]
  transit_gateway_id = aws_ec2_transit_gateway.example.id
  vpc_id             = aws_vpc.example.id
}

resource "aws_ec2_transit_gateway_route_table" "example" {
  transit_gateway_id = aws_ec2_transit_gateway.example.id
}

# Correct: Reference the VPC attachment ID directly
resource "aws_ec2_transit_gateway_route_table_propagation" "example" {
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.example.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.example.id
}
```

~> **NOTE:** When the `transit_gateway_attachment_id` changes (for example, when a VPC attachment is replaced), this resource will be recreated. This is the correct behavior to maintain consistency between the attachment and its route table propagation.

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `transit_gateway_attachment_id` - (Required) Identifier of EC2 Transit Gateway Attachment.
* `transit_gateway_route_table_id` - (Required) Identifier of EC2 Transit Gateway Route Table.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - EC2 Transit Gateway Route Table identifier combined with EC2 Transit Gateway Attachment identifier
* `resource_id` - Identifier of the resource
* `resource_type` - Type of the resource

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ec2_transit_gateway_route_table_propagation` using the EC2 Transit Gateway Route Table identifier, an underscore, and the EC2 Transit Gateway Attachment identifier. For example:

```terraform
import {
  to = aws_ec2_transit_gateway_route_table_propagation.example
  id = "tgw-rtb-12345678_tgw-attach-87654321"
}
```

Using `terraform import`, import `aws_ec2_transit_gateway_route_table_propagation` using the EC2 Transit Gateway Route Table identifier, an underscore, and the EC2 Transit Gateway Attachment identifier. For example:

```console
% terraform import aws_ec2_transit_gateway_route_table_propagation.example tgw-rtb-12345678_tgw-attach-87654321
```
