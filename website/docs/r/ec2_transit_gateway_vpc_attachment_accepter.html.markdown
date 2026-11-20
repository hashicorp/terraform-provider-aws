---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_vpc_attachment_accepter"
description: |-
  Manages the accepter's side of an EC2 Transit Gateway VPC Attachment
---

# Resource: aws_ec2_transit_gateway_vpc_attachment_accepter

Manages the accepter's side of an EC2 Transit Gateway VPC Attachment.

When a cross-account (requester's AWS account differs from the accepter's AWS account) EC2 Transit Gateway VPC Attachment
is created, an EC2 Transit Gateway VPC Attachment resource is automatically created in the accepter's account.
The requester can use the `aws_ec2_transit_gateway_vpc_attachment` resource to manage its side of the connection
and the accepter can use the `aws_ec2_transit_gateway_vpc_attachment_accepter` resource to "adopt" its side of the
connection into management.

## Example Usage

### Basic Usage

```terraform
resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "example" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.example.id

  tags = {
    Name = "Example cross-account attachment"
  }
}
```

### Cross-Account Usage

This example creates a Transit Gateway in one AWS account (the accepter), shares it with a second AWS account (the VPC owner) using Resource Access Manager (RAM), attaches a VPC in the second account to the Transit Gateway, and accepts the attachment in the first account.

~> **Note:** This example requires both AWS accounts to be in the same AWS Organization with [RAM sharing with AWS Organizations enabled](https://docs.aws.amazon.com/ram/latest/userguide/getting-started-sharing.html#getting-started-sharing-orgs). For accounts outside the Organization, set `allow_external_principals = true` on the `aws_ram_resource_share`, accept the RAM invitation in the VPC owner's account with the [`aws_ram_resource_share_accepter` resource](/docs/providers/aws/r/ram_resource_share_accepter.html), and add the accepter to the attachment's `depends_on`.

```terraform
provider "aws" {
  # Accepter's (Transit Gateway owner's) credentials.
}

provider "aws" {
  alias = "vpc_owner"

  # VPC owner's credentials.
}

data "aws_caller_identity" "vpc_owner" {
  provider = aws.vpc_owner
}

data "aws_availability_zones" "available" {
  provider = aws.vpc_owner

  state = "available"
}

# Accepter's side: create and share the Transit Gateway.
resource "aws_ec2_transit_gateway" "example" {
  tags = {
    Name = "example"
  }
}

resource "aws_ram_resource_share" "example" {
  name = "example"
}

resource "aws_ram_resource_association" "example" {
  resource_arn       = aws_ec2_transit_gateway.example.arn
  resource_share_arn = aws_ram_resource_share.example.arn
}

resource "aws_ram_principal_association" "example" {
  principal          = data.aws_caller_identity.vpc_owner.account_id
  resource_share_arn = aws_ram_resource_share.example.arn
}

# VPC owner's side: create the VPC and the attachment.
resource "aws_vpc" "example" {
  provider = aws.vpc_owner

  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "example" {
  provider = aws.vpc_owner

  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.example.id
}

resource "aws_ec2_transit_gateway_vpc_attachment" "example" {
  provider = aws.vpc_owner

  subnet_ids         = [aws_subnet.example.id]
  transit_gateway_id = aws_ec2_transit_gateway.example.id
  vpc_id             = aws_vpc.example.id

  # The Transit Gateway must be shared before the attachment can be created.
  depends_on = [
    aws_ram_principal_association.example,
    aws_ram_resource_association.example,
  ]

  tags = {
    Name = "example"
    Side = "Creator"
  }
}

# Accepter's side: accept the attachment.
resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "example" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.example.id

  tags = {
    Name = "example"
    Side = "Accepter"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `transit_gateway_attachment_id` - (Required) The ID of the EC2 Transit Gateway Attachment to manage.
* `transit_gateway_default_route_table_association` - (Optional) Boolean whether the VPC Attachment should be associated with the EC2 Transit Gateway association default route table. Default value: `true`.
* `transit_gateway_default_route_table_propagation` - (Optional) Boolean whether the VPC Attachment should propagate routes with the EC2 Transit Gateway propagation default route table. Default value: `true`.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway VPC Attachment. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - EC2 Transit Gateway Attachment identifier
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `appliance_mode_support` - Whether Appliance Mode support is enabled. Valid values: `disable`, `enable`.
* `dns_support` - Whether DNS support is enabled. Valid values: `disable`, `enable`.
* `security_group_referencing_support` - Whether Security Group Referencing Support is enabled. Valid values: `disable`, `enable`.
* `ipv6_support` - Whether IPv6 support is enabled. Valid values: `disable`, `enable`.
* `subnet_ids` - Identifiers of EC2 Subnets.
* `transit_gateway_id` - Identifier of EC2 Transit Gateway.
* `vpc_id` - Identifier of EC2 VPC.
* `vpc_owner_id` - Identifier of the AWS account that owns the EC2 VPC.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ec2_transit_gateway_vpc_attachment_accepter` using the EC2 Transit Gateway Attachment identifier. For example:

```terraform
import {
  to = aws_ec2_transit_gateway_vpc_attachment_accepter.example
  id = "tgw-attach-12345678"
}
```

Using `terraform import`, import `aws_ec2_transit_gateway_vpc_attachment_accepter` using the EC2 Transit Gateway Attachment identifier. For example:

```console
% terraform import aws_ec2_transit_gateway_vpc_attachment_accepter.example tgw-attach-12345678
```
